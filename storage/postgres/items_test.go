package postgres_test

import (
	"context"
	"log"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	tableSlug  = "test_table"
	tableLabel = "Test table"
	projectId  = "a95e3cba-bfc0-46c8-aeaa-ef75ee9986ed"
)

var tableFields = []map[string]any{
	{
		"slug":       "user_id_auth",
		"required":   false,
		"label":      "User ID Auth",
		"type":       "string",
		"index":      "btree",
		"attributes": "{}",
		"is_visible": true,
		"unique":     false,
		"automatic":  false,
		"is_search":  false,
	},
	{
		"slug":       "first_name",
		"required":   true,
		"label":      "First Name",
		"type":       "string",
		"index":      "btree",
		"attributes": "{}",
		"is_visible": true,
		"unique":     false,
		"automatic":  false,
		"is_search":  true,
	},
	{
		"slug":       "last_name",
		"required":   true,
		"label":      "Last Name",
		"type":       "string",
		"index":      "btree",
		"attributes": "{}",
		"is_visible": true,
		"unique":     false,
		"automatic":  false,
		"is_search":  true,
	},
}

func TestItems(t *testing.T) {

	// log.Println("RUNNING .......................")

	// var (
	// 	cfg = config.Config{
	// 		PostgresUser:           "company_service",
	// 		PostgresPassword:       "uVah9foo",
	// 		PostgresHost:           "postgresql01.u-code.io",
	// 		PostgresPort:           30032,
	// 		PostgresDatabase:       "test",
	// 		PostgresMaxConnections: 10,
	// 	}

	// 	log2 = logger.NewLogger("111", "debug")
	// )

	// storage, err := NewPostgres(context.Background(), cfg, nil, log2)
	// assert.NoError(t, err)

	// dbURL := fmt.Sprintf(
	// 	"postgres://%v:%v@%v:%v/%v?sslmode=disable",
	// 	"company_service",
	// 	"uVah9foo",
	// 	"postgresql01.u-code.io",
	// 	30032,
	// 	"test",
	// )

	// config, err := pgxpool.ParseConfig(dbURL)
	// assert.NoError(t, err)

	// config.MaxConns = 10

	// pool, err := pgxpool.NewWithConfig(context.Background(), config)
	// assert.NoError(t, err)

	// ––––––––––––––––––– Adding to pool –––––––––––––––––––––––
	// psqlpool.Add(projectId, &psqlpool.Pool{Db: pool})

	// ––––––––––––––-–––– Create Table ––––––––––––––-––––––––––
	tableId, err := CreateTable()
	assert.NoError(t, err)

	// ––––––––––––––-–––– Create Field ––––––––––––––-––––––––––
	err = CreateField()
	assert.NoError(t, err)

	// ––––––––––––––––––– Creat Items ––––––––––––––––––––––––––
	id, userIdAuth, err := CreateItem(nil)
	assert.NoError(t, err)

	// ––––––––––––––––––– Update Items –––––––––––––––––––––––––
	err = UpdateItem(t, id)
	assert.NoError(t, err)

	// ––––––––––––––––––– Update Items –––––––––––––––––––––––––
	err = GetSingleItem(id)
	assert.NoError(t, err)

	// ––––––––––––––––––– Multiple Update ––––––––––––––––––––––
	id2, err := MultipleUpdate(id)
	assert.NoError(t, err)

	// ––––––––––––––––––– Upsert Many ––––––––––––––––––––––––––
	id3, err := UpsertMany(id2)
	assert.NoError(t, err)

	// ––––––––––––––––––– Update By UserId Auth ––––––––––––––––
	err = UpdateByUserIdAuth(userIdAuth)
	assert.NoError(t, err)

	// ––––––––––––––––––– Update UserId Auth –––––––––––––––––––
	err = UpdateUserIdAuth(id2)
	assert.NoError(t, err)

	// ––––––––––––––––––– Delete Many ––––––––––––––––––––––––––
	err = DeleteMany(id, id3)
	assert.NoError(t, err)

	// ––––––––––––––––––– Delete Items –––––––––––––––––––––––––
	err = DeleteItem(id2)
	assert.NoError(t, err)

	// ––––––––––––––––––– DELETE TALE ––––––––––––––––––––––––––
	err = DeleteTable(tableId)
	assert.NoError(t, err)

	log.Println("DONE")
}
func CreateTable() (string, error) {
	request := &nb.CreateTableRequest{
		Label:     tableLabel,
		Slug:      tableSlug,
		ProjectId: projectId,
	}

	resp, err := strg.Table().Create(context.Background(), request)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func DeleteTable(id string) error {
	return strg.Table().Delete(context.Background(), &nb.TablePrimaryKey{
		Id:        id,
		ProjectId: projectId,
	})
}

func CreateField() error {

	// var (
	// 	ctx     = context.Background()
	// 	tableId string

	// 	query = `SELECT id FROM "table" WHERE slug = $1`
	// )

	// tx, err := conn.Begin(ctx)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to begin transaction")
	// }

	// err = tx.QueryRow(ctx, query, tableSlug).Scan(&tableId)
	// if err != nil {
	// 	return errors.Wrap(err, "failed to get table by slug")
	// }

	// for _, field := range tableFields {
	// 	query = `INSERT INTO "field" (
	// 		id,
	// 	"table_id",
	// 	"required",
	// 	"slug",
	// 	"label",
	// 	"type",
	// 	"index",
	// 	"attributes",
	// 	"is_visible",
	// 	"unique",
	// 	"automatic",
	// 	"is_search"
	// 		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	// 	_, err = tx.Exec(ctx, query,
	// 		uuid.New().String(),
	// 		tableId,
	// 		false,
	// 		field["slug"],
	// 		field["label"],
	// 		field["type"],
	// 		field["index"],
	// 		field["attributes"],
	// 		field["is_visible"],
	// 		field["unique"],
	// 		field["automatic"],
	// 		field["is_search"],
	// 	)
	// 	if err != nil {
	// 		return errors.Wrap(err, "failed to insert field")
	// 	}

	// 	query = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" VARCHAR`, tableSlug, field["slug"])
	// 	_, err = tx.Exec(ctx, query)
	// 	if err != nil {
	// 		return errors.Wrap(err, "failed to insert field")
	// 	}
	// }

	// tx.Commit(ctx)
	return nil
}

func CreateItem(dataMap map[string]any) (string, string, error) {

	var (
		id         = uuid.NewString()
		userIdAuth = uuid.NewString()

		// dataMap = map[string]any{
		// 	"from_auth_service": false,
		// 	"guid":              id,
		// 	"first_name":        "Fazliddin",
		// 	"last_name":         "Xayrullaev",
		// 	"user_id_auth":      userIdAuth,
		// }
	)

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().Create(context.Background(), request)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create item")
	}

	return id, userIdAuth, nil
}

