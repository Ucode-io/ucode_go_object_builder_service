package postgres

import (
	"context"
	"fmt"
	"log"
	"testing"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestCreateFunction(t *testing.T) {

	log.Println("RUNNING .......................")

	var (
		cfg = config.Config{
			PostgresUser:           "company_service",
			PostgresPassword:       "uVah9foo",
			PostgresHost:           "postgresql01.u-code.io",
			PostgresPort:           30032,
			PostgresDatabase:       "test",
			PostgresMaxConnections: 10,
		}

		log2 = logger.NewLogger("111", "debug")
	)

	storage, err := NewPostgres(context.Background(), cfg, nil, log2)
	assert.NoError(t, err)

	dbURL := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=disable",
		"company_service",
		"uVah9foo",
		"postgresql01.u-code.io",
		30032,
		"test",
	)

	config, err := pgxpool.ParseConfig(dbURL)
	assert.NoError(t, err)

	config.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	assert.NoError(t, err)

	// ––––––––––––––––– Adding to pool ––––––––––––––––––
	psqlpool.Add(projectId, &psqlpool.Pool{Db: pool})

	// ––––––––––––––- Create Table ––––––––––––––-–
	tableId, err := CreateTable(pool, storage)
	assert.NoError(t, err)

	// ––––––––––––––- Create Field ––––––––––––––-–
	err = CreateField(pool)
	assert.NoError(t, err)

	// ––––––––––––––––– Creat Items –––––––––––––––––
	id, userIdAuth, err := CreateItem(storage)
	assert.NoError(t, err)

	// ––––––––––––––––– Update Items –––––––––––––––––
	err = UpdateItem(storage, id)
	assert.NoError(t, err)

	// ––––––––––––––––– Update Items –––––––––––––––––
	err = GetSingleItem(storage, id)
	assert.NoError(t, err)

	// ––––––––––––––––––– Multiple Update ––––––––––––––
	id2, err := MultipleUpdate(storage, id)
	assert.NoError(t, err)

	// ––––––––––––––––– Upsert Many –––––––––––––––––
	id3, err := UpsertMany(storage, id2)
	assert.NoError(t, err)

	// ––––––––––––––––––– Update By UserId Auth ––––––––––––––––
	err = UpdateByUserIdAuth(storage, userIdAuth)
	assert.NoError(t, err)

	// –––––––––––––––––– Update UserId Auth ––––––––––––––––
	err = UpdateUserIdAuth(storage, id2)
	assert.NoError(t, err)

	// ––––––––––––––––– Delete Many –––––––––––––––––
	err = DeleteMany(storage, id, id3)
	assert.NoError(t, err)

	// ––––––––––––––––– Delete Items –––––––––––––––––
	err = DeleteItem(storage, id2)
	assert.NoError(t, err)

	// –––––––––––––––– DELETE TALE –––––––––––––––
	err = DeleteTable(storage, tableId)
	assert.NoError(t, err)

	log.Println("DONE")
}
func CreateTable(conn *pgxpool.Pool, storage storage.StorageI) (string, error) {
	request := &nb.CreateTableRequest{
		Label:     tableLabel,
		Slug:      tableSlug,
		ProjectId: projectId,
	}

	resp, err := storage.Table().Create(context.Background(), request)
	if err != nil {
		return "", err
	}

	return resp.Id, nil
}

