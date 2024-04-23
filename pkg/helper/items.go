package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/models"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/spf13/cast"
)

func PrepareToCreateInObjectBuilder(ctx context.Context, conn *pgxpool.Pool, req *nb.CommonMessage) (map[string]interface{}, []map[string]interface{}, error) {

	var (
		response = make(map[string]interface{})
		tableId  string
	)

	data, err := ConvertStructToMap(req.Data)
	if err != nil {
		return map[string]interface{}{}, []map[string]interface{}{}, err
	}

	response = data

	query := `SELECT id FROM "table" WHERE slug = $1`

	err = conn.QueryRow(ctx, query, req.TableSlug).Scan(&tableId)
	if err != nil {
		return map[string]interface{}{}, []map[string]interface{}{}, err
	}

	// * RANDOM_NUMBER
	{
		randomNumbers, err := GetFieldByType(ctx, conn, tableId, "RANDOM_NUMBERS")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		randNum := GenerateRandomNumber(cast.ToString(randomNumbers.Attributes["prefix"]), cast.ToInt(randomNumbers.Attributes["digit_number"]))

		isExists, err := IsExists(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomNumbers.Slug, FieldValue: randNum})
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		if isExists {
			return PrepareToCreateInObjectBuilder(ctx, conn, req)
		} else {
			response[randomNumbers.Slug] = randNum
		}
	}

	// * RANDOM_TEXT
	{
		randomText, err := GetFieldByType(ctx, conn, tableId, "RANDOM_TEXT")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		randText := GenerateRandomString(cast.ToString(randomText.Attributes["prefix"]), cast.ToInt(randomText.Attributes["digit_number"]))

		isExists, err := IsExists(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomText.Slug, FieldValue: randText})
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		if isExists {
			return PrepareToCreateInObjectBuilder(ctx, conn, req)
		} else {
			response[randomText.Slug] = randText
		}
	}

	// * RANDOM_UUID
	{
		randomUuid, err := GetFieldByType(ctx, conn, tableId, "RANDOM_UUID")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		response[randomUuid.Slug] = uuid.NewString()
	}

	// * MANUAL_STRING
	{
		manual, err := GetFieldByType(ctx, conn, tableId, "MANUAL_STRING")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}
		fields, err := ConvertStructToMap(req.Data)
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		text := cast.ToString(manual.Attributes["formula"])

		query := `SELECT "slug" FROM "field" WHERE table_id = $1 ORDER BY LENGTH("slug") DESC`

		rows, err := conn.Query(ctx, query, tableId)
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				slug  string
				value interface{}
			)

			err := rows.Scan(&slug)
			if err != nil {
				return map[string]interface{}{}, []map[string]interface{}{}, err
			}

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

	// * INCREMENT_ID
	{

		incrementField, err := GetFieldByType(ctx, conn, tableId, "INCREMENT_ID")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		incrementBy := 0

		query := `UPDATE "incrementseqs" SET increment_by = increment_by + 1  WHERE table_slug = $1 AND field_slug = $2 RETURNING increment_by AS old_value`

		err = conn.QueryRow(ctx, query, req.TableSlug, incrementField.Slug).Scan(&incrementBy)
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		response[incrementField.Slug] = cast.ToString(incrementField.Attributes["perfix"]) + "-" + fmt.Sprintf("%09d", incrementBy)
	}

	// * INCREMENT_NUMBER
	{
		incrementNum, err := GetFieldByType(ctx, conn, tableId, "INCREMENT_NUMBER")
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		incNum := 0

		query := fmt.Sprintf(`SELECT %s FROM %s ORDER BY created_at DESC LIMIT 1`, incrementNum.Slug, req.TableSlug)

		err = conn.QueryRow(ctx, query).Scan(&incNum)
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		format := "%d"

		if cast.ToInt(incrementNum.Attributes["digit_number"]) > 0 {
			format = "%0" + cast.ToString(incrementNum.Attributes["digit_number"]) + "d"
		}

		response[incrementNum.Slug] = cast.ToString(incrementNum.Attributes["perfix"]) + fmt.Sprintf(format, incNum)
	}

	query = `SELECT
	"id",
	"type",
	"attributes",
	"relation_id",
	"autofill_table",
	"autofill_field,
	"slug"
FROM "field" WHERE table_id = $1`

	rows, err := conn.Query(ctx, query, tableId)
	if err != nil {
		return map[string]interface{}{}, []map[string]interface{}{}, err
	}
	defer rows.Close()

	fields := []models.Field{}

	for rows.Next() {
		field := models.Field{}

		var (
			atr = []byte{}
		)

		err = rows.Scan(
			&field.Id,
			&field.Type,
			&atr,
			&field.RelationId,
			&field.TableSlug,
			&field.AutofillTable,
			&field.AutofillField,
			&field.Slug,
		)
		if err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return map[string]interface{}{}, []map[string]interface{}{}, err
		}

		fields = append(fields, field)
	}

	// * AUTOFILL
	{
		for _, field := range fields {

			attributes, err := ConvertStructToMap(req.Data)
			if err != nil {
				return map[string]interface{}{}, []map[string]interface{}{}, err
			}

			_, ok := response[field.Slug]
			if !ok && field.AutofillField != "" && field.AutofillTable != "" {
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

			_, ok2 := attributes["defaultValue"]
			defaultValues := cast.ToSlice(attributes["default_values"])
			if ok2 && !ok {
				_, fOk := FIELD_TYPES[field.Type]
				if fOk {
					response[field.Type] = cast.ToInt(attributes["defaultValue"])
				} else if field.Type == "DATE_TIME" || field.Type == "DATE" {
					response[field.Type] = time.Now().Format(time.RFC3339)
				} else if field.Type == "SWITCH" {
					defaultValue := strings.ToLower(cast.ToString(attributes["defaultValue"]))

					if defaultValue == "true" {
						response[field.Type] = true
					} else if defaultValue == "false" {
						response[field.Type] = false
					}
				} else {
					response[field.Type] = attributes["defaultValue"]
				}
			} else if len(defaultValues) > 0 && !ok {
				response[field.Type] = defaultValues[0]
			}
		}
	}

	query = `SELECT table_to, table_from FROM "relation" WHERE id = $1`

	appendMany2ManyObjects := []map[string]interface{}{}
	// * AppendMany2ManyObjects
	{
		for _, field := range fields {
			var (
				tableTo, tableFrom string
			)

			if field.Type == "LOOKUPS" {
				_, ok := response[field.Slug]
				if ok {
					err = conn.QueryRow(ctx, query, field.RelationId).Scan(&tableTo, &tableFrom)
					if err != nil {
						return map[string]interface{}{}, []map[string]interface{}{}, err
					}

					appendMany2Many := make(map[string]interface{})

					appendMany2Many = map[string]interface{}{
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

func PrepareToUpdateInObjectBuilder(ctx context.Context, conn *pgxpool.Pool, req *nb.CommonMessage) error {

	data, err := ConvertStructToMap(req.Data)
	if err != nil {
		return err
	}

	oldData, err := GetItem(ctx, conn, req.TableSlug, cast.ToString(data["guid"]))
	if err != nil {
		return err
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
		return err
	}
	defer rows.Close()

	for rows.Next() {
		id := ""

		err := rows.Scan(&id)
		if err != nil {
			return err
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

	rows, err = conn.Query(ctx, query, pq.Array(relationIds))
	if err != nil {
		return err
	}

	for rows.Next() {
		rel := models.Relation{}

		err = rows.Scan(
			&rel.Id,
			&rel.TableTo,
			&rel.TableFrom,
		)
		if err != nil {
			return err
		}

		relationMap[rel.Id] = rel
	}

	query = `SELECT "id", "type", "attributes", "slug", "relation_id", "required", FROM "field" as f JOIN "table" as t ON t.id = f.table_id WHERE t.slug = $1`

	rows, err = conn.Query(ctx, query)
	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			field      = models.Field{}
			attributes = []byte{}
		)

		err = rows.Scan(
			&field.Id,
			&field.Type,
			&attributes,
			&field.Slug,
			&field.RelationId,
			&field.Required,
		)
		if err != nil {
			return err
		}

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
				return fmt.Errorf("multiselect field is required")
			}
		}
	}

	fieldTypes["guid"] = "String"

	return nil
}

func GetItem(ctx context.Context, conn *pgxpool.Pool, tableSlug, guid string) (map[string]interface{}, error) {

	query := `SELECT * FROM %s WHERE guid = $1`

	rows, err := conn.Query(ctx, query, guid)
	if err != nil {
		return map[string]interface{}{}, err
	}
	defer rows.Close()

	data := make(map[string]interface{})

	for rows.Next() {

		values, err := rows.Values()
		if err != nil {
			log.Fatalf("Error getting row values: %v", err)
		}

		for i, value := range values {

			if strings.Contains(string(rows.FieldDescriptions()[i].Name), "_id") {
				if arr, ok := value.([16]uint8); ok {
					value = ConvertGuid(arr)
				}
			}

			data[string(rows.FieldDescriptions()[i].Name)] = value
		}
	}

	return data, nil
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