func UpdateItem(t *testing.T, itemId string) error {
	relationId, _, err := CreateItem(nil)
	assert.NoError(t, err)

	dataMap := map[string]any{
		"from_auth_service": false,
		"guid":              itemId,
		"first_name":        fakeData.Person().FirstName(),
		"last_name":         fakeData.Person().LastName(),
		"author_id":         relationId,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().Update(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to create item")
	}

	return nil
}

func GetSingleItem(itemId string) error {
	dataMap := map[string]any{
		"from_auth_service": false,
		"id":                itemId,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().GetSingle(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to get item")
	}

	return nil
}

func DeleteItem(itemId string) error {
	dataMap := map[string]any{
		"from_auth_service": false,
		"id":                itemId,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().Delete(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to get item")
	}

	return nil
}

func DeleteMany(id, id2 string) error {
	var dataMap = map[string]any{
		"from_auth_service": false,
		"ids":               []string{id, id2},
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().DeleteMany(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed DeleteMany")
	}

	return nil
}

func MultipleUpdate(existingId string) (string, error) {
	newId := uuid.NewString()

	objects := []map[string]any{
		{
			"from_auth_service": false,
			"guid":              existingId,
			"first_name":        "UpdatedFirst",
			"last_name":         "UpdatedLast",
			"is_new":            false,
		},
		{
			"from_auth_service": false,
			"guid":              newId,
			"first_name":        "NewFirst",
			"last_name":         "NewLast",
			"is_new":            true,
		},
	}

	dataMap := map[string]any{
		"objects": objects,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().MultipleUpdate(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "failed MultipleUpdate")
	}

	return newId, nil
}

func UpsertMany(existingId string) (string, error) {
	var (
		newId = uuid.NewString()

		objects = []map[string]any{
			{
				"guid":       existingId,
				"first_name": "UpsertedFirst",
				"last_name":  "UpsertedLast",
			},
			{
				"guid":       newId,
				"first_name": "UpsertedFirst2",
				"last_name":  "UpsertedLast2",
			},
		}

		dataMap = map[string]any{
			"field_slug": "guid",
			"fields":     []string{"guid", "first_name", "last_name"},
			"objects":    objects,
		}
	)
	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert map to struct")
	}

	request := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	err = strg.Items().UpsertMany(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "failed UpsertMany")
	}

	return newId, nil
}

func UpdateUserIdAuth(id string) error {
	var (
		ctx = context.Background()

		newAuth = uuid.NewString()
		conn, _ = psqlpool.Get(projectId)
		tx, _   = conn.Begin(ctx)
	)

	req := &models.ItemsChangeGuid{
		Tx:        tx,
		TableSlug: tableSlug,
		OldId:     id,
		NewId:     newAuth,
	}

	if err := strg.Items().UpdateUserIdAuth(ctx, req); err != nil {
		return errors.Wrap(err, "UpdateUserIdAuth returned error in dependent test")
	}

	tx.Commit(ctx)

	return nil
}

func UpdateByUserIdAuth(userAuth string) error {
	ctx := context.Background()

	dataMap := map[string]any{
		"id":         userAuth,
		"first_name": "UpdatedByAuthDependent",
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	if err != nil {
		return errors.Wrap(err, "failed convert map to struct for UpdateByUserIdAuthDependent")
	}

	req := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().UpdateByUserIdAuth(ctx, req)
	if err != nil {
		return errors.Wrap(err, "UpdateByUserIdAuth returned error in dependent test")
	}

	return nil
}
