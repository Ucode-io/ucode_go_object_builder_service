package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
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
			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting client types")
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
				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning client types")
			}

			data = append(data, clientType)
		}

		jsonBytes, err := json.Marshal(data)
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
			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting roles")
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
				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning roles")
			}

			data = append(data, role)
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling roles")
		}

		var dataStruct structpb.Struct
		jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

		err = json.Unmarshal(jsonBytes, &dataStruct)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling roles")
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
		fields, relationsFields, decodedFields []models.Field
		views                                  = []models.View{}
		params                                 = make(map[string]interface{})
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
		f.relation_id,
		f."is_search",
		f.enable_multilanguage
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
			field                            = models.Field{}
			attributes                       = []byte{}
			relationIdNull, autofillField    sql.NullString
			defaultStr, index, autofillTable sql.NullString
			atr                              = make(map[string]interface{})
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
			&field.IsSearch,
			&field.EnableMultilanguage,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		field.RelationId = relationIdNull.String
		field.AutofillField = autofillField.String
		field.AutofillTable = autofillTable.String
		field.Default = defaultStr.String
		field.Index = index.String
		newAtrb := make(map[string]interface{})

		if err := json.Unmarshal(attributes, &atr); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		if field.Type == "LOOKUP" || field.Type == "LOOKUPS" {

			view, err := helper.ViewFindOne(ctx, helper.RelationHelper{
				Conn:       conn,
				RelationID: field.RelationId,
			})
			if err != nil {
				return resp, errors.Wrap(err, "error while getting view by relation id")
			}

			newAtrb, err = helper.ConvertStructToMap(view.Attributes)
			if err != nil {
				return resp, errors.Wrap(err, "error while converting struct to map")
			}

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
				return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relation by id")
			}
			defer relationRows.Close()

			for relationRows.Next() {
				var (
					viewFields         []string
					tableFrom, tableTo string
				)
				err = relationRows.Scan(&viewFields, &tableFrom, &tableTo)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relation rows")
				}

				atr["relation_data"] = map[string]interface{}{
					"view_fields": viewFields,
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
							f.relation_id,
							f."is_search",
							f.enable_multilanguage
						FROM "field" as f 
						WHERE f.id = $1
					`

					err = conn.QueryRow(ctx, query, id).Scan(
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
						&field.IsSearch,
						&field.EnableMultilanguage,
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

					fieldObjects = append(fieldObjects, field)
				}

				field.ViewFields = fieldObjects
			}
		}

		for key, val := range newAtrb {
			atr[key] = val
		}

		attributes, _ = json.Marshal(atr)

		if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		fields = append(fields, field)
	}

	query = `SELECT 
		"id",
		"attributes",
		"table_slug",
		"type",
		"columns",
		"order",
		COALESCE("time_interval", 0),
		COALESCE("group_fields"::varchar[], '{}'),
		"name",
		"main_field",
		"quick_filters",
		"users",
		"view_fields",
		"calendar_from_slug",
		"calendar_to_slug",
		"multiple_insert",
		"status_field_slug",
		"is_editable",
		"relation_table_slug",
		"relation_id",
		"multiple_insert_field",
		"updated_fields",
		"table_label",
		"default_limit",
		"default_editable",
		"name_uz",
		"name_en"
	FROM "view" WHERE "table_slug" = $1 ORDER BY "order" ASC`

	viewRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting views by table slug")
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var (
			attributes                               []byte
			view                                     = models.View{}
			Name, MainField, CalendarFromSlug        sql.NullString
			RelationTableSlug, StatusFieldSlug       sql.NullString
			MultipleInsertField, RelationId          sql.NullString
			TableLabel, DefaultLimit, CalendarToSlug sql.NullString
			NameUz, NameEn, QuickFilters             sql.NullString
		)

		err := viewRows.Scan(
			&view.Id,
			&attributes,
			&view.TableSlug,
			&view.Type,
			&view.Columns,
			&view.Order,
			&view.TimeInterval,
			&view.GroupFields,
			&Name,
			&MainField,
			&QuickFilters,
			&view.Users,
			&view.ViewFields,
			&CalendarFromSlug,
			&CalendarToSlug,
			&view.MultipleInsert,
			&StatusFieldSlug,
			&view.IsEditable,
			&RelationTableSlug,
			&RelationId,
			&MultipleInsertField,
			&view.UpdatedFields,
			&TableLabel,
			&DefaultLimit,
			&view.DefaultEditable,
			&NameUz,
			&NameEn,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning views")
		}

		view.Name = Name.String
		view.MainField = MainField.String
		view.CalendarFromSlug = CalendarFromSlug.String
		view.CalendarToSlug = CalendarToSlug.String
		view.StatusFieldSlug = StatusFieldSlug.String
		view.RelationTableSlug = RelationTableSlug.String
		view.RelationId = RelationId.String
		view.MultipleInsertField = MultipleInsertField.String
		view.TableLabel = TableLabel.String
		view.DefaultLimit = DefaultLimit.String
		view.NameUz = NameUz.String
		view.NameEn = NameEn.String

		if QuickFilters.Valid {
			err = json.Unmarshal([]byte(QuickFilters.String), &view.QuickFilters)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling quick filters")
			}
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

	fieldsWithPermissions, _, err := helper.AddPermissionToField1(ctx, helper.AddPermissionToFieldRequest{Conn: conn, Fields: fields, RoleId: cast.ToString(params["role_id_from_token"]), TableSlug: req.TableSlug})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while adding permissions to fields")
	}

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
				body, err := json.Marshal(tempViewFields)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling view fields")
				}
				if err := json.Unmarshal(body, &elementField.ViewFields); err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling view fields")
				}
			}

			atrb, err = helper.ConvertStructToMap(elementField.Attributes)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "convert struct to map")
			}

			atrb["enable_multilanguage"] = elementField.EnableMultilanguage
			atrb["enable_multi_language"] = elementField.EnableMultilanguage

			strc, err := helper.ConvertMapToStruct(atrb)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "convert map to struct")
			}

			elementField.Attributes = strc

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
		params                = make(map[string]interface{})
		views                 = []models.View{}
		fieldsMap             = make(map[string]models.Field)
		fields, decodedFields []models.Field
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	roleIdFromToken := cast.ToString(params["role_id_from_token"])

	query := `
		SELECT 
			f.id,
			f."table_id",
			t.slug,
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
		LEFT JOIN "relation" r ON r.id = f.relation_id
		WHERE t."slug" = $1
	`

	rows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field                            = models.Field{}
			attributes                       = []byte{}
			relationIdNull, autofillField    sql.NullString
			defaultStr, index, autofillTable sql.NullString
			atrb                             = make(map[string]interface{})
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.TableSlug,
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

		if err := json.Unmarshal(attributes, &atrb); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		attributes, _ = json.Marshal(atrb)

		if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		fields = append(fields, field)
		fieldsMap[field.Slug] = field
	}

	fieldsWithPermissions, _, err := helper.AddPermissionToField1(ctx, helper.AddPermissionToFieldRequest{Conn: conn, RoleId: roleIdFromToken, TableSlug: req.TableSlug, Fields: fields})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while adding permissions to fields")
	}

	rquery := `SELECT 
			f.id,
			f."table_id",
			t.slug,
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
	
	FROM field f 
	JOIN "table" t ON t.id = f.table_id
	JOIN relation r ON r.id = $1 WHERE f.id::text = ANY(r.view_fields)`

	reqlationQ := `
	SELECT
		r.id,
		r.table_from,
		r.table_to,
		r.field_from,
		r.field_to,
		r.type,
		r.relation_field_slug,
		r.editable,
		r.is_user_id_default,
		r.is_system,
		r.object_id_from_jwt,
		r.cascading_tree_table_slug,
		r.cascading_tree_field_slug,
		r.view_fields
	FROM
		relation r
	WHERE  r.id = $1`

	for _, el := range fieldsWithPermissions {
		if el.Attributes != nil && !(el.Type == "LOOKUP" || el.Type == "LOOKUPS" || el.Type == "DYNAMIC") {
			decodedFields = append(decodedFields, el)
		} else {
			elementField := el
			viewFields := []models.Field{}

			if el.RelationId != "" {
				relation := models.RelationBody{}

				err = conn.QueryRow(ctx, reqlationQ, el.RelationId).Scan(
					&relation.Id,
					&relation.TableFrom,
					&relation.TableTo,
					&relation.FieldFrom,
					&relation.FieldTo,
					&relation.Type,
					&relation.RelationFieldSlug,
					&relation.Editable,
					&relation.IsUserIdDefault,
					&relation.IsSystem,
					&relation.ObjectIdFromJwt,
					&relation.CascadingTreeTableSlug,
					&relation.CascadingTreeFieldSlug,
					&relation.ViewFields,
				)

				if err != nil {
					if !strings.Contains(err.Error(), "no rows") {
						return nil, errors.Wrap(err, "error while scanning relation")
					}
				} else {
					elementField.RelationData = relation

					if relation.TableFrom != req.TableSlug {
						elementField.TableSlug = relation.TableFrom
					} else {
						elementField.TableSlug = relation.TableTo
					}

					frows, err := conn.Query(ctx, rquery, el.RelationId)
					if err != nil {
						return &nb.CommonMessage{}, errors.Wrap(err, "error while getting relation fields")
					}
					defer frows.Close()

					for frows.Next() {
						var (
							vf                               = models.Field{}
							attributes                       = []byte{}
							relationIdNull, autofillField    sql.NullString
							defaultStr, index, autofillTable sql.NullString
						)

						err = frows.Scan(
							&vf.Id,
							&vf.TableId,
							&vf.TableSlug,
							&vf.Required,
							&vf.Slug,
							&vf.Label,
							&defaultStr,
							&vf.Type,
							&index,
							&attributes,
							&vf.IsVisible,
							&autofillField,
							&autofillTable,
							&vf.Unique,
							&vf.Automatic,
							&relationIdNull,
						)
						if err != nil {
							return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning relation fields")
						}

						vf.RelationId = relationIdNull.String
						vf.AutofillField = autofillField.String
						vf.AutofillTable = autofillTable.String
						vf.Default = defaultStr.String
						vf.Index = index.String

						if err := json.Unmarshal(attributes, &vf.Attributes); err != nil {
							return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling relation field attributes")
						}

						viewFields = append(viewFields, vf)
					}
				}

			}

			elementField.ViewFields = viewFields
			decodedFields = append(decodedFields, elementField)
		}
	}

	query = `SELECT 
		"id",
		"attributes",
		"table_slug",
		"type",
		"columns",
		"order",
		COALESCE("time_interval", 0),
		COALESCE("group_fields"::varchar[], '{}'),
		"name",
		"main_field",
		"quick_filters",
		"users",
		"view_fields",
		"calendar_from_slug",
		"calendar_to_slug",
		"multiple_insert",
		"status_field_slug",
		"is_editable",
		"relation_table_slug",
		"relation_id",
		"multiple_insert_field",
		"updated_fields",
		"table_label",
		"default_limit",
		"default_editable",
		"name_uz",
		"name_en"
	FROM "view" WHERE "table_slug" = $1 ORDER BY "order" ASC`

	viewRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting views by table slug")
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var (
			attributes                                        []byte
			view                                              = models.View{}
			Name, MainField, CalendarFromSlug, CalendarToSlug sql.NullString
			StatusFieldSlug, RelationTableSlug, RelationId    sql.NullString
			MultipleInsertField, TableLabel, DefaultLimit     sql.NullString
			NameUz, NameEn, QuickFilters                      sql.NullString
		)

		err := viewRows.Scan(
			&view.Id,
			&attributes,
			&view.TableSlug,
			&view.Type,
			&view.Columns,
			&view.Order,
			&view.TimeInterval,
			&view.GroupFields,
			&Name,
			&MainField,
			&QuickFilters,
			&view.Users,
			&view.ViewFields,
			&CalendarFromSlug,
			&CalendarToSlug,
			&view.MultipleInsert,
			&StatusFieldSlug,
			&view.IsEditable,
			&RelationTableSlug,
			&RelationId,
			&MultipleInsertField,
			&view.UpdatedFields,
			&TableLabel,
			&DefaultLimit,
			&view.DefaultEditable,
			&NameUz,
			&NameEn,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning views")
		}

		view.Name = Name.String
		view.MainField = MainField.String
		view.CalendarFromSlug = CalendarFromSlug.String
		view.CalendarToSlug = CalendarToSlug.String
		view.StatusFieldSlug = StatusFieldSlug.String
		view.RelationTableSlug = RelationTableSlug.String
		view.RelationId = RelationId.String
		view.MultipleInsertField = MultipleInsertField.String
		view.TableLabel = TableLabel.String
		view.DefaultLimit = DefaultLimit.String
		view.NameUz = NameUz.String
		view.NameEn = NameEn.String

		if QuickFilters.Valid {
			err = json.Unmarshal([]byte(QuickFilters.String), &view.QuickFilters)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling quick filters")
			}
		}

		if view.Columns == nil {
			view.Columns = []string{}
		}

		if err := json.Unmarshal(attributes, &view.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling view attributes")
		}

		views = append(views, view)
	}

	items, count, err := helper.GetItems(ctx, conn, models.GetItemsBody{
		TableSlug: req.TableSlug,
		Params:    params,
		FieldsMap: fieldsMap,
	})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting items")
	}

	repsonse := map[string]interface{}{
		"fields":   decodedFields,
		"views":    views,
		"count":    count,
		"response": items,
	}

	newResp, err := helper.ConvertMapToStruct(repsonse)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          newResp,
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
		params       = make(map[string]interface{})
		searchFields = []string{}
		fields       = make(map[string]models.Field)
		fieldsArr    []models.Field
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	query := `
		SELECT 
			f.type, 
			f.slug, 
			f.attributes,
			f.is_search
		FROM "field" f 
		JOIN "table" t ON t.id = f.table_id 
		WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
			&fBody.IsSearch,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		if fBody.IsSearch && helper.FIELD_TYPES[fBody.Type] == "VARCHAR" {
			searchFields = append(searchFields, fBody.Slug)
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling field attributes")
		}

		fields[fBody.Slug] = fBody
		fieldsArr = append(fieldsArr, fBody)
	}

	items, count, err := helper.GetItemsGetList(ctx, conn, models.GetItemsBody{
		TableSlug:    req.TableSlug,
		Params:       params,
		FieldsMap:    fields,
		SearchFields: searchFields,
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

	var (
		params    = make(map[string]interface{})
		fields    = make(map[string]models.Field)
		fieldsArr = []models.Field{}
	)

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

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while marshalling request data")
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling request data")
	}

	query := `SELECT f.type, f.slug, f.attributes FROM "field" f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer fieldRows.Close()

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

func (o *objectBuilderRepo) GetListInExcel(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		params    = make(map[string]interface{})
		fields    = make(map[string]models.Field)
		fieldsArr = []models.Field{}
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	fieldIds := cast.ToStringSlice(params["field_ids"])

	delete(params, "field_ids")

	query := `SELECT f.type, f.slug, f.attributes, f.label FROM "field" f WHERE f.id = ANY ($1)`

	fieldRows, err := conn.Query(ctx, query, pq.Array(fieldIds))
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

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
					timeF, err := time.Parse("2006-01-02", strings.Split(cast.ToString(item[f.Slug]), " ")[0])
					if err != nil {
						return &nb.CommonMessage{}, err
					}

					item[f.Slug] = timeF.Format("02.01.2006")
				} else if f.Type == "DATE_TIME" {
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
					return &nb.CommonMessage{}, err
				}
				letterCount++
			}
		}
	}

	var (
		filename        = fmt.Sprintf("report_%d.xlsx", time.Now().Unix())
		filepath        = "./" + filename
		cfg             = config.Load()
		endpoint        = cfg.MinioHost
		accessKeyID     = cfg.MinioAccessKeyID
		secretAccessKey = cfg.MinioSecretKey
	)

	err = file.SaveAs(filename)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, err = minioClient.FPutObject(
		ctx,
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
	defer rows.Close()

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

func (o *objectBuilderRepo) UpdateWithQuery(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	var (
		whereQuery = req.Data.AsMap()["postgres_query"] // this is how developer send request to object builder: "postgres_query"
		setClauses []string
		args       []interface{}
		i          = 1
	)

	for col, val := range req.Data.AsMap() {
		if col == "postgres_query" || col == "guid" {
			continue
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	var (
		setSQL = strings.Join(setClauses, ", ")
		query  = fmt.Sprintf("UPDATE %s SET %s WHERE %s", req.TableSlug, setSQL, whereQuery)
	)

	_, err := o.db.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{}, nil
}

func (o *objectBuilderRepo) GroupByColumns(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		viewAttributes    = make(map[string]interface{})
		atrb              = []byte{}
		fieldMap, grField = make(map[string]string), make(map[string]string)
		fields, grFields  []string
		tableSlug         = req.TableSlug
		query             string
		newResp           = []map[string]interface{}{}
	)

	queryV := `SELECT attributes, group_fields FROM view WHERE id = $1`

	reqData, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "convert req data")
	}

	err = conn.QueryRow(ctx, queryV, reqData["builder_service_view_id"]).Scan(&atrb, &grFields)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get view attributes")
	}

	if err := json.Unmarshal(atrb, &viewAttributes); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "unmarshal view attributes")
	}

	groupFields := cast.ToStringSlice(viewAttributes["group_by_columns"])

	if len(groupFields) == 0 && len(grFields) > 0 {
		groupFields = grFields
	}

	queryF := `SELECT f.id, f.type, f.slug, COALESCE(f.relation_id::varchar, '') FROM field f JOIN "table" t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, queryF, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			id, ftype, slug, relationId string
		)

		err = fieldRows.Scan(&id, &ftype, &slug, &relationId)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "scan field")
		}

		if relationId != "" {
			id = relationId
		}

		fieldMap[slug] = ftype
		fields = append(fields, slug)
		grField[id] = slug
	}

	for i, gr := range groupFields {
		groupFields[i] = grField[gr]
	}

	reversedField := groupFields

	for i, j := 0, len(reversedField)-1; i < j; i, j = i+1, j-1 {
		reversedField[i], reversedField[j] = reversedField[j], reversedField[i]
	}

	innerQuery := `SELECT `
	innerQuery += strings.Join(groupFields, ",")
	innerQuery += `, jsonb_agg(jsonb_build_object( `

	for _, f := range fields {
		innerQuery += fmt.Sprintf(` '%s', %s,`, f, f)
	}

	innerQuery = strings.TrimRight(innerQuery, ",")
	innerQuery += fmt.Sprintf(`)) as data FROM "%s" GROUP BY %s`, tableSlug, strings.Join(groupFields, ","))
	lastSlug := reversedField[0]

	for i := range reversedField {
		if i == 0 {
			continue
		}

		query += `SELECT `
		gr := ""

		for j := i; j < len(reversedField); j++ {
			gr += fmt.Sprintf(`%s,`, reversedField[j])
			query += fmt.Sprintf(`%s,`, reversedField[j])
		}

		gr = strings.TrimRight(gr, ",")

		query += fmt.Sprintf(`jsonb_agg(jsonb_build_object( '%s', %s, 'data', data)) as data FROM ( %s ) subquery GROUP BY %s`, lastSlug, lastSlug, innerQuery, gr)
		lastSlug = reversedField[i]
		innerQuery = query
		if i != len(reversedField)-1 {
			query = ""
		}
	}

	if query == "" {
		query = innerQuery
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer rows.Close()

	for rows.Next() {
		data := make(map[string]interface{})
		values, err := rows.Values()
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		for i, value := range values {

			if strings.Contains(string(rows.FieldDescriptions()[i].Name), "_id") || string(rows.FieldDescriptions()[i].Name) == "guid" {
				if arr, ok := value.([16]uint8); ok {
					value = helper.ConvertGuid(arr)
				}
			}

			data[string(rows.FieldDescriptions()[i].Name)] = value
		}

		newResp = append(newResp, data)
	}

	data := make([]interface{}, len(newResp))
	for i, d := range newResp {
		data[i] = d
	}

	addGroupByType(conn, data, fieldMap, map[string]map[string]interface{}{})

	newData := map[string]interface{}{
		"response": newResp,
	}

	res, err := helper.ConvertMapToStruct(newData)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		Data:      res,
		TableSlug: req.TableSlug,
	}, nil
}

