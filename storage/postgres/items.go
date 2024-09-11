package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/config"
	pa "ucode/ucode_go_object_builder_service/genproto/auth_service"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type itemsRepo struct {
	db *pgxpool.Pool
}

func NewItemsRepo(db *pgxpool.Pool) storage.ItemsRepoI {
	return &itemsRepo{
		db: db,
	}
}

var Ftype = map[string]string{
	"INCREMENT_NUMBER": "INCREMENT_NUMBER",
	"INCREMENT_ID":     "INCREMENT_ID",
	"MANUAL_STRING":    "MANUAL_STRING",
	"RANDOM_UUID":      "RANDOM_UUID",
	"RANDOM_TEXT":      "RANDOM_TEXT",
	"RANDOM_NUMBER":    "RANDOM_NUMBER",
}

func (i *itemsRepo) Create(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	var (
		conn            = psqlpool.Get(req.GetProjectId())
		args            = []interface{}{}
		argCount        = 3
		tableSlugs      = []string{}
		fieldM          = make(map[string]helper.FieldBody)
		query, valQuery string
		fields          = []models.Field{}
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}
	defer tx.Rollback(ctx)

	fQuery := ` SELECT
		f."id",
		f."type",
		f."attributes",
		f."relation_id",
		f."autofill_table",
		f."autofill_field",
		f."slug"
	FROM "field" f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := tx.Query(ctx, fQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			field                                    = models.Field{}
			atr                                      = []byte{}
			autoFillTable, autoFillField, relationId sql.NullString
			attributes                               = make(map[string]interface{})
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&atr,
			&relationId,
			&autoFillTable,
			&autoFillField,
			&field.Slug,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}
		if err := json.Unmarshal(atr, &attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}

		tableSlugs = append(tableSlugs, field.Slug)

		if _, ok := Ftype[field.Type]; ok {
			fieldM[field.Type] = helper.FieldBody{
				Slug:       field.Slug,
				Attributes: attributes,
			}
		}

		field.AutofillField = autoFillField.String
		field.AutofillTable = autoFillTable.String
		field.RelationId = relationId.String

		fields = append(fields, field)
	}

	reqBody := helper.CreateBody{
		FieldMap:   fieldM,
		Fields:     fields,
		TableSlugs: tableSlugs,
	}

	data, appendMany2Many, err := helper.PrepareToCreateInObjectBuilderWithTx(ctx, tx, req, reqBody)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while preparing to create in object builder")
	}

	if req.TableSlug != "client_type" && req.TableSlug != "role" {
		query = fmt.Sprintf(`INSERT INTO "%s" (guid, folder_id`, req.TableSlug)
		valQuery = ") VALUES ($1, $2"
	} else {
		argCount--
		query = fmt.Sprintf(`INSERT INTO "%s" (guid`, req.TableSlug)
		valQuery = ") VALUES ($1,"
	}

	guid := cast.ToString(data["guid"])
	var folderId interface{}

	if helper.IsEmpty(data["guid"]) {
		guid = uuid.NewString()
	}
	if helper.IsEmpty(data["folder_id"]) {
		folderId = nil
	} else {
		folderId = data["folder_id"]
	}

	if req.TableSlug != "client_type" && req.TableSlug != "role" {
		args = append(args, guid, folderId)
	} else {
		args = append(args, guid)
	}

	delete(data, "guid")
	delete(data, "folder_id")

	for _, fieldSlug := range tableSlugs {
		if exist := config.SkipFields[fieldSlug]; exist {
			continue
		}

		if strings.Contains(fieldSlug, "_id") && !strings.Contains(fieldSlug, "_ids") && strings.Contains(fieldSlug, req.TableSlug) {
			_, ok := data[fieldSlug]
			if ok {
				id := cast.ToStringSlice(data[fieldSlug])[0]
				query += fmt.Sprintf(", %s", fieldSlug)
				args = append(args, id)
				if argCount != 2 {
					valQuery += ","
				}

				valQuery += fmt.Sprintf(" $%d", argCount)
				argCount++
			}
		} else {
			val, ok := data[fieldSlug]
			if ok {
				if strVal, isString := val.(string); isString {
					const inputLayout = "02.01.2006 15:04"
					const outputLayout = "2006-01-02 15:04:05"

					if t, err := time.Parse(inputLayout, strVal); err == nil {
						val = t.Format(outputLayout)
					}
				}

				query += fmt.Sprintf(", %s", fieldSlug)
				args = append(args, val)
				if argCount != 2 {
					valQuery += ","
				}

				valQuery += fmt.Sprintf(" $%d", argCount)
				argCount++
			}
		}
	}

	if len(args) == 1 {
		valQuery = strings.TrimRight(valQuery, ",")
	}

	query = query + valQuery + ")"

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing query")
	}

	// ! Skip AppendMany2Many

	var (
		tableData       = models.Table{}
		attr            = []byte{}
		tableAttributes = make(map[string]interface{})
	)

	query = `SELECT 
		id,
		slug,
		is_login_table,
		attributes
	FROM "table" WHERE slug = $1
	`

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(
		&tableData.Id,
		&tableData.Slug,
		&tableData.IsLoginTable,
		&attr,
	)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning table")
	}

	if tableData.IsLoginTable && !cast.ToBool(data["from_auth_service"]) {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}

		_, ok := tableAttributes["auth_info"]
		if ok {
			count := 0
			authInfo := cast.ToStringMap(tableAttributes["auth_info"])
			loginStarg := cast.ToStringSlice(authInfo["login_strategy"])

			if cast.ToString(authInfo["client_type_id"]) == "" || cast.ToString(authInfo["role_id"]) == "" {
				return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
			}

			for _, ls := range loginStarg {
				if ls == "login" {
					if cast.ToString(authInfo["login"]) == "" || cast.ToString(authInfo["password"]) == "" {
						return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given login password")
					}
				} else if ls == "email" {
					if cast.ToString(authInfo["email"]) == "" {
						return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
					}
				} else if ls == "phone" {
					if cast.ToString(authInfo["phone"]) == "" {
						return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
					}
				}
			}

			query = `SELECT COUNT(*) FROM "client_type" WHERE guid = $1 AND ( table_slug = $2 OR name = 'ADMIN')`

			err = tx.QueryRow(ctx, query, data["client_type_id"], req.TableSlug).Scan(&count)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning count")
			}

			if count != 0 {
				data["authInfo"] = authInfo
				data["create_user"] = true
			}
		}
	} else {
		data["create_user"] = false
	}

	err = helper.AppendMany2Many(ctx, tx, appendMany2Many)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while appending many2many")
	}

	data["guid"] = guid
	if req.TableSlug != "client_type" && req.TableSlug != "role" {
		data["folder_id"] = folderId
	}
	newData, err := helper.ConvertMapToStruct(data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while committing")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		Data:      newData,
	}, nil
}

func (i *itemsRepo) Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	var (
		conn     = psqlpool.Get(req.GetProjectId())
		args     = []interface{}{}
		argCount = 2
		guid     string
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}
	defer tx.Rollback(ctx)

	data, err := helper.PrepareToUpdateInObjectBuilder(ctx, req, conn)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while preparing to update in object builder")
	}

	_, ok := data["guid"]
	if !ok {
		data["guid"] = data["id"]
	}
	guid = cast.ToString(data["guid"])
	_, ok = data["auth_guid"]
	if ok {
		data["guid"] = data["auth_guid"]
	}

	args = append(args, guid)

	query := fmt.Sprintf(`UPDATE "%s" SET `, req.TableSlug)

	fieldQuery := `SELECT f.slug, f.type FROM "field" as f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := tx.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var (
			fieldSlug, fieldType string
		)

		err = fieldRows.Scan(&fieldSlug, &fieldType)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}
		val, ok := data[fieldSlug]
		if fieldType == "MULTISELECT" {
			switch val.(type) {
			case string:
				val = []string{cast.ToString(val)}
			}
		} else if fieldType == "DATE_TIME_WITHOUT_TIME_ZONE" {
			switch val.(type) {
			case string:
				val = helper.ConvertTimestamp2DB(cast.ToString(val))
			}
		}
		if ok {
			query += fmt.Sprintf(`%s=$%d, `, fieldSlug, argCount)
			argCount++
			args = append(args, val)
		}
	}

	query = strings.TrimRight(query, ", ")

	query += " WHERE guid = $1"

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing query")
	}

	output, err := helper.GetItemWithTx(ctx, req.TableSlug, guid, conn)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
	}

	response, err := helper.ConvertMapToStruct(output)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while committing")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      response,
	}, nil
}

