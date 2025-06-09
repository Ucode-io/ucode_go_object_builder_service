package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

func (o *objectBuilderRepo) GetBoardStructure(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to get database connection")
	}

	params := &models.BoardDataParams{}
	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to marshak request")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to unmarshal request")
	}

	var (
		groupByField    = params.GroupBy.Field
		subgroupByField = params.SubgroupBy.Field
		hasSubgroup     = subgroupByField != ""
		groups          = make([]models.BoardGroup, 0)
		subgroups       = make([]models.BoardSubgroup, 0)
		response        = map[string]any{}
		noGroupValue    = "Unassigned"
	)

	if groupByField == "" {
		return nil, helper.HandleDatabaseError(errors.New("group_by field is required"), o.logger, "GetBoardStructure: Group by is required")
	}

	query := fmt.Sprintf(`
		SELECT
			unnest(ARRAY[%s]::TEXT[]) AS name,
			COUNT(*) AS count
		FROM %s
		GROUP BY name
	`,
		pq.QuoteIdentifier(groupByField),
		pq.QuoteIdentifier(req.TableSlug),
	)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to query db")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name  sql.NullString
			count int
		)
		if err := rows.Scan(&name, &count); err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to scan row")
		}
		if !name.Valid {
			name.String = noGroupValue
		}
		groups = append(groups, models.BoardGroup{
			Name:  name.String,
			Count: count,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Error after row iteration")
	}

	if hasSubgroup {
		query = fmt.Sprintf(`
			SELECT DISTINCT %s FROM %s ORDER BY %s
		`,
			pq.QuoteIdentifier(subgroupByField),
			pq.QuoteIdentifier(req.TableSlug),
			pq.QuoteIdentifier(subgroupByField),
		)
		rows, err := conn.Query(ctx, query)
		if err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to query db")
		}
		defer rows.Close()

		for rows.Next() {
			var name sql.NullString
			if err := rows.Scan(&name); err != nil {
				return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to scan row")
			}
			if !name.Valid {
				name.String = noGroupValue
			}
			subgroups = append(subgroups, models.BoardSubgroup{Name: name.String})
		}
		if err := rows.Err(); err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Error after row iteration")
		}
	}

	response["groups"] = groups
	response["subgroups"] = subgroups

	respData, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardStructure: Failed to convert response")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      respData,
	}, nil
}

func (o *objectBuilderRepo) GetBoardData(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to get database connection")
	}

	params := &models.BoardDataParams{}
	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to marshal request")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to unmarshal request")
	}

	var (
		groupByField    = params.GroupBy.Field
		subgroupByField = params.SubgroupBy.Field
		hasSubgroup     = subgroupByField != ""
		fields          = params.Fields
		limit           = params.Limit
		offset          = params.Offset
		orderBy         = "created_at"
		noGroupValue    = "Unassigned"

		groups    = make(map[string][]any)
		subgroups = make(map[string]map[string][]any)
		response  = map[string]any{}
	)

	if len(fields) == 0 {
		return nil, helper.HandleDatabaseError(errors.New("fields are required"), o.logger, "GetBoardData: Fields are required")
	}
	if groupByField == "" {
		return nil, helper.HandleDatabaseError(errors.New("group_by field is required"), o.logger, "GetBoardData: Group by is required")
	}
	if hasSubgroup {
		orderBy = subgroupByField
	}

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM %s
		WHERE deleted_at IS NULL
		ORDER BY %s ASC, board_order ASC, updated_at DESC
		OFFSET $1
		LIMIT $2
	`,
		joinColumnsWithPrefix(fields, ""),
		pq.QuoteIdentifier(req.TableSlug),
		pq.QuoteIdentifier(orderBy),
	)

	rows, err := conn.Query(ctx, query, offset, limit)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to query db")
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	for rows.Next() {
		row, err := processRow(rows, columns)
		if err != nil {
			return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to process row")
		}

		groupValues := toStringSlice(row[groupByField], noGroupValue)

		var subgroupValues []string
		if hasSubgroup {
			subgroupValues = toStringSlice(row[subgroupByField], noGroupValue)
		}

		for _, groupValue := range groupValues {
			if hasSubgroup {
				for _, subgroupValue := range subgroupValues {
					if _, exists := subgroups[subgroupValue]; !exists {
						subgroups[subgroupValue] = make(map[string][]any)
					}
					subgroups[subgroupValue][groupValue] = append(subgroups[subgroupValue][groupValue], row)
				}
			} else {
				groups[groupValue] = append(groups[groupValue], row)
			}
		}
	}

	var count int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, req.TableSlug)
	err = conn.QueryRow(ctx, countQuery).Scan(&count)
	if err != nil {
		return &nb.CommonMessage{}, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to get count")
	}

	response["response"] = groups
	response["count"] = count
	if hasSubgroup {
		response["response"] = subgroups
	}

	respData, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "GetBoardData: Failed to convert response")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      respData,
	}, nil
}

func toStringSlice(value any, noGroupValue string) []string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return []string{noGroupValue}
		}
		return []string{v}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return []string{fmt.Sprint(v)}
	case []any:
		slice := cast.ToStringSlice(v)
		if len(slice) == 0 {
			return []string{noGroupValue}
		}
		return slice
	default:
		return []string{noGroupValue}
	}
}
