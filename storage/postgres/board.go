package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type BoardResult struct {
	Groups []BoardGroup `json:"groups"`
}

type BoardGroup struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type BoardSubgroup struct {
	Name string `json:"name"`
}

func (o *objectBuilderRepo) GetBoardStructure(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to get database connection")
	}

	params := map[string]any{}

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to parse request")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to unmarshal request")
	}
	groupMap, ok := params["group"].(map[string]any)
	if !ok {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Group is not provided")
	}
	groupField, ok := groupMap["field"].(string)
	if !ok {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Group field is not provided")
	}

	query := fmt.Sprintf(`
        SELECT 
            unnest(%s) AS name,
            COUNT(*) AS count
        FROM %s
        WHERE cardinality(%s) > 0
        GROUP BY name
        ORDER BY count DESC`,
		pq.QuoteIdentifier(groupField),
		pq.QuoteIdentifier(req.TableSlug),
		pq.QuoteIdentifier(groupField))
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to query db")
	}
	defer rows.Close()

	var groups []BoardGroup
	for rows.Next() {
		var group BoardGroup
		if err := rows.Scan(&group.Name, &group.Count); err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to scan row")
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Error after row iteration")
	}

	var subgroups []BoardSubgroup
	subgroupMap, ok := params["subgroup"].(map[string]any)
	if ok {
		subgroupField, ok := subgroupMap["field"].(string)
		if ok {
			query = fmt.Sprintf(`
				SELECT DISTINCT %s FROM %s ORDER BY %s
			`,
				pq.QuoteIdentifier(subgroupField),
				pq.QuoteIdentifier(req.TableSlug),
				pq.QuoteIdentifier(subgroupField),
			)
			rows, err := conn.Query(ctx, query)
			if err != nil {
				return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to query db")
			}
			defer rows.Close()

			for rows.Next() {
				var subgroup BoardSubgroup
				if err := rows.Scan(&subgroup.Name); err != nil {
					return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to scan row")
				}
				subgroups = append(subgroups, subgroup)
			}

			if err := rows.Err(); err != nil {
				return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Error after row iteration")
			}
		}
	}

	response := map[string]any{"response": map[string]any{"groups": groups, "subgroups": subgroups}}
	respData, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to convert response")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      respData,
	}, nil
}

type BoardDataParams struct {
	GroupBy    GroupBy    `json:"group_by"`
	SubgroupBy SubgroupBy `json:"subgroup_by"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	Fields     []string   `json:"fields"`
}

type GroupBy struct {
	Field string `json:"field"`
}

type SubgroupBy struct {
	Field string `json:"field"`
}

func (o *objectBuilderRepo) GetBoardData(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to get database connection")
	}

	params, err := parseRequest(req)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to parse request")
	}

	hasSubgroup := params.SubgroupBy.Field != ""

	orderBy := params.GroupBy.Field
	if hasSubgroup {
		orderBy = params.SubgroupBy.Field
	}

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM %s
		WHERE deleted_at IS NULL
		ORDER BY %s, created_at
		OFFSET %d
		LIMIT %d
	`,
		joinColumnsWithPrefix(params.Fields, ""),
		req.TableSlug,
		orderBy,
		params.Offset,
		params.Limit,
	)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to query db")
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var (
		groups    = make(map[string][]any)
		subgroups = make(map[string]map[string][]any)
		output    any
	)
	for rows.Next() {
		row, err := processRow(rows, columns)
		if err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to process row")
		}
		groupArray := cast.ToStringSlice(row[params.GroupBy.Field])
		if len(groupArray) == 0 {
			continue
		}
		groupValue := groupArray[0]

		if hasSubgroup {
			subgroupVal := cast.ToString(row[params.SubgroupBy.Field])
			if subgroupVal == "" {
				continue
			}

			if _, exists := subgroups[subgroupVal]; !exists {
				subgroups[subgroupVal] = make(map[string][]any)
			}

			subgroups[subgroupVal][groupValue] = append(subgroups[subgroupVal][groupValue], row)
		} else {
			groups[groupValue] = append(groups[groupValue], row)
		}
	}

	output = groups
	if hasSubgroup {
		output = subgroups
	}

	response := map[string]any{"response": output}
	respData, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to convert response")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      respData,
	}, nil
}

func parseRequest(req *nb.CommonMessage) (*BoardDataParams, error) {
	params := &BoardDataParams{}
	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, err
	}

	if len(params.Fields) == 0 {
		return nil, errors.New("fields are required")
	}
	if params.GroupBy.Field == "" {
		return nil, errors.New("group_by field is required")
	}
	if params.Limit == 0 {
		params.Limit = 100
	}

	return params, nil
}

// func (o *objectBuilderRepo) GetBoardData(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
// 	// Get database connection
// 	conn, err := psqlpool.Get(req.GetProjectId())
// 	if err != nil {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to get database connection")
// 	}

// 	// Parse request data into params map
// 	params := map[string]any{}
// 	paramBody, err := json.Marshal(req.Data)
// 	if err != nil {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to parse request")
// 	}
// 	if err := json.Unmarshal(paramBody, &params); err != nil {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Failed to unmarshal request")
// 	}

// 	// Extract and validate group and group field
// 	groupMap, ok := params["group_by"].(map[string]any)
// 	if !ok {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Group is not provided")
// 	}
// 	groupField, ok := groupMap["field"].(string)
// 	if !ok {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardGroups: Group field is not provided")
// 	}

// 	// Extract and validate subgroup field if provided
// 	var subGroupField string
// 	if sg, ok := params["subgroup_by"]; ok {
// 		subGroupMap, ok := sg.(map[string]any)
// 		if !ok {
// 			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Subgroup is not provided correctly")
// 		}
// 		subGroupField, ok = subGroupMap["field"].(string)
// 		if !ok {
// 			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Subgroup field is not provided")
// 		}
// 	}

