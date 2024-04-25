package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
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
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	var (
		args     = []interface{}{}
		argCount = 1
	)

	data, _, err := helper.PrepareToCreateInObjectBuilder(ctx, conn, req)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	fieldQuery := `SELECT f.slug FROM "field" as f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	query := fmt.Sprintf(`INSERT INTO %s (guid`, req.TableSlug)

	val, ok := data["guid"]
	if !ok {
		val = uuid.NewString()
	}

	args = append(args, val)

	delete(data, "guid")

	for fieldRows.Next() {
		fieldSlug := ""

		err = fieldRows.Scan(&fieldSlug)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if fieldSlug == "guid" {
			continue
		}

		val, ok := data[fieldSlug]
		if ok {
			query += fmt.Sprintf(", %s", fieldSlug)
			args = append(args, val)
			argCount++
		}
	}

	query += ") VALUES ("

	for i := 0; i < argCount; i++ {
		if i != 0 {
			query += ","
		}
		query += fmt.Sprintf(" $%d", i+1)
	}

	query += ")"

	// tx, err := conn.Begin(ctx)
	// if err != nil {
	// 	return &nb.CommonMessage{}, err
	// }

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, err
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

	err = conn.QueryRow(ctx, query, req.TableSlug).Scan(
		&tableData.Id,
		&tableData.Slug,
		&tableData.IsLoginTable,
		&attr,
	)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	if tableData.IsLoginTable && !cast.ToBool(data["from_auth_service"]) {
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
				return &nb.CommonMessage{}, fmt.Errorf("this table is auth table. Auth information not fully given")
			}

			query = `SELECT COUNT(*) FROM "client_type" WHERE guid = $1 AND table_slug = $2`

			err = conn.QueryRow(ctx, query, authInfo["client_type_id"], req.TableSlug).Scan(&count)
			if err != nil {
				return &nb.CommonMessage{}, err
			}
			if count != 0 {
				data["authInfo"] = authInfo
			}
		}
	} else {
		data["create_user"] = false
	}

	newData, err := helper.ConvertMapToStruct(data)

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		Data:      newData,
	}, nil
}

func (i *itemsRepo) Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var (
		args     = []interface{}{}
		argCount = 2
		guid     string
	)

	data, err := helper.PrepareToUpdateInObjectBuilder(ctx, conn, req)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, ok := data["guid"]
	if !ok {
		data["guid"] = data["id"]
	}
	// data["id"] = data["guid"]
	guid = cast.ToString(data["guid"])
	_, ok = data["auth_guid"]
	if ok {
		data["guid"] = data["auth_guid"]
	}

	args = append(args, guid)

	query := fmt.Sprintf(`UPDATE %s SET `, req.TableSlug)

	fieldQuery := `SELECT f.slug FROM "field" as f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		fieldSlug := ""

		err = fieldRows.Scan(&fieldSlug)
		if err != nil {
			return &nb.CommonMessage{}, err
		}
		val, ok := data[fieldSlug]
		if ok {
			query += fmt.Sprintf(`%s=$%d, `, fieldSlug, argCount)
			argCount++
			args = append(args, val)
		}
	}

	query = strings.TrimRight(query, ", ")

	query += " WHERE guid = $1"

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return &nb.CommonMessage{}, nil
	}

	// ! skip append/delete many2many

	return &nb.CommonMessage{}, nil
}

func (i *itemsRepo) GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	output, err := helper.GetItem(ctx, conn, req.TableSlug, cast.ToString(data["id"]))
	if err != nil {
		return &nb.CommonMessage{}, err
	}

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

	var (
		attributeTableFromSlugs       = []string{}
		attributeTableFromRelationIds = []string{}

		relationFieldTablesMap = make(map[string]interface{})
		relationFieldTableIds  = []string{}
	)

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

	query = `SELECT id, slug FROM "table" WHERE slug IN ($1)`

	tableRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromSlugs))
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer tableRows.Close()

	for tableRows.Next() {
		table := models.Table{}

		err = tableRows.Scan(&table.Id, &table.Slug)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		relationFieldTableIds = append(relationFieldTableIds, table.Id)
		relationFieldTablesMap[table.Slug] = table
	}

	query = `SELECT slug, table_id, relation_id FROM "field" WHERE relation_id IN ($1) AND table_id IN ($2)`

	relationFieldRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds), pq.Array(relationFieldTableIds))
	if err != nil {
		return &nb.CommonMessage{}, err
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
			return &nb.CommonMessage{}, err
		}

		relationFieldsMap[field.RelationId+"_"+field.TableId] = field.Slug
	}

	query = `SELECT id, type, field_from FROM "relation" WHERE id IN ($1)`

	dynamicRows, err := conn.Query(ctx, query, pq.Array(attributeTableFromRelationIds))
	if err != nil {
		return &nb.CommonMessage{}, err
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

		dynamicRelationsMap[relation.Id] = relation
	}

	isChanged := false

	for _, field := range fields {

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
					return &nb.CommonMessage{}, err
				}
				if output[field.Slug] != resultFormula {
					isChanged = true
				}
				output[field.Slug] = resultFormula
			}
		}
	}

	newBody, err := helper.ConvertMapToStruct(output)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	if isChanged {

		go i.Update(ctx, &nb.CommonMessage{
			ProjectId: req.ProjectId,
			TableSlug: req.TableSlug,
			Data:      newBody,
		})
	}

	return &nb.CommonMessage{
		ProjectId: req.ProjectId,
		TableSlug: req.TableSlug,
		Data:      newBody,
	}, err
}

func (i *itemsRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	return &nb.CommonMessage{}, nil
}

// func (i *itemsRepo) Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)

// func (i *itemsRepo) DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
