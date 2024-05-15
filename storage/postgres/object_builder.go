package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	excel "github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type objectBuilderRepo struct {
	db *pgxpool.Pool
}

func NewObjectBuilder(db *pgxpool.Pool) storage.ObjectBuilderRepoI {
	return &objectBuilderRepo{
		db: db,
	}
}

func (o *objectBuilderRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.GetProjectId())

	conn := psqlpool.Get(req.GetProjectId())

	if req.TableSlug == "client_type" {
		query := `
			SELECT
				"guid",
				"project_id",
				"name",
				"self_register",
				"self_recover",
				"client_platform_ids",
				"confirm_by",
				"is_system"
			FROM client_type
		`

		rows, err := conn.Query(ctx, query)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		defer rows.Close()

		data := make([]models.ClientType, 0)
		for rows.Next() {
			var clientType models.ClientType

			err = rows.Scan(
				&clientType.Guid,
				&clientType.ProjectId,
				&clientType.Name,
				&clientType.SelfRegister,
				&clientType.SelfRecover,
				&clientType.ClientPlatformIds,
				&clientType.ConfirmBy,
				&clientType.IsSystem,
			)
			if err != nil {
				return &nb.CommonMessage{}, err
			}

			data = append(data, clientType)
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		var dataStruct structpb.Struct
		jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

		err = json.Unmarshal(jsonBytes, &dataStruct)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		return &nb.CommonMessage{
			TableSlug:     req.TableSlug,
			ProjectId:     req.ProjectId,
			Data:          &dataStruct,
			VersionId:     req.VersionId,
			CustomMessage: req.CustomMessage,
			IsCached:      req.IsCached,
		}, err
	} else {
		query := `
			SELECT
				"guid",
				"project_id",
				"name",
				"client_platform_id",
				"client_type_id",
				"is_system"
			FROM "role"
		`

		rows, err := conn.Query(ctx, query)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		defer rows.Close()

		data := make([]models.Role, 0)
		for rows.Next() {
			var role models.Role

			err = rows.Scan(
				&role.Guid,
				&role.ProjectId,
				&role.Name,
				&role.ClientPlatformId,
				&role.ClientTypeId,
				&role.IsSystem,
			)
			if err != nil {
				return &nb.CommonMessage{}, err
			}

			data = append(data, role)
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		var dataStruct structpb.Struct
		jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

		err = json.Unmarshal(jsonBytes, &dataStruct)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		return &nb.CommonMessage{
			TableSlug:     req.TableSlug,
			ProjectId:     req.ProjectId,
			Data:          &dataStruct,
			VersionId:     req.VersionId,
			CustomMessage: req.CustomMessage,
			IsCached:      req.IsCached,
		}, err
	}
}

func (o *objectBuilderRepo) GetListConnection(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.GetProjectId())

	conn := psqlpool.Get(req.GetProjectId())

	query := `
		SELECT
			"guid",
			"table_slug",
			"view_slug",
			"view_label",
			"name",
			"type",
			"icon",
			"main_table_slug",
			"field_slug",
			"client_type_id"
		FROM "connection"
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer rows.Close()

	data := make([]models.Connection, 0)
	for rows.Next() {
		var connection models.Connection

		err = rows.Scan(
			&connection.Guid,
			&connection.TableSlug,
			&connection.ViewSlug,
			&connection.ViewLabel,
			&connection.Name,
			&connection.Type,
			&connection.Icon,
			&connection.MainTableSlug,
			&connection.FieldSlug,
			&connection.ClientTypeId,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		data = append(data, connection)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	var dataStruct structpb.Struct
	jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

	err = json.Unmarshal(jsonBytes, &dataStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          &dataStruct,
		VersionId:     req.VersionId,
		CustomMessage: req.CustomMessage,
		IsCached:      req.IsCached,
	}, err
}

func (o *objectBuilderRepo) GetTableDetails(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		fields = []models.Field{}
		// relations       = []models.Relation{}
		views           = []models.View{}
		params          = make(map[string]interface{})
		relationsFields = []models.Field{}
	)

	body, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error marshalling request data")
	}
	if err := json.Unmarshal(body, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error unmarshalling request data")
	}

	query := `SELECT 
		f.id,
		f."table_id",
		f."required",
		f."slug",
		f."label",
		f."default",
		f."type",
		f."index",
		f."attributes",
		f."is_visible",
		f.autofill_field,
		f.autofill_table,
		f."unique",
		f."automatic",
		f.relation_id
	FROM "field" as f 
	JOIN "table" as t ON f.table_id = t.id 
	WHERE t.slug = $1`

	rows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field             = models.Field{}
			attributes        = []byte{}
			relationIdNull    sql.NullString
			autofillField     sql.NullString
			autofillTable     sql.NullString
			defaultStr, index sql.NullString
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.Required,
			&field.Slug,
			&field.Label,
			&defaultStr,
			&field.Type,
			&index,
			&attributes,
			&field.IsVisible,
			&autofillField,
			&autofillTable,
			&field.Unique,
			&field.Automatic,
			&relationIdNull,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		field.RelationId = relationIdNull.String
		field.AutofillField = autofillField.String
		field.AutofillTable = autofillTable.String
		field.Default = defaultStr.String
		field.Index = index.String

		if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
			query := `
				SELECT
					"view_fields",
					"table_from",
					"table_to"
				FROM "relation" r
				WHERE id = $1
			`

			relationRows, err := conn.Query(ctx, query, field.RelationId)
			if err != nil {
				return &nb.CommonMessage{}, err
			}
			defer relationRows.Close()

			for relationRows.Next() {
				var (
					viewFields []string
					tableFrom  string
					tableTo    string
				)
				err = relationRows.Scan(&viewFields, &tableFrom, &tableTo)
				if err != nil {
					return &nb.CommonMessage{}, err
				}

				if tableFrom != req.TableSlug {
					field.TableSlug = tableFrom
				} else {
					field.TableSlug = tableTo
				}

				var fieldObjects []models.Field
				for _, id := range viewFields {
					var field models.Field

					query = `
						SELECT 
							f.id,
							f."table_id",
							f."required",
							f."slug",
							f."label",
							f."default",
							f."type",
							f."index",
							f."attributes",
							f."is_visible",
							f.autofill_field,
							f.autofill_table,
							f."unique",
							f."automatic",
							f.relation_id
						FROM "field" as f 
						WHERE f.id = $1
					`

					err = conn.QueryRow(context.Background(), query, id).Scan(
						&field.Id,
						&field.TableId,
						&field.Required,
						&field.Slug,
						&field.Label,
						&defaultStr,
						&field.Type,
						&index,
						&attributes,
						&field.IsVisible,
						&autofillField,
						&autofillTable,
						&field.Unique,
						&field.Automatic,
						&relationIdNull,
					)
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					field.RelationId = relationIdNull.String
					field.AutofillField = autofillField.String
					field.AutofillTable = autofillTable.String
					field.Default = defaultStr.String
					field.Index = index.String

					if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
						return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
					}

					fieldObjects = append(fieldObjects, field)
				}

				field.ViewFields = fieldObjects
			}
		}

		fields = append(fields, field)
	}

	// query = `
	// 	SELECT
	// 		"id",
	// 		"table_from",
	// 		"table_to",
	// 		"type",
	// 		"view_fields"
	// 	FROM "relation" r
	// 	WHERE "table_from" = $1 OR "table_to" = $1
	// `

	// relationRows, err := conn.Query(ctx, query, req.TableSlug)
	// if err != nil {
	// 	return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relations by table slug")
	// }
	// defer relationRows.Close()

	// for relationRows.Next() {
	// 	relation := models.Relation{}

	// 	err = relationRows.Scan(
	// 		&relation.Id,
	// 		&relation.TableFrom,
	// 		&relation.TableTo,
	// 		&relation.Type,
	// 		&relation.ViewFields,
	// 	)
	// 	if err != nil {
	// 		return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relations")
	// 	}

	// 	relations = append(relations, relation)
	// }

	query = `SELECT 
		"id",
		"attributes",
		"table_slug",
		"type",
		"columns"
	FROM "view" WHERE "table_slug" = $1`

	viewRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting views by table slug")
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var (
			attributes []byte
			view       = models.View{}
		)

		err := viewRows.Scan(
			&view.Id,
			&attributes,
			&view.TableSlug,
			&view.Type,
			&view.Columns,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning views")
		}

		if view.Columns == nil {
			view.Columns = []string{}
		}

		if err := json.Unmarshal(attributes, &view.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling view attributes")
		}

		views = append(views, view)
	}

	query = `SELECT 
		"guid",
		"role_id",
		"view_id",
		"view",
		"edit",
		"delete"
	FROM "view_permission" WHERE "view_id" = $1 AND "role_id" = $2`

	for _, view := range views {

		vp := models.ViewPermission{}

		err = conn.QueryRow(ctx, query, view.Id, cast.ToString(params["role_id_from_token"])).Scan(
			&vp.Guid,
			&vp.RoleId,
			&vp.ViewId,
			&vp.View,
			&vp.Edit,
			&vp.Delete,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning view permissions")
		}

		view.Attributes["view_permission"] = vp
	}

	// if cast.ToBool(params["with_relations"]) {

	// 	var (
	// 		relationTableToSlugs = []string{}
	// 		relationTableIds     = []string{}
	// 		relationTablesMap    = make(map[string]models.Table)
	// 	)

	// 	for _, relation := range relations {
	// 		if relation.Type != "Many2Dynamic" {
	// 			if relation.Type == "Many2Many" && relation.TableTo == req.TableSlug {
	// 				relation.TableTo = relation.TableFrom
	// 			}

	// 			relationTableToSlugs = append(relationTableToSlugs, relation.TableTo)
	// 		}
	// 	}

	// 	if len(relationTableToSlugs) > 0 {

	// 		query = `SELECT
	// 			"id",
	// 			"slug",
	// 			"label"
	// 		FROM "table" WHERE "slug" IN ($1)`

	// 		rows, err := conn.Query(ctx, query, pq.Array(relationTableToSlugs))
	// 		if err != nil {
	// 			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting tables by slugs")
	// 		}
	// 		defer rows.Close()

	// 		for rows.Next() {
	// 			table := models.Table{}

	// 			err = rows.Scan(
	// 				&table.Id,
	// 				&table.Slug,
	// 				&table.Label,
	// 			)
	// 			if err != nil {
	// 				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning tables")
	// 			}

	// 			relationTableIds = append(relationTableIds, table.Id)
	// 			_, ok := relationTablesMap[table.Slug]
	// 			if !ok {
	// 				relationTablesMap[table.Slug] = table
	// 			}
	// 		}
	// 	}

	// 	var (
	// 		relationFieldSlugsR = []string{}
	// 		relationFieldsMap   = make(map[string][]models.Field)
	// 	)
	// 	if len(relationTableIds) > 0 {
	// 		query = `SELECT
	// 			id,
	// 			type,
	// 			slug,
	// 			table_id
	// 		FROM "field" WHERE table_id IN ($1)`

	// 		rows, err := conn.Query(ctx, query, pq.Array(relationTableIds))
	// 		if err != nil {
	// 			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table ids")
	// 		}
	// 		defer rows.Close()

	// 		for rows.Next() {
	// 			field := models.Field{}

	// 			err = rows.Scan(
	// 				&field.Id,
	// 				&field.Type,
	// 				&field.Slug,
	// 				&field.TableId,
	// 			)
	// 			if err != nil {
	// 				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
	// 			}

	// 			if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
	// 				tableSlug := ""
	// 				if field.Type == "LOOKUP" {
	// 					tableSlug = field.Slug[:len(field.Slug)-3]
	// 				} else {
	// 					tableSlug = field.Slug[:len(field.Slug)-4]
	// 				}

	// 				relationFieldSlugsR = append(relationFieldSlugsR, tableSlug)
	// 			}

	// 			_, ok := relationFieldsMap[field.TableId]
	// 			if ok {
	// 				relationFieldsMap[field.TableId] = append(relationFieldsMap[field.TableId], field)
	// 			} else {
	// 				relationFieldsMap[field.TableId] = []models.Field{field}
	// 			}
	// 		}
	// 	}

	// 	var (
	// 		childRelationsMap      = make(map[string]models.Relation)
	// 		viewFieldIds           = []string{}
	// 		viewFieldsMap          = make(map[string]models.Field)
	// 		childRelationTablesMap = make(map[string]string)
	// 	)

	// 	if len(relationTableToSlugs) > 0 && len(relationFieldSlugsR) > 0 {
	// 		query = `SELECT
	// 			"id",
	// 			"table_from",
	// 			"table_to",
	// 			"type",
	// 			view_fields
	// 		FROM "relation" WHERE "table_from" IN ($1) AND "table_to" IN ($2)`

	// 		rows, err := conn.Query(ctx, query, pq.Array(relationTableToSlugs), pq.Array(relationFieldSlugsR))
	// 		if err != nil {
	// 			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relations by table slugs and field slugs")
	// 		}
	// 		defer rows.Close()

	// 		for rows.Next() {
	// 			relation := models.Relation{}

	// 			err = rows.Scan(
	// 				&relation.Id,
	// 				&relation.TableFrom,
	// 				&relation.TableTo,
	// 				&relation.Type,
	// 				pq.Array(&relation.ViewFields),
	// 			)
	// 			if err != nil {
	// 				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relations")
	// 			}

	// 			_, ok := childRelationsMap[relation.TableFrom+"_"+relation.TableTo]
	// 			if !ok {
	// 				childRelationsMap[relation.TableFrom+"_"+relation.TableTo] = relation
	// 			}

	// 			viewFieldIds = relation.ViewFields
	// 		}
	// 	}

	// 	if len(viewFieldIds) > 0 {
	// 		query = `SELECT
	// 			id,
	// 			"table_id",
	// 			"required",
	// 			"slug",
	// 			"label",
	// 			"default",
	// 			"type",
	// 			"index",
	// 			"attributes",
	// 			"is_visible",
	// 			autofill_field,
	// 			autofill_table,
	// 			"unique",
	// 			"automatic",
	// 			relation_id
	// 		FROM "field" WHERE id IN ($1)`

	// 		rows, err := conn.Query(ctx, query, pq.Array(viewFieldIds))
	// 		if err != nil {
	// 			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by ids")
	// 		}
	// 		defer rows.Close()

	// 		for rows.Next() {
	// 			var (
	// 				field          = models.Field{}
	// 				attributes     = []byte{}
	// 				relationIdNull sql.NullString
	// 			)
	// 			err = rows.Scan(
	// 				&field.Id,
	// 				&field.TableId,
	// 				&field.Required,
	// 				&field.Slug,
	// 				&field.Label,
	// 				&field.Default,
	// 				&field.Type,
	// 				&field.Index,
	// 				&attributes,
	// 				&field.IsVisible,
	// 				&field.AutofillField,
	// 				&field.AutofillTable,
	// 				&field.Unique,
	// 				&field.Automatic,
	// 				&relationIdNull,
	// 			)
	// 			if err != nil {
	// 				return &nb.CommonMessage{}, err
	// 			}

	// 			field.RelationId = relationIdNull.String

	// 			if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
	// 				return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
	// 			}

	// 			viewFieldsMap[field.Id] = field
	// 		}
	// 	}

	// 	if len(relationFieldSlugsR) > 0 {
	// 		query = `SELECT
	// 			"slug",
	// 			"label"
	// 		FROM "table" WHERE slug IN ($1)`

	// 		rows, err := conn.Query(ctx, query, pq.Array(relationFieldSlugsR))
	// 		if err != nil {
	// 			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting tables by slugs")
	// 		}
	// 		defer rows.Close()

	// 		for rows.Next() {
	// 			var (
	// 				tableSlug  string
	// 				tableLabel string
	// 			)

	// 			err = rows.Scan(
	// 				&tableSlug,
	// 				&tableLabel,
	// 			)
	// 			if err != nil {
	// 				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning tables")
	// 			}

	// 			_, ok := childRelationTablesMap[tableSlug]
	// 			if !ok {
	// 				childRelationTablesMap[tableSlug] = tableLabel
	// 			}
	// 		}
	// 	}

	// 	for _, relation := range relations {
	// 		if relation.Type != "Many2Dynamic" {
	// 			if relation.Type == "Many2Many" && relation.TableTo == req.TableSlug {
	// 				relation.TableTo = relation.TableFrom
	// 			}
	// 		}

	// 		relationTable := relationTablesMap[relation.TableTo]
	// 		tableRelationFields := relationFieldsMap[relationTable.Id]

	// 		for _, field := range tableRelationFields {
	// 			changedField := models.Field{}
	// 			if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
	// 				var (
	// 					viewFields = []models.Field{}
	// 				)
	// 				tableSlug := ""
	// 				if field.Type == "LOOKUP" {
	// 					tableSlug = field.Slug[:len(field.Slug)-3]
	// 				} else {
	// 					tableSlug = field.Slug[:len(field.Slug)-4]
	// 				}

	// 				childRelation, ok := childRelationsMap[relationTable.Slug+"_"+tableSlug]
	// 				if ok {
	// 					for _, view_field := range childRelation.ViewFields {
	// 						viewField, ok := viewFieldsMap[view_field]
	// 						if ok {
	// 							viewFields = append(viewFields, viewField)
	// 						}
	// 					}
	// 				}

	// 				field.ViewFields = viewFields
	// 				field.Label = childRelationTablesMap[tableSlug]
	// 				changedField = field
	// 				changedField.PathSlug = relationTable.Slug + "_id_data." + field.Slug
	// 				changedField.TableSlug = tableSlug
	// 				relationsFields = append(relationsFields, changedField)
	// 			} else {
	// 				changedField = field
	// 				changedField.PathSlug = relationTable.Slug + "_id_data." + field.Slug
	// 				relationsFields = append(relationsFields, changedField)
	// 			}
	// 		}
	// 	}
	// }

	fieldsWithPermissions, _, err := helper.AddPermissionToField1(ctx, helper.AddPermissionToFieldRequest{Conn: conn, Fields: fields, RoleId: cast.ToString(params["role_id_from_token"]), TableSlug: req.TableSlug})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while adding permissions to fields")
	}

	decodedFields := []models.Field{}

	for _, element := range fieldsWithPermissions {
		if element.Attributes != nil && !(element.Type == "LOOKUP" || element.Type == "LOOKUPS" || element.Type == "DYNAMIC") {
			decodedFields = append(decodedFields, element)
		} else {
			elementField := element

			atrb, err := helper.ConvertStructToMap(element.Attributes)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
			}

			// queryR := `
			// SELECT
			// 	r.id,
			// 	r.table_from,
			// 	r.table_to,
			// 	r.field_from,
			// 	r.field_to,
			// 	r.type,
			// 	r.relation_field_slug,
			// 	r.editable,
			// 	r.is_user_id_default,
			// 	r.object_id_from_jwt,
			// 	r.cascading_tree_table_slug,
			// 	r.cascading_tree_field_slug,
			// 	r.view_fields
			// FROM
			// 	relation r
			// WHERE  r.id = $1`

			// relation := models.RelationBody{}

			// err = conn.QueryRow(ctx, queryR, elementField.RelationId).Scan(
			// 	&relation.Id,
			// 	&relation.TableFrom,
			// 	&relation.TableTo,
			// 	&relation.FieldFrom,
			// 	&relation.FieldTo,
			// 	&relation.Type,
			// 	&relation.RelationFieldSlug,
			// 	&relation.Editable,
			// 	&relation.IsUserIdDefault,
			// 	&relation.ObjectIdFromJwt,
			// 	&relation.CascadingTreeTableSlug,
			// 	&relation.CascadingTreeFieldSlug,
			// 	&relation.ViewFields,
			// )
			// if err != nil {
			// 	return nil, err
			// }

			// elementField.RelationData = relation

			tempViewFields := cast.ToSlice(atrb["view_fields"])

			if len(tempViewFields) > 0 {
				// skip language settings

				body, err := json.Marshal(tempViewFields)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling view fields")
				}
				if err := json.Unmarshal(body, &elementField.ViewFields); err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling view fields")
				}
			}

			decodedFields = append(decodedFields, elementField)
		}
	}

	repsonse := map[string]interface{}{
		"fields":          decodedFields,
		"views":           views,
		"relation_fields": relationsFields,
	}

	newResp, err := helper.ConvertMapToStruct(repsonse)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		Data:      newResp,
	}, nil
}