func (i *itemsRepo) GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
	}

	output, err := helper.GetItem(ctx, conn, req.TableSlug, cast.ToString(data["id"]))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
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
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields")
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
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		field.AutofillField = autoFillField.String
		field.AutofillTable = autoFillTable.String
		field.RelationId = relationId.String
		field.Default = defaultNull.String
		field.Index = index.String

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}

		fields = append(fields, field)
	}

	var (
		attributeTableFromSlugs       = []string{}
		attributeTableFromRelationIds = []string{}
		relationFieldTablesMap        = make(map[string]interface{})
		relationFieldTableIds         = []string{}
	)

	for _, field := range fields {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
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
		return &nb.CommonMessage{}, errors.Wrap(err, "error while querying")
	}
	defer tableRows.Close()

	for tableRows.Next() {
		table := models.Table{}

		err = tableRows.Scan(&table.Id, &table.Slug)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
		}

		relationFieldTableIds = append(relationFieldTableIds, table.Id)
		relationFieldTablesMap[table.Slug] = table
	}

	query = `SELECT slug, table_id, relation_id FROM "field" WHERE relation_id IN ($1) AND table_id IN ($2)`

	relationFieldRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds), pq.Array(relationFieldTableIds))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while querying")
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
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
		}

		relationFieldsMap[field.RelationId+"_"+field.TableId] = field.Slug
	}

	query = `SELECT id, type, field_from FROM "relation" WHERE id IN ($1)`

	dynamicRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds))
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while querying")
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
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
		}

		dynamicRelationsMap[relation.Id] = relation
	}

	isChanged := false

	for _, field := range fields {
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
				_, ok := resp[cast.ToString(output["guid"])]
				if ok {
					output[field.Slug] = resp[cast.ToString(output["guid"])]
					isChanged = true
				} else {
					output[field.Slug] = 0
					isChanged = true
				}
			}
		} else if field.Type == "FORMULA_FRONTEND" {
			_, ok := attributes["formula"]
			if ok {
				resultFormula, err := helper.CalculateFormulaFrontend(attributes, fields, output)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while calculating formula frontend")
				}
				if output[field.Slug] != resultFormula {
					isChanged = true
				}
				output[field.Slug] = resultFormula
			}
		}
	}

	response := make(map[string]interface{})

	response["response"] = output
	response["fields"] = fields

	newBody, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	if isChanged {
		go i.Update(ctx, &nb.CommonMessage{
			ProjectId: req.ProjectId,
			TableSlug: req.TableSlug,
			Data:      newBody,
		})
	}

	// ? SKIP ...

	return &nb.CommonMessage{
		ProjectId: req.ProjectId,
		TableSlug: req.TableSlug,
		Data:      newBody,
	}, err
}

