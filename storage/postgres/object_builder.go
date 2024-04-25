package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
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

	conn := o.db

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
}

func (o *objectBuilderRepo) GetListConnection(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.GetProjectId())

	conn := o.db

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
	// conn := psqlpool.Get(req.GetProjectId())
	// defer conn.Close()

	conn := o.db

	var (
		fields          = []models.Field{}
		relations       = []models.Relation{}
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

		fields = append(fields, field)
	}

	query = `SELECT 
		"id",
		"table_from",
		"table_to",
		"type"
	FROM "relation" r,
	jsonb_array_elements(r."dynamic_tables") AS dt
	WHERE "table_from" = $1 OR "table_to" = $1 OR dt->>'table_slug' = $1;
	`

	relationRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relations by table slug")
	}
	defer relationRows.Close()

	for relationRows.Next() {
		relation := models.Relation{}

		err = relationRows.Scan(
			&relation.Id,
			&relation.TableFrom,
			&relation.TableTo,
			&relation.Type,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relations")
		}

		relations = append(relations, relation)
	}

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

	if cast.ToBool(params["with_relations"]) {

		var (
			relationTableToSlugs = []string{}
			relationTableIds     = []string{}
			relationTablesMap    = make(map[string]models.Table)
		)

		for _, relation := range relations {
			if relation.Type != "Many2Dynamic" {
				if relation.Type == "Many2Many" && relation.TableTo == req.TableSlug {
					relation.TableTo = relation.TableFrom
				}

				relationTableToSlugs = append(relationTableToSlugs, relation.TableTo)
			}
		}

		if len(relationTableToSlugs) > 0 {

			query = `SELECT 
				"id",
				"slug",
				"label"
			FROM "table" WHERE "slug" IN ($1)`

			rows, err := conn.Query(ctx, query, pq.Array(relationTableToSlugs))
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting tables by slugs")
			}
			defer rows.Close()

			for rows.Next() {
				table := models.Table{}

				err = rows.Scan(
					&table.Id,
					&table.Slug,
					&table.Label,
				)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning tables")
				}

				relationTableIds = append(relationTableIds, table.Id)
				_, ok := relationTablesMap[table.Slug]
				if !ok {
					relationTablesMap[table.Slug] = table
				}
			}
		}

		var (
			relationFieldSlugsR = []string{}
			relationFieldsMap   = make(map[string][]models.Field)
		)
		if len(relationTableIds) > 0 {
			query = `SELECT
				id,
				type,
				slug,
				table_id
			FROM "field" WHERE table_id IN ($1)`

			rows, err := conn.Query(ctx, query, pq.Array(relationTableIds))
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table ids")
			}
			defer rows.Close()

			for rows.Next() {
				field := models.Field{}

				err = rows.Scan(
					&field.Id,
					&field.Type,
					&field.Slug,
					&field.TableId,
				)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
				}

				if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
					tableSlug := ""
					if field.Type == "LOOKUP" {
						tableSlug = field.Slug[:len(field.Slug)-3]
					} else {
						tableSlug = field.Slug[:len(field.Slug)-4]
					}

					relationFieldSlugsR = append(relationFieldSlugsR, tableSlug)
				}

				_, ok := relationFieldsMap[field.TableId]
				if ok {
					relationFieldsMap[field.TableId] = append(relationFieldsMap[field.TableId], field)
				} else {
					relationFieldsMap[field.TableId] = []models.Field{field}
				}
			}
		}

		var (
			childRelationsMap      = make(map[string]models.Relation)
			viewFieldIds           = []string{}
			viewFieldsMap          = make(map[string]models.Field)
			childRelationTablesMap = make(map[string]string)
		)

		if len(relationTableToSlugs) > 0 && len(relationFieldSlugsR) > 0 {
			query = `SELECT 
				"id",
				"table_from",
				"table_to",
				"type",
				view_fields
			FROM "relation" WHERE "table_from" IN ($1) AND "table_to" IN ($2)`

			rows, err := conn.Query(ctx, query, pq.Array(relationTableToSlugs), pq.Array(relationFieldSlugsR))
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relations by table slugs and field slugs")
			}
			defer rows.Close()

			for rows.Next() {
				relation := models.Relation{}

				err = rows.Scan(
					&relation.Id,
					&relation.TableFrom,
					&relation.TableTo,
					&relation.Type,
					pq.Array(&relation.ViewFields),
				)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relations")
				}

				_, ok := childRelationsMap[relation.TableFrom+"_"+relation.TableTo]
				if !ok {
					childRelationsMap[relation.TableFrom+"_"+relation.TableTo] = relation
				}

				viewFieldIds = relation.ViewFields
			}
		}

		if len(viewFieldIds) > 0 {
			query = `SELECT 
				id,
				"table_id",
				"required",
				"slug",
				"label",
				"default",
				"type",
				"index",
				"attributes",
				"is_visible",
				autofill_field,
				autofill_table,
				"unique",
				"automatic",
				relation_id
			FROM "field" WHERE id IN ($1)`

			rows, err := conn.Query(ctx, query, pq.Array(viewFieldIds))
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by ids")
			}
			defer rows.Close()

			for rows.Next() {
				var (
					field          = models.Field{}
					attributes     = []byte{}
					relationIdNull sql.NullString
				)
				err = rows.Scan(
					&field.Id,
					&field.TableId,
					&field.Required,
					&field.Slug,
					&field.Label,
					&field.Default,
					&field.Type,
					&field.Index,
					&attributes,
					&field.IsVisible,
					&field.AutofillField,
					&field.AutofillTable,
					&field.Unique,
					&field.Automatic,
					&relationIdNull,
				)
				if err != nil {
					return &nb.CommonMessage{}, err
				}

				field.RelationId = relationIdNull.String

				if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
				}

				viewFieldsMap[field.Id] = field
			}
		}

		if len(relationFieldSlugsR) > 0 {
			query = `SELECT 
				"slug",
				"label"
			FROM "table" WHERE slug IN ($1)`

			rows, err := conn.Query(ctx, query, pq.Array(relationFieldSlugsR))
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting tables by slugs")
			}
			defer rows.Close()

			for rows.Next() {
				var (
					tableSlug  string
					tableLabel string
				)

				err = rows.Scan(
					&tableSlug,
					&tableLabel,
				)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning tables")
				}

				_, ok := childRelationTablesMap[tableSlug]
				if !ok {
					childRelationTablesMap[tableSlug] = tableLabel
				}
			}
		}

		for _, relation := range relations {
			if relation.Type != "Many2Dynamic" {
				if relation.Type == "Many2Many" && relation.TableTo == req.TableSlug {
					relation.TableTo = relation.TableFrom
				}
			}

			relationTable := relationTablesMap[relation.TableTo]
			tableRelationFields := relationFieldsMap[relationTable.Id]

			for _, field := range tableRelationFields {
				changedField := models.Field{}
				if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {
					var (
						viewFields = []models.Field{}
					)
					tableSlug := ""
					if field.Type == "LOOKUP" {
						tableSlug = field.Slug[:len(field.Slug)-3]
					} else {
						tableSlug = field.Slug[:len(field.Slug)-4]
					}

					childRelation, ok := childRelationsMap[relationTable.Slug+"_"+tableSlug]
					if ok {
						for _, view_field := range childRelation.ViewFields {
							viewField, ok := viewFieldsMap[view_field]
							if ok {
								viewFields = append(viewFields, viewField)
							}
						}
					}

					field.ViewFields = viewFields
					field.Label = childRelationTablesMap[tableSlug]
					changedField = field
					changedField.PathSlug = relationTable.Slug + "_id_data." + field.Slug
					changedField.TableSlug = tableSlug
					relationsFields = append(relationsFields, changedField)
				} else {
					changedField = field
					changedField.PathSlug = relationTable.Slug + "_id_data." + field.Slug
					relationsFields = append(relationsFields, changedField)
				}
			}
		}
	}

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
	conn := o.db

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
		WHERE f.table_id IN (
			SELECT id FROM "table" WHERE slug = $1
		)
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
	conn := o.db

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
		// limit  = cast.ToInt32(params["limit"])
		// offset = cast.ToInt32(params["offset"])
		// languageSetting       = cast.ToString("language_setting")
		clientTypeIdFromToken = cast.ToString(params["client_type_id_from_token"])
		// roleIdFromToken       = cast.ToString(params["role_id_from_token"])
	)
	delete(params, "limit")
	delete(params, "offset")
	delete(params, "language_setting")
	delete(params, "client_type_id_from_token")
	delete(params, "role_id_from_token")
	params["client_type_id"] = clientTypeIdFromToken

	items, err := helper.GetItems(ctx, conn, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	itemsStruct, err := helper.ConvertMapToStruct(items)
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

func AddPermissionToField(ctx context.Context, conn *pgxpool.Pool, fields []models.Field, roleId string, tableSlug string) ([]models.Field, error) {

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
				newAtb, err := helper.ConvertMapToStruct(decoded)
				if err != nil {
					return []models.Field{}, err
				}
				field.Attributes = newAtb
			} else {
				atributes := map[string]interface{}{
					"field_permission": fieldPer,
				}

				newAtb, err := helper.ConvertMapToStruct(atributes)
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
