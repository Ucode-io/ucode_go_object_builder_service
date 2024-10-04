package helper

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cast"
)

func PrepareToCreateInObjectBuilderWithTx(ctx context.Context, conn pgx.Tx, req *nb.CommonMessage, reqBody CreateBody) (map[string]interface{}, []map[string]interface{}, error) {
	var (
		fieldM     = reqBody.FieldMap
		tableSlugs = reqBody.TableSlugs
		fields     = reqBody.Fields
	)

	data, err := ConvertStructToMap(req.Data)
	if err != nil {
		return map[string]interface{}{}, []map[string]interface{}{}, err
	}

	response := data

	// * RANDOM_NUMBER
	{
		randomNumbers, ok := fieldM["RANDOM_NUMBERS"]
		if ok {
			for {
				randNum := GenerateRandomNumber(cast.ToString(randomNumbers.Attributes["prefix"]), cast.ToInt(randomNumbers.Attributes["digit_number"]))

				isExists, err := IsExistsWithTx(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomNumbers.Slug, FieldValue: randNum})
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
				isExists, err := IsExistsWithTx(ctx, conn, IsExistsBody{TableSlug: req.TableSlug, FieldSlug: randomText.Slug, FieldValue: randText})
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

				slug := splitArr[0]
				if !strings.Contains(slug, "_id") {
					slug += "_id"
				}

				if IsEmpty(response[slug]) {
					continue
				}

				var (
					query    = fmt.Sprintf(`SELECT %s FROM "%s" WHERE guid = '%s'`, field.AutofillField, splitArr[0], response[slug])
					autofill interface{}
				)

				if err = conn.QueryRow(ctx, query).Scan(&autofill); err != nil {
					return map[string]interface{}{}, []map[string]interface{}{}, err
				}

				if autofill != nil {
					response[field.Slug] = autofill
					continue
				}
			}

			_, ok := response[field.Slug]
			ok2 := !IsEmpty(response[field.Slug])

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
				delete(response, field.Slug)
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