func DeleteAllTable(conn *pgxpool.Pool) error {
	var (
		ctx     = context.Background()
		tableId string

		query = `SELECT id FROM "table" WHERE slug = $1`
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	err = tx.QueryRow(ctx, query, tableSlug).Scan(&tableId)
	if err != nil {
		return errors.Wrap(err, "failed to get table by slug")
	}

	query = `DELETE FROM "field" WHERE table_id = $1`
	_, err = tx.Exec(ctx, query, tableId)
	if err != nil {
		return errors.Wrap(err, "failed to delete field")
	}

	query = `DELETE FROM "table" WHERE id = $1`
	_, err = tx.Exec(ctx, query, tableId)
	if err != nil {
		return errors.Wrap(err, "failed to delete table")
	}

	query = `DROP TABLE IF EXISTS ` + tableSlug
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	tx.Commit(ctx)

	return nil

}

func DeleteTable(storage storage.StorageI, id string) error {
	return storage.Table().Delete(context.Background(), &nb.TablePrimaryKey{
		Id:        id,
		ProjectId: projectId,
	})
}

func CreateField(conn *pgxpool.Pool) error {

	var (
		ctx     = context.Background()
		tableId string

		query = `SELECT id FROM "table" WHERE slug = $1`
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	err = tx.QueryRow(ctx, query, tableSlug).Scan(&tableId)
	if err != nil {
		return errors.Wrap(err, "failed to get table by slug")
	}

	for _, field := range tableFields {
		query = `INSERT INTO "field" (
			id,
		"table_id",
		"required",
		"slug",
		"label",
		"type",
		"index",
		"attributes",
		"is_visible",
		"unique",
		"automatic",
		"is_search"
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err = tx.Exec(ctx, query,
			uuid.New().String(),
			tableId,
			false,
			field["slug"],
			field["label"],
			field["type"],
			field["index"],
			field["attributes"],
			field["is_visible"],
			field["unique"],
			field["automatic"],
			field["is_search"],
		)
		if err != nil {
			return errors.Wrap(err, "failed to insert field")
		}

		query = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" VARCHAR`, tableSlug, field["slug"])
		_, err = tx.Exec(ctx, query)
		if err != nil {
			return errors.Wrap(err, "failed to insert field")
		}
	}

	tx.Commit(ctx)
	return nil
}

func CreateItem(storage storage.StorageI) (string, string, error) {

	var (
		id         = uuid.NewString()
		userIdAuth = uuid.NewString()

		dataMap = map[string]any{
			"from_auth_service": false,
			"guid":              id,
			"first_name":        "Fazliddin",
			"last_name":         "Xayrullaev",
			"user_id_auth":      userIdAuth,
		}
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

	_, err = storage.Items().Create(context.Background(), request)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create item")
	}

	return id, userIdAuth, nil
}

func UpdateItem(storage storage.StorageI, itemId string) error {
	dataMap := map[string]any{
		"from_auth_service": false,
		"guid":              itemId,
		"first_name":        "Kimdur",
		"last_name":         "Kimdurov",
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

	_, err = storage.Items().Update(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to create item")
	}

	return nil
}

func GetSingleItem(storage storage.StorageI, itemId string) error {
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

	_, err = storage.Items().GetSingle(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to get item")
	}

	return nil
}

func DeleteItem(storage storage.StorageI, itemId string) error {
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

	_, err = storage.Items().Delete(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed to get item")
	}

	return nil
}

func DeleteMany(storage storage.StorageI, id, id2 string) error {
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

	_, err = storage.Items().DeleteMany(context.Background(), request)
	if err != nil {
		return errors.Wrap(err, "failed DeleteMany")
	}

	return nil
}

func MultipleUpdate(storage storage.StorageI, existingId string) (string, error) {
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

	_, err = storage.Items().MultipleUpdate(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "failed MultipleUpdate")
	}

	return newId, nil
}

func UpsertMany(storage storage.StorageI, existingId string) (string, error) {
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

	err = storage.Items().UpsertMany(context.Background(), request)
	if err != nil {
		return "", errors.Wrap(err, "failed UpsertMany")
	}

	return newId, nil
}

func UpdateUserIdAuth(storage storage.StorageI, id string) error {
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

	if err := storage.Items().UpdateUserIdAuth(ctx, req); err != nil {
		return errors.Wrap(err, "UpdateUserIdAuth returned error in dependent test")
	}

	tx.Commit(ctx)

	return nil
}

func UpdateByUserIdAuth(storage storage.StorageI, userAuth string) error {
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

	_, err = storage.Items().UpdateByUserIdAuth(ctx, req)
	if err != nil {
		return errors.Wrap(err, "UpdateByUserIdAuth returned error in dependent test")
	}

	return nil
}