func (o *objectBuilderRepo) GetAll(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		params = make(map[string]interface{})
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	var (
		languageSetting       = cast.ToString("language_setting")
		clientTypeIdFromToken = cast.ToString(params["client_type_id_from_token"])
		roleIdFromToken       = cast.ToString(params["role_id_from_token"])

		fields = []models.Field{}
	)
	delete(params, "limit")
	delete(params, "offset")
	delete(params, "language_setting")
	delete(params, "client_type_id_from_token")
	delete(params, "role_id_from_token")
	params["client_type_id"] = clientTypeIdFromToken

	// _, err = helper.GetRecordPermission(ctx, helper.GetRecordPermissionRequest{Conn: conn, TableSlug: req.TableSlug, RoleId: roleIdFromToken})
	// if err != nil && err != pgx.ErrNoRows {
	// 	return &nb.CommonMessage{}, err
	// }

	// for key := range params {
	// 	if (key == req.TableSlug+"_id" || key == req.TableSlug+"_ids") && params[key] != "" && !cast.ToBool(params["is_recursive"]) {
	// 		params["guid"] = params[key]
	// 	}
	// }

	query := `
		SELECT 
			f.id,
			f."table_id",
			f."required",
			f."slug",
			f."label",
			f."default",
			f."type",
			f."index",
			f."attributes",
			f."is_visible",
			f.autofill_field,
			f.autofill_table,
			f."unique",
			f."automatic",
			f.relation_id
		FROM "field" as f 
		JOIN "table" as t ON t."id" = f."table_id"
		WHERE t."slug" = $1
	`

	rows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field             = models.Field{}
			attributes        = []byte{}
			relationIdNull    sql.NullString
			autofillField     sql.NullString
			autofillTable     sql.NullString
			defaultStr, index sql.NullString
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.Required,
			&field.Slug,
			&field.Label,
			&defaultStr,
			&field.Type,
			&index,
			&attributes,
			&field.IsVisible,
			&autofillField,
			&autofillTable,
			&field.Unique,
			&field.Automatic,
			&relationIdNull,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		field.RelationId = relationIdNull.String
		field.AutofillField = autofillField.String
		field.AutofillTable = autofillTable.String
		field.Default = defaultStr.String
		field.Index = index.String

		if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, err
		}

		fields = append(fields, field)
	}

	fieldsWithPermissions, _, err := helper.AddPermissionToField1(ctx, helper.AddPermissionToFieldRequest{Conn: conn, RoleId: roleIdFromToken, TableSlug: req.TableSlug, Fields: fields})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	decodedFields := []models.Field{}
	for _, el := range fieldsWithPermissions {
		if el.Attributes != nil && !(el.Type == "LOOKUP" || el.Type == "LOOKUPS" || el.Type == "DYNAMIC") {
			decodedFields = append(decodedFields, el)
		} else {
			elementField := el

			attrb, err := helper.ConvertStructToMap(elementField.Attributes)
			if err != nil {
				return &nb.CommonMessage{}, err
			}

			tempViewFields := cast.ToSlice(attrb["view_fields"])
			viewFields := []models.Field{}
			if len(tempViewFields) > 0 {
				if languageSetting != "" {
					for _, el := range tempViewFields {
						if el != nil && el.(models.Field).Slug != "" && strings.HasSuffix(el.(models.Field).Slug, "_"+languageSetting) && el.(models.Field).EnableMultilanguage {
							viewFields = append(viewFields, el.(models.Field))
						} else if el != nil && !el.(models.Field).EnableMultilanguage {
							viewFields = append(viewFields, el.(models.Field))
						}
					}
				} else {
					for _, el := range tempViewFields {
						viewFields = append(viewFields, el.(models.Field))
					}
				}
			}
			elementField.ViewFields = viewFields
			decodedFields = append(decodedFields, elementField)
		}
	}

	fieldBytes, err := json.Marshal(decodedFields)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	// views, err := helper.GetViewWithPermission(ctx, &helper.GetViewWithPermissionReq{Conn: conn, TableSlug: req.TableSlug, RoleId: roleIdFromToken})
	// if err != nil {
	// 	return &nb.CommonMessage{}, err
	// }

	// viewBytes, err := json.Marshal(views)
	// if err != nil {
	// 	return &nb.CommonMessage{}, err
	// }

	fieldJsonBytes := fmt.Sprintf(`{"fields": %s}`, fieldBytes)
	var dataStruct structpb.Struct
	err = json.Unmarshal([]byte(fieldJsonBytes), &dataStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          &dataStruct,
		IsCached:      req.IsCached,
		CustomMessage: req.CustomMessage,
	}, nil
}

