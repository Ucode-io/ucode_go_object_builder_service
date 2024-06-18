package helper

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/google/uuid"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/spf13/cast"
)

func PrepareToCreateInObjectBuilder(ctx context.Context, conn *pgxpool.Pool, req *nb.CommonMessage, reqBody CreateBody) (map[string]interface{}, []map[string]interface{}, error) {

	// defer conn.Close()

	var (
		response   = make(map[string]interface{})
		fieldM     = reqBody.FieldMap
		tableSlugs = reqBody.TableSlugs
		fields     = reqBody.Fields
	)

	data, err := ConvertStructToMap(req.Data)
	if err != nil {
		return map[string]interface{}{}, []map[string]interface{}{}, err
	}

	response = data

	// return map[string]interface{}{}, []map[string]interface{}{}, err

	// * RANDOM_NUMBER
	{
		randomNumbers, ok := fieldM["RANDOM_NUMBERS"]
		if ok {
			for {
				randNum := GenerateRandomNumber(cast.ToString(randomNumbers.Attributes["prefix"]), cast.ToInt(randomNumbers.Attributes["digit_number"]))

				isExists, err := IsExists(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomNumbers.Slug, FieldValue: randNum})
				if err != nil {
					return map[string]interface{}{}, []map[string]interface{}{}, err
				}

				if !isExists {
					response[randomNumbers.Slug] = randNum
					break
				}
			}

		}
	}

	// * RANDOM_TEXT
	{
		randomText, ok := fieldM["RANDOM_TEXT"]
		if ok {
			for {
				randText := GenerateRandomString(cast.ToString(randomText.Attributes["prefix"]), cast.ToInt(randomText.Attributes["digit_number"]))
				isExists, err := IsExists(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomText.Slug, FieldValue: randText})
				if err != nil {
					return map[string]interface{}{}, []map[string]interface{}{}, err
				}

				if randText != "" {
					if !isExists {
						response[randomText.Slug] = randText
						break
					}
				}
			}

		}
	}

	// * RANDOM_UUID
	{
		randomUuid, ok := fieldM["RANDOM_UUID"]
		if ok {
			response[randomUuid.Slug] = uuid.NewString()
		}
	}

	// * MANUAL_STRING
	{
		manual, ok := fieldM["MANUAL_STRING"]
		if ok {
			fields, err := ConvertStructToMap(req.Data)
			if err != nil {
				return map[string]interface{}{}, []map[string]interface{}{}, err
			}

			text := cast.ToString(manual.Attributes["formula"])

			for _, slug := range tableSlugs {
				var (
					value interface{}
				)

				switch v := fields[slug].(type) {
				case bool:
					value = strconv.FormatBool(v)
				case string:
					value = v
				case []interface{}:
					if len(v) > 0 {
						if str, ok := v[0].(string); ok {
							value = str
						}
					}
				case int, float64:
					value = v
				default:
					value = v
				}

				text = strings.ReplaceAll(text, slug, cast.ToString(value))
			}

			response[manual.Slug] = text
		}
	}

	// * INCREMENT_ID
	{
		incrementField, ok := fieldM["INCREMENT_ID"]
		if ok {
			incrementBy := 0

			query := `UPDATE "incrementseqs" SET increment_by = increment_by + 1  WHERE table_slug = $1 AND field_slug = $2 RETURNING increment_by AS old_value`

			err = conn.QueryRow(ctx, query, req.TableSlug, incrementField.Slug).Scan(&incrementBy)
			if err != nil {
				return map[string]interface{}{}, []map[string]interface{}{}, err
			}

			response[incrementField.Slug] = cast.ToString(incrementField.Attributes["prefix"]) + "-" + fmt.Sprintf("%09d", incrementBy)
		}
	}

	// * INCREMENT_NUMBER
	{
		incrementNum, ok := fieldM["INCREMENT_NUMBER"]
		if ok {
			delete(response, incrementNum.Slug)
		}
	}

	// * AUTOFILL
	{
		for _, field := range fields {

			attributes, err := ConvertStructToMap(field.Attributes)
			if err != nil {
				return map[string]interface{}{}, []map[string]interface{}{}, err
			}

			if field.AutofillField != "" && field.AutofillTable != "" {

				splitArr := strings.Split(field.AutofillTable, "#")
				query := fmt.Sprintf(`SELECT %s FROM %s WHERE guid = '%s'`, field.AutofillField, splitArr[0], response[splitArr[0]+"_id"])

				var (
					autofill interface{}
				)

				err = conn.QueryRow(ctx, query).Scan(&autofill)
				if err != nil {
					return map[string]interface{}{}, []map[string]interface{}{}, err
				}

				if autofill != nil {
					response[field.Slug] = autofill
				}
			}

			_, ok := response[field.Slug]
			_, ok2 := attributes["defaultValue"]

			defaultValues := cast.ToSlice(attributes["default_values"])
			ftype := FIELD_TYPES[field.Type]
			if ok2 && !ok {
				if ftype == "FLOAT" {
					response[field.Slug] = cast.ToInt(attributes["defaultValue"])
				} else if field.Type == "DATE_TIME" || field.Type == "DATE" {
					response[field.Slug] = time.Now().Format(time.RFC3339)
				} else if field.Type == "SWITCH" {
					defaultValue := strings.ToLower(cast.ToString(attributes["defaultValue"]))

					if defaultValue == "true" {
						response[field.Slug] = true
					} else if defaultValue == "false" {
						response[field.Slug] = false
					}
				} else if ftype == "TEXT[]" {
					response[field.Slug] = "{}"
				} else if field.Type == "FORMULA_FRONTEND" {
					continue
				} else {
					response[field.Slug] = attributes["defaultValue"]
				}
			} else if len(defaultValues) > 0 && !ok {
				response[field.Slug] = defaultValues[0]
			} else if field.Type == "FORMULA_FRONTEND" {
				response[field.Slug] = cast.ToString(response[field.Slug])
			} else if IsEmpty(response[field.Slug]) {
				switch ftype {
				case "FLOAT":
					response[field.Slug] = 0
				case "TEXT[]":
					response[field.Slug] = []string{}
				case "VARCHAR":
					response[field.Slug] = ""
				case "DATE", "TIMESTAMP":
					response[field.Slug] = nil
				}

				if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
					delete(response, field.Slug)
				}
			}
		}
	}

	query := `SELECT table_to, table_from FROM "relation" WHERE id = $1`

	appendMany2ManyObjects := []map[string]interface{}{}
	// * AppendMany2ManyObjects
	{
		for _, field := range fields {
			var (
				tableTo, tableFrom string
			)

			if field.Type == "LOOKUPS" {
				if strings.Contains(field.Slug, "_ids") {

					_, ok := response[field.Slug]
					if ok {
						err = conn.QueryRow(ctx, query, field.RelationId).Scan(&tableTo, &tableFrom)
						if err != nil {
							return map[string]interface{}{}, []map[string]interface{}{}, err
						}

						// appendMany2Many := make(map[string]interface{})

						appendMany2Many := map[string]interface{}{
							"project_id": req.ProjectId,
							"id_from":    response["guid"],
							"id_to":      response[field.Slug],
							"table_from": req.TableSlug,
						}
						if tableTo == req.TableSlug {
							appendMany2Many["table_to"] = tableFrom
						} else if tableFrom == req.TableSlug {
							appendMany2Many["table_to"] = tableTo
						}

						appendMany2ManyObjects = append(appendMany2ManyObjects, appendMany2Many)
					}
				}
			}
		}
	}

	return response, appendMany2ManyObjects, nil
}

