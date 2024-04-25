package helper

import (
	"context"
	"encoding/json"
	"strings"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type GetFieldBySlugReq struct {
	Conn    *pgxpool.Pool
	Slug    string
	TableId string
}

type AddPermissionToFieldRequest struct {
	Conn      *pgxpool.Pool
	Fields    []models.Field
	RoleId    string
	TableSlug string
}

func GetFieldBySlug(ctx context.Context, req GetFieldBySlugReq) (map[string]interface{}, error) {

	query := `SELECT id, type, attributes FROM "field" WHERE slug = $1 AND table_id = $2`

	var (
		id, ftype  string
		attributes []byte
	)

	err := req.Conn.QueryRow(ctx, query, req.Slug, req.TableId).Scan(&id, &req.Slug)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":         id,
		"type":       ftype,
		"attributes": attributes,
	}, nil
}

func AddPermissionToField1(ctx context.Context, req AddPermissionToFieldRequest) ([]models.Field, map[string]int, error) {
	var (
		fieldPermissionMap         = make(map[string]models.FieldPermission)
		relationFieldPermissionMap = make(map[string]string)
		unusedFieldsSlugs          = make(map[string]int)
		fieldsWithPermissions      = []models.Field{}
		fieldIds                   = []string{}

		tableId string
	)

	query := `SELECT "id" FROM "table" WHERE "slug" = $1`

	err := req.Conn.QueryRow(ctx, query, req.TableSlug).Scan(&tableId)
	if err != nil {
		return []models.Field{}, map[string]int{}, err
	}

	for _, field := range req.Fields {
		var (
			fieldId    string
			relationId string
		)
		if strings.Contains(field.Id, "#") {
			relationId = strings.Split(field.Id, "#")[1]

			query = `SELECT "id" FROM "field" WHERE relation_id = $1 AND table_id = $2`

			err = req.Conn.QueryRow(ctx, query, relationId, tableId).Scan(&fieldId)
			if err != nil {
				return []models.Field{}, map[string]int{}, err
			}

			if fieldId != "" {
				relationFieldPermissionMap[relationId] = fieldId
				fieldIds = append(fieldIds, fieldId)
				continue
			}
		} else {
			fieldIds = append(fieldIds, field.Id)
		}
	}

	if len(fieldIds) > 0 {
		query := `
			SELECT
				"guid",
				"role_id",
				"label",
				"table_slug",
				"field_id",
				"edit_permission",
				"view_permission"
			FROM "field_permission" 
			WHERE field_id = ANY($1) AND role_id = $2 AND table_slug = $3
		`

		rows, err := req.Conn.Query(ctx, query, pq.Array(fieldIds), req.RoleId, req.TableSlug)
		if err != nil {
			return []models.Field{}, map[string]int{}, err
		}
		defer rows.Close()

		for rows.Next() {
			fp := models.FieldPermission{}

			err := rows.Scan(
				&fp.Guid,
				&fp.RoleId,
				&fp.Label,
				&fp.TableSlug,
				&fp.FieldId,
				&fp.EditPermission,
				&fp.ViewPermission,
			)
			if err != nil {
				return []models.Field{}, map[string]int{}, err
			}

			fieldPermissionMap[fp.FieldId] = fp
		}
	}

	for _, field := range req.Fields {
		id := field.Id
		if strings.Contains(id, "#") {
			id = relationFieldPermissionMap[strings.Split(id, "#")[1]]
		}
		fieldPer, ok := fieldPermissionMap[id]

		if ok && req.RoleId != "" {
			if field.Attributes != nil {
				decoded := make(map[string]interface{})
				body, err := json.Marshal(field.Attributes)
				if err != nil {
					return []models.Field{}, map[string]int{}, err
				}
				if err := json.Unmarshal(body, &decoded); err != nil {
					return []models.Field{}, map[string]int{}, err
				}
				decoded["field_permission"] = fieldPer
				newAtb, err := ConvertMapToStruct(decoded)
				if err != nil {
					return []models.Field{}, map[string]int{}, err
				}
				field.Attributes = newAtb
			} else {
				atributes := map[string]interface{}{
					"field_permission": fieldPer,
				}

				newAtb, err := ConvertMapToStruct(atributes)
				if err != nil {
					return []models.Field{}, map[string]int{}, err
				}

				field.Attributes = newAtb
			}
			if !fieldPer.ViewPermission {
				unusedFieldsSlugs[field.Slug] = 0
				continue
			}
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		} else if req.RoleId == "" {
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		} else {
			unusedFieldsSlugs[field.Slug] = 0
		}
	}
	return fieldsWithPermissions, unusedFieldsSlugs, nil
}