func (i *itemsRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	return &nb.CommonMessage{}, nil
}

func (i *itemsRepo) Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	var (
		conn       = psqlpool.Get(req.GetProjectId())
		table      = models.Table{}
		atr        = []byte{}
		attributes = make(map[string]interface{})
	)

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
	}

	id := cast.ToString(data["id"])

	response, err := helper.GetItem(ctx, conn, req.TableSlug, id)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
	}

	query := `SELECT slug, attributes, is_login_table, soft_delete FROM "table" WHERE slug = $1`

	err = conn.QueryRow(ctx, query, req.TableSlug).Scan(
		&table.Slug,
		&atr,
		&table.IsLoginTable,
		&table.SoftDelete,
	)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
	}

	if err := json.Unmarshal(atr, &attributes); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling")
	}

	_, ok := attributes["auth_info"]
	if ok {
		response["delete_user"] = true

		authInfo := cast.ToStringMap(attributes["auth_info"])
		_, clienType := response[cast.ToString(authInfo["client_type_id"])]
		_, role := response[cast.ToString(authInfo["role_id"])]

		if !clienType && !role {
			return &nb.CommonMessage{}, errors.Wrap(fmt.Errorf("this table is auth table. auth information not fully given"), "error while checking auth table")
		}

		query := `SELECT COUNT(*) FROM client_type WHERE guid = $1 AND table_slug = $2`
		count := 0

		err = conn.QueryRow(ctx, query, response[cast.ToString(authInfo["client_type_id"])], req.TableSlug).Scan(
			&count,
		)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
		}

		if count != 0 {
			data["login_data"] = true
		}
	}

	query = fmt.Sprintf(`UPDATE "%s" SET deleted_at = CURRENT_TIMESTAMP WHERE guid = $1`, req.TableSlug)

	_, err = conn.Exec(ctx, query, id)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing")
	}

	response["attributes"] = attributes

	newRes, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      newRes,
	}, nil
}

