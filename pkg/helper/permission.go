package helper

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/structpb"
)

func AddPermissionToField(ctx context.Context, conn *pgxpool.Pool, fields []*new_object_builder_service.FieldResponse, roleId string, tableSlug string, projectID string) ([]*new_object_builder_service.FieldResponse, error) {
	unusedFieldsSlugs := make(map[string]int)
	var fieldsWithPermissions []*new_object_builder_service.FieldResponse
	fieldPermissionMap := make(map[string]interface{})
	relationFieldPermissionMap := make(map[string]string)
	var fieldIds []string

	for _, field := range fields {
		if strings.Contains(field.Id, "#") {
			parts := strings.Split(field.Id, "#")
			relationID := parts[1]

			query := `
				SELECT id FROM field WHERE relation_id = $1 AND table_id = (SELECT id FROM table WHERE slug = $2)
			`
			var fieldID string
			err := conn.QueryRow(ctx, query, relationID, tableSlug).Scan(&fieldID)
			if err != nil {
				continue
			}

			relationFieldPermissionMap[relationID] = fieldID
			fieldIds = append(fieldIds, fieldID)
			continue
		}

		fieldIds = append(fieldIds, field.Id)
	}

	query := `
		SELECT
			field_id,
			view_permission,
			field_permission
		FROM field_permission
		WHERE field_id = ANY($1)
			AND role_id = $2
			AND table_slug = $3
	`
	rows, err := conn.Query(ctx, query, fieldIds, roleId, tableSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fieldID string
		var viewPermission bool
		var fieldPermission json.RawMessage

		if err := rows.Scan(&fieldID, &viewPermission, &fieldPermission); err != nil {
			return nil, err
		}

		fieldPermissionMap[fieldID] = map[string]interface{}{
			"view_permission":  viewPermission,
			"field_permission": fieldPermission,
		}
	}

	for _, field := range fields {
		var id string

		if strings.Contains(field.Id, "#") {
			parts := strings.Split(field.Id, "#")
			id = relationFieldPermissionMap[parts[1]]
		} else {
			id = field.Id
		}

		fieldPer, ok := fieldPermissionMap[id]

		if ok && roleId != "" {
			if field.Attributes != nil {
				decodedAttributes := make(map[string]interface{})
				attributesBytes, err := field.Attributes.MarshalJSON()
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal(attributesBytes, &decodedAttributes)
				if err != nil {
					return nil, err
				}
				decodedAttributes["field_permission"] = fieldPer.(map[string]interface{})["field_permission"]
				encodedAttributes, err := json.Marshal(decodedAttributes)
				if err != nil {
					return nil, err
				}
				var structAttributes *structpb.Struct
				err = jsonpb.UnmarshalString(string(encodedAttributes), structAttributes)
				if err != nil {
					return nil, err
				}
				field.Attributes = structAttributes
			} else {
				attributes := map[string]interface{}{
					"field_permission": fieldPer.(map[string]interface{})["field_permission"],
				}
				encodedAttributes, err := json.Marshal(attributes)
				if err != nil {
					return nil, err
				}
				var structAttributes *structpb.Struct
				err = jsonpb.UnmarshalString(string(encodedAttributes), structAttributes)
				if err != nil {
					return nil, err
				}
				field.Attributes = structAttributes
			}
			if fieldPerMap, ok := fieldPer.(map[string]interface{}); ok {
				if !fieldPerMap["view_permission"].(bool) {
					unusedFieldsSlugs[field.Slug] = 0
					continue
				}
			}
			fieldsWithPermissions = append(fieldsWithPermissions, field)
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		} else if roleId == "" || !ok {
			fieldsWithPermissions = append(fieldsWithPermissions, field)
		} else {
			unusedFieldsSlugs[field.Slug] = 0
		}
	}

	return fieldsWithPermissions, nil
}

func AddPermissionToTab(ctx context.Context, relation map[string]interface{}, conn *pgxpool.Pool, roleId string, tableSlug string, projectID string) (map[string]interface{}, error) {

	if projectID == "" {
		fmt.Println("WARNING: Using default project ID in [helper.addPermission.toRelationTab]...")
	}

	query := `
        SELECT 
			"guid"
			"role_id"
			"view_id"
			"view"
			"edit"
			"delete"
		FROM view_relation_permission
        WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
    `
	var (
		guid    string
		role_id string
		view_id string
		view    bool
		edit    bool
		delete  bool
	)
	err := conn.QueryRow(ctx, query, roleId, tableSlug, relation["id"]).Scan(&guid, &role_id, &view_id, &view, &edit, &delete)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	encodedPermission := make(map[string]interface{})
	encodedPermission["guid"] = guid
	encodedPermission["role_id"] = roleId
	encodedPermission["view_id"] = view_id
	encodedPermission["edit"] = edit
	encodedPermission["delete"] = delete

	relation["permission"] = encodedPermission

	return relation, nil
}