func GetFieldByType(ctx context.Context, conn *pgxpool.Pool, tableId, fieldType string) (FieldBody, error) {

	var (
		slug       string
		body       []byte
		attributes = make(map[string]interface{})
	)

	query := `SELECT 
		"slug",
		"attributes"
	FROM "field" WHERE table_id = $1 AND "type" = $2`

	err := conn.QueryRow(ctx, query, tableId, fieldType).Scan(&slug, &body)
	if err != nil {
		return FieldBody{}, err
	}
	if err := json.Unmarshal(body, &attributes); err != nil {
		return FieldBody{}, err
	}

	return FieldBody{Slug: slug, Attributes: attributes}, nil
}

func IsExists(ctx context.Context, conn *pgxpool.Pool, req IsExistsBody) (bool, error) {

	count := 0

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s=$1`, req.TableSlug, req.FieldSlug)

	err := conn.QueryRow(ctx, query, req.FieldValue).Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

type FieldBody struct {
	Slug       string
	Attributes map[string]interface{}
}

type IsExistsBody struct {
	TableSlug  string
	FieldSlug  string
	FieldValue interface{}
}

type CreateBody struct {
	FieldMap   map[string]FieldBody
	TableId    string
	Fields     []models.Field
	TableSlugs []string
}

func PrepareToUpdateInObjectBuilder(ctx context.Context, conn *pgxpool.Pool, req *nb.CommonMessage) (map[string]interface{}, error) {

	data, err := ConvertStructToMap(req.Data)
	if err != nil {
		return map[string]interface{}{}, err
	}

	oldData, err := GetItem(ctx, conn, req.TableSlug, cast.ToString(data["guid"]))
	if err != nil {
		return map[string]interface{}{}, err
	}

	var (
		event       = make(map[string]interface{})
		relationIds []string
	)

	event["payload"] = map[string]interface{}{
		"data":       data,
		"table_slug": req.TableSlug,
	}

	query := `SELECT f.relation_id FROM "field" as f JOIN "table" as t ON t.id = f.table_id WHERE t.slug = $1 AND f.type = 'LOOKUPS'`

	rows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer rows.Close()

	for rows.Next() {
		id := ""

		err := rows.Scan(&id)
		if err != nil {
			return map[string]interface{}{}, err
		}

		relationIds = append(relationIds, id)
	}

	var (
		relationMap                      = make(map[string]models.Relation)
		fieldTypes                       = make(map[string]string)
		appendMany2Many, deleteMany2Many = []map[string]interface{}{}, []map[string]interface{}{}
		dataToAnalytics                  = make(map[string]interface{})
	)

	query = `SELECT id, table_to, table_from FROM "relation" WHERE id IN ($1)`

	relationRows, err := conn.Query(ctx, query, pq.Array(relationIds))
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer relationRows.Close()

	for relationRows.Next() {
		rel := models.Relation{}

		err = relationRows.Scan(
			&rel.Id,
			&rel.TableTo,
			&rel.TableFrom,
		)
		if err != nil {
			return map[string]interface{}{}, err
		}

		relationMap[rel.Id] = rel
	}

	query = `SELECT f."id", f."type", f."attributes", f."slug", f."relation_id", f."required" FROM "field" as f JOIN "table" as t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			field      = models.Field{}
			attributes = []byte{}
			relationId sql.NullString
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&attributes,
			&field.Slug,
			&relationId,
			&field.Required,
		)
		if err != nil {
			return map[string]interface{}{}, err
		}

		field.RelationId = relationId.String

		fType := FIELD_TYPES[field.Type]
		fieldTypes[field.Slug] = fType

		if field.Type == "LOOKUPS" {

			var newIds, deletedIds []string

			_, ok := data[field.Slug]
			if ok {
				olderArr := cast.ToStringSlice(oldData[field.Slug])
				newArr := cast.ToStringSlice(data[field.Slug])

				if len(newArr) > 0 {
					for _, val := range newArr {
						found := false
						for _, oldVal := range olderArr {
							if val == oldVal {
								found = true
								break
							}
						}
						if !found {
							newIds = append(newIds, val)
						}
					}

					for _, oldVal := range olderArr {
						found := false
						for _, val := range newArr {
							if oldVal == val {
								found = true
								break
							}
						}
						if !found && !Contains(deletedIds, oldVal) {
							deletedIds = append(deletedIds, oldVal)
						}
					}
				}
			}

			relation := relationMap[field.RelationId]

			if len(newIds) > 0 {
				appendMany2ManyObj := make(map[string]interface{})

				appendMany2ManyObj = map[string]interface{}{
					"project_id": req.ProjectId,
					"id_from":    data["guid"],
					"id_to":      newIds,
					"table_from": req.TableSlug,
				}

				if relation.TableTo == req.TableSlug {
					appendMany2ManyObj["table_to"] = relation.TableFrom
				} else if relation.TableFrom == req.TableSlug {
					appendMany2ManyObj["table_to"] = relation.TableTo
				}

				appendMany2Many = append(appendMany2Many, appendMany2ManyObj)
			}
			if len(deletedIds) > 0 {
				deleteMany2ManyObj := make(map[string]interface{})

				deleteMany2ManyObj = map[string]interface{}{
					"project_id": req.ProjectId,
					"id_from":    data["guid"],
					"id_to":      deletedIds,
					"table_from": req.TableSlug,
				}

				if relation.TableTo == req.TableSlug {
					deleteMany2ManyObj["table_to"] = relation.TableFrom
				} else if relation.TableFrom == req.TableSlug {
					deleteMany2ManyObj["table_to"] = relation.TableTo
				}

				deleteMany2Many = append(deleteMany2Many, deleteMany2ManyObj)
			}
			dataToAnalytics[field.Slug] = data[field.Slug]
		} else if field.Type == "MULTISELECT" {
			val, ok := data[field.Slug]
			if field.Required && (!ok || len(cast.ToSlice(val)) == 0) {
				return map[string]interface{}{}, fmt.Errorf("multiselect field is required")
			}
		}
	}

	fieldTypes["guid"] = "String"

	return data, nil
}

