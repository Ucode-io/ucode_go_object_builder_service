package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (i *itemsRepo) Create(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	var (
		args     = []interface{}{}
		argCount = 1
	)

	data, _, err := helper.PrepareToCreateInObjectBuilder(ctx, conn, req)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	query := fmt.Sprintf(`INSERT INTO %s (guid`, req.TableSlug)

	val, ok := data["guid"]
	if !ok {
		val = uuid.NewString()
	}

	args = append(args, val)

	delete(data, "guid")

	for key, val := range data {
		query += fmt.Sprintf(", %s", key)
		args = append(args, val)
	}

	query += ") VALUES ("

	for i := 0; i < argCount; i++ {
		if i != 0 {
			query += ","
		}
		query += fmt.Sprintf(" $%d", i+1)
	}

	query += ")"

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	// ! Skip AppendMany2Many

	var (
		tableData       = models.Table{}
		FromAuthService bool
		attr            = []byte{}
		tableAttributes = make(map[string]interface{})
	)

	query = `SELECT 
		id,
		slug,
		is_login_table,
		from_auth_service,
		attributes
	FROM "table" WHERE slug = $1
	`

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(
		&tableData.Id,
		&tableData.Slug,
		&tableData.IsLoginTable,
		&FromAuthService,
		&attr,
	)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	if tableData.IsLoginTable && !FromAuthService {
		if err := json.Unmarshal(attr, &tableAttributes); err != nil {
			return &nb.CommonMessage{}, err
		}
		_, ok := tableAttributes["auth_info"]
		if ok {

			count := 0

			authInfo := cast.ToStringMap(tableAttributes["auth_info"])
			if cast.ToString(authInfo["client_type_id"]) != "" ||
				cast.ToString(authInfo["role_id"]) != "" || cast.ToString(authInfo["login"]) != "" ||
				cast.ToString(authInfo["email"]) != "" || cast.ToString(authInfo["phone"]) != "" {
				return &nb.CommonMessage{}, fmt.Errorf("This table is auth table. Auth information not fully given")
			}

			query = `SELECT COUNT(*) FROM "client_type" WHERE guid = $1 AND table_slug = $2`

			err = tx.QueryRow(ctx, query, authInfo["client_type_id"], req.TableSlug).Scan(&count)
			if err != nil {
				return &nb.CommonMessage{}, err
			}
			if count != 0 {
				data["authInfo"] = authInfo
			}
		}
	}

	newData, err := helper.ConvertMapToStruct(data)

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		Data:      newData,
	}, nil
}

func (i *itemsRepo) Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	var (
		data     = make(map[string]interface{})
		args     = []interface{}{}
		argCount = 2
		guid     string
	)

	body, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return &nb.CommonMessage{}, err
	}

	_, ok := data["guid"]
	if !ok {
		data["guid"] = data["id"]
	}
	data["id"] = data["guid"]
	guid = cast.ToString(data["guid"])
	_, ok = data["auth_guid"]
	if ok {
		data["guid"] = data["auth_guid"]
	}

	args = append(args, guid)

	query := fmt.Sprintf(`UPDATE %s SET `, req.TableSlug)

	for key, val := range data {
		query += fmt.Sprintf(`%s=$%d, `, key, argCount)
		argCount++
		args = append(args, val)
	}

	query = strings.TrimRight(query, ", ")

	query += " WHERE guid = $1"

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, nil
	}

	// ! skip append/delete many2many

	return &nb.CommonMessage{}, nil
}

func (i *itemsRepo) GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	// data, err := helper.ConvertStructToMap(req.Data)
	// if err != nil {
	// 	return &nb.CommonMessage{}, err
	// }

	// output, err := helper.GetItem(ctx, conn, req.TableSlug, cast.ToString(data["guid"]))
	// if err != nil {
	// 	return &nb.CommonMessage{}, err
	// }

	query := `SELECT f.id, f.type, f.slug, f.attributes FROM "field" as f JOIN "table" as t ON t.id = f.table_id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	fields := []models.Field{}

	for fieldRows.Next() {
		var (
			field = models.Field{}
			atr   = []byte{}
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&field.Slug,
			&atr,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return &nb.CommonMessage{}, err
		}

		fields = append(fields, field)
	}

	attributeTableFromSlugs := []string{}
	attributeTableFromRelationIds := []string{}

	for _, field := range fields {
		attributes, err := helper.ConvertStructToMap(field.Attributes)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		if field.Type == "FORMULA" {
			if cast.ToString(attributes["table_from"]) != "" && cast.ToString(attributes["sum_field"]) != "" {
				attributeTableFromSlugs = append(attributeTableFromSlugs, strings.Split(cast.ToString(attributes["table_from"]), "#")[0])
				attributeTableFromRelationIds = append(attributeTableFromRelationIds, strings.Split(cast.ToString(attributes["table_from"]), "#")[1])
			}
		}
	}

	return &nb.CommonMessage{}, err
}

func (i *itemsRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)

func (i *itemsRepo) Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)

func (i *itemsRepo) DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