func (o *objectBuilderRepo) GetList2(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn := psqlpool.Get(req.GetProjectId())

	if req.TableSlug == "template" {
		response := map[string]interface{}{
			"count":    0,
			"response": []string{},
		}

		responseStruct, err := helper.ConvertMapToStruct(response)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
		}

		return &nb.CommonMessage{Data: responseStruct, TableSlug: req.TableSlug}, nil
	}

	var (
		params = make(map[string]interface{})
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	var (
	// limit  = cast.ToInt32(params["limit"])
	// offset = cast.ToInt32(params["offset"])
	// languageSetting       = cast.ToString("language_setting")
	// clientTypeIdFromToken = cast.ToString(params["client_type_id_from_token"])
	// roleIdFromToken       = cast.ToString(params["role_id_from_token"])
	)
	// delete(params, "limit")
	// delete(params, "offset")
	// delete(params, "language_setting")
	// delete(params, "client_type_id_from_token")
	// delete(params, "role_id_from_token")
	// params["client_type_id"] = clientTypeIdFromToken

	query := `SELECT f.type, f.slug, f.attributes FROM "field" f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer fieldRows.Close()

	fields := make(map[string]models.Field)
	fieldsArr := []models.Field{}

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		fields[fBody.Slug] = fBody
		fieldsArr = append(fieldsArr, fBody)
	}

	items, count, err := helper.GetItems(ctx, conn, models.GetItemsBody{
		TableSlug: req.TableSlug,
		Params:    params,
		FieldsMap: fields,
	})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting items")
	}

	for _, field := range fieldsArr {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
		}

		if field.Type == "FORMULA" {
			_, tFrom := attributes["table_from"]
			_, sF := attributes["sum_field"]

			if tFrom && sF {
				resp, err := helper.CalculateFormulaBackend(ctx, conn, attributes, req.TableSlug)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while calculating formula backend")
				}

				for _, i := range items {
					i[field.Slug] = resp[cast.ToString(i["guid"])]
				}
			}
		} else if field.Type == "FORMULA_FRONTEND" {
			_, ok := attributes["formula"]
			if ok {
				for _, i := range items {
					resultFormula, err := helper.CalculateFormulaFrontend(attributes, fieldsArr, i)
					if err != nil {
						return &nb.CommonMessage{}, errors.Wrap(err, "error while calculating formula frontend")
					}

					i[field.Slug] = resultFormula
				}
			}
		}
	}

	response := map[string]interface{}{
		"count":    count,
		"response": items,
	}

	itemsStruct, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          itemsStruct,
		IsCached:      req.IsCached,
		CustomMessage: req.CustomMessage,
	}, nil
}

func (o *objectBuilderRepo) GetListSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	if req.TableSlug == "template" {
		response := map[string]interface{}{
			"count":    0,
			"response": []string{},
		}

		responseStruct, err := helper.ConvertMapToStruct(response)
		if err != nil {
			return &nb.CommonMessage{}, nil
		}

		return &nb.CommonMessage{Data: responseStruct, TableSlug: req.TableSlug}, nil
	}

	var (
		params = make(map[string]interface{})
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	query := `SELECT f.type, f.slug, f.attributes FROM "field" f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	fields := make(map[string]models.Field)
	fieldsArr := []models.Field{}

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.CommonMessage{}, err
		}

		fields[fBody.Slug] = fBody
		fieldsArr = append(fieldsArr, fBody)
	}

	items, count, err := helper.GetItems(ctx, conn, models.GetItemsBody{
		TableSlug: req.TableSlug,
		Params:    params,
		FieldsMap: fields,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	for _, field := range fieldsArr {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if field.Type == "FORMULA" {
			_, tFrom := attributes["table_from"]
			_, sF := attributes["sum_field"]

			if tFrom && sF {
				resp, err := helper.CalculateFormulaBackend(ctx, conn, attributes, req.TableSlug)
				if err != nil {
					return &nb.CommonMessage{}, err
				}

				for _, i := range items {
					i[field.Slug] = resp[cast.ToString(i["guid"])]
				}
			}
		} else if field.Type == "FORMULA_FRONTEND" {
			_, ok := attributes["formula"]
			if ok {
				for _, i := range items {
					resultFormula, err := helper.CalculateFormulaFrontend(attributes, fieldsArr, i)
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					i[field.Slug] = resultFormula
				}
			}
		}
	}

	response := map[string]interface{}{
		"count":    count,
		"response": items,
	}

	itemsStruct, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          itemsStruct,
		IsCached:      req.IsCached,
		CustomMessage: req.CustomMessage,
	}, nil
}