// 	// Extract pagination parameters
// 	limit := 300
// 	if l, ok := params["limit"]; ok {
// 		limit = cast.ToInt(l)
// 	}
// 	offset := 0
// 	if o, ok := params["offset"]; ok {
// 		offset = cast.ToInt(o)
// 	}

// 	// Common query parts
// 	tableName := pq.QuoteIdentifier(req.TableSlug)
// 	quotedGroupField := pq.QuoteIdentifier(groupField)

// 	// Base SELECT clause
// 	baseSelect := fmt.Sprintf(`
// 		SELECT
// 			guid,
// 			%s as subgroup_value,
// 			unnest(%s) as group_value,
// 			to_jsonb(%s.*) as row_data,
// 			ROW_NUMBER() OVER (
// 				PARTITION BY %s, unnest(%s)
// 				ORDER BY created_at DESC
// 			) as group_row_num,
// 			ROW_NUMBER() OVER (ORDER BY %s, unnest(%s), created_at DESC) as global_row_num
// 		FROM %s
// 		WHERE %s IS NOT NULL AND cardinality(%s) > 0`,
// 		pq.QuoteIdentifier(subGroupField),
// 		quotedGroupField,
// 		tableName,
// 		pq.QuoteIdentifier(subGroupField),
// 		quotedGroupField,
// 		pq.QuoteIdentifier(subGroupField),
// 		quotedGroupField,
// 		tableName,
// 		quotedGroupField,
// 		quotedGroupField,
// 	)

// 	// If no subgroup, simplify the base SELECT
// 	if subGroupField == "" {
// 		baseSelect = fmt.Sprintf(`
// 			SELECT
// 				guid,
// 				unnest(%s) as group_value,
// 				to_jsonb(%s.*) as row_data,
// 				ROW_NUMBER() OVER (
// 					PARTITION BY unnest(%s)
// 					ORDER BY created_at DESC
// 				) as group_row_num,
// 				ROW_NUMBER() OVER (ORDER BY unnest(%s), created_at DESC) as global_row_num
// 			FROM %s
// 			WHERE %s IS NOT NULL AND cardinality(%s) > 0`,
// 			quotedGroupField,
// 			tableName,
// 			quotedGroupField,
// 			quotedGroupField,
// 			tableName,
// 			quotedGroupField,
// 			quotedGroupField,
// 		)
// 	}

// 	// Common CTEs
// 	commonCTEs := `
// 		paginated AS (
// 			SELECT *
// 			FROM base_data
// 			WHERE global_row_num > $1 AND global_row_num <= $1 + $2
// 		),
// 		total AS (
// 			SELECT COUNT(*) as total_count
// 			FROM base_data
// 		)`

// 	// Build the appropriate GROUP BY part based on subgroup presence
// 	var groupByPart, finalSelect string
// 	if subGroupField == "" {
// 		groupByPart = `
// 			group_data AS (
// 				SELECT
// 					group_value,
// 					jsonb_agg(row_data ORDER BY group_row_num) as items
// 				FROM paginated
// 				GROUP BY group_value
// 			)`
// 		finalSelect = `
// 			SELECT
// 				COALESCE(
// 					jsonb_object_agg(
// 						group_value,
// 						CASE
// 							WHEN items IS NULL THEN '[]'::jsonb
// 							ELSE items
// 						END
// 					),
// 					'{}'::jsonb
// 				) as groups,
// 				(SELECT total_count FROM total) as total_count
// 			FROM group_data`
// 	} else {
// 		groupByPart = `
// 			group_data AS (
// 				SELECT
// 					subgroup_value,
// 					group_value,
// 					jsonb_agg(row_data ORDER BY group_row_num) as items
// 				FROM paginated
// 				GROUP BY subgroup_value, group_value
// 			),
// 			subgrouped_data AS (
// 				SELECT
// 					subgroup_value,
// 					COALESCE(
// 						jsonb_object_agg(
// 							group_value,
// 							CASE
// 								WHEN items IS NULL THEN '[]'::jsonb
// 								ELSE items
// 							END
// 						),
// 						'{}'::jsonb
// 					) as group_items
// 				FROM group_data
// 				GROUP BY subgroup_value
// 			)`
// 		finalSelect = `
// 			SELECT
// 				COALESCE(
// 					jsonb_object_agg(
// 						subgroup_value,
// 						group_items
// 					),
// 					'{}'::jsonb
// 				) as groups,
// 				(SELECT total_count FROM total) as total_count
// 			FROM subgrouped_data`
// 	}

// 	// Combine all parts into the final query
// 	query := fmt.Sprintf(`
// 		WITH RECURSIVE
// 		base_data AS (%s),
// 		%s,
// 		%s
// 		%s`,
// 		baseSelect,
// 		commonCTEs,
// 		groupByPart,
// 		finalSelect,
// 	)

// 	fmt.Println("QUERY->", query)

// 	// Execute the query
// 	var (
// 		groups     json.RawMessage
// 		totalCount int
// 	)
// 	err = conn.QueryRow(ctx, query, offset, limit).Scan(&groups, &totalCount)
// 	if err != nil {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to execute query")
// 	}

// 	// Prepare the response, wrapping groups in "data"
// 	response := map[string]any{
// 		"data":  groups,
// 		"count": totalCount,
// 	}
// 	respData, err := helper.ConvertMapToStruct(response)
// 	if err != nil {
// 		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to convert response")
// 	}

// 	return &nb.CommonMessage{
// 		TableSlug: req.TableSlug,
// 		ProjectId: req.ProjectId,
// 		Data:      respData,
// 	}, nil
// }
