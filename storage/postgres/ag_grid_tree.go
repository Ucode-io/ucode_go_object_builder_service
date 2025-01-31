package postgres

import (
	"context"
	"encoding/json"
	"fmt"
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
		conn       = psqlpool.Get(req.ProjectId)
		params     = make(map[string]interface{})
		childField = req.TableSlug + "_id"
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	fields, ok := params["fields"].([]interface{})
	if !ok {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while type asserting fields")
	}
	stringFields := cast.ToStringSlice(fields)

	query := fmt.Sprintf(`
		WITH RECURSIVE hierarchy AS (
			SELECT 
				%s,
				ARRAY[guid] AS path
			FROM %s
			WHERE %s IS NULL

			UNION ALL

			SELECT 
				%s,
				parent.path || child.guid AS path
			FROM %s child
			INNER JOIN hierarchy parent ON child.%s = parent.guid
		)
		SELECT %s, path FROM hierarchy
	`,
		joinColumns(stringFields, ""),
		req.TableSlug,
		childField,
		joinColumns(stringFields, "child."),
		req.TableSlug,
		childField,
		joinColumns(stringFields, ""),
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

	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("error getting row values: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := values[i]
			if val == nil {
				rowMap[colName] = nil
			} else if colName == config.GUID || strings.Contains(colName, config.ID) {
				if arr, ok := val.([16]byte); ok {
					rowMap[colName] = helper.ConvertGuid(arr)
				}
			} else if colName == config.PATH {
				if arr, ok := val.([]interface{}); ok {
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

	response := map[string]interface{}{
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