func addGroupByType(conn *pgxpool.Pool, data interface{}, typeMap map[string]string, cache map[string]map[string]interface{}) {
	switch v := data.(type) {
	case []interface{}:
		for _, item := range v {
			addGroupByType(conn, item, typeMap, cache)
		}
	case map[string]interface{}:
		for key, value := range v {
			if strings.Contains(key, "_id") {

				body, ok := cache[cast.ToString(value)]
				if !ok {
					body, err := helper.GetItem(context.Background(), conn, strings.ReplaceAll(key, "_id", ""), cast.ToString(value))
					if err != nil {
						return
					}

					v[key+"_data"] = body
					cache[cast.ToString(value)] = body
				} else {
					v[key+"_data"] = body
				}
			}
			_, last := v["data"]
			if last {
				if typeVal, exists := typeMap[key]; exists {
					if strings.Contains(key, "_id") {
						v["group_by_slug"] = strings.ReplaceAll(key, "_id", "")

						body, ok := cache[cast.ToString(value)]
						if !ok {
							body, err := helper.GetItem(context.Background(), conn, strings.ReplaceAll(key, "_id", ""), cast.ToString(value))
							if err != nil {
								return
							}

							v[key+"_data"] = body
							cache[cast.ToString(value)] = body
						} else {
							v[key+"_data"] = body
						}
					}

					v["group_by_type"] = typeVal
					v["label"] = v[key]
				}
			}
			addGroupByType(conn, value, typeMap, cache)
		}
	}
}