func (i *itemsRepo) UpdateGuid(ctx context.Context, req *models.ItemsChangeGuid) error {
	conn := psqlpool.Get(req.ProjectId)

	query := fmt.Sprintf(`UPDATE "%s" SET guid = $2 WHERE guid = $1`, req.TableSlug)

	_, err := conn.Exec(ctx, query, req.OldId, req.NewId)
	if err != nil {
		return errors.Wrap(err, "error while executing query")
	}

	return nil
}

func (i *itemsRepo) DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *models.DeleteUsers, err error) {
	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &models.DeleteUsers{}, errors.Wrap(err, "error while converting struct to map")
	}

	var (
		conn       = psqlpool.Get(req.GetProjectId())
		table      = models.Table{}
		atr        = []byte{}
		attributes = make(map[string]interface{})
		users      = []*pa.DeleteManyUserRequest_User{}
		isDelete   bool
		ids        = cast.ToStringSlice(data["ids"])
	)

	query := `SELECT slug, attributes, is_login_table, soft_delete FROM "table" WHERE slug = $1`

	err = conn.QueryRow(ctx, query, req.TableSlug).Scan(
		&table.Slug,
		&atr,
		&table.IsLoginTable,
		&table.SoftDelete,
	)
	if err != nil {
		return &models.DeleteUsers{}, errors.Wrap(err, "error while scanning")
	}

	if err := json.Unmarshal(atr, &attributes); err != nil {
		return &models.DeleteUsers{}, errors.Wrap(err, "error while unmarshalling")
	}

	_, ok := attributes["auth_info"]
	if table.IsLoginTable && ok {
		isDelete = true

		authInfo := cast.ToStringMap(attributes["auth_info"])

		clientType := cast.ToString(authInfo["client_type_id"])
		role := cast.ToString(authInfo["role_id"])

		query = fmt.Sprintf(`SELECT guid, %s, %s FROM %s WHERE guid = ANY($1)`, clientType, role, req.TableSlug)

		rows, err := conn.Query(ctx, query, ids)
		if err != nil {
			return &models.DeleteUsers{}, errors.Wrap(err, "error while querying")
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id, roleId, clientTypeId string
			)

			err = rows.Scan(
				&id,
				&clientTypeId,
				&roleId,
			)
			if err != nil {
				return &models.DeleteUsers{}, errors.Wrap(err, "error while scanning")
			}

			users = append(users, &pa.DeleteManyUserRequest_User{
				UserId:       id,
				RoleId:       roleId,
				ClientTypeId: clientTypeId,
			})
		}
	}

	if table.SoftDelete {
		query = fmt.Sprintf(`UPDATE %s SET deleted_at = CURRENT_TIMESTAMP WHERE guid = ANY($1)`, req.TableSlug)
	} else {
		query = fmt.Sprintf(`DELETE FROM %s WHERE guid = ANY($1)`, req.TableSlug)
	}

	_, err = conn.Exec(ctx, query, ids)
	if err != nil {
		return &models.DeleteUsers{}, errors.Wrap(err, "error while executing")
	}

	return &models.DeleteUsers{
		IsDelete:      isDelete,
		Users:         users,
		ProjectId:     cast.ToString(data["company_service_project_id"]),
		EnvironmentId: cast.ToString(data["company_service_environment_id"]),
	}, nil
}

