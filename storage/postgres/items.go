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
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/person"
	"ucode/ucode_go_object_builder_service/pkg/security"
	"ucode/ucode_go_object_builder_service/pkg/util"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type itemsRepo struct {
	db         *psqlpool.Pool
	grpcClient client.ServiceManagerI
}

func NewItemsRepo(db *psqlpool.Pool, grpcClient client.ServiceManagerI) storage.ItemsRepoI {
	return &itemsRepo{
		db:         db,
		grpcClient: grpcClient,
	}
}

func (i *itemsRepo) Create(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.Create")
	defer dbSpan.Finish()

	var (
		conn            = psqlpool.Get(req.GetProjectId())
		fieldM          = make(map[string]helper.FieldBody)
		tableData       = models.Table{}
		fields          = []models.Field{}
		args            = []any{}
		tableSlugs      = []string{}
		attr            = []byte{}
		argCount        = 3
		query, valQuery string
		isSystemTable   sql.NullBool
		authInfo        models.AuthInfo
		tableAttributes models.TableAttributes
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	body, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error marshalling request data")
	}

	query = `SELECT id, slug, is_login_table, attributes FROM "table" WHERE slug = $1 `

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(&tableData.Id, &tableData.Slug, &tableData.IsLoginTable, &attr)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning table")
	}

	fQuery := ` SELECT
		f."id",
		f."type",
		f."attributes",
		f."relation_id",
		f."autofill_table",
		f."autofill_field",
		f."slug",
		t."is_system"
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
			attributes                               = make(map[string]any)
			autoFillTable, autoFillField, relationId sql.NullString
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&atr,
			&relationId,
			&autoFillTable,
			&autoFillField,
			&field.Slug,
			&isSystemTable,
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

		if config.Ftype[field.Type] {
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

	if cast.ToBool(body["from_auth_service"]) {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}

		authInfo = tableAttributes.AuthInfo

		var (
			login         = cast.ToString(body["login"])
			phone         = cast.ToString(body["phone"])
			email         = cast.ToString(body["email"])
			roleId        = cast.ToString(body["role_id"])
			clientTypeId  = cast.ToString(body["client_type_id"])
			loginStrategy = cast.ToStringSlice(authInfo.LoginStrategy)
			password      = cast.ToString(body["password"])
		)

		if len(authInfo.ClientTypeID) == 0 || len(authInfo.RoleID) == 0 {
			return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information fully given")
		}

		delete(body, "client_type_id")
		delete(body, "role_id")
		body[authInfo.ClientTypeID] = clientTypeId
		body[authInfo.RoleID] = roleId

		for _, ls := range loginStrategy {
			if ls == "login" {
				if authInfo.Login == "" || authInfo.Password == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given login password")
				}
				delete(body, "login")
				delete(body, "password")
				body[authInfo.Login] = login
				body[authInfo.Password] = cast.ToString(body["password"])
				hashedPassword, err := security.HashPasswordBcrypt(password)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while hashing password")
				}
				password = hashedPassword
			} else if ls == "email" {
				if authInfo.Email == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given phone")
				}
				delete(body, "email")
				body[authInfo.Email] = email
			} else if ls == "phone" {
				if authInfo.Phone == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given email")
				}
				delete(body, "phone")
				body[authInfo.Phone] = phone
			}
		}

		err = i.InsertPersonTable(ctx, &models.PersonRequest{
			Tx:           tx,
			Guid:         cast.ToString(body["guid"]),
			UserIdAuth:   cast.ToString(body["user_id_auth"]),
			Login:        cast.ToString(body[authInfo.Login]),
			Password:     password,
			Email:        cast.ToString(body[authInfo.Email]),
			Phone:        cast.ToString(body[authInfo.Phone]),
			RoleId:       cast.ToString(body[authInfo.RoleID]),
			ClientTypeId: cast.ToString(body[authInfo.ClientTypeID]),
			FullName:     cast.ToString(body["name"]),
			Image:        cast.ToString(body["photo"]),
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while inserting to person")
		}
	}

	data, appendMany2Many, err := helper.PrepareToCreateInObjectBuilderWithTx(ctx, tx, req, helper.CreateBody{
		FieldMap:   fieldM,
		Fields:     fields,
		TableSlugs: tableSlugs,
	})
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while preparing to create in object builder")
	}

	if !isSystemTable.Bool {
		query = fmt.Sprintf(`INSERT INTO "%s" (guid, folder_id`, req.TableSlug)
		valQuery = ") VALUES ($1, $2"
	} else {
		argCount--
		query = fmt.Sprintf(`INSERT INTO "%s" (guid`, req.TableSlug)
		valQuery = ") VALUES ($1,"
	}

	var (
		guid     = cast.ToString(data["guid"])
		folderId any
	)

	if helper.IsEmpty(data["guid"]) {
		guid = uuid.NewString()
	}

	if helper.IsEmpty(data["folder_id"]) {
		folderId = nil
	} else {
		folderId = data["folder_id"]
	}

	if !isSystemTable.Bool {
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

	if tableData.IsLoginTable && !cast.ToBool(data["from_auth_service"]) {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributes")
		}

		authInfo = tableAttributes.AuthInfo

		var (
			count         = 0
			loginStrategy = cast.ToStringSlice(authInfo.LoginStrategy)
		)

		if len(authInfo.ClientTypeID) == 0 || len(authInfo.RoleID) == 0 {
			return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
		}

		for _, ls := range loginStrategy {
			if ls == "login" {
				if authInfo.Login == "" || authInfo.Password == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given login password")
				}
			} else if ls == "email" {
				if authInfo.Email == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
				}
			} else if ls == "phone" {
				if authInfo.Phone == "" {
					return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
				}
			}
		}

		query = `SELECT COUNT(*) FROM "client_type" WHERE guid = $1 AND ( table_slug = $2 OR name = 'ADMIN')`

		err = tx.QueryRow(ctx, query, data[config.CLIENT_TYPE_ID], req.TableSlug).Scan(&count)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning count")
		}

		if count != 0 {
			user, err := i.grpcClient.SyncUserService().CreateUser(ctx, &pa.CreateSyncUserRequest{
				Login:         cast.ToString(data[authInfo.Login]),
				Email:         cast.ToString(data[authInfo.Email]),
				Phone:         cast.ToString(data[authInfo.Phone]),
				Password:      cast.ToString(body[authInfo.Password]),
				RoleId:        cast.ToString(data[config.ROLE_ID]),
				ClientTypeId:  cast.ToString(data[config.CLIENT_TYPE_ID]),
				Invite:        cast.ToBool(data["invite"]),
				ProjectId:     cast.ToString(body["company_service_project_id"]),
				EnvironmentId: cast.ToString(body["company_service_environment_id"]),
				LoginStrategy: authInfo.LoginStrategy,
			})
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while creating auth user")
			}

			err = i.UpdateUserIdAuth(ctx, &models.ItemsChangeGuid{
				TableSlug: req.TableSlug,
				ProjectId: req.ProjectId,
				OldId:     guid,
				NewId:     user.UserId,
				Tx:        tx,
			})
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while updating guid")
			}

			err = i.InsertPersonTable(ctx, &models.PersonRequest{
				Tx:           tx,
				Guid:         guid,
				UserIdAuth:   user.UserId,
				Login:        cast.ToString(data[authInfo.Login]),
				Password:     cast.ToString(data[authInfo.Password]),
				Email:        cast.ToString(data[authInfo.Email]),
				Phone:        cast.ToString(data[authInfo.Phone]),
				RoleId:       cast.ToString(data[config.ROLE_ID]),
				ClientTypeId: cast.ToString(data[config.CLIENT_TYPE_ID]),
			})
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while inserting to person")
			}
		}
	}

	if config.PersonTable[tableData.Slug] {
		if data[config.CLIENT_TYPE_ID] == nil || data[config.ROLE_ID] == nil {
			return &nb.CommonMessage{}, errors.New(config.ErrAuthInfo)
		}

		var (
			loginTableSlug sql.NullString
			clinetTypeId   string = data[config.CLIENT_TYPE_ID].(string)
		)

		query = `SELECT table_slug FROM "client_type" WHERE guid = $1`

		if err = tx.QueryRow(ctx, query, clinetTypeId).Scan(&loginTableSlug); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when select client_type for person")
		}

		var (
			tableAttributes, tableId sql.NullString
			tableAttributesMap       = make(map[string]any)
		)

		if !loginTableSlug.Valid {
			loginTableSlug.String = config.USER
		}

		query = `SELECT id, attributes FROM "table" WHERE slug = $1`

		if err = tx.QueryRow(ctx, query, loginTableSlug).Scan(&tableId, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when select table for person")
		}

		if err = json.Unmarshal([]byte(tableAttributes.String), &tableAttributesMap); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when unmarshal login table attributes")
		}

		user, err := i.grpcClient.SyncUserService().CreateUser(ctx, &pa.CreateSyncUserRequest{
			Login:         cast.ToString(data["login"]),
			Email:         cast.ToString(data["email"]),
			Phone:         cast.ToString(data["phone_number"]),
			Invite:        cast.ToBool(data["invite"]),
			Password:      cast.ToString(body["password"]),
			ProjectId:     cast.ToString(body["company_service_project_id"]),
			EnvironmentId: cast.ToString(body["company_service_environment_id"]),
			RoleId:        cast.ToString(data[config.ROLE_ID]),
			ClientTypeId:  cast.ToString(data[config.CLIENT_TYPE_ID]),
			LoginStrategy: []string{"login", "phone", "email"},
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while creating auth user")
		}

		err = i.UpdateUserIdAuth(ctx, &models.ItemsChangeGuid{
			TableSlug: req.TableSlug,
			ProjectId: req.ProjectId,
			OldId:     guid,
			NewId:     user.UserId,
			Tx:        tx,
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while updating guid")
		}

		err = person.CreateSyncWithLoginTable(ctx, models.CreateSyncWithLoginTableRequest{
			Tx:                 tx,
			Guid:               guid,
			UserIdAuth:         user.UserId,
			LoginTableSlug:     loginTableSlug.String,
			Data:               data,
			TableAttributesMap: tableAttributesMap,
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when sync with login table")
		}
	}

	err = helper.AppendMany2Many(ctx, tx, appendMany2Many)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while appending many2many")
	}

	data["guid"] = guid

	newData, err := helper.ConvertMapToStruct(data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	if err = tx.Commit(ctx); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while committing")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		Data:      newData,
	}, nil
}

func (i *itemsRepo) Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.Update")
	defer dbSpan.Finish()

	var (
		conn            = psqlpool.Get(req.GetProjectId())
		argCount        = 2
		args            = []any{}
		attr            = []byte{}
		guid            string
		isLoginTable    bool
		tableAttributes models.TableAttributes
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	data, err := helper.PrepareToUpdateInObjectBuilder(ctx, req, tx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while preparing to update in object builder")
	}

	if _, ok := data["guid"]; !ok {
		data["guid"] = data["id"]
	}

	guid = cast.ToString(data["guid"])

	if authGuid, ok := data["auth_guid"]; ok {
		data["guid"] = authGuid
	}

	args = append(args, guid)

	query := fmt.Sprintf(`UPDATE "%s" SET `, req.TableSlug)

	fieldQuery := `
		SELECT 
			f.slug, 
			f.type, 
			t.attributes, 
			t.is_login_table
		FROM "field" as f 
		JOIN "table" as t 
		ON f.table_id = t.id 
		WHERE t.slug = $1 AND f.slug != 'user_id_auth'`

	fieldRows, err := tx.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var fieldSlug, fieldType string

		if err = fieldRows.Scan(&fieldSlug, &fieldType, &attr, &isLoginTable); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		val, ok := data[fieldSlug]
		switch fieldType {
		case "MULTISELECT":
			switch val.(type) {
			case string:
				val = []string{cast.ToString(val)}
			}
		case "DATE_TIME_WITHOUT_TIME_ZONE":
			switch val.(type) {
			case string:
				val = helper.ConvertTimestamp2DB(cast.ToString(val))
			}
		case "FORMULA_FRONTEND":
			val = cast.ToString(val)
		case "PASSWORD":
			if ok {
				password := cast.ToString(val)
				err = util.ValidStrongPassword(password)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "strong password checker")
				}

				if len(password) != config.BcryptHashPasswordLength {
					hashedPassword, err := security.HashPasswordBcrypt(password)
					if err != nil {
						return &nb.CommonMessage{}, errors.Wrap(err, "error when hash password")
					}
					val = hashedPassword
				}
			}
		}

		if ok {
			query += fmt.Sprintf(`%s=$%d, `, fieldSlug, argCount)
			argCount++
			args = append(args, val)
		}
	}

	if isLoginTable {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributs")
		}

		response, err := helper.GetItem(ctx, conn, req.TableSlug, guid, false)
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
		}

		var (
			count    = 0
			authInfo = tableAttributes.AuthInfo
		)

		if response[authInfo.ClientTypeID] == nil || response[authInfo.RoleID] == nil {
			return &nb.CommonMessage{}, errors.New(config.ErrAuthInfo)
		}

		clientTypeQuery := `SELECT COUNT(*) FROM "client_type" WHERE guid = $1 AND ( table_slug = $2 OR name = 'ADMIN')`

		if err = tx.QueryRow(ctx, clientTypeQuery, response[authInfo.ClientTypeID], req.TableSlug).Scan(&count); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning count")
		}

		if count != 0 {
			var (
				email    = cast.ToString(data[authInfo.Email])
				login    = cast.ToString(data[authInfo.Login])
				phone    = cast.ToString(data[authInfo.Phone])
				password = cast.ToString(data[authInfo.Password])
			)

			updateUserRequest := &pa.UpdateSyncUserRequest{
				Guid:         cast.ToString(response["user_id_auth"]),
				EnvId:        req.EnvId,
				RoleId:       cast.ToString(response[config.ROLE_ID]),
				ProjectId:    req.CompanyProjectId,
				ClientTypeId: cast.ToString(response[config.CLIENT_TYPE_ID]),
			}

			personTableRequest := &models.PersonRequest{
				Tx:           tx,
				Guid:         guid,
				UserIdAuth:   cast.ToString(response["user_id_auth"]),
				RoleId:       cast.ToString(response[config.ROLE_ID]),
				ClientTypeId: cast.ToString(response[config.CLIENT_TYPE_ID]),
				Login:        login,
				Password:     password,
				Phone:        phone,
				Email:        email,
			}

			if len(password) != config.BcryptHashPasswordLength && len(password) != 0 {
				err = util.ValidStrongPassword(password)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "strong password checker")
				}
				updateUserRequest.Password = password
				hashedPassword, err := security.HashPasswordBcrypt(password)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "error while hashing password")
				}

				personTableRequest.Password = hashedPassword
			}

			if email != cast.ToString(response[authInfo.Email]) {
				updateUserRequest.Email = email
			}

			if login != cast.ToString(response[authInfo.Login]) {
				updateUserRequest.Login = login
			}

			if phone != cast.ToString(response[authInfo.Phone]) {
				updateUserRequest.Phone = phone
			}

			user, err := i.grpcClient.SyncUserService().UpdateUser(ctx, updateUserRequest)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while updating user")
			}

			err = i.UpdateUserIdAuth(ctx, &models.ItemsChangeGuid{
				TableSlug: req.TableSlug,
				OldId:     guid,
				NewId:     user.GetUserId(),
				Tx:        tx,
			})
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while updating guid")
			}

			personTableRequest.UserIdAuth = user.GetUserId()

			err = i.UpsertPersonTable(ctx, personTableRequest)
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while updating person")
			}
		} else {
			return &nb.CommonMessage{}, errors.New(config.ErrAuthInfo)
		}
	}

	query = strings.TrimRight(query, ", ")
	query += " WHERE guid = $1"

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing query")
	}

	if config.PersonTable[req.GetTableSlug()] {
		err = person.UpdateSyncWithLoginTable(ctx, i.grpcClient, models.UpdateSyncWithLoginTableRequest{
			Tx:        tx,
			Guid:      guid,
			EnvId:     req.GetEnvId(),
			ProjectId: req.GetCompanyProjectId(),
			Data:      data,
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when sync person with auth and login table")
		}
	}

	output, err := helper.GetItemWithTx(ctx, tx, req.TableSlug, guid, false)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
	}

	response, err := helper.ConvertMapToStruct(output)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	if err = tx.Commit(ctx); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while committing")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      response,
	}, nil
}

func (i *itemsRepo) GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.GetSingle")
	defer dbSpan.Finish()

	var (
		conn     = psqlpool.Get(req.GetProjectId())
		fromAuth bool
	)

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
	}

	id, ok := data["id"].(string)
	if ok {
		fromAuth = false
	} else {
		id = cast.ToString(data["user_id_auth"])
		fromAuth = true
	}

	output, err := helper.GetItem(ctx, conn, req.TableSlug, id, fromAuth)
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
		relationFieldTableIds         = []string{}
		relationFieldTablesMap        = make(map[string]any)
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

		if field.Type == "DATE_TIME_WITHOUT_TIME_ZONE" {
			if val, ok := output[field.Slug]; ok {
				time := cast.ToTime(val)
				output[field.Slug] = time.Format(config.TimeLayoutItems)
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

	response := make(map[string]any)

	response["response"] = output
	response["fields"] = fields

	newBody, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting map to struct")
	}

	if isChanged {
		go func() {
			_, _ = i.Update(ctx, &nb.CommonMessage{
				ProjectId: req.ProjectId,
				TableSlug: req.TableSlug,
				Data:      newBody,
			})
		}()
	}

	// ? SKIP ...

	return &nb.CommonMessage{
		ProjectId: req.ProjectId,
		TableSlug: req.TableSlug,
		Data:      newBody,
	}, err
}

func (i *itemsRepo) Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.Delete")
	defer dbSpan.Finish()

	var (
		conn       = psqlpool.Get(req.GetProjectId())
		table      = models.Table{}
		atr        = []byte{}
		query      string
		attributes models.TableAttributes
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while converting struct to map")
	}

	var id = cast.ToString(data["id"])

	response, err := helper.GetItemWithTx(ctx, tx, req.TableSlug, id, false)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting item")
	}

	query = `SELECT slug, attributes, is_login_table, soft_delete FROM "table" WHERE slug = $1`

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(
		&table.Slug,
		&atr,
		&table.IsLoginTable,
		&table.SoftDelete,
	)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
	}

	if table.SoftDelete {
		query = fmt.Sprintf(`UPDATE "%s" SET deleted_at = CURRENT_TIMESTAMP WHERE guid = $1`, req.TableSlug)
	} else {
		query = fmt.Sprintf(`DELETE FROM "%s" WHERE guid = $1`, req.TableSlug)
	}

	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing")
	}

	if table.IsLoginTable {
		if err := json.Unmarshal(atr, &attributes); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while unmarshalling attributs")
		}

		var (
			authInfo     = attributes.AuthInfo
			count        = 0
			clientTypeId = response[authInfo.ClientTypeID]
			userIdAuth   = cast.ToString(response["user_id_auth"])
		)

		if userIdAuth == "" {
			return &nb.CommonMessage{}, errors.New(config.ErrInvalidUserId)
		}

		if clientTypeId == nil || response[authInfo.RoleID] == nil {
			return &nb.CommonMessage{}, errors.New(config.ErrAuthInfo)
		}

		query = `SELECT COUNT(*) FROM client_type WHERE guid = $1 AND table_slug = $2`

		if err = tx.QueryRow(ctx, query, clientTypeId, req.TableSlug).Scan(&count); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning")
		}

		err = i.DeletePesonTable(ctx, &models.PersonRequest{Tx: tx, Guid: id})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while deleting person")
		}

		if count != 0 {
			_, err = i.grpcClient.SyncUserService().DeleteUser(ctx, &pa.DeleteSyncUserRequest{
				UserId:       userIdAuth,
				ClientTypeId: cast.ToString(clientTypeId),
			})
			if err != nil {
				return &nb.CommonMessage{}, errors.Wrap(err, "error while deleting user from auth service")
			}
		}
	}

	if config.PersonTable[req.GetTableSlug()] {
		err = person.DeleteSyncWithLoginTable(ctx, i.grpcClient, models.DeleteSyncWithLoginTableRequest{
			Tx:       tx,
			Id:       id,
			Response: response,
		})
		if err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "when sync person with auth and login table")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while committing")
	}

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

func (i *itemsRepo) DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *models.DeleteUsers, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.DeleteMany")
	defer dbSpan.Finish()

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return nil, errors.Wrap(err, "error while converting struct to map")
	}

	var (
		ids             = cast.ToStringSlice(data["ids"])
		conn            = psqlpool.Get(req.GetProjectId())
		query           string
		table           models.Table
		tableAttributes models.TableAttributes
		attr            []byte
		users           []*pa.DeleteManyUserRequest_User
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error while beginning transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query = `SELECT slug, attributes, is_login_table, soft_delete FROM "table" WHERE slug = $1`

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(
		&table.Slug,
		&attr,
		&table.IsLoginTable,
		&table.SoftDelete,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error while scanning")
	}

	if !table.IsLoginTable && !config.PersonTable[table.Slug] {
		if table.SoftDelete {
			query = fmt.Sprintf(`UPDATE "%s" SET deleted_at = CURRENT_TIMESTAMP WHERE guid = ANY($1)`, req.TableSlug)
		} else {
			query = fmt.Sprintf(`DELETE FROM "%s" WHERE guid = ANY($1)`, req.TableSlug)
		}

		_, err = tx.Exec(ctx, query, ids)
		if err != nil {
			return nil, errors.Wrap(err, "error while executing")
		}
	}

	if table.IsLoginTable {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return nil, errors.Wrap(err, "error while unmarshalling attributs")
		}

		var authInfo = tableAttributes.AuthInfo

		query = fmt.Sprintf(`
			SELECT 
				user_id_auth, 
				%s, 
				%s 
			FROM "%s" WHERE guid = ANY($1)`, authInfo.ClientTypeID, authInfo.RoleID, req.TableSlug,
		)

		rows, err := tx.Query(ctx, query, ids)
		if err != nil {
			return nil, errors.Wrap(err, "error while querying")
		}

		defer rows.Close()

		for rows.Next() {
			var id, roleId, clientTypeId string

			err = rows.Scan(
				&id,
				&clientTypeId,
				&roleId,
			)
			if err != nil {
				return nil, errors.Wrap(err, "error while scanning")
			}

			users = append(users, &pa.DeleteManyUserRequest_User{
				UserId:       id,
				RoleId:       roleId,
				ClientTypeId: clientTypeId,
			})
		}

		if table.SoftDelete {
			query = fmt.Sprintf(`UPDATE "%s" SET deleted_at = CURRENT_TIMESTAMP WHERE guid = ANY($1)`, req.TableSlug)
		} else {
			query = fmt.Sprintf(`DELETE FROM "%s" WHERE guid = ANY($1)`, req.TableSlug)
		}

		_, err = tx.Exec(ctx, query, ids)
		if err != nil {
			return nil, errors.Wrap(err, "error while executing")
		}

		err = i.DeleteManyPersonTable(ctx, &models.PersonRequest{Tx: tx, Ids: ids})
		if err != nil {
			return nil, errors.Wrap(err, "error while deleting many person")
		}

		_, err = i.grpcClient.SyncUserService().DeleteManyUser(ctx, &pa.DeleteManyUserRequest{
			Users:         users,
			ProjectId:     cast.ToString(data["company_service_project_id"]),
			EnvironmentId: cast.ToString(data["company_service_environment_id"]),
		})
		if err != nil {
			return nil, errors.Wrap(err, "error while deleting users from auth service")
		}
	}

	if config.PersonTable[table.Slug] {
		err = person.DeleteManySyncWithLoginTable(ctx, i.grpcClient, models.DeleteManySyncWithLoginTableRequest{
			Tx:    tx,
			Ids:   ids,
			Table: &table,
			Data:  data,
		})
		if err != nil {
			return nil, errors.Wrap(err, "when sync person with auth and login table")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "error while committing")
	}

	return &models.DeleteUsers{
		Users: users,
	}, nil
}

func (i *itemsRepo) MultipleUpdate(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.MultipleUpdate")
	defer dbSpan.Finish()

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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.UpsertMany")
	defer dbSpan.Finish()

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return errors.Wrap(err, "upsertMany convert req")
	}

	if _, ok := data["field_slug"]; !ok {
		return errors.Wrap(errors.New("field_slug required"), "field_slug required")
	}

	if _, ok := data["fields"]; !ok {
		return errors.Wrap(errors.New("fields required"), "fields required")
	}

	var (
		conn = psqlpool.Get(req.GetProjectId())

		objects    = cast.ToSlice(data["objects"])
		fieldSlug  = cast.ToString(data["field_slug"])
		fieldsReq  = cast.ToStringSlice(data["fields"])
		fieldSlugs = make([]models.Field, 0)

		insertQuery = fmt.Sprintf(`INSERT INTO "%s" (`, req.TableSlug)
		valuesQuery = " ) VALUES "
		updateQuery = fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET ", fieldSlug)
		args        []any
		argCount    = 1
	)

	fieldRows, err := conn.Query(ctx, `
		SELECT f.slug, f.type 
		FROM "field" as f 
		JOIN "table" as t 
		ON f.table_id = t.id 
		WHERE t.slug = $1 
		AND f.slug = ANY($2::text[])`,
		req.TableSlug, fieldsReq)
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

func (i *itemsRepo) UpdateByUserIdAuth(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.Update")
	defer dbSpan.Finish()

	var (
		argCount     = 2
		guid         string
		attr         = []byte{}
		args         = []any{}
		isLoginTable bool
		conn         = psqlpool.Get(req.GetProjectId())
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while beginning transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	data, err := helper.PrepareToUpdateInObjectBuilderFromAuth(ctx, req, tx)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while preparing to update in object builder")
	}

	if _, ok := data["guid"]; !ok {
		data["user_id_auth"] = data["id"]
	}
	guid = cast.ToString(data["id"])

	if authGuid, ok := data["auth_guid"]; ok {
		data["user_id_auth"] = authGuid
	}

	args = append(args, guid)

	query := fmt.Sprintf(`UPDATE "%s" SET `, req.TableSlug)

	fieldQuery := `
		SELECT 
			f.slug, f.type, t.attributes, t.is_login_table
		FROM "field" as f 
		JOIN "table" as t 
		ON f.table_id = t.id 
		WHERE t.slug = $1 AND f.slug != 'user_id_auth'`

	fieldRows, err := tx.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while getting fields")
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		var fieldSlug, fieldType string

		if err = fieldRows.Scan(&fieldSlug, &fieldType, &attr, &isLoginTable); err != nil {
			return &nb.CommonMessage{}, errors.Wrap(err, "error while scanning fields")
		}

		val, ok := data[fieldSlug]
		switch fieldType {
		case "MULTISELECT":
			switch val.(type) {
			case string:
				val = []string{cast.ToString(val)}
			}
		case "DATE_TIME_WITHOUT_TIME_ZONE":
			switch val.(type) {
			case string:
				val = helper.ConvertTimestamp2DB(cast.ToString(val))
			}
		case "FORMULA_FRONTEND":
			val = cast.ToString(val)
		case "PASSWORD":
			if ok {
				password := cast.ToString(val)
				err = util.ValidStrongPassword(password)
				if err != nil {
					return &nb.CommonMessage{}, errors.Wrap(err, "strong password checker")
				}

				if len(password) != config.BcryptHashPasswordLength {
					hashedPassword, err := security.HashPasswordBcrypt(password)
					if err != nil {
						return &nb.CommonMessage{}, errors.Wrap(err, "error when hash password")
					}
					val = hashedPassword
				}
			}
		}

		if ok {
			query += fmt.Sprintf(`%s=$%d, `, fieldSlug, argCount)
			argCount++
			args = append(args, val)
		}
	}

	query = strings.TrimRight(query, ", ")
	query += " WHERE user_id_auth = $1"

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, errors.Wrap(err, "error while executing query")
	}

	output, err := helper.GetItemWithTx(ctx, tx, req.TableSlug, guid, true)
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

func (i *itemsRepo) UpdateUserIdAuth(ctx context.Context, req *models.ItemsChangeGuid) error {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "items.UpdateUserIdAuth")
	defer dbSpan.Finish()

	var query = fmt.Sprintf(`UPDATE "%s" SET user_id_auth = $2 WHERE guid = $1`, req.TableSlug)

	_, err := req.Tx.Exec(ctx, query, req.OldId, req.NewId)
	if err != nil {
		return errors.Wrap(err, "error while executing query")
	}

	return nil
}

func (i *itemsRepo) InsertPersonTable(ctx context.Context, req *models.PersonRequest) error {
	var query = `INSERT INTO "person" (
		guid, 
		login, 
		password, 
		email, 
		phone_number, 
		user_id_auth, 
		client_type_id, 
		role_id,
		full_name,
		image
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := req.Tx.Exec(ctx, query,
		req.Guid,
		req.Login,
		req.Password,
		req.Email,
		req.Phone,
		req.UserIdAuth,
		req.ClientTypeId,
		req.RoleId,
		req.FullName,
		req.Image,
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *itemsRepo) UpsertPersonTable(ctx context.Context, req *models.PersonRequest) error {
	var query = `INSERT INTO "person" (
		guid, 
		login, 
		password, 
		email, 
		phone_number, 
		user_id_auth, 
		client_type_id, 
		role_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (guid) DO UPDATE SET 
		login = EXCLUDED.login,
		password = CASE WHEN EXCLUDED.password != '' THEN EXCLUDED.password ELSE person.password END,
		email = CASE WHEN EXCLUDED.email != '' THEN EXCLUDED.email ELSE person.email END,
		phone_number = CASE WHEN EXCLUDED.phone_number != '' THEN EXCLUDED.phone_number ELSE person.phone_number END,
		user_id_auth = EXCLUDED.user_id_auth,
		client_type_id = EXCLUDED.client_type_id,
		role_id = EXCLUDED.role_id`

	_, err := req.Tx.Exec(ctx, query,
		req.Guid,
		req.Login,
		req.Password,
		req.Email,
		req.Phone,
		req.UserIdAuth,
		req.ClientTypeId,
		req.RoleId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (i *itemsRepo) DeletePesonTable(ctx context.Context, req *models.PersonRequest) error {
	var query = `DELETE FROM "person" WHERE guid = $1`

	_, err := req.Tx.Exec(ctx, query, req.Guid)
	if err != nil {
		return err
	}

	return nil
}

func (i *itemsRepo) DeleteManyPersonTable(ctx context.Context, req *models.PersonRequest) error {
	var query = `DELETE FROM "person" WHERE guid = ANY($1)`

	_, err := req.Tx.Exec(ctx, query, req.Ids)
	if err != nil {
		return err
	}

	return nil
}