func (o *objectBuilderRepo) UpdateWithParams(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	params := cast.ToStringMap(data["params"])
	delete(data, "params")
	fields := []string{}

	queryField := `SELECT f.slug FROM field f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, queryField, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		slug := ""

		err = fieldRows.Scan(&slug)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		fields = append(fields, slug)
	}

	argCount := 1
	args := []interface{}{}

	query := fmt.Sprintf(`UPDATE %s SET `, req.TableSlug)

	filter := " WHERE 1=1 "

	for _, slug := range fields {
		arg, ok := data[slug]
		if ok {
			query += fmt.Sprintf(` %s = $%d,`, slug, argCount)
			argCount++
			args = append(args, arg)
		}

		val, ok := params[slug]
		if ok {

			switch val.(type) {
			case map[string]interface{}:
				newOrder := cast.ToStringMap(val)

				for k, v := range newOrder {

					switch v.(type) {
					case string:
						if cast.ToString(v) == "" {
							continue
						}
					}

					if k == "$gt" {
						filter += fmt.Sprintf(" AND %s > $%d ", slug, argCount)
					} else if k == "$gte" {
						filter += fmt.Sprintf(" AND %s >= $%d ", slug, argCount)
					} else if k == "$lt" {
						filter += fmt.Sprintf(" AND %s < $%d ", slug, argCount)
					} else if k == "$lte" {
						filter += fmt.Sprintf(" AND %s <= $%d ", slug, argCount)
					}

					args = append(args, v)
					argCount++
				}
			default:
				filter += fmt.Sprintf(` AND %s = $%d`, slug, argCount)
				argCount++
				args = append(args, val)
			}
		}
	}

	query = strings.TrimRight(query, ",")
	query = query + filter

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
	}, nil
}

func (o *objectBuilderRepo) GetListV2(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		tableSlugs, tableSlugsTable, searchFields []string
		fields                                    = make(map[string]interface{})
		tableOrderBy                              bool
		args, result                              []interface{}
		count, argCount                           = 0, 1
		filter, limit, offset                     = " WHERE deleted_at IS NULL ", " LIMIT 20 ", " OFFSET 0"
		order, searchCondition                    = " ORDER BY a.created_at DESC ", " OR "
		query                                     = `SELECT jsonb_build_object( `
		fquery                                    = `SELECT f.slug, f.type, t.order_by, f.is_search FROM field f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`
	)

	params, _ := helper.ConvertStructToMap(req.Data)

	fieldRows, err := conn.Query(ctx, fquery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields by table slug")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			slug, ftype string
			isSearch    bool
		)

		err := fieldRows.Scan(&slug, &ftype, &tableOrderBy, &isSearch)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		if ftype == "DATE_TIME_WITHOUT_TIME_ZONE" {
			query += fmt.Sprintf(`'%s', TO_CHAR(a.%s, 'DD.MM.YYYY HH24:MI'),`, slug, slug)
			continue
		}

		query += fmt.Sprintf(`'%s', a.%s,`, slug, slug)
		fields[slug] = ftype

		if strings.Contains(slug, "_id") && !strings.Contains(slug, req.TableSlug) && ftype == "LOOKUP" {
			tableSlugs = append(tableSlugs, slug)
			parts := strings.Split(slug, "_")
			if len(parts) > 2 {
				lastPart := parts[len(parts)-1]
				if _, err := strconv.Atoi(lastPart); err == nil {
					slug = strings.ReplaceAll(slug, fmt.Sprintf("_%v", lastPart), "")
				}
			}
			tableSlugsTable = append(tableSlugsTable, strings.ReplaceAll(slug, "_id", ""))
		}

		if helper.FIELD_TYPES[ftype] == "VARCHAR" && isSearch {
			searchFields = append(searchFields, slug)
		}
	}

	_, ok := params["with_relations"]

	if cast.ToBool(params["with_relations"]) || !ok {
		for i, slug := range tableSlugs {
			as := fmt.Sprintf("r%d", i+1)

			query += fmt.Sprintf(`'%s_data', (
				SELECT row_to_json(%s)
				FROM "%s" %s WHERE %s.guid = a.%s
			),`, slug, as, tableSlugsTable[i], as, as, slug)
		}
	}

	query = strings.TrimRight(query, ",")
	query += fmt.Sprintf(`) AS DATA FROM "%s" a`, req.TableSlug)

	if !tableOrderBy {
		order = " ORDER BY a.created_at ASC "
	}

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
				if k == "created_at" {
					continue
				}
				oType := " ASC"
				if cast.ToInt(v) == -1 {
					oType = " DESC"
				}

				if counter == 0 {
					order += fmt.Sprintf(" a.%s"+oType, k)
				} else {
					order += fmt.Sprintf(", a.%s"+oType, k)
				}
				counter++
			}
		} else {
			_, ok := fields[key]
			if ok {
				switch val.(type) {
				case []string:
					filter += fmt.Sprintf(" AND a.%s IN($%d) ", key, argCount)
					args = append(args, pq.Array(val))
				case int, float32, float64, int32:
					filter += fmt.Sprintf(" AND a.%s = $%d ", key, argCount)
					args = append(args, val)
				case []interface{}:
					if fields[key] == "MULTISELECT" {
						filter += fmt.Sprintf(" AND a.%s && $%d", key, argCount)
						args = append(args, pq.Array(val))
					} else {
						filter += fmt.Sprintf(" AND a.%s = ANY($%d) ", key, argCount)
						args = append(args, pq.Array(val))
					}
				case map[string]interface{}:
					newOrder := cast.ToStringMap(val)

					for k, v := range newOrder {
						switch v.(type) {
						case string:
							if cast.ToString(v) == "" {
								continue
							}
						}

						if k == "$gt" {
							filter += fmt.Sprintf(" AND a.%s > $%d ", key, argCount)
						} else if k == "$gte" {
							filter += fmt.Sprintf(" AND a.%s >= $%d ", key, argCount)
						} else if k == "$lt" {
							filter += fmt.Sprintf(" AND a.%s < $%d ", key, argCount)
						} else if k == "$lte" {
							filter += fmt.Sprintf(" AND a.%s <= $%d ", key, argCount)
						} else if k == "$in" {
							filter += fmt.Sprintf(" AND a.%s::varchar = ANY($%d)", key, argCount)
						}

						args = append(args, val)
						argCount++
					}
				default:
					if strings.Contains(key, "_id") || key == "guid" {
						if req.TableSlug == "client_type" {
							filter += " AND a.guid = ANY($1::uuid[]) "

							args = append(args, pq.Array(cast.ToStringSlice(val)))
						} else {
							filter += fmt.Sprintf(" AND a.%s = $%d ", key, argCount)
							args = append(args, val)
						}
					} else {
						val = escapeSpecialCharacters(cast.ToString(val))
						filter += fmt.Sprintf(" AND a.%s ~* $%d ", key, argCount)
						args = append(args, val)
					}
				}
				argCount++
			}
		}
	}

	searchValue := cast.ToString(params["search"])
	if len(searchValue) > 0 {
		for idx, val := range searchFields {
			if idx == 0 {
				filter += " AND ("
				searchCondition = ""
			} else {
				searchCondition = " OR "
			}
			filter += fmt.Sprintf(" %s a.%s ~* $%d ", searchCondition, val, argCount)
			args = append(args, searchValue)
			argCount++

			if idx == len(searchFields)-1 {
				filter += " ) "
			}
		}
	}

	query += filter + order + limit + offset

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting rows")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			data interface{}
			temp = make(map[string]interface{})
		)

		values, err := rows.Values()
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting values")
		}

		for i, value := range values {
			temp[rows.FieldDescriptions()[i].Name] = value
			data = temp["data"]
		}

		result = append(result, data)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s" AS a %s`, req.TableSlug, filter)
	err = conn.QueryRow(ctx, countQuery, args...).Scan(&count)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting count")
	}

	rr := map[string]interface{}{
		"response": result,
		"count":    count,
	}

	response, _ := helper.ConvertMapToStruct(rr)

	return &nb.CommonMessage{
		Data: response,
	}, nil
}

