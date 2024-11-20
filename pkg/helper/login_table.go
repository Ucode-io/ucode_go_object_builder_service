package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func UpsertLoginTableField(ctx context.Context, req models.Field) error {
	var (
		tx                 = req.Tx
		fieldId, fieldType string
		query              = `SELECT id, type FROM field where slug = $1 AND table_id = $2`
	)

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, query, req.Slug, req.TableId).Scan(&fieldId, &fieldType)
	if err != nil && err != pgx.ErrNoRows {
		return err
	} else if err == pgx.ErrNoRows {
		fieldId = uuid.NewString()

		query = `INSERT INTO "field" (
			id,
			"table_id",
			"required",
			"slug",
			"label",
			"default",
			"type",
			"attributes",
			"index"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = tx.Exec(ctx, query, fieldId, req.TableId, req.Required, req.Slug, req.Label, req.Default, req.Type, attributes, req.Index)
		if err != nil {
			return err
		}

		var (
			body, data                            []byte
			ids, valueStrings                     []string
			values                                []any
			tableSlug, layoutId, tabId, sectionId string
			sectionCount                          int32
			fields                                = []models.SectionFields{}
		)

		query = `SELECT is_changed_by_host, slug FROM "table" WHERE id = $1`

		err = tx.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
		if err != nil {
			return err
		}

		query = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableSlug, req.Slug, GetDataType(req.Type))

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}

		data, err = ChangeHostname(data)
		if err != nil {
			return err
		}

		query = `UPDATE "table" SET 
			is_changed = true,
			is_changed_by_host = $1
		WHERE id = $2
		`

		_, err = tx.Exec(ctx, query, data, req.TableId)
		if err != nil {
			return err
		}

		query = `SELECT guid FROM "role"`

		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var id string

			err = rows.Scan(&id)
			if err != nil {
				return err
			}

			ids = append(ids, id)
		}

		query = `INSERT INTO "field_permission" (
    		"edit_permission",
    		"view_permission",
    		"table_slug",
    		"field_id",
    		"label",
    		"role_id"
		) VALUES 
		`

		for i, id := range ids {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
				1+i*6, 2+i*6, 3+i*6, 4+i*6, 5+i*6, 6+i*6))

			// Add the corresponding values for each column
			values = append(values, true, true, tableSlug, fieldId, req.Label, id)
		}

		query += strings.Join(valueStrings, ", ")

		_, err = tx.Exec(ctx, query, values...)
		if err != nil {
			return err
		}

		query = `SELECT id FROM "layout" WHERE table_id = $1`
		err = tx.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
		if err != nil && err != pgx.ErrNoRows {
			return err
		} else if err == pgx.ErrNoRows {
			return nil
		}

		query = `SELECT id FROM "tab" WHERE "layout_id" = $1 and type = 'section'`
		err = tx.QueryRow(ctx, query, layoutId).Scan(&tabId)
		if err != nil && err != pgx.ErrNoRows {
			return err
		} else if err == pgx.ErrNoRows {
			return nil
		}

		query = `SELECT id, fields FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
		err = tx.QueryRow(ctx, query, tabId).Scan(&sectionId, &body)
		if err != nil {
			return err
		} else if err == pgx.ErrNoRows {
			return nil
		}

		queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
		err = tx.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
		if err != nil && err != pgx.ErrNoRows {
			return err
		} else if err == pgx.ErrNoRows {
			return nil
		}

		if err := json.Unmarshal(body, &fields); err != nil {
			return err
		}

		if len(fields) < 3 {
			query = `UPDATE "section" SET fields = $2 WHERE id = $1`

			fields = append(fields, models.SectionFields{
				Id:    fieldId,
				Order: len(fields) + 1,
			})

			reqBody, err := json.Marshal(fields)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query, sectionId, reqBody)
			if err != nil {
				return err
			}
		} else {
			query = `INSERT INTO "section" ("order", "column", label, table_id, tab_id, fields) VALUES ($1, $2, $3, $4, $5, $6)`

			sectionId = uuid.NewString()

			fields := []models.SectionFields{{Id: fieldId, Order: 1}}

			reqBody, err := json.Marshal(fields)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query, sectionCount+1, "SINGLE", "Info", req.TableId, tabId, reqBody)
			if err != nil {
				return err
			}
		}

		return nil
	}

	query = `UPDATE "field" SET
		"default" = $2,
		"label" = $3,
		"required" = $4,
		"type" = $5,
		"attributes" = $6,
		"index" = $7
	WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, fieldId, req.Default, req.Label, req.Required, req.Type, attributes, req.Index)
	if err != nil {
		return err
	}

	if req.Type != fieldType {
		query = fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, req.TableSlug, req.Slug)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}

		fieldType := GetDataType(req.Type)

		query = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN %s %s`, req.TableSlug, req.Slug, fieldType)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateLoginTableRelation(ctx context.Context, data *models.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error) {
	var (
		fieldFrom, fieldTo string
		autoFilters        []byte
	)

	resp = &nb.CreateRelationRequest{}

	if len(data.Id) == 0 {
		data.Id = uuid.New().String()
	}

	tx := data.Tx

	roles, err := RolesFind(ctx, RelationHelper{Tx: tx})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find roles")
	}

	switch data.Type {
	case config.MANY2ONE:
		fieldFrom = data.TableTo + "_id"
		fieldTo = "id"
		table, err := TableFindOneTx(ctx, tx, data.TableFrom)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find table_from")
		}

		exists, result, err := CheckRelationFieldExists(ctx, RelationHelper{
			Tx:        tx,
			FieldName: fieldFrom,
			TableID:   table.Id,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to check relation field exists")
		}
		if exists {
			fieldFrom = result
		}

		field, err := UpsertField(ctx, RelationHelper{
			Tx: tx,
			Field: &nb.CreateFieldRequest{
				Id:         data.RelationFieldId,
				TableId:    table.Id,
				Slug:       fieldFrom,
				Label:      "FROM " + data.TableFrom + " TO " + data.TableTo,
				Type:       "LOOKUP",
				RelationId: data.Id,
				Attributes: data.Attributes,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to upsert field")
		}

		layout, err := LayoutFindOne(ctx, RelationHelper{
			Tx:      tx,
			TableID: table.Id,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to find layout")
		}

		if layout != nil {
			tab, err := TabFindOne(ctx, RelationHelper{
				Tx:       tx,
				LayoutID: layout.GetId(),
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to find tab")
			}

			if tab == nil {
				tab, err = TabCreate(ctx, RelationHelper{
					Tx:        tx,
					LayoutID:  layout.GetId(),
					TableSlug: table.GetSlug(),
					Order:     1,
					Label:     "Tab",
					Type:      "section",
				})
				if err != nil {
					return nil, errors.Wrap(err, "failed to create tab")
				}
			}

			sections, err := SectionFind(ctx, RelationHelper{
				Tx:    tx,
				TabID: tab.Id,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to find sections")
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
						Attributes:      data.Attributes,
					},
				}
				err = SectionCreate(ctx, RelationHelper{
					Tx:      tx,
					Order:   len(sections) + 1,
					Fields:  fields,
					TableID: table.Id,
					TabID:   tab.Id,
				})
				if err != nil {
					return nil, errors.Wrap(err, "failed to create section")
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
					fields := []*nb.FieldForSection{}

					fields = append(fields, sections[0].Fields...)

					fields = append(fields, &nb.FieldForSection{
						Id:              fmt.Sprintf("%s#%s", data.TableTo, data.Id),
						Order:           int32(countColumns) + 1,
						FieldName:       "",
						RelationType:    config.MANY2ONE,
						IsVisibleLayout: true,
						ShowLabel:       true,
						Attributes:      data.Attributes,
					},
					)

					err = SectionFindOneAndUpdate(ctx, RelationHelper{
						Tx:        tx,
						SectionID: sections[0].Id,
						Fields:    fields,
					})
					if err != nil {
						return nil, errors.Wrap(err, "failed to find one and update section")
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
							Attributes:      data.Attributes,
						},
					}
					err = SectionCreate(ctx, RelationHelper{
						Tx:      tx,
						Fields:  fields,
						TableID: table.Id,
						TabID:   tab.Id,
						Order:   len(sections) + 1,
					})
					if err != nil {
						return nil, errors.Wrap(err, "failed to create section")
					}
				}
			}
		}

		err = RelationFieldPermission(ctx, RelationHelper{
			Tx:        tx,
			FieldID:   field.Id,
			TableSlug: data.TableFrom,
			Label:     "FROM " + data.TableFrom + " TO " + data.TableTo,
			RoleIDs:   roles,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create relation field permission")
		}
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
			"cascading_tree_field_slug",
			"auto_filters" 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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

	if data.AutoFilters != nil || len(data.AutoFilters) == 0 {
		autoFilters, err = json.Marshal(data.AutoFilters)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal")
		}
	} else {
		autoFilters = []byte(`[{}]`)
	}

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
		autoFilters,
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
		return nil, errors.Wrap(err, "failed to insert relation")
	}

	var tableSlugs []string

	tableTo, err := TableFindOneTx(ctx, tx, data.TableTo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find table_to")
	}
	tableFrom, err := TableFindOneTx(ctx, tx, data.TableFrom)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find table_from")
	}

	viewRequest := &nb.CreateViewRequest{
		Id:         uuid.NewString(),
		Type:       data.ViewType,
		RelationId: resp.Id,
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
		MainField:    "",
		QuickFilters: data.QuickFilters,
		Users:        []string{},
		Name:         "",
		Columns: func() []string {
			if len(data.Columns) == 0 {
				return []string{}
			}
			return data.Columns
		}(),
		MultipleInsert:      data.MultipleInsert,
		IsEditable:          data.IsEditable,
		RelationTableSlug:   data.RelationFieldSlug,
		MultipleInsertField: data.MultipleInsertField,
		UpdatedFields: func() []string {
			if len(data.UpdatedFields) == 0 {
				return []string{}
			}
			return data.UpdatedFields
		}(),
		DefaultLimit:    data.DefaultLimit,
		DefaultEditable: data.DefaultEditable,
	}

	err = ViewCreate(ctx, RelationHelper{
		Tx:   tx,
		View: viewRequest,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create view")
	}

	tableSlugs = append(tableSlugs, data.TableTo)

	layout, err := LayoutFindOne(ctx, RelationHelper{
		Tx:      tx,
		TableID: tableTo.Id,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find layout")
	}

	if layout != nil {
		tabs, err := TabFind(ctx, RelationHelper{
			Tx:       tx,
			LayoutID: layout.Id,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to find tabs")
		}

		var label string
		if tableFrom != nil && tableFrom.Label != "" {
			label = tableFrom.Label
		} else {
			label = "Relation tab" + data.TableFrom
		}

		_, err = TabCreate(ctx, RelationHelper{
			Tx:         tx,
			Order:      len(tabs) + 1,
			Label:      label,
			Type:       "relation",
			LayoutID:   layout.Id,
			RelationID: resp.Id,
			Attributes: data.Attributes,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tab")
		}

		err = ViewRelationPermission(ctx, RelationHelper{
			Tx:         tx,
			TableSlug:  tableTo.Slug,
			RelationID: resp.Id,
			RoleIDs:    roles,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create view relation permission")
		}
	}

	err = TableUpdateMany(ctx, tx, tableSlugs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update many tables")
	}

	err = ExecRelation(ctx, RelationHelper{
		Tx:           tx,
		TableFrom:    data.TableFrom,
		TableTo:      data.TableTo,
		FieldFrom:    fieldFrom,
		FieldTo:      fieldTo,
		RelationType: data.Type,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec relation")
	}

	resp.Attributes = data.Attributes

	return resp, nil
}
