package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

func (o *objectBuilderRepo) AgGridTree(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	var (
		tableSlugs, tableSlugsTable []string
		params                      = make(map[string]any)
		childField                  = req.TableSlug + "_id"
	)

	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, err
	}

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	fields, ok := params["fields"].([]any)
	if !ok {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while type asserting fields")
	}
	stringFields := cast.ToStringSlice(fields)

	fieldQuery := `
		SELECT 
			f.slug, 
			f.type, 
			f.is_search 
		FROM field f 
		JOIN "table" t ON t.id = f.table_id 
		WHERE t.slug = $1 AND type IN ('LOOKUP', 'LOOKUPS')`
	fieldRows, err := conn.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			slug, ftype string
			isSearch    bool
		)

		err := fieldRows.Scan(&slug, &ftype, &isSearch)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		if strings.Contains(slug, "_id") && !strings.Contains(slug, req.TableSlug) && ftype == "LOOKUP" {
			tableSlugs = append(tableSlugs, slug)
			parts := strings.Split(slug, "_")
			if len(parts) > 2 {
				lastPart := parts[len(parts)-1]
				if _, err := strconv.Atoi(lastPart); err == nil {
					slug = strings.ReplaceAll(slug, fmt.Sprintf("_%v", lastPart), "")
				}
			}
			tableSlugsTable = append(tableSlugsTable, strings.ReplaceAll(slug, "_id", ""))
		}
	}

	var baseRelations, recursiveRelations, selectRelations string
	for i, slug := range tableSlugs {
		alias := fmt.Sprintf("r%d", i+1)
		fieldName := slug + "_data"

		baseRelations += fmt.Sprintf(`,
            (SELECT row_to_json(%s) 
             FROM "%s" %s 
             WHERE %s.guid = g.%s
            ) AS %s`,
			alias,
			tableSlugsTable[i],
			alias,
			alias,
			slug,
			fieldName)

		recursiveRelations += fmt.Sprintf(`,
            (SELECT row_to_json(%s) 
             FROM "%s" %s 
             WHERE %s.guid = child.%s
            ) AS %s`,
			alias,
			tableSlugsTable[i],
			alias,
			alias,
			slug,
			fieldName)

		selectRelations += fmt.Sprintf(`, %s`, fieldName)
	}

	query := fmt.Sprintf(`
        WITH RECURSIVE hierarchy AS (
            SELECT 
                %s,
                ARRAY[g.guid] AS path
                %s
            FROM %s g
            WHERE g.%s IS NULL

            UNION ALL

            SELECT 
                %s,
                parent.path || child.guid AS path
                %s
            FROM %s child
            INNER JOIN hierarchy parent ON child.%s = parent.guid
        )
        SELECT %s, path %s FROM hierarchy h
    `,
		joinColumns(stringFields, "g."),
		baseRelations,
		req.TableSlug,
		childField,
		joinColumns(stringFields, "child."),
		recursiveRelations,
		req.TableSlug,
		childField,
		joinColumns(stringFields, "h."),
		selectRelations,
	)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while querying database")
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var results []map[string]any

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("error getting row values: %w", err)
		}

		rowMap := make(map[string]any)
		for i, colName := range columns {
			val := values[i]
			if val == nil {
				rowMap[colName] = nil
			} else if (colName == config.GUID || strings.Contains(colName, config.ID)) && !strings.Contains(colName, "_id_data") {
				if arr, ok := val.([16]byte); ok {
					rowMap[colName] = helper.ConvertGuid(arr)
				}
			} else if colName == config.PATH {
				if arr, ok := val.([]any); ok {
					var guidList []string
					for _, guid := range arr {
						if guidArr, ok := guid.([16]byte); ok {
							guidList = append(guidList, helper.ConvertGuid(guidArr))
						}
					}
					rowMap[colName] = guidList
				}
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(err, "error while iterating over rows")
	}

	response := map[string]any{
		"response": results,
	}
	newResp, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, errors.Wrap(err, "error while converting map to struct")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      newResp,
	}, nil
}

func joinColumns(columns []string, prefix string) string {
	prefixedColumns := make([]string, len(columns))
	for i, col := range columns {
		prefixedColumns[i] = prefix + col
	}
	return strings.Join(prefixedColumns, ", ")
}