func (o *objectBuilderRepo) GetSingleSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "convert req data")
	}

	output, err := helper.GetItem(ctx, conn, req.TableSlug, cast.ToString(data["id"]))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get item by id")
	}

	query := `SELECT 
		f."id",
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
	FROM "field" as f JOIN "table" as t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get fields by table slug")
	}
	defer fieldRows.Close()

	fields := []models.Field{}

	for fieldRows.Next() {
		var (
			field                          = models.Field{}
			atr                            = []byte{}
			autoFillField, autoFillTable   sql.NullString
			relationId, defaultNull, index sql.NullString
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.TableId,
			&field.Required,
			&field.Slug,
			&field.Label,
			&defaultNull,
			&field.Type,
			&index,
			&atr,
			&field.IsVisible,
			&autoFillField,
			&autoFillTable,
			&field.Unique,
			&field.Automatic,
			&relationId,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "scan fields")
		}

		field.AutofillField = autoFillField.String
		field.AutofillTable = autoFillTable.String
		field.RelationId = relationId.String
		field.Default = defaultNull.String
		field.Index = index.String

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "unmarshal attributes")
		}

		fields = append(fields, field)
	}

	var (
		attributeTableFromSlugs, attributeTableFromRelationIds []string
		relationFieldTablesMap                                 = make(map[string]interface{})
		relationFieldTableIds                                  = []string{}
	)

	for _, field := range fields {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "convert struct to map")
		}
		if field.Type == "FORMULA" {
			if cast.ToString(attributes["table_from"]) != "" && cast.ToString(attributes["sum_field"]) != "" {
				attributeTableFromSlugs = append(attributeTableFromSlugs, strings.Split(cast.ToString(attributes["table_from"]), "#")[0])
				attributeTableFromRelationIds = append(attributeTableFromRelationIds, strings.Split(cast.ToString(attributes["table_from"]), "#")[1])
			}
		}
	}

	query = `SELECT id, slug FROM "table" WHERE slug IN ($1)`

	tableRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromSlugs))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get tables by slugs")
	}
	defer tableRows.Close()

	for tableRows.Next() {
		table := models.Table{}

		err = tableRows.Scan(&table.Id, &table.Slug)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "scan tables")
		}

		relationFieldTableIds = append(relationFieldTableIds, table.Id)
		relationFieldTablesMap[table.Slug] = table
	}

	query = `SELECT slug, table_id, relation_id FROM "field" WHERE relation_id IN ($1) AND table_id IN ($2)`

	relationFieldRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds), pq.Array(relationFieldTableIds))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get relation fields")
	}
	defer relationFieldRows.Close()

	relationFieldsMap := make(map[string]string)

	for relationFieldRows.Next() {
		field := models.Field{}

		err = relationFieldRows.Scan(
			&field.Slug,
			&field.TableId,
			&field.RelationId,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "scan relation fields")
		}

		relationFieldsMap[field.RelationId+"_"+field.TableId] = field.Slug
	}

	query = `SELECT id, type, field_from FROM "relation" WHERE id IN ($1)`

	dynamicRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "get dynamic rows")
	}
	defer dynamicRows.Close()

	dynamicRelationsMap := make(map[string]models.Relation)

	for dynamicRows.Next() {
		relation := models.Relation{}

		err = dynamicRows.Scan(
			&relation.Id,
			&relation.Type,
			&relation.FieldFrom,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "scan dynamic rows")
		}

		dynamicRelationsMap[relation.Id] = relation
	}

	for _, field := range fields {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "convert struct to map")
		}

		if field.Type == "FORMULA" {
			_, tFrom := attributes["table_from"]
			_, sF := attributes["sum_field"]
			if tFrom && sF {
				resp, err := helper.CalculateFormulaBackend(ctx, conn, attributes, req.TableSlug)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "calculate formula backend")
				}
				_, ok := resp[cast.ToString(output["guid"])]
				if ok {
					output[field.Slug] = resp[cast.ToString(output["guid"])]
				} else {
					output[field.Slug] = 0
				}
			}
		} else if field.Type == "FORMULA_FRONTEND" {
			_, ok := attributes["formula"]
			if ok {
				resultFormula, err := helper.CalculateFormulaFrontend(attributes, fields, output)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "calculate formula frontend")
				}
				output[field.Slug] = resultFormula
			}
		}
	}

	response := make(map[string]interface{})
	response["response"] = output

	newBody, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "convert map to struct")
	}

	return &nb.CommonMessage{
		ProjectId: req.ProjectId,
		TableSlug: req.TableSlug,
		Data:      newBody,
	}, err
}

func (o *objectBuilderRepo) GetAllForDocx(ctx context.Context, req *nb.CommonMessage) (resp map[string]interface{}, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		params    = make(map[string]interface{})
		views     = []models.View{}
		fieldsMap = make(map[string]models.Field)
		item      = make(map[string]interface{})
		items     = make([]map[string]interface{}, 0)
		count     = 0
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, err
	}

	additionalFields := cast.ToStringMap(params["additional_fields"])
	fmt.Println("additional fields", additionalFields)

	delete(params, "table_slugs")
	delete(params, "additional_fields")

	var (
		roleIdFromToken = cast.ToString(params["role_id_from_token"])

		fields = []models.Field{}
	)

	query := `
		SELECT 
			f.id,
			f."table_id",
			t.slug,
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
		LEFT JOIN "relation" r ON r.id = f.relation_id
		WHERE t."slug" = $1
	`

	rows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return nil, err
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
			atrb              = make(map[string]interface{})
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.TableSlug,
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
			return nil, err
		}

		field.RelationId = relationIdNull.String
		field.AutofillField = autofillField.String
		field.AutofillTable = autofillTable.String
		field.Default = defaultStr.String
		field.Index = index.String

		if err := json.Unmarshal(attributes, &atrb); err != nil {
			return nil, err
		}

		attributes, _ = json.Marshal(atrb)

		if err := json.Unmarshal(attributes, &field.Attributes); err != nil {
			return nil, err
		}

		fields = append(fields, field)
		fieldsMap[field.Slug] = field
	}

	fieldsWithPermissions, _, err := helper.AddPermissionToField1(ctx, helper.AddPermissionToFieldRequest{Conn: conn, RoleId: roleIdFromToken, TableSlug: req.TableSlug, Fields: fields})
	if err != nil {
		return nil, err
	}

	rquery := `SELECT 
			f.id,
			f."table_id",
			t.slug,
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
	
	FROM field f 
	JOIN "table" t ON t.id = f.table_id
	JOIN relation r ON r.id = $1 WHERE f.id::text = ANY(r.view_fields)`

	reqlationQ := `
	SELECT
		r.id,
		r.table_from,
		r.table_to,
		r.field_from,
		r.field_to,
		r.type,
		r.relation_field_slug,
		r.editable,
		r.is_user_id_default,
		r.is_system,
		r.object_id_from_jwt,
		r.cascading_tree_table_slug,
		r.cascading_tree_field_slug,
		r.view_fields
	FROM
		relation r
	WHERE  r.id = $1`

	decodedFields := []models.Field{}
	for _, el := range fieldsWithPermissions {
		if el.Attributes != nil && !(el.Type == "LOOKUP" || el.Type == "LOOKUPS" || el.Type == "DYNAMIC") {
			decodedFields = append(decodedFields, el)
		} else {
			elementField := el
			viewFields := []models.Field{}

			if el.RelationId != "" {

				relation := models.RelationBody{}

				err = conn.QueryRow(ctx, reqlationQ, el.RelationId).Scan(
					&relation.Id,
					&relation.TableFrom,
					&relation.TableTo,
					&relation.FieldFrom,
					&relation.FieldTo,
					&relation.Type,
					&relation.RelationFieldSlug,
					&relation.Editable,
					&relation.IsUserIdDefault,
					&relation.IsSystem,
					&relation.ObjectIdFromJwt,
					&relation.CascadingTreeTableSlug,
					&relation.CascadingTreeFieldSlug,
					&relation.ViewFields,
				)

				if err != nil {
					if !strings.Contains(err.Error(), "no rows") {
						return nil, err
					}
				} else {
					elementField.RelationData = relation

					if relation.TableFrom != req.TableSlug {
						elementField.TableSlug = relation.TableFrom
					} else {
						elementField.TableSlug = relation.TableTo
					}

					frows, err := conn.Query(ctx, rquery, el.RelationId)
					if err != nil {
						return nil, err
					}
					defer frows.Close()

					for frows.Next() {
						var (
							vf                = models.Field{}
							attributes        = []byte{}
							relationIdNull    sql.NullString
							autofillField     sql.NullString
							autofillTable     sql.NullString
							defaultStr, index sql.NullString
						)

						err = frows.Scan(
							&vf.Id,
							&vf.TableId,
							&vf.TableSlug,
							&vf.Required,
							&vf.Slug,
							&vf.Label,
							&defaultStr,
							&vf.Type,
							&index,
							&attributes,
							&vf.IsVisible,
							&autofillField,
							&autofillTable,
							&vf.Unique,
							&vf.Automatic,
							&relationIdNull,
						)
						if err != nil {
							return nil, err
						}

						vf.RelationId = relationIdNull.String
						vf.AutofillField = autofillField.String
						vf.AutofillTable = autofillTable.String
						vf.Default = defaultStr.String
						vf.Index = index.String

						if err := json.Unmarshal(attributes, &vf.Attributes); err != nil {
							return nil, err
						}

						viewFields = append(viewFields, vf)
					}
				}

			}

			elementField.ViewFields = viewFields
			decodedFields = append(decodedFields, elementField)
		}
	}

	query = `SELECT 
		"id",
		"attributes",
		"table_slug",
		"type",
		"columns",
		"order",
		COALESCE("time_interval", 0),
		COALESCE("group_fields"::varchar[], '{}'),
		"name",
		"main_field",
		"quick_filters",
		"users",
		"view_fields",
		"calendar_from_slug",
		"calendar_to_slug",
		"multiple_insert",
		"status_field_slug",
		"is_editable",
		"relation_table_slug",
		"relation_id",
		"multiple_insert_field",
		"updated_fields",
		"table_label",
		"default_limit",
		"default_editable",
		"name_uz",
		"name_en"
	FROM "view" WHERE "table_slug" = $1 ORDER BY "order" ASC`

	viewRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return nil, errors.Wrap(err, "error while getting views by table slug")
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var (
			attributes          []byte
			view                = models.View{}
			Name                sql.NullString
			MainField           sql.NullString
			CalendarFromSlug    sql.NullString
			CalendarToSlug      sql.NullString
			StatusFieldSlug     sql.NullString
			RelationTableSlug   sql.NullString
			RelationId          sql.NullString
			MultipleInsertField sql.NullString
			TableLabel          sql.NullString
			DefaultLimit        sql.NullString
			NameUz              sql.NullString
			NameEn              sql.NullString
			QuickFilters        sql.NullString
		)

		err := viewRows.Scan(
			&view.Id,
			&attributes,
			&view.TableSlug,
			&view.Type,
			&view.Columns,
			&view.Order,
			&view.TimeInterval,
			&view.GroupFields,
			&Name,
			&MainField,
			&QuickFilters,
			&view.Users,
			&view.ViewFields,
			&CalendarFromSlug,
			&CalendarToSlug,
			&view.MultipleInsert,
			&StatusFieldSlug,
			&view.IsEditable,
			&RelationTableSlug,
			&RelationId,
			&MultipleInsertField,
			&view.UpdatedFields,
			&TableLabel,
			&DefaultLimit,
			&view.DefaultEditable,
			&NameUz,
			&NameEn,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error while scanning views")
		}

		view.Name = Name.String
		view.MainField = MainField.String
		view.CalendarFromSlug = CalendarFromSlug.String
		view.CalendarToSlug = CalendarToSlug.String
		view.StatusFieldSlug = StatusFieldSlug.String
		view.RelationTableSlug = RelationTableSlug.String
		view.RelationId = RelationId.String
		view.MultipleInsertField = MultipleInsertField.String
		view.TableLabel = TableLabel.String
		view.DefaultLimit = DefaultLimit.String
		view.NameUz = NameUz.String
		view.NameEn = NameEn.String

		if QuickFilters.Valid {
			err = json.Unmarshal([]byte(QuickFilters.String), &view.QuickFilters)
			if err != nil {
				return nil, errors.Wrap(err, "error while unmarshalling quick filters")
			}
		}

		if view.Columns == nil {
			view.Columns = []string{}
		}

		if err := json.Unmarshal(attributes, &view.Attributes); err != nil {
			return nil, errors.Wrap(err, "error while unmarshalling view attributes")
		}

		views = append(views, view)
	}

	response := map[string]interface{}{
		"count": count,
	}

	if _, ok := params[req.TableSlug+"_id"]; ok {
		item, err = helper.GetItem(ctx, conn, req.TableSlug, cast.ToString(params[req.TableSlug+"_id"]))
		if err != nil {
			return nil, errors.Wrap(err, "error while getting item")
		}

		additionalItems := make(map[string]interface{})
		for key, value := range additionalFields {
			if key != "folder_id" {
				additionalItem := make(map[string]interface{})

				additionalItem, err = helper.GetItem(ctx, conn, strings.TrimSuffix(key, "_id"), cast.ToString(value))
				if err != nil {
					return nil, errors.Wrap(err, "error while getting additional item")
				}

				additionalItems[key+"_data"] = additionalItem
			}
		}
		response["additional_items"] = additionalItems
		response["response"] = item
	} else {
		items, count, err = helper.GetItems(ctx, conn, models.GetItemsBody{
			TableSlug: req.TableSlug,
			Params:    params,
			FieldsMap: fieldsMap,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error while getting items")
		}
		response["response"] = items
	}

	return response, nil
}

func (o *objectBuilderRepo) GetAllFieldsForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	var (
		fields = []models.Field{}
	)

	query := `select f.table_id, f.label, f.slug from field f join "table" t on t.id = f.table_id where t.slug = $1`

	rows, err := conn.Query(ctx, query, req.GetTableSlug())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field = models.Field{}
		)

		if err = rows.Scan(&field.TableId, &field.Label, &field.Slug); err != nil {
			return nil, err
		}

		fields = append(fields, field)
	}

	item := map[string]interface{}{
		"fields":     fields,
		"relations:": []interface{}{},
	}

	res, err := helper.ConvertMapToStruct(item)
	if err != nil {
		return nil, err
	}

	return &nb.CommonMessage{
		TableSlug: req.GetTableSlug(),
		Data:      res,
	}, nil
}

func (o *objectBuilderRepo) GetListForDocxMultiTables(ctx context.Context, req *nb.CommonForDocxMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	params, _ := helper.ConvertStructToMap(req.Data)

	query := "WITH combined_data AS ("
	tableOrderBy := false
	fields := make(map[string]map[string]interface{})
	searchFields := make(map[string][]string)
	tableSubqueries := make([]string, len(req.GetTableSlugs()))

	for i, tableSlug := range req.GetTableSlugs() {
		fquery := `SELECT f.slug, f.type, t.order_by, f.is_search 
                   FROM field f 
                   JOIN "table" t ON t.id = f.table_id 
                   WHERE t.slug = $1`
		fieldRows, err := conn.Query(ctx, fquery, tableSlug)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		defer fieldRows.Close()

		fields[tableSlug] = make(map[string]interface{})
		searchFields[tableSlug] = []string{}

		tableSubqueries[i] = "SELECT jsonb_build_object("
		for fieldRows.Next() {
			var (
				slug, ftype string
				isSearch    bool
			)

			err := fieldRows.Scan(&slug, &ftype, &tableOrderBy, &isSearch)
			if err != nil {
				return &nb.CommonMessage{}, err
			}

			tableSubqueries[i] += fmt.Sprintf(`'%s', %s.%s,`, slug, tableSlug, slug)
			fields[tableSlug][slug] = ftype

			if helper.FIELD_TYPES[ftype] == "VARCHAR" && isSearch {
				searchFields[tableSlug] = append(searchFields[tableSlug], slug)
			}
		}

		if cast.ToBool(params["with_relations"]) {
			for j, slug := range req.GetTableSlugs() {
				as := fmt.Sprintf("r%d", j+1)
				tableSubqueries[i] += fmt.Sprintf(`'%s_id_data', (
                    SELECT row_to_json(%s)
                    FROM %s %s WHERE %s.guid = %s.%s_id
                ),`, slug, as, slug, as, as, tableSlug, slug)
			}
		}

		tableSubqueries[i] += fmt.Sprintf(`'table_slug', '%s'`, tableSlug)
		tableSubqueries[i] += fmt.Sprintf(`) AS data from %s`, tableSlug)
	}

	query += strings.Join(tableSubqueries, " UNION ALL ") + ")"

	query += " SELECT DISTINCT data FROM combined_data WHERE 1=1"

	filter := ""
	limit := " LIMIT 200"
	offset := " OFFSET 0"
	args := []interface{}{}
	argCount := 1

	for key, val := range params {
		for _, tableSlug := range req.GetTableSlugs() {
			if _, ok := fields[tableSlug][key]; ok {
				switch val.(type) {
				case []string:
					filter += fmt.Sprintf(" AND %s.%s IN($%d) ", tableSlug, key, argCount)
					args = append(args, pq.Array(val))
				case int, float32, float64, int32:
					filter += fmt.Sprintf(" AND %s.%s = $%d ", tableSlug, key, argCount)
					args = append(args, val)
				case []interface{}:
					if fields[tableSlug][key] == "MULTISELECT" {
						filter += fmt.Sprintf(" AND %s.%s && $%d", tableSlug, key, argCount)
						args = append(args, pq.Array(val))
					} else {
						filter += fmt.Sprintf(" AND %s.%s = ANY($%d) ", tableSlug, key, argCount)
						args = append(args, pq.Array(val))
					}
				case map[string]interface{}:
					newOrder := cast.ToStringMap(val)
					for k, v := range newOrder {
						switch v.(type) {
						case string:
							if cast.ToString(v) == "" {
								continue
							}
						}
						switch k {
						case "$gt":
							filter += fmt.Sprintf(" AND %s.%s > $%d ", tableSlug, key, argCount)
						case "$gte":
							filter += fmt.Sprintf(" AND %s.%s >= $%d ", tableSlug, key, argCount)
						case "$lt":
							filter += fmt.Sprintf(" AND %s.%s < $%d ", tableSlug, key, argCount)
						case "$lte":
							filter += fmt.Sprintf(" AND %s.%s <= $%d ", tableSlug, key, argCount)
						case "$in":
							filter += fmt.Sprintf(" AND %s.%s::varchar = ANY($%d)", tableSlug, key, argCount)
						}
						args = append(args, val)
						argCount++
					}
				default:
					if strings.Contains(key, "_id") || key == "guid" {
						//filter += fmt.Sprintf(" AND %s.%s = $%d ", tableSlug, key, argCount)
						filter += fmt.Sprintf(" AND data->>'%s' = $%d ", key, argCount)
						args = append(args, val)
					} else {
						filter += fmt.Sprintf(" AND %s.%s ~* $%d ", tableSlug, key, argCount)
						args = append(args, val)
					}
				}
				argCount++
			}
		}
	}

	searchValue := cast.ToString(params["search"])
	if len(searchValue) > 0 {
		for _, tableSlug := range req.GetTableSlugs() {
			for idx, val := range searchFields[tableSlug] {
				if idx == 0 {
					filter += " AND ("
				} else {
					filter += " OR "
				}
				filter += fmt.Sprintf(" %s.%s ~* $%d ", tableSlug, val, argCount)
				args = append(args, searchValue)
				argCount++
			}
		}
		filter += " ) "
	}

	query += filter + limit + offset

	fmt.Println("query:  docx new", query)
	fmt.Println("query args", args)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		for _, value := range values {
			res, _ := helper.ConvertMapToStruct(value.(map[string]interface{}))
			for j, val := range value.(map[string]interface{}) {
				if j == "table_slug" {
					if arr, ok := result[val.(string)]; ok {
						arr = append(arr.([]interface{}), res)
						result[val.(string)] = arr
					} else {
						result[val.(string)] = []interface{}{res}
					}
					break
				}
			}
		}
	}

	response, _ := helper.ConvertMapToStruct(result)

	return &nb.CommonMessage{
		Data: response,
	}, nil
}

func (o *objectBuilderRepo) GetListForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	conn := psqlpool.Get(req.GetProjectId())

	params, _ := helper.ConvertStructToMap(req.Data)

	fquery := `SELECT f.slug, f.type, t.order_by, f.is_search FROM field f JOIN "table" t ON t.id = f.table_id WHERE t.slug = $1`
	query := `SELECT jsonb_build_object( `

	tableSlugs := []string{}
	tableOrderBy := false
	fields := make(map[string]interface{})
	searchFields := []string{}

	fieldRows, err := conn.Query(ctx, fquery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			slug, ftype string
			isSearch    bool
		)

		err := fieldRows.Scan(&slug, &ftype, &tableOrderBy, &isSearch)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		query += fmt.Sprintf(`'%s', a.%s,`, slug, slug)
		fields[slug] = ftype

		if strings.Contains(slug, "_id") && !strings.Contains(slug, req.TableSlug) && ftype == "LOOKUP" {
			tableSlugs = append(tableSlugs, strings.ReplaceAll(slug, "_id", ""))
		}

		if helper.FIELD_TYPES[ftype] == "VARCHAR" && isSearch {
			searchFields = append(searchFields, slug)
		}
	}

	_, ok := params["with_relations"]

	if cast.ToBool(params["with_relations"]) || !ok {

		for i, slug := range tableSlugs {

			as := fmt.Sprintf("r%d", i+1)

			query += fmt.Sprintf(`'%s_id_data', (
				SELECT row_to_json(%s)
				FROM %s %s WHERE %s.guid = a.%s_id
			),`, slug, as, slug, as, as, slug)

		}
	}

	query = strings.TrimRight(query, ",")

	query += fmt.Sprintf(`) AS DATA FROM %s a`, req.TableSlug)

	filter := " WHERE 1=1 "
	limit := " LIMIT 20 "
	offset := " OFFSET 0"
	order := " ORDER BY a.created_at DESC "
	searchCondition := " OR "
	args := []interface{}{}
	argCount := 1

	if !tableOrderBy {
		order = " ORDER BY a.created_at ASC "
	}

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
				if k == "created_at" {
					continue
				}
				oType := " ASC"
				if cast.ToInt(v) == -1 {
					oType = " DESC"
				}

				if counter == 0 {
					order += fmt.Sprintf(" a.%s"+oType, k)
				} else {
					order += fmt.Sprintf(", a.%s"+oType, k)
				}
				counter++
			}
		} else {
			_, ok := fields[key]

			if ok {
				switch val.(type) {
				case []string:
					filter += fmt.Sprintf(" AND a.%s IN($%d) ", key, argCount)
					args = append(args, pq.Array(val))
				case int, float32, float64, int32:
					filter += fmt.Sprintf(" AND a.%s = $%d ", key, argCount)
					args = append(args, val)
				case []interface{}:
					if fields[key] == "MULTISELECT" {
						filter += fmt.Sprintf(" AND a.%s && $%d", key, argCount)
						args = append(args, pq.Array(val))
					} else {
						filter += fmt.Sprintf(" AND a.%s = ANY($%d) ", key, argCount)
						args = append(args, pq.Array(val))
					}
				case map[string]interface{}:
					newOrder := cast.ToStringMap(val)

					for k, v := range newOrder {
						switch v.(type) {
						case string:
							if cast.ToString(v) == "" {
								continue
							}
						}

						if k == "$gt" {
							filter += fmt.Sprintf(" AND a.%s > $%d ", key, argCount)
						} else if k == "$gte" {
							filter += fmt.Sprintf(" AND a.%s >= $%d ", key, argCount)
						} else if k == "$lt" {
							filter += fmt.Sprintf(" AND a.%s < $%d ", key, argCount)
						} else if k == "$lte" {
							filter += fmt.Sprintf(" AND a.%s <= $%d ", key, argCount)
						} else if k == "$in" {
							filter += fmt.Sprintf(" AND a.%s::varchar = ANY($%d)", key, argCount)
						}

						args = append(args, val)

						argCount++
					}
				default:
					if strings.Contains(key, "_id") || key == "guid" {
						if req.TableSlug == "client_type" {
							filter += " AND a.guid = ANY($1::uuid[]) "

							args = append(args, pq.Array(cast.ToStringSlice(val)))
						} else {
							filter += fmt.Sprintf(" AND a.%s = $%d ", key, argCount)
							args = append(args, val)
						}
					} else {
						filter += fmt.Sprintf(" AND a.%s ~* $%d ", key, argCount)
						args = append(args, val)
					}
				}

				argCount++
			}
		}
	}

	searchValue := cast.ToString(params["search"])
	if len(searchValue) > 0 {
		for idx, val := range searchFields {
			if idx == 0 {
				filter += " AND ("
				searchCondition = ""
			} else {
				searchCondition = " OR "
			}
			filter += fmt.Sprintf(" %s a.%s ~* $%d ", searchCondition, val, argCount)
			args = append(args, searchValue)
			argCount++

			if idx == len(searchFields)-1 {
				filter += " ) "
			}
		}
	}

	// countQuery += filter
	query += filter + order + limit + offset

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer rows.Close()

	result := []interface{}{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		var (
			data interface{}
			temp = make(map[string]interface{})
		)

		for i, value := range values {
			temp[rows.FieldDescriptions()[i].Name] = value
			data = temp["data"]
		}

		result = append(result, data)
	}

	rr := map[string]interface{}{
		"response": result,
	}

	response, _ := helper.ConvertMapToStruct(rr)

	return &nb.CommonMessage{
		Data: response,
	}, nil
}

func escapeSpecialCharacters(input string) string {
	return regexp.QuoteMeta(input)
}

var letters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
var sh = "Sheet1"
