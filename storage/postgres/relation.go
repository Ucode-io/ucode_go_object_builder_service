package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

type relationRepo struct {
	db *pgxpool.Pool
}

func NewRelationRepo(db *pgxpool.Pool) storage.RelationRepoI {
	return &relationRepo{
		db: db,
	}
}

func (r *relationRepo) Create(ctx context.Context, data *nb.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error) {
	conn := r.db
	defer conn.Close()
	var (
		fieldFrom, fieldTo string
	)

	resp = &nb.CreateRelationRequest{}

	if data.Id == "" {
		data.Id = uuid.New().String()
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	roles, err := helper.RolesFind(ctx, helper.RelationHelper{
		Tx: tx,
	})
	if err != nil {
		return nil, err
	}

	switch data.Type {
	case config.MANY2DYNAMIC:
	case config.MANY2MANY:
	case config.MANY2ONE:
		fieldFrom = data.TableTo + "_id"
		fieldTo = "id"
		table, err := helper.TableFindOne(ctx, conn, data.TableFrom)
		if err != nil {
			return nil, err
		}

		exists, result, err := helper.CheckRelationFieldExists(ctx, helper.RelationHelper{
			Tx:        tx,
			FieldName: fieldFrom,
			TableID:   table.Id,
		})
		if err != nil {
			return nil, err
		}
		if exists {
			fieldFrom = result
		}

		field, err := helper.UpsertField(ctx, helper.RelationHelper{
			Tx: tx,
			Field: &nb.CreateFieldRequest{
				Id:         data.RelationFieldId,
				TableId:    table.Id,
				Slug:       fieldFrom,
				Label:      "FROM " + data.TableFrom + " TO " + data.TableTo,
				Type:       "LOOKUP",
				RelationId: data.Id,
			},
		})
		if err != nil {
			return nil, err
		}

		layout, err := helper.LayoutFindOne(ctx, helper.RelationHelper{
			Tx:      tx,
			TableID: table.Id,
		})
		if err != nil {
			return nil, err
		}

		if layout != nil {
			tab, err := helper.TabFindOne(ctx, helper.RelationHelper{
				Tx:       tx,
				LayoutID: layout.GetId(),
			})
			if err != nil {
				return nil, err
			}

			if tab == nil {
				tab, err := helper.TabCreate(ctx, helper.RelationHelper{
					Tx:        tx,
					LayoutID:  layout.GetId(),
					TableSlug: table.GetSlug(),
					Order:     1,
					Label:     "Tab",
					Type:      "section",
				})
				if err != nil {
					return nil, err
				}

				sections, err := helper.SectionFind(ctx, helper.RelationHelper{
					Tx:    tx,
					TabID: tab.Id,
				})
				if err != nil {
					return nil, err
				}

				if len(sections) == 0 {
					fields := []*nb.FieldForSection{
						{
							Id:              fmt.Sprintf("%s#%s", data.TableTo, data.Id),
							Order:           1,
							FieldName:       "",
							RelationType:    config.MANY2ONE,
							IsVisibleLayout: true,
							ShowLabel:       true,
						},
					}
					err = helper.SectionCreate(ctx, helper.RelationHelper{
						Tx:           tx,
						TabID:        tab.Id,
						SectionOrder: len(sections) + 1,
						TableID:      table.Id,
						Fields:       fields,
					})
					if err != nil {
						return nil, err
					}
				}

				if len(sections) > 0 {
					countColumns := 0
					if len(sections[0].Fields) > 0 {
						countColumns = len(sections[0].Fields)
					}

					sectionColumnCount := 3
					if table.SectionColumnCount != 0 {
						sectionColumnCount = int(table.SectionColumnCount)
					}

					if countColumns < int(sectionColumnCount) {
						fields := []*nb.FieldForSection{
							{
								Id:              fmt.Sprintf("%s#%s", data.TableTo, data.Id),
								Order:           int32(countColumns) + 1,
								FieldName:       "",
								RelationType:    config.MANY2ONE,
								IsVisibleLayout: true,
								ShowLabel:       true,
							},
						}
						err = helper.SectionFindOneAndUpdate(ctx, helper.RelationHelper{
							Tx:        tx,
							SectionID: sections[0].Id,
							Fields:    fields,
						})
						if err != nil {
							return nil, err
						}
					} else {
						fields := []*nb.FieldForSection{
							{
								Id:              fmt.Sprintf("%s#%s", data.TableTo, data.Id),
								Order:           1,
								FieldName:       "",
								RelationType:    config.MANY2ONE,
								IsVisibleLayout: true,
								ShowLabel:       true,
							},
						}
						err = helper.SectionCreate(ctx, helper.RelationHelper{
							Tx:           tx,
							Fields:       fields,
							TableID:      table.Id,
							TabID:        tab.Id,
							SectionOrder: len(sections) + 1,
						})
						if err != nil {
							return nil, err
						}
					}

				}
			}
		}

		err = helper.RelationFieldPermission(ctx, helper.RelationHelper{
			Tx:        tx,
			FieldID:   field.Id,
			TableSlug: data.TableFrom,
			Label:     "FROM " + data.TableFrom + " TO " + data.TableTo,
			RoleIDs:   roles,
		})
		if err != nil {
			return nil, err
		}

	case config.ONE2ONE:
	}

	query := `
		INSERT INTO "relation" (
			"id", 
			"table_from", 
			"table_to", 
			"field_from", 
			"field_to", 
			"type",
			"view_fields", 
			"relation_field_slug", 
			"dynamic_tables", 
			"editable",
			"is_user_id_default", 
			"is_system", 
			"object_id_from_jwt",
			"cascading_tree_table_slug", 
			"cascading_tree_field_slug" 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING 
			"id", 
			"type",
			"relation_field_slug", 
			"dynamic_tables", 
			"editable",
			"is_user_id_default", 
			"object_id_from_jwt",
			"cascading_tree_table_slug", 
			"cascading_tree_field_slug"
	`

	err = tx.QueryRow(ctx, query,
		data.Id,
		data.TableFrom,
		data.TableTo,
		fieldFrom,
		fieldTo,
		data.Type,
		data.ViewFields,
		data.RelationFieldSlug,
		data.DynamicTables,
		data.Editable,
		data.IsUserIdDefault,
		false,
		data.ObjectIdFromJwt,
		data.CascadingTreeTableSlug,
		data.CascadingTreeFieldSlug,
	).Scan(
		&resp.Id,
		&resp.Type,
		&resp.RelationFieldSlug,
		&resp.DynamicTables,
		&resp.Editable,
		&resp.IsUserIdDefault,
		&resp.ObjectIdFromJwt,
		&resp.CascadingTreeTableSlug,
		&resp.CascadingTreeFieldSlug,
	)
	if err != nil {
		return nil, err
	}

	var tableSlugs []string
	if resp.Type == config.MANY2DYNAMIC {

	} else {
		tableTo, err := helper.TableFindOne(ctx, conn, data.TableTo)
		if err != nil {
			return nil, err
		}
		tableFrom, err := helper.TableFindOne(ctx, conn, data.TableFrom)
		if err != nil {
			return nil, err
		}

		viewRequest := &nb.CreateViewRequest{
			Id:         uuid.NewString(),
			Type:       data.ViewType,
			RelationId: resp.Id,
			// Name: data.,
			Attributes: data.Attributes,
			TableSlug:  "",
			GroupFields: func() []string {
				if len(data.GroupFields) == 0 {
					return []string{}
				}
				return data.GroupFields
			}(),
			ViewFields: func() []string {
				if len(data.ViewFields) == 0 {
					return []string{}
				}
				return data.ViewFields
			}(),
			MainField: "",
			// DisableDates: data.DisableDates,
			QuickFilters: data.QuickFilters,
			Users:        []string{},
			Name:         "",
			Columns: func() []string {
				if len(data.Columns) == 0 {
					return []string{}
				}
				return data.Columns
			}(),
			// CalendarFromSlug: data.CalendarFromSlug,
			// CalendarToSlug:   data.CalendarToSlug,
			// TimeInterval: data.TimeInterval,
			MultipleInsert: data.MultipleInsert,
			// StatusFieldSlug: data.StatusFieldSlug,
			IsEditable:          data.IsEditable,
			RelationTableSlug:   data.RelationFieldSlug,
			MultipleInsertField: data.MultipleInsertField,
			UpdatedFields: func() []string {
				if len(data.UpdatedFields) == 0 {
					return []string{}
				}
				return data.UpdatedFields
			}(),
			// AppId: data.AppId,
			// TableLabel: data.TableLabel,
			DefaultLimit:    data.DefaultLimit,
			DefaultEditable: data.DefaultEditable,
			// Order: data.Order,
		}

		err = helper.ViewCreate(ctx, helper.RelationHelper{
			Tx:   tx,
			View: viewRequest,
		})
		if err != nil {
			return nil, err
		}

		tableSlugs = append(tableSlugs, data.TableTo)

		layout, err := helper.LayoutFindOne(ctx, helper.RelationHelper{
			Tx:      tx,
			TableID: tableTo.Id,
		})
		if err != nil {
			return nil, err
		}

		if layout != nil {
			tabs, err := helper.TabFind(ctx, helper.RelationHelper{
				Tx:       tx,
				LayoutID: layout.Id,
			})
			if err != nil {
				return nil, err
			}

			var label string
			if tableFrom != nil && tableFrom.Label != "" {
				label = tableFrom.Label
			} else {
				label = "Relation tab" + data.TableFrom
			}

			_, err = helper.TabCreate(ctx, helper.RelationHelper{
				Tx:         tx,
				Order:      len(tabs) + 1,
				Label:      label,
				Type:       "relation",
				LayoutID:   layout.Id,
				RelationID: resp.Id,
			})
			if err != nil {
				return nil, err
			}

			err = helper.ViewRelationPermission(ctx, helper.RelationHelper{
				Tx:         tx,
				TableSlug:  tableTo.Slug,
				RelationID: resp.Id,
				RoleIDs:    roles,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	err = helper.TableUpdateMany(ctx, tx, tableSlugs)
	if err != nil {
		return nil, err
	}

	err = helper.ExecRelation(ctx, helper.RelationHelper{
		Tx:        tx,
		TableFrom: data.TableFrom,
		TableTo:   data.TableTo,
	})
	if err != nil {
		return nil, err
	}

	resp.Attributes = data.Attributes

	return resp, nil
}

func (r *relationRepo) GetByID(ctx context.Context, data *nb.RelationPrimaryKey) (resp *nb.RelationForGetAll, err error) {
	conn := r.db

	query := `
		SELECT
    		r.id,
    		r.table_from,
    		r.table_to,
    		r.field_from,
    		r.field_to,
    		r.type,
    		r.relation_field_slug,
    		r.dynamic_tables,
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

	var (
		tableFromSlug, tableToSlug string
		dynamicTables              sql.NullString
		viewFields                 []string
	)

	resp = &nb.RelationForGetAll{}

	err = conn.QueryRow(ctx, query, data.Id).Scan(
		&resp.Id,
		&tableFromSlug,
		&tableToSlug,
		&resp.FieldFrom,
		&resp.FieldTo,
		&resp.Type,
		&resp.RelationFieldSlug,
		&dynamicTables,
		&resp.Editable,
		&resp.IsUserIdDefault,
		&resp.IsSystem,
		&resp.ObjectIdFromJwt,
		&resp.CascadingTreeTableSlug,
		&resp.CascadingTreeFieldSlug,
		&viewFields,
	)
	if err != nil {
		return nil, err
	}

	if len(viewFields) > 0 {
		query = `
			SELECT 
				"id",
				"table_id",
				"required",
				"slug",
				"label",
				"default",
				"type",
				"index",
				"is_visible",
				autofill_field,
				autofill_table,
				"unique",
				"automatic",
				relation_id
			FROM "field" WHERE id = ANY($1)`

		rows, err := conn.Query(ctx, query, pq.Array(viewFields))
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				field             = nb.Field{}
				autoFillFieldNull sql.NullString
				autoFillTableNull sql.NullString
				relationIdNull    sql.NullString
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
				&field.IsVisible,
				&autoFillFieldNull,
				&autoFillTableNull,
				&field.Unique,
				&field.Automatic,
				&relationIdNull,
			)
			if err != nil {
				return nil, err
			}

			field.AutofillField = autoFillFieldNull.String
			field.AutofillTable = autoFillTableNull.String
			field.RelationField = relationIdNull.String
			field.Default = defaultStr.String
			field.Index = index.String

			resp.ViewFields = append(resp.ViewFields, &field)
		}
	}

	if dynamicTables.Valid {
		err = json.Unmarshal([]byte(dynamicTables.String), &resp.DynamicTables)
		if err != nil {
			return resp, err
		}
	}

	tableFrom, err := helper.TableFindOne(ctx, conn, tableFromSlug)
	if err != nil {
		return nil, err
	}
	tableTo, err := helper.TableFindOne(ctx, conn, tableToSlug)
	if err != nil {
		return nil, err
	}

	resp.TableFrom = tableFrom
	resp.TableTo = tableTo

	view, err := helper.ViewFindOne(ctx, helper.RelationHelper{
		Conn:       conn,
		RelationID: resp.Id,
	})
	if err != nil {
		return nil, err
	}

	if view != nil {
		resp.Title = view.Name
		resp.Columns = view.Columns
		resp.QuickFilters = view.QuickFilters
		resp.GroupFields = view.GroupFields
		resp.IsEditable = view.IsEditable
		resp.RelationTableSlug = view.RelationTableSlug
		resp.ViewType = view.Type
		resp.Id = view.RelationId
		resp.DefaultValues = view.DefaultValues
		resp.DefaultLimit = view.DefaultLimit
		resp.MultipleInsert = view.MultipleInsert
		resp.MultipleInsertField = view.MultipleInsertField
		resp.UpdatedFields = view.UpdatedFields
		resp.Creatable = view.Creatable
		resp.DefaultEditable = view.DefaultEditable
		resp.FunctionPath = view.FunctionPath
		resp.Attributes = view.Attributes
	}

	return resp, nil
}

func (r *relationRepo) GetList(ctx context.Context, data *nb.GetAllRelationsRequest) (resp *nb.GetAllRelationsResponse, err error) {
	conn := r.db

	if data.TableSlug == "" {
		table, err := helper.TableFindOne(ctx, conn, data.TableId)
		if err != nil {
			return nil, err
		}
		data.TableSlug = table.Slug
	}

	var (
		tableFromSlug, tableToSlug string
		relations                  []*nb.RelationForGetAll
	)

	resp = &nb.GetAllRelationsResponse{}

	params := make(map[string]interface{})
	params["table_slug"] = data.TableSlug

	query := `
		SELECT
    		r.id,
    		r.table_from,
    		r.table_to,
    		r.field_from,
    		r.field_to,
    		r.type,
    		r.relation_field_slug,
    		r.dynamic_tables,
    		r.editable,
    		r.is_user_id_default,
    		r.is_system,
    		r.object_id_from_jwt,
    		r.cascading_tree_table_slug,
    		r.cascading_tree_field_slug,
    		jsonb_agg(field.*) AS view_fields
		FROM
		    relation r
		INNER JOIN
		    field ON r.id = field.relation_id
		WHERE  r.table_from = :table_slug OR r.table_to = :table_slug
			OR r.dynamic_tables->>'table_slug' = :table_slug
		GROUP BY r.id `

	query += ` ORDER BY r.created_at DESC `

	if data.Limit > 0 {
		query += ` LIMIT :limit `
		params["limit"] = data.Limit
	}

	if data.Offset >= 0 {
		query += ` OFFSET :offset `
		params["offset"] = data.Offset
	}

	query, args := helper.ReplaceQueryParams(query, params)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			viewFields, dynamicTables sql.NullString
		)
		relation := &nb.RelationForGetAll{}

		err := rows.Scan(
			&relation.Id,
			&tableFromSlug,
			&tableToSlug,
			&relation.FieldFrom,
			&relation.FieldTo,
			&relation.Type,
			&relation.RelationFieldSlug,
			&dynamicTables,
			&relation.Editable,
			&relation.IsUserIdDefault,
			&relation.IsSystem,
			&relation.ObjectIdFromJwt,
			&relation.CascadingTreeTableSlug,
			&relation.CascadingTreeFieldSlug,
			&viewFields,
		)
		if err != nil {
			return resp, err
		}

		if viewFields.Valid {
			err = json.Unmarshal([]byte(viewFields.String), &relation.ViewFields)
			if err != nil {
				return resp, err
			}
		}

		if dynamicTables.Valid {
			err = json.Unmarshal([]byte(dynamicTables.String), &relation.DynamicTables)
			if err != nil {
				return resp, err
			}
		}

		relations = append(relations, relation)
	}

	if len(relations) == 0 {
		return resp, nil
	}

	tableFrom, err := helper.TableFindOne(ctx, conn, tableFromSlug)
	if err != nil {
		return resp, err
	}

	for i := 0; i < len(relations); i++ {
		relations[i].TableFrom = tableFrom
		tableTo, err := helper.TableFindOne(ctx, conn, tableToSlug)
		if err != nil {
			return resp, err
		}
		relations[i].TableTo = tableTo

		view, err := helper.ViewFindOne(ctx, helper.RelationHelper{
			Conn:       conn,
			RelationID: relations[i].Id,
		})
		if err != nil {
			return resp, err
		}

		if view != nil {
			relations[i].Title = view.Name
			relations[i].Columns = view.Columns
			relations[i].QuickFilters = view.QuickFilters
			relations[i].GroupFields = view.GroupFields
			relations[i].IsEditable = view.IsEditable
			relations[i].RelationTableSlug = view.RelationTableSlug
			relations[i].ViewType = view.Type
			// relations[i].Summaries = view.Summaries
			relations[i].Id = view.RelationId
			relations[i].DefaultValues = view.DefaultValues
			// relations[i].ActionRelations = view.ActionRelations
			relations[i].DefaultLimit = view.DefaultLimit
			relations[i].MultipleInsert = view.MultipleInsert
			relations[i].MultipleInsertField = view.MultipleInsertField
			relations[i].UpdatedFields = view.UpdatedFields
			relations[i].Creatable = view.Creatable
			relations[i].DefaultEditable = view.DefaultEditable
			relations[i].FunctionPath = view.FunctionPath
			relations[i].Attributes = view.Attributes
		}
	}

	query = `SELECT COUNT(*) FROM "relation" WHERE table_from = $1`

	err = conn.QueryRow(ctx, query, data.TableSlug).Scan(&resp.Count)
	if err != nil {
		return resp, err
	}

	resp.Relations = relations

	return resp, nil
}

func (r *relationRepo) Update(ctx context.Context, data *nb.UpdateRelationRequest) (resp *nb.RelationForGetAll, err error) {
	return resp, err
}

func (r *relationRepo) Delete(ctx context.Context, data *nb.RelationPrimaryKey) (err error) {
	conn := r.db

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	query := `
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
    		r.cascading_tree_field_slug
		FROM
		    relation r
		WHERE  r.id = $1`

	var (
		tableFromSlug, tableToSlug string
		columns                    []string
		isExists                   bool
	)

	relation := &nb.RelationForGetAll{}

	err = tx.QueryRow(ctx, query, data.Id).Scan(
		&relation.Id,
		&tableFromSlug,
		&tableToSlug,
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
	)
	if err != nil {
		return errors.Wrap(err, "relation not found")
	}

	fmt.Printf("RELATION: %+v\n", relation)

	if relation.IsSystem {
		return errors.New("system relations cannot be deleted")
	}

	field, err := helper.FieldFindOne(ctx, helper.RelationHelper{
		Tx:         tx,
		RelationID: data.Id,
	})
	if err != nil {
		return errors.Wrap(err, "failed to find field")
	}

	if field == nil {
		return errors.New("field not found")
	}

	// if relation.Type == config.ONE2MANY {
	// } else if relation.Type == config.MANY2MANY {
	// } else if relation.Type == config.RECURSIVE {
	// } else {}

	viewDeleteQuery := `DELETE FROM view WHERE relation_id = $1`
	_, err = tx.Exec(ctx, viewDeleteQuery, data.Id)
	if err != nil {
		return errors.Wrap(err, "failed to delete views")
	}

	existsColumnView, err := helper.ViewFindOneByTableSlug(ctx, helper.RelationHelper{
		Tx:        tx,
		TableSlug: tableFromSlug,
	})
	if err != nil {
		return errors.Wrap(err, "failed to find column view")
	}

	if existsColumnView != nil && len(existsColumnView.Columns) > 0 {
		for _, id := range existsColumnView.Columns {
			if id == field.Id {
				isExists = true
				continue
			} else if id != "" {
				columns = append(columns, id)
			}
		}

		if isExists {
			viewFindOneAndUpdate := `
				UPDATE view SET 
					columns = $1
				WHERE id = $2`
			_, err = tx.Exec(ctx, viewFindOneAndUpdate, columns, existsColumnView.Id)
			if err != nil {
				return errors.Wrap(err, "failed to update view")
			}
		}
	}

	//table updatemany is_changed_by_host

	query = `DELETE FROM relation WHERE id = $1`
	rows, err := tx.Exec(ctx, query, data.Id)
	if err != nil {
		return errors.Wrap(err, "failed to delete relation")
	}

	if rows.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}

	fmt.Println("RELATION DELETED", rows.RowsAffected())

	err = helper.TabDeleteMany(ctx, helper.RelationHelper{
		Tx:         tx,
		RelationID: data.Id,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete tabs")
	}

	return nil
}

func (r *relationRepo) GetSingleViewForRelation(ctx context.Context, req models.ReqForViewRelation) (resp models.RelationForView, err error) {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return resp, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	var tableId string

	table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Slug: req.TableSlug})
	if err != nil {
		return resp, err
	}
	if table != nil {
		cast.ToString(table["id"])
		table["id"] = tableId
	}

	fieldResp := &nb.Field{}

	query := `
    SELECT 
        r.id AS relation_id,
        r.table_from AS table_from,
        r.table_to AS  table_to,
        r.field_from AS field_from,
        r.field_to AS field_to ,
        r.type AS type ,
        r.view_fields AS view_fields ,
        r.relation_field_slug AS relation_field_slug ,
        r.editable AS editable ,
        r.is_user_id_default AS is_user_id_default ,
        r.cascading_tree_table_slug AS cascading_tree_table_slug ,
        r.cascading_tree_field_slug AS cascading_tree_field_slug ,

        f.table_id AS table_id ,
        f.required AS required ,
        f.slug AS slug ,
        f.label AS label ,
        f.default AS default ,
        f.type AS type ,
        f.index AS index ,
        f.is_visible AS is_visible ,
        f.is_search AS is_search ,
        f.autofill_field AS autofill_field ,
        f.autofill_table AS autofill_table ,
        f.relation_id AS relation_id ,
        f.unique AS unique ,
        f.automatic AS automatic ,
        f.enable_multilanguage AS enable_multilanguage 
    FROM 
        relation r
    LEFT JOIN 
        field f ON f.relation_id = r.id
    WHERE 
        r.id = $1
    ORDER BY 
        r.created_at DESC
    LIMIT 1
`

	rows, err := conn.Query(ctx, query, req.Id)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	var (
		defaultNull         sql.NullString
		index               sql.NullString
		AutofillField       sql.NullString
		AutofillTable       sql.NullString
		RelationId          sql.NullString
		Unique              sql.NullBool
		Automatic           sql.NullBool
		EnableMultilanguage sql.NullBool
	)
	for rows.Next() {
		resp = models.RelationForView{}
		if err := rows.Scan(
			&resp.Id,
			&resp.TableFrom,
			&resp.TableTo,
			&resp.FieldFrom,
			&resp.FieldTo,
			&resp.Type,
			&resp.ViewFields,
			&resp.RelationFieldSlug,
			&resp.Editable,
			&resp.IsUserIdDefault,
			// &resp.ObjectIdFromJwt,
			&resp.CascadingTreeTableSlug,
			&resp.CascadingTreeFieldSlug,

			&fieldResp.TableId,
			&fieldResp.Required,
			&fieldResp.Slug,
			&fieldResp.Label,
			&defaultNull,
			&fieldResp.Type,
			&index,
			&fieldResp.IsVisible,
			&fieldResp.IsSearch,
			&AutofillField,
			&AutofillTable,
			&RelationId,
			&Unique,
			&Automatic,
			&EnableMultilanguage,
		); err != nil {
			return resp, err
		}

		if defaultNull.Valid {
			fieldResp.Default = defaultNull.String
		}
		if index.Valid {
			fieldResp.Index = index.String
		}

		if AutofillField.Valid {
			fieldResp.AutofillField = AutofillField.String
			// if err := json.Unmarshal([]byte(attributes), &fieldResp.Attributes); err != nil {
			// 	return resp, err
			// }
			if AutofillTable.Valid {
				fieldResp.AutofillTable = AutofillTable.String
			}
			if RelationId.Valid {
				fieldResp.RelationId = RelationId.String
			}
			if Unique.Valid {
				fieldResp.Unique = Unique.Bool
			}
			if Automatic.Valid {
				fieldResp.Automatic = Automatic.Bool
			}
			if EnableMultilanguage.Valid {

				fieldResp.EnableMultilanguage = EnableMultilanguage.Bool
			}

		}

		if resp.Id == "" {
			return resp, errors.New("no data found")
		}
		if err := rows.Err(); err != nil {
			return resp, err
		}

		tableFrom, err := helper.TableVer(ctx, helper.TableVerReq{
			Conn: conn,
			Slug: req.TableSlug,
		})
		if err != nil {
			return resp, err
		}
		if resp.Type == config.MANY2DYNAMIC {
			for _, dynamicTable := range resp.DynamicTables {

				if dynamicTable.TableSlug == req.TableSlug || cast.ToString(table["slug"]) == req.TableSlug {
					tableTo, err := helper.TableVer(ctx, helper.TableVerReq{
						Conn: conn,
						Slug: dynamicTable.TableSlug,
					})
					if err != nil {
						return resp, err
					}
					view := &nb.View{}
					err = conn.QueryRow(ctx, `
					SELECT * FROM view
					WHERE relation_id = $1
				`, req.Id).Scan(
						&view.RelationId,
						&view.TableSlug,
					)
					if err != nil {
						return resp, err
					}

					viewFieldsInDynamicTable := []nb.Field{}
					for _, fieldID := range dynamicTable.ViewFields {

						query := `SELECT id, slug, table_id , attributes,  FROM field WHERE id = $1`
						view_field := &nb.Field{}
						err = conn.QueryRow(ctx, query, fieldID).Scan(
							&view_field.Id,
							&view_field.Slug,
							&view_field.TableId,
							&view_field.Attributes,
						)
						if err != nil {
							return resp, err
						}

						viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, nb.Field{
							Id:         view_field.Id,
							Slug:       view_field.Slug,
							TableId:    view_field.TableId,
							Attributes: view_field.Attributes,
						})

						responseRelation := map[string]interface{}{
							"id":                        resp.Id,
							"table_from":                tableFrom,
							"table_to":                  tableTo,
							"type":                      resp.Type,
							"view_fields":               viewFieldsInDynamicTable,
							"editable":                  resp.Editable,
							"dynamic_tables":            resp.DynamicTables,
							"relation_field_slug":       resp.RelationFieldSlug,
							"auto_filters":              resp.AutoFilters,
							"is_user_id_default":        resp.IsUserIdDefault,
							"cascadings":                resp.Cascadings,
							"object_id_from_jwt":        resp.ObjectIdFromJwt,
							"cascading_tree_table_slug": resp.CascadingTreeTableSlug,
							"cascading_tree_field_slug": resp.CascadingTreeFieldSlug,
						}

						responseRelation["title"] = view.Name
						responseRelation["columns"] = view.Columns
						responseRelation["quick_filters"] = view.QuickFilters
						responseRelation["group_fields"] = view.GroupFields
						responseRelation["is_editable"] = view.IsEditable
						responseRelation["relation_table_slug"] = view.RelationTableSlug
						responseRelation["view_type"] = view.Type
						responseRelation["summaries"] = view.Summaries
						responseRelation["relation_id"] = view.RelationId
						responseRelation["default_values"] = view.DefaultValues
						responseRelation["action_relations"] = view.ActionRelations
						responseRelation["default_limit"] = view.DefaultLimit
						responseRelation["multiple_insert"] = view.MultipleInsert
						responseRelation["multiple_insert_field"] = view.MultipleInsertField
						responseRelation["updated_fields"] = view.UpdatedFields
						responseRelation["attributes"] = view.Attributes
					}
				}
			}
		}
		tableTo, err := helper.TableVer(ctx, helper.TableVerReq{
			Conn: conn,
			Slug: req.TableSlug,
		})
		if err != nil {
			return resp, err
		}
		query = `
		SELECT relation_id, table_slug FROM view
		WHERE relation_id = $1
	`
		view := &nb.View{}
		err = conn.QueryRow(ctx, query, req.Id).Scan(
			&view.RelationId,
			&view.TableSlug,
		)
		if err != nil {
			return resp, err
		}
		responseRelation := map[string]interface{}{
			"id":                        resp.Id,
			"table_from":                tableFrom,
			"table_to":                  tableTo,
			"type":                      resp.Type,
			"view_fields":               resp.ViewFields,
			"editable":                  resp.Editable,
			"dynamic_tables":            resp.DynamicTables,
			"relation_field_slug":       resp.RelationFieldSlug,
			"auto_filters":              resp.AutoFilters,
			"is_user_id_default":        resp.IsUserIdDefault,
			"cascadings":                resp.Cascadings,
			"object_id_from_jwt":        resp.ObjectIdFromJwt,
			"cascading_tree_table_slug": resp.CascadingTreeTableSlug,
			"cascading_tree_field_slug": resp.CascadingTreeFieldSlug,
		}
		responseRelation["title"] = view.Name
		responseRelation["columns"] = view.Columns
		responseRelation["quick_filters"] = view.QuickFilters
		responseRelation["group_fields"] = view.GroupFields
		responseRelation["is_editable"] = view.IsEditable
		responseRelation["relation_table_slug"] = view.RelationTableSlug
		responseRelation["view_type"] = view.Type
		responseRelation["summaries"] = view.Summaries
		responseRelation["relation_id"] = view.RelationId
		responseRelation["default_values"] = view.DefaultValues
		responseRelation["action_relations"] = view.ActionRelations
		responseRelation["default_limit"] = view.DefaultLimit
		responseRelation["multiple_insert"] = view.MultipleInsert
		responseRelation["multiple_insert_field"] = view.MultipleInsertField
		responseRelation["updated_fields"] = view.UpdatedFields
		responseRelation["attributes"] = view.Attributes
		relationTabWithPermission, err := helper.AddPermissionToTab(ctx, responseRelation, conn, req.RoleId, req.TableSlug, req.ProjectId)
		if err != nil {
			return resp, err
		}

		resp.Id = cast.ToString(relationTabWithPermission["id"])
	}
	return resp, nil

}
