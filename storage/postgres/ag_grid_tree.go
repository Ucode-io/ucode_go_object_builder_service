package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func (o *objectBuilderRepo) AgGridTree(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	var (
		conn   = psqlpool.Get(req.ProjectId)
		params = make(map[string]any)
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}

	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	var requestAgGrid models.RequestAgGrid

	if startRow, ok := params["startRow"].(float64); ok {
		requestAgGrid.StartRow = int64(startRow)
	}
	if endRow, ok := params["endRow"].(float64); ok {
		requestAgGrid.EndRow = int64(endRow)
	}
	if rowGroupCols, ok := params["rowGroupCols"].([]interface{}); ok {
		for _, col := range rowGroupCols {
			if colMap, ok := col.(map[string]interface{}); ok {
				requestAgGrid.RowGroupCols = append(requestAgGrid.RowGroupCols, mapToColumnVO(colMap))
			}
		}
	}
	if valueCols, ok := params["valueCols"].([]interface{}); ok {
		for _, col := range valueCols {
			if colMap, ok := col.(map[string]interface{}); ok {
				requestAgGrid.ValueCols = append(requestAgGrid.ValueCols, mapToColumnVO(colMap))
			}
		}
	}
	if pivotCols, ok := params["pivotCols"].([]interface{}); ok {
		for _, col := range pivotCols {
			if colMap, ok := col.(map[string]interface{}); ok {
				requestAgGrid.PivotCols = append(requestAgGrid.PivotCols, mapToColumnVO(colMap))
			}
		}
	}
	if pivotMode, ok := params["pivotMode"].(bool); ok {
		requestAgGrid.PivotMode = pivotMode
	}
	if groupKeys, ok := params["groupKeys"].([]interface{}); ok {
		for _, key := range groupKeys {
			if keyStr, ok := key.(string); ok {
				requestAgGrid.GroupKeys = append(requestAgGrid.GroupKeys, keyStr)
			}
		}
	}
	if filterModel, ok := params["filterModel"].(map[string]interface{}); ok {
		requestAgGrid.FilterModel = filterModel
	}
	if sortModel, ok := params["sortModel"].([]interface{}); ok {
		for _, sortItem := range sortModel {
			if sortMap, ok := sortItem.(map[string]interface{}); ok {
				requestAgGrid.SortModel = append(requestAgGrid.SortModel, sortMap)
			}
		}
	}

	selectSQL := createSelectSQL(requestAgGrid)
	fromSQL := fmt.Sprintf("FROM %s ", req.TableSlug)
	whereSQL := createWhereSQL(requestAgGrid)
	limitSQL := createLimitSQL(requestAgGrid)
	orderBySQL := createOrderBySQL(requestAgGrid)
	groupBySQL := createGroupBySQL(requestAgGrid)

	SQL := fmt.Sprintf("%s %s %s %s %s %s", selectSQL, fromSQL, whereSQL, groupBySQL, orderBySQL, limitSQL)

	rows, err := conn.Query(ctx, SQL)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing query")
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
			} else if colName == "guid" {
				if arr, ok := val.([16]uint8); ok {
					rowMap[colName] = helper.ConvertGuid(arr)
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

func createSelectSQL(r models.RequestAgGrid) string {
	rowGroupCols := r.RowGroupCols
	valueCols := r.ValueCols
	groupKeys := r.GroupKeys

	isDoingGrouping := isDoingGrouping(rowGroupCols, groupKeys)
	if isDoingGrouping {
		groupKeysLength := len(groupKeys)
		rowGroupCol := rowGroupCols[groupKeysLength]
		colsToSelect := make([]interface{}, 0)
		colsToSelect = append(colsToSelect, fmt.Sprintf(`"%s"`, rowGroupCol.Field))

		for _, v := range valueCols {
			s := fmt.Sprintf(`%s("%s") AS "%s"`, v.AggFunc, v.Field, v.Field)
			colsToSelect = append(colsToSelect, s)
		}

		strs := make([]string, len(colsToSelect))
		for i, v := range colsToSelect {
			strs[i] = v.(string)
		}
		part := strings.Join(strs, ", ")

		return fmt.Sprintf("SELECT %s", part)
	}

	return "SELECT *"
}

func createWhereSQL(r models.RequestAgGrid) string {
	rowGroupCols := r.RowGroupCols
	groupKeys := r.GroupKeys
	filterModel := r.FilterModel

	whereParts := make([]string, 0)

	if len(groupKeys) > 0 {
		for k, v := range groupKeys {
			colName := rowGroupCols[k].Field
			part := fmt.Sprintf(`"%s" = '%s'`, colName, v)
			whereParts = append(whereParts, part)
		}
	}

	if filterModel != nil {
		for i, v := range filterModel {
			inRange := v.(map[string]interface{})
			operator := inRange["operator"]
			if operator == "AND" || operator == "OR" {
				partRange := make([]string, 0)
				for i2, v2 := range inRange {
					if i2 == "filterType" || i2 == "operator" {
						continue
					}

					createFilterSQL := createFilterSQL(i, v2.(map[string]interface{}))
					partRange = append(partRange, createFilterSQL)
				}

				strs := make([]string, 0)
				for _, v3 := range partRange {
					strs = append(strs, v3)
				}
				part := strings.Join(strs, fmt.Sprintf(" %s ", operator.(string)))

				wherePartRange := fmt.Sprintf(" %s ", part)
				whereParts = append(whereParts, wherePartRange)
			} else {
				createFilterSQL := createFilterSQL(i, v.(map[string]interface{}))
				whereParts = append(whereParts, createFilterSQL)
			}
		}
	}

	if len(whereParts) > 0 {
		strs := make([]string, len(whereParts))
		for i, v := range whereParts {
			strs[i] = v
		}
		part := strings.Join(strs, " AND ")

		return fmt.Sprintf(" WHERE %s ", part)
	}

	return ""
}

func createLimitSQL(r models.RequestAgGrid) string {
	startRow := r.StartRow
	endRow := r.EndRow
	pageSize := endRow - startRow

	return fmt.Sprintf("LIMIT %v OFFSET %v", (pageSize + 1), startRow)
}

func createOrderBySQL(r models.RequestAgGrid) string {
	rowGroupCols := r.RowGroupCols
	groupKeys := r.GroupKeys
	sortModel := r.SortModel
	grouping := isDoingGrouping(rowGroupCols, groupKeys)

	sortParts := make([]string, 0)
	if len(sortModel) != 0 {
		groupColIds := make([]string, 0)
		for _, v := range rowGroupCols {
			id := v.ID
			groupColIds = append(groupColIds, id)
			break
		}

		for _, v := range sortModel {
			var groupColIdsIndexOf int
			for ig, vg := range groupColIds {
				if v["colId"] == vg {
					groupColIdsIndexOf = ig
					break
				} else {
					groupColIdsIndexOf = -1
					break
				}
			}

			if grouping && groupColIdsIndexOf < 0 {
				// ignore
			} else {
				part := fmt.Sprintf(`"%s" %s`, v["colId"], v["sort"])
				sortParts = append(sortParts, part)
			}
		}
	}

	if len(sortParts) > 0 {
		strs := make([]string, len(sortParts))
		for i, v := range sortParts {
			strs[i] = v
		}
		part := strings.Join(strs, ", ")
		return fmt.Sprintf(` ORDER BY %s`, part)
	}

	return ""
}

func createGroupBySQL(r models.RequestAgGrid) string {
	rowGroupCols := r.RowGroupCols
	groupKeys := r.GroupKeys

	isDoingGrouping := isDoingGrouping(rowGroupCols, groupKeys)
	if isDoingGrouping {
		colsToGroupBy := make([]interface{}, 0)
		rowGroupCol := rowGroupCols[len(groupKeys)]
		field := fmt.Sprintf(`"%s"`, rowGroupCol.Field)
		colsToGroupBy = append(colsToGroupBy, field)

		strs := make([]string, len(colsToGroupBy))
		for i, v := range colsToGroupBy {
			strs[i] = v.(string)
		}

		part := strings.Join(strs, ", ")
		return fmt.Sprintf(` GROUP BY %s`, part)
	}

	// select all columns
	return ""
}

func createFilterSQL(key string, item map[string]interface{}) string {
	switch item["filterType"] {
	case "text":
		return createTextFilterSQL(key, item)
	case "number":
		return createNumberFilterSQL(key, item)
	case "date":
		return createDateFilterSQL(key, item)
	case "dateTime":
		return createDateTimeFilterSQL(key, item)
	default:
		log.Println("unkonwn filter type: ", item["filterType"])
		return ""
	}
}

func createTextFilterSQL(key string, item map[string]interface{}) string {
	switch item["type"] {
	case "equals":
		return fmt.Sprintf(`lower("%s"::TEXT) = trim(lower('%s'))`, key, item["filter"])
	case "notEqual":
		return fmt.Sprintf(`lower("%s"::TEXT) != trim(lower('%s'))`, key, item["filter"])
	case "contains":
		return fmt.Sprintf(`lower("%s"::TEXT) LIKE '%s' || trim(lower('%s')) || '%s'`, key, "%", item["filter"], "%")
	case "notContains":
		return fmt.Sprintf(`lower("%s"::TEXT) NOT LIKE '%s' || trim(lower('%s')) || '%s'`, key, "%", item["filter"], "%")
	case "startsWith":
		return fmt.Sprintf(`lower("%s"::TEXT) LIKE trim(lower('%s')) || '%s'`, key, item["filter"], "%")
	case "endsWith":
		return fmt.Sprintf(`lower("%s"::TEXT) LIKE '%s' || trim(lower('%s'))`, key, "%", item["filter"])
	default:
		log.Println("unknown text filter type: ", item["type"])
		return "true"
	}
}

func createNumberFilterSQL(key string, item map[string]interface{}) string {
	switch item["type"] {
	case "equals":
		return fmt.Sprintf(`"%s" = %v`, key, item["filter"])
	case "notEqual":
		return fmt.Sprintf(`"%s" != %v`, key, item["filter"])
	case "greaterThan":
		return fmt.Sprintf(`"%s" > %v`, key, item["filter"])
	case "greaterThanOrEqual":
		return fmt.Sprintf(`"%s" >= %v`, key, item["filter"])
	case "lessThan":
		return fmt.Sprintf(`"%s" < %v`, key, item["filter"])
	case "lessThanOrEqual":
		return fmt.Sprintf(`"%s" <= %v`, key, item["filter"])
	case "inRange":
		return fmt.Sprintf(`("%s" >= %v AND "%s" <= %v)`, key, item["filter"], key, item["filterTo"])
	default:
		log.Println("unknown number filter type: ", item["type"])
		return "true"
	}
}

func createDateFilterSQL(key string, item map[string]interface{}) string {
	switch item["type"] {
	case "equals":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') = '%v'`, key, item["dateFrom"])
	case "notEqual":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') != '%v'`, key, item["dateFrom"])
	case "greaterThan":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') > '%v'`, key, item["dateFrom"])
	case "greaterThanOrEqual":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') >= '%v'`, key, item["dateFrom"])
	case "lessThan":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') < '%v'`, key, item["dateFrom"])
	case "lessThanOrEqual":
		return fmt.Sprintf(`to_char(%s, 'YYYY-MM-DD') <= '%v'`, key, item["dateFrom"])
	case "inRange":
		return fmt.Sprintf(`(to_char(%s, 'YYYY-MM-DD') >= '%v' AND to_char(%s, 'YYYY-MM-DD') <= '%v')`, key, item["dateFrom"], key, item["dateTo"])
	default:
		log.Println("unknown date filter type: ", item["type"])
		return "true"
	}
}

func createDateTimeFilterSQL(key string, item map[string]interface{}) string {
	switch item["type"] {
	case "equals":
		return fmt.Sprintf(`%s::TIMESTAMP = '%v'`, key, item["dateFrom"])
	case "notEqual":
		return fmt.Sprintf(`%s::TIMESTAMP != '%v'`, key, item["dateFrom"])
	case "greaterThan":
		return fmt.Sprintf(`%s::TIMESTAMP > '%v'`, key, item["dateFrom"])
	case "greaterThanOrEqual":
		return fmt.Sprintf(`%s::TIMESTAMP >= '%v'`, key, item["dateFrom"])
	case "lessThan":
		return fmt.Sprintf(`%s::TIMESTAMP < '%v'`, key, item["dateFrom"])
	case "lessThanOrEqual":
		return fmt.Sprintf(`%s::TIMESTAMP <= '%v'`, key, item["dateFrom"])
	case "inRange":
		return fmt.Sprintf(`%s::TIMESTAMP BETWEEN %v::TIMESTAMP AND %v::TIMESTAMP`, key, item["dateFrom"], item["dateTo"])
	default:
		log.Println("unknown date filter type: ", item["type"])
		return "true"
	}
}

func isDoingGrouping(r []models.ColumnVO, g []string) bool {
	return len(r) > len(g)
}

func mapToColumnVO(data map[string]interface{}) models.ColumnVO {
	return models.ColumnVO{
		ID:          toString(data["id"]),
		DisplayName: toString(data["displayName"]),
		Field:       toString(data["field"]),
		AggFunc:     toString(data["aggFunc"]),
	}
}

func toString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