func (i *itemsRepo) MultipleUpdate(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	for _, obj := range cast.ToSlice(data["objects"]) {
		object := cast.ToStringMap(obj)

		newObj, err := helper.ConvertMapToStruct(object)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		isNew := object["is_new"]
		if !cast.ToBool(isNew) {
			_, err := i.Update(ctx, &nb.CommonMessage{
				ProjectId: req.ProjectId,
				TableSlug: req.TableSlug,
				Data:      newObj,
			})
			if err != nil {
				return &nb.CommonMessage{}, err
			}

		} else {
			_, err := i.Create(ctx, &nb.CommonMessage{
				ProjectId: req.ProjectId,
				TableSlug: req.TableSlug,
				Data:      newObj,
			})
			if err != nil {
				return &nb.CommonMessage{}, err
			}
		}
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
	}, nil
}

func (i *itemsRepo) UpsertMany(ctx context.Context, req *nb.CommonMessage) error {
	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return errors.Wrap(err, "upsertMany convert req")
	}

	var (
		conn = psqlpool.Get(req.GetProjectId())

		objects    = cast.ToSlice(data["objects"])
		fieldSlug  = data["field_slug"].(string)
		fieldSlugs = make([]models.Field, 0)

		insertQuery = fmt.Sprintf(`INSERT INTO "%s" (`, req.TableSlug)
		valuesQuery = " ) VALUES "
		updateQuery = fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET ", fieldSlug)
		args        []interface{}
		argCount    = 1
	)

	fieldRows, err := conn.Query(ctx, `SELECT f.slug, f.type FROM "field" as f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`, req.TableSlug)
	if err != nil {
		return errors.Wrap(err, "upsertMany get fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		field := models.Field{}
		err = fieldRows.Scan(&field.Slug, &field.Type)
		if err != nil {
			return errors.Wrap(err, "upsertMany fields scan")
		}
		fieldSlugs = append(fieldSlugs, field)
	}

	for _, field := range fieldSlugs {
		if exist := config.SkipFields[field.Slug]; exist {
			continue
		}
		insertQuery += fmt.Sprintf(`%s, `, field.Slug)
		updateQuery += fmt.Sprintf(`%s = EXCLUDED.%s, `, field.Slug, field.Slug)
	}

	insertQuery = insertQuery[:len(insertQuery)-2]
	updateQuery = updateQuery[:len(updateQuery)-2]

	for _, obj := range objects {
		data := cast.ToStringMap(obj)
		valuesQuery += "("
		for _, field := range fieldSlugs {
			if exist := config.SkipFields[field.Slug]; exist {
				continue
			}

			val, ok := data[field.Slug]
			if ok {
				if field.Type == "MULTISELECT" {
					switch val.(type) {
					case string:
						val = []string{cast.ToString(val)}
					}
				} else if field.Type == "DATE_TIME_WITHOUT_TIME_ZONE" {
					switch val.(type) {
					case string:
						val = helper.ConvertTimestamp2DB(cast.ToString(val))
					}
				}

				valuesQuery += fmt.Sprintf(`$%d, `, argCount)
				args = append(args, val)
				argCount++
			} else {
				valuesQuery += "NULL, "
			}
		}

		valuesQuery = valuesQuery[:len(valuesQuery)-2] + "), "
	}

	valuesQuery = valuesQuery[:len(valuesQuery)-2]

	var query = insertQuery + valuesQuery + updateQuery

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "upsertMany execute query")
	}

	return nil
}