func GetItem(ctx context.Context, conn *pgxpool.Pool, tableSlug, guid string) (map[string]interface{}, error) {

	query := fmt.Sprintf(`SELECT * FROM "%s" WHERE guid = $1`, tableSlug)

	rows, err := conn.Query(ctx, query, guid)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer rows.Close()

	data := make(map[string]interface{})

	for rows.Next() {

		values, err := rows.Values()
		if err != nil {
			return map[string]interface{}{}, err
		}

		for i, value := range values {

			if strings.Contains(string(rows.FieldDescriptions()[i].Name), "_id") || string(rows.FieldDescriptions()[i].Name) == "guid" {
				if arr, ok := value.([16]uint8); ok {
					value = ConvertGuid(arr)
				}
			}

			data[string(rows.FieldDescriptions()[i].Name)] = value
		}
	}

	return data, nil
}

func GetItems(ctx context.Context, conn *pgxpool.Pool, req models.GetItemsBody) ([]map[string]interface{}, int, error) {
	const maxRetries = 3
	var (
		relations   []models.Relation
		relationMap = make(map[string]map[string]interface{})
	)

	tableSlug := req.TableSlug
	params := req.Params
	fields := req.FieldsMap
	searchCondition := " OR "
	order := " ORDER BY created_at DESC "

	table, err := TableFindOne(ctx, conn, tableSlug)
	if err != nil {
		return nil, 0, err
	}

	if !table.OrderBy {
		order = " ORDER BY created_at ASC "
	}

	query := fmt.Sprintf(`SELECT * FROM %s `, tableSlug)

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s `, tableSlug)
	filter := " WHERE 1=1 "
	limit := " LIMIT 20 "
	offset := " OFFSET 0"

	if tableSlug == "user" {
		query = `SELECT * FROM "user" `

		countQuery = `SELECT COUNT(*) FROM "user" `
	}

	args := []interface{}{}
	argCount := 1

	for key, val := range params {
		if key == "limit" {
			limit = fmt.Sprintf(" LIMIT %d ", cast.ToInt(val))
		} else if key == "offset" {
			offset = fmt.Sprintf(" OFFSET %d ", cast.ToInt(val))
		} else if key == "order" {
			orders := cast.ToStringMap(val)
			counter := 0

			if len(orders) > 0 {
				order = " ORDER BY "
			}

			for k, v := range orders {
				oType := " ASC"
				if cast.ToInt(v) == -1 {
					oType = " DESC"
				}

				if counter == 0 {
					order += fmt.Sprintf(" %s"+oType, k)
				} else {
					order += fmt.Sprintf(", %s"+oType, k)
				}
				counter++
			}
		} else {
			_, ok := fields[key]

			if ok {

				switch val.(type) {
				case []string:
					filter += fmt.Sprintf(" AND %s IN($%d) ", key, argCount)
					args = append(args, pq.Array(val))
				case int, float32, float64, int32:
					filter += fmt.Sprintf(" AND %s = $%d ", key, argCount)
					args = append(args, val)
				case []interface{}:
					if fields[key].Type == "MULTISELECT" {
						filter += fmt.Sprintf(" AND %s && $%d", key, argCount)
						args = append(args, pq.Array(val))
					} else {
						filter += fmt.Sprintf(" AND %s = ANY($%d) ", key, argCount)
						args = append(args, pq.Array(val))
					}

				case map[string]interface{}:
					newOrder := cast.ToStringMap(val)

					for k, val := range newOrder {

						switch val.(type) {
						case string:
							if cast.ToString(val) == "" {
								continue
							}
						case int, float32, float64, int32:
							if cast.ToFloat32(val) == 0 {
								continue
							}
						}

						if k == "$gt" {
							filter += fmt.Sprintf(" AND %s > $%d ", key, argCount)
						} else if k == "$gte" {
							filter += fmt.Sprintf(" AND %s >= $%d ", key, argCount)
						} else if k == "$lt" {
							filter += fmt.Sprintf(" AND %s < $%d ", key, argCount)
						} else if k == "$lte" {
							filter += fmt.Sprintf(" AND %s <= $%d ", key, argCount)
						} else if k == "$in" {
							filter += fmt.Sprintf(" AND %s::varchar = ANY($%d)", key, argCount)
						}

						args = append(args, val)

						argCount++
					}

					argCount--
				default:
					if strings.Contains(key, "_id") || key == "guid" {
						if tableSlug == "client_type" {
							filter += " AND guid = ANY($1::uuid[]) "

							args = append(args, pq.Array(cast.ToStringSlice(val)))
						} else {
							filter += fmt.Sprintf(" AND %s = $%d ", key, argCount)
							args = append(args, val)
						}
					} else {
						filter += fmt.Sprintf(" AND %s ~* $%d ", key, argCount)
						args = append(args, val)
					}
				}

				argCount++
			}
		}
	}

	searchValue := cast.ToString(params["search"])
	if len(searchValue) > 0 {
		for idx, val := range req.SearchFields {
			if idx == 0 {
				filter += " AND ("
				searchCondition = ""
			} else {
				searchCondition = " OR "
			}
			filter += fmt.Sprintf(" %s %s ~* $%d ", searchCondition, val, argCount)
			args = append(args, searchValue)
			argCount++

			if idx == len(req.SearchFields)-1 {
				filter += " ) "
			}
		}
	}

	countQuery += filter
	query += filter + order + limit + offset

	// fmt.Println(query)
	// fmt.Println(args...)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		rows, err := conn.Query(ctx, query, args...)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.SQLState() == "0A000" {
				continue
			}
			return nil, 0, err
		}
		defer rows.Close()

		var result []map[string]interface{}

		skipFields := map[string]bool{
			"created_at": true,
			"updated_at": true,
		}

		withRelations := cast.ToBool(params["with_relations"])
		if withRelations {
			relationQuery := `
			SELECT
				id,
				table_from,
				table_to,
				field_from,
				type
			FROM
				relation
			WHERE  table_from = $1 OR table_to = $1 `

			relRows, err := conn.Query(context.Background(), relationQuery, tableSlug)
			if err != nil {
				return nil, 0, err
			}
			defer relRows.Close()

			for relRows.Next() {
				relation := models.Relation{}

				err := relRows.Scan(
					&relation.Id,
					&relation.TableFrom,
					&relation.TableTo,
					&relation.FieldFrom,
					&relation.Type,
				)
				if err != nil {
					return nil, 0, err
				}

				if relation.Type == config.MANY2MANY || relation.Type == config.MANY2DYNAMIC || relation.Type == config.RECURSIVE {
					continue
				}

				relations = append(relations, relation)
			}
		}

	outerloop:
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, 0, err
			}

			data := make(map[string]interface{}, len(values))

			for i, value := range values {
				fieldName := string(rows.FieldDescriptions()[i].Name)

				if skipFields[fieldName] {
					continue
				}
				if fieldName == "deleted_at" && value != nil {
					continue outerloop
				}
				if strings.Contains(fieldName, "_id") || fieldName == "guid" {

					if arr, ok := value.([16]uint8); ok {
						value = ConvertGuid(arr)
					}

					if tableSlug == "client_type" {
						if arr, ok := value.([]interface{}); ok {
							ids := []interface{}{}
							for _, a := range arr {
								ids = append(ids, ConvertGuid(a.([16]uint8)))
							}

							value = ids
						}
					}
				}
				data[fieldName] = value
			}

			if len(relations) > 0 {
				for _, relation := range relations {
					joinId := cast.ToString(data[relation.TableTo+"_id"])
					if _, ok := relationMap[joinId]; ok {
						data[relation.TableTo+"_id_data"] = relationMap[joinId]
						continue
					}
					relationData, err := GetItem(ctx, conn, relation.TableTo, joinId)
					if err != nil {
						return nil, 0, err
					}

					data[relation.TableTo+"_id_data"] = relationData
					relationMap[joinId] = relationData
				}
			}

			result = append(result, data)
		}

		if err = rows.Err(); err != nil {
			if err.Error() == "ERROR: cached plan must not change result type (SQLSTATE 0A000)" {
				continue
			}
			return nil, 0, err
		}

		count := 0
		err = conn.QueryRow(ctx, countQuery, args...).Scan(&count)
		if err != nil {
			return nil, 0, err
		}

		return result, count, nil
	}

	return nil, 0, fmt.Errorf("failed to execute query after %d attempts due to cached plan changes", maxRetries)
}

func ConvertGuid(arr [16]uint8) string {
	guidString := fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		arr[0], arr[1], arr[2], arr[3],
		arr[4], arr[5],
		arr[6], arr[7],
		arr[8], arr[9],
		arr[10], arr[11], arr[12], arr[13], arr[14], arr[15])

	return guidString
}

func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func CalculateFormulaBackend(ctx context.Context, conn *pgxpool.Pool, attributes map[string]interface{}, tableSlug string) (map[string]float32, error) {

	var (
		query    string
		response = make(map[string]float32)

		relationField = tableSlug + "_id"
		table         = strings.Split(cast.ToString(attributes["table_from"]), "#")[0]
		field         = cast.ToString(attributes["sum_field"])

		round = cast.ToInt(attributes["number_of_rounds"])
	)

	// ! SKIP formula_filter
	// formulaFilter := cast.ToSlice(attributes["formula_filters"])
	// for _, v := range formulaFilter {
	// 	el := cast.ToStringMap(v)
	// }

	switch cast.ToString(attributes["type"]) {
	case "SUMM":
		query = fmt.Sprintf(`SELECT %s, SUM(%s) FROM %s GROUP BY %s`, relationField, field, table, relationField)
	case "MAX":
		query = fmt.Sprintf(`SELECT %s, MAX(%s) FROM %s GROUP BY %s`, relationField, field, table, relationField)
	case "AVG":
		query = fmt.Sprintf(`SELECT %s, AVG(%s) FROM %s GROUP BY %s`, relationField, field, table, relationField)
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return map[string]float32{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id  string
			num float32
		)

		err = rows.Scan(&id, &num)
		if err != nil {
			return map[string]float32{}, err
		}

		if round > 0 {
			format := "%." + fmt.Sprint(round) + "f"

			num = cast.ToFloat32(fmt.Sprintf(format, num))
		}

		response[id] = num
	}

	return response, nil
}

func CalculateFormulaFrontend(attributes map[string]interface{}, fields []models.Field, object map[string]interface{}) (interface{}, error) {

	computedFormula := attributes["formula"].(string)

	for _, el := range fields {

		value, ok := object[el.Slug]
		if !ok {
			value = 0
		}

		valueStr := fmt.Sprintf("%v", value)

		computedFormula = strings.ReplaceAll(computedFormula, el.Slug, valueStr)
	}

	result, err := callJS(computedFormula)
	if err != nil {
		return nil, err
	}

	// expression, err := govaluate.NewEvaluableExpression(computedFormula)
	// if err != nil {
	// 	return "", err
	// }

	// result, err := expression.Evaluate(nil)
	// if err != nil {
	// 	return "", err
	// }

	return result, nil
}

func AppendMany2Many(ctx context.Context, conn *pgxpool.Pool, req []map[string]interface{}) error {

	for _, data := range req {

		idTo := cast.ToStringSlice(data["id_to"])
		idTos := []string{}
		idFrom := cast.ToString(data["id_from"])

		query := fmt.Sprintf(`SELECT %s_ids FROM %s WHERE guid = $1`, data["table_to"], data["table_from"])

		err := conn.QueryRow(ctx, query, idFrom).Scan(&idTos)
		if err != nil {
			return err
		}

		for _, id := range idTo {
			if len(idTos) > 0 {
				if !contains(idTos, id) {
					idTos = append(idTos, id)
				}
			} else {
				idTos = []string{id}
			}
		}

		query = fmt.Sprintf(`UPDATE %s SET %s_ids=$1 WHERE guid = $2`, data["table_from"], data["table_to"])
		_, err = conn.Exec(ctx, query, pq.Array(idTos), idFrom)
		if err != nil {
			return err
		}

		for _, id := range idTo {
			ids := []string{}
			query := fmt.Sprintf(`SELECT %s_ids FROM %s WHERE guid = $1`, data["table_from"], data["table_to"])

			err = conn.QueryRow(ctx, query, id).Scan(&ids)
			if err != nil {
				return err
			}

			if len(ids) > 0 {
				if !contains(ids, idFrom) {
					ids = append(ids, idFrom)
				}
			} else {
				ids = []string{idFrom}
			}

			query = fmt.Sprintf(`UPDATE %s SET %s_ids=$1 WHERE guid = $2`, data["table_to"], data["table_from"])
			_, err = conn.Exec(ctx, query, pq.Array(ids), id)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func contains(arr []string, t string) bool {
	for _, v := range arr {
		if v == t {
			return true
		}
	}
	return false
}

func AddPermissionToFieldv2(ctx context.Context, conn *pgxpool.Pool, fields []models.Field, roleId string, tableSlug string) ([]models.Field, error) {

	var (
		fieldPermissionMap         = make(map[string]models.FieldPermission)
		relationFieldPermissionMap = make(map[string]string)
		fieldIds                   = []string{}
		tableId                    string
		fieldsWithPermissions      = []models.Field{}
	)

	for _, field := range fields {
		fieldId := ""
		if strings.Contains(field.Id, "#") {
			query := `SELECT "id" FROM "table" WHERE "slug" = $1`

			err := conn.QueryRow(ctx, query, tableSlug).Scan(&tableId)
			if err != nil {
				return []models.Field{}, err
			}
			relationID := strings.Split(field.Id, "#")[1]

			query = `SELECT "id" FROM "field" WHERE relation_id = $1 AND table_id = $2`

			err = conn.QueryRow(ctx, query, relationID, tableId).Scan(&fieldId)
			if err != nil {
				return []models.Field{}, err
			}

			if fieldId != "" {
				relationFieldPermissionMap[relationID] = fieldId
				fieldIds = append(fieldIds, fieldId)
				continue
			}
		} else {
			fieldIds = append(fieldIds, field.Id)
		}
	}

	if len(fieldIds) > 0 {
		query := `SELECT
			"guid",
			"role_id",
			"label",
			"table_slug",
			"field_id",
			"edit_permission",
			"view_permission"
		FROM "field_permission" WHERE field_id IN ($1) AND role_id = $2 AND table_slug = $3`

		rows, err := conn.Query(ctx, query, pq.Array(fieldIds), roleId, tableSlug)
		if err != nil {
			return []models.Field{}, err
		}
		defer rows.Close()

		for rows.Next() {
			fp := models.FieldPermission{}

			err = rows.Scan(
				&fp.Guid,
				&fp.RoleId,
				&fp.Label,
				&fp.TableSlug,
				&fp.FieldId,
				&fp.EditPermission,
				&fp.ViewPermission,
			)
			if err != nil {
				return []models.Field{}, err
			}

			fieldPermissionMap[fp.FieldId] = fp
		}
	}

	for _, field := range fields {
		id := field.Id
		if strings.Contains(id, "#") {
			id = relationFieldPermissionMap[strings.Split(id, "#")[1]]
		}
		fieldPer, ok := fieldPermissionMap[id]

		if ok && roleId != "" {

			if field.Attributes != nil {
				decoded := make(map[string]interface{})
				body, err := json.Marshal(field.Attributes)
				if err != nil {
					return []models.Field{}, err
				}
				if err := json.Unmarshal(body, &decoded); err != nil {
					return []models.Field{}, err
				}
				decoded["field_permission"] = fieldPer
				newAtb, err := ConvertMapToStruct(decoded)
				if err != nil {
					return []models.Field{}, err
				}
				field.Attributes = newAtb
			} else {
				atributes := map[string]interface{}{
					"field_permission": fieldPer,
				}

				newAtb, err := ConvertMapToStruct(atributes)
				if err != nil {
					return []models.Field{}, err
				}

				field.Attributes = newAtb
			}
			if !fieldPer.ViewPermission {
				continue
			}
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		} else if roleId == "" {
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		}
	}

	return fieldsWithPermissions, nil
}

func callJS(value string) (string, error) {

	cmd := exec.Command("node", "/js/pkg/js_parser/frontend_formula.js", value)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(output))

	return result, nil
}

func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice, reflect.Chan:
		return v.Len() == 0 || v.IsNil()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil() || IsEmpty(v.Elem().Interface())
	case reflect.Struct:
		return reflect.DeepEqual(value, reflect.Zero(v.Type()).Interface())
	default:
		return reflect.DeepEqual(value, reflect.Zero(v.Type()).Interface())
	}
}