func (o *objectBuilderRepo) GetListInExcel(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	conn := psqlpool.Get(req.GetProjectId())

	var (
		params = make(map[string]interface{})
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	fieldIds := cast.ToStringSlice(params["field_ids"])

	// fieldIds := []string{"fe5a4241-9767-4b2e-9bca-6ff988ae87f2", "28ec1137-3d87-4827-8117-04b013b5e347"}

	delete(params, "field_ids")

	query := `SELECT f.type, f.slug, f.attributes, f.label FROM "field" f WHERE f.id = ANY ($1)`

	fieldRows, err := conn.Query(ctx, query, pq.Array(fieldIds))
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	fields := make(map[string]models.Field)
	fieldsArr := []models.Field{}

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
			&fBody.Label,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.CommonMessage{}, err
		}

		fields[fBody.Slug] = fBody
		fieldsArr = append(fieldsArr, fBody)
	}

	items, _, err := helper.GetItems(ctx, conn, models.GetItemsBody{
		TableSlug: req.TableSlug,
		Params:    params,
		FieldsMap: fields,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	file := excel.NewFile()

	for i, field := range fieldsArr {
		err := file.SetCellValue(sh, letters[i]+"1", field.Label)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
	}

	for i, item := range items {
		letterCount := 0
		column := fmt.Sprint(i + 2)

		for _, f := range fieldsArr {
			_, ok := fields[f.Slug]
			if ok {

				if f.Type == "MULTI_LINE" {

					re := regexp.MustCompile(`<[^>]+>`)

					result := re.ReplaceAllString(cast.ToString(item[f.Slug]), "")

					item[f.Slug] = result
				} else if f.Type == "DATE" {

					fmt.Println(cast.ToString(item[f.Slug]))

					timeF, err := time.Parse("2006-01-02", strings.Split(cast.ToString(item[f.Slug]), " ")[0])
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					item[f.Slug] = timeF.Format("02.01.2006")
				} else if f.Type == "DATE_TIME" {

					fmt.Println(cast.ToString(item[f.Slug]))

					newTime := strings.Split(cast.ToString(item[f.Slug]), " ")[0] + " " + strings.Split(cast.ToString(item[f.Slug]), " ")[1]

					timeF, err := time.Parse("2006-01-02 15:04:05", newTime)
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					item[f.Slug] = timeF.Format("02.01.2006 15:04")
				} else if f.Type == "MULTISELECT" {
					attributes, err := helper.ConvertStructToMap(f.Attributes)
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					multiselectValue := ""

					_, ok := attributes["options"]
					if ok {
						options := cast.ToSlice(attributes["options"])
						values := cast.ToStringSlice(item[f.Slug])

						if len(options) > 0 && len(values) > 0 {
							for _, val := range values {
								for _, op := range options {
									opt := cast.ToStringMap(op)
									if val == cast.ToString(opt["value"]) {
										if cast.ToString(opt["label"]) != "" {
											multiselectValue += cast.ToString(opt["label"]) + ","
										} else {
											multiselectValue += cast.ToString(opt["value"]) + ","
										}
									}
								}
							}
						}
					}

					item[f.Slug] = strings.TrimRight(multiselectValue, ",")
				}

				err = file.SetCellValue(sh, letters[letterCount]+column, item[f.Slug])
				if err != nil {
					fmt.Println(err)
					return &nb.CommonMessage{}, err
				}
				letterCount++
			}
		}

		// for key, value := range item {
		// 	_, ok := fields[key]
		// 	if ok {
		// 		err = file.SetCellValue(sh, letters[letterCount]+column, value)
		// 		if err != nil {
		// 			fmt.Println(err)
		// 			return &nb.CommonMessage{}, err
		// 		}
		// 		letterCount++
		// 	}
		// }
	}

	filename := fmt.Sprintf("report_%d.xlsx", time.Now().Unix())
	filepath := "./" + filename

	err = file.SaveAs(filename)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	cfg := config.Load()

	endpoint := "cdn-api.ucode.run"
	accessKeyID := cfg.MinioAccessKeyID
	secretAccessKey := cfg.MinioSecretKey

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, err = minioClient.FPutObject(
		context.Background(),
		"reports",
		filename,
		filepath,
		minio.PutObjectOptions{ContentType: ""},
	)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	err = os.Remove(filepath)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	link := fmt.Sprintf("%s/reports/%s", endpoint, filename)
	respExcel := map[string]string{
		"link": link,
	}

	marshledInputMap, err := json.Marshal(respExcel)
	outputStruct := &structpb.Struct{}
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	err = protojson.Unmarshal(marshledInputMap, outputStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{TableSlug: req.TableSlug, Data: outputStruct}, nil
}

var letters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
var sh = "Sheet1"

func (o *objectBuilderRepo) TestApi(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	query := `
		SELECT
			guid,
			name,
			birth_date,
			net_worth,
			weight,
			age,
			married
		FROM bingo
		OFFSET 0 LIMIT 20
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	type BingoData struct {
		Guid      string
		Name      string
		BirthDate time.Time
		NetWorth  float64
		Weight    float64
		Age       float64
		Married   bool
	}

	var result []BingoData
	for rows.Next() {
		var data BingoData
		err := rows.Scan(
			&data.Guid,
			&data.Name,
			&data.BirthDate,
			&data.NetWorth,
			&data.Weight,
			&data.Age,
			&data.Married,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, data)
	}

	response := map[string]interface{}{
		"count":    0,
		"response": result,
	}

	itemsStruct, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          itemsStruct,
		IsCached:      req.IsCached,
		CustomMessage: req.CustomMessage,
	}, nil
}
