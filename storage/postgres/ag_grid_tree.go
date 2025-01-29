package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/structpb"
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
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnValues[i] = new(interface{})
		}
		if err := rows.Scan(columnValues...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := *(columnValues[i].(*interface{}))
			if val == nil {
				rowMap[colName] = nil
			} else if colName == "guid" || strings.Contains(colName, "_id") {
				if arr, ok := val.([16]uint8); ok {
					rowMap[colName] = helper.ConvertGuid(arr)
				}
			} else if colName == "path" {
				if arr, ok := val.([]any); ok {
					var guidList []string
					for _, guid := range arr {
						guidList = append(guidList, helper.ConvertGuid(guid.([16]uint8)))
					}
					rowMap[colName] = guidList
				}
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling client types")
	}

	var dataStruct structpb.Struct
	jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

	err = json.Unmarshal(jsonBytes, &dataStruct)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling client types")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      &dataStruct,
	}, nil
}

func joinColumns(columns []string, prefix string) string {
	prefixedColumns := make([]string, len(columns))
	for i, col := range columns {
		prefixedColumns[i] = prefix + col
	}
	return strings.Join(prefixedColumns, ", ")
}
