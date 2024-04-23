package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
)

type sectionRepo struct {
	db *pgxpool.Pool
}

func NewSectionRepo(db *pgxpool.Pool) storage.SectionRepoI {
	return &sectionRepo{
		db: db,
	}
}

func (s *sectionRepo) GetViewRelation(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetViewRelationResponse, err error) {

	return &nb.GetViewRelationResponse{}, nil
}

func (s *sectionRepo) GetAll(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetAllSectionsResponse, err error) {

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp = &nb.GetAllSectionsResponse{}

	var tableId string
	if req.TableId == "" {
		err := conn.QueryRow(ctx, `SELECT id, slug FROM "table" WHERE slug = $1", req.TableSlug`).Scan(&tableId, &req.TableSlug)
		if err != nil {
			return nil, err
		}
	}
	fieldRes := []*nb.FieldResponse{}
	var tabID string
	if req.TabId != "" {
		tabID = req.TabId
	} else {
		return nil, errors.New("req.TabId is empty or nil")
	}
	rows, err := conn.Query(ctx, `SELECT 
			section.Id,
			section.label,
			section.order,
			section.column,
			section.icon,
			section_field.id,
			section_field.field_name,
			section_field.relation_type,
			section_field.column,
			section_field.order
		FROM section 
		JOIN section_field ON section.id = section_field.section_id
		WHERE section.tab_id = $1`, tabID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []*nb.Section
	for rows.Next() {

		var section nb.Section
		var sectionField nb.FieldForSection
		var relationType sql.NullString
		var column sql.NullInt32

		err = rows.Scan(
			&section.Id,
			&section.Label,
			&section.Order,
			&section.Column,
			&section.Icon,
			&sectionField.Id,
			&sectionField.FieldName,
			&relationType,
			&column,
			&sectionField.Order,
		)

		if err != nil {
			return nil, err
		}

		sections = append(sections, &section)
		section.Fields = append(section.Fields, &sectionField, &nb.FieldForSection{
			Id:           sectionField.Id,
			Column:       column.Int32,
			Order:        sectionField.Order,
			FieldName:    sectionField.FieldName,
			RelationType: relationType.String,
		},
		)

	}

	var fieldAsAttribute []*nb.Field

	for _, section := range sections {
		for _, fieldReq := range section.Fields {
			field := &nb.FieldResponse{}
			field.IsVisibleLayout = fieldReq.IsVisibleLayout
			if strings.Contains(fieldReq.Id, "#") {
				field.Id = fieldReq.Id
				field.Label = fieldReq.FieldName
				field.Order = fieldReq.Order
				field.RelationType = fieldReq.RelationType
				relationID := strings.Split(fieldReq.Id, "#")[1]
				var fieldResp nb.Field
				err := conn.QueryRow(ctx, "SELECT relation_id, table_id FROM field WHERE relation_id = $1 AND table_id = $2", relationID, tableId).Scan(&fieldResp.RelationId, &fieldResp.TableId)
				if err != nil {
					return nil, err
				}
				if relationID != "" {
					field.Slug = fieldResp.Slug
					field.Required = fieldResp.Required
				}
				var relation nb.RelationForGetAll
				err = conn.QueryRow(ctx, "SELECT id FROM relation WHERE id = $1", relationID).Scan(&relation.Id)
				if err != nil {
					return nil, err
				}
				var viewOfRelation nb.View
				err = conn.QueryRow(ctx, "SELECT id, view_fields FROM view WHERE relation_id = $1", relation.Id).Scan(&viewOfRelation.Id, &viewOfRelation.ViewFields)
				if err != nil {
					return nil, err
				}
				var viewFieldIds []string
				for _, field := range relation.ViewFields {
					viewFieldIds = append(viewFieldIds, field.Id)

					if viewOfRelation.ViewFields != nil && len(viewOfRelation.ViewFields) > 0 {
						viewFieldIds = viewOfRelation.ViewFields
					}

				}

				for _, fieldID := range viewFieldIds {
					var field nb.Field

					err := conn.QueryRow(ctx, `SELECT 
					"id",
					"table_id",
					"required",
					"slug",
					"label",
					"default",
					"type",
					"index",
					"attributes",
					"is_visible",
					"is_system",
					"is_search",
					"autofill_field",
					"autofill_table",
					"relation_id",
					"unique",
					"automatic",
					"enable_multilanguage"  FROM field WHERE id = $1`, fieldID).Scan(&field.Id,
						&field.TableId,
						&field.Required,
						&field.Slug,
						&field.Label,
						&field.Default,
						&field.Type,
						&field.Index,
						&field.Attributes,
						&field.IsVisible,
						&field.IsSystem,
						&field.IsSearch,
						&field.AutofillField,
						&field.AutofillTable,
						&field.RelationId,
						&field.Unique,
						&field.Automatic,
						&field.EnableMultilanguage,
					)
					if err != nil {
						return nil, err
					}
					if field.Id != "" {
						if req.LanguageSetting != "" && field.EnableMultilanguage {
							if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
								fieldAsAttribute = append(fieldAsAttribute, &field)
							} else {
								continue
							}
						} else {
							fieldAsAttribute = append(fieldAsAttribute, &field)
						}
					}
				}
				field := &nb.FieldResponse{}
				field.IsEditable = viewOfRelation.IsEditable
				field.IsVisibleLayout = fieldReq.IsVisibleLayout

				var tableFields []*nb.Field
				rows, err := conn.Query(ctx, "SELECT id, autofill_table, autofill_field, slug FROM field WHERE table_slug = $1", req.TableSlug)
				if err != nil {
					return nil, err
				}
				defer rows.Close()
				for rows.Next() {
					field := nb.Field{}
					err := rows.Scan(&field.Id, &field.AutofillTable, &field.AutofillField, &field.Slug)
					if err != nil {
						return nil, err
					}
					tableFields = append(tableFields, &field)
				}
				autofillFields := []map[string]interface{}{}
				for _, field := range tableFields {
					autoFillTable := field.AutofillTable
					splitedAutoFillTable := make([]string, 0)
					if strings.Contains(field.AutofillTable, "#") {
						splitedAutoFillTable = strings.Split(field.AutofillTable, "#")
						autoFillTable = splitedAutoFillTable[0]
					}
					if field.AutofillField != "" && autoFillTable != "" && autoFillTable == strings.Split(fieldReq.Id, "#")[0] {
						autofill := map[string]interface{}{
							"field_from": field.AutofillField,
							"field_to":   field.Slug,
							"automatic":  field.Automatic,
						}
						if fieldResp.Slug == splitedAutoFillTable[1] {
							autofillFields = append(autofillFields, autofill)
						}

						originalAttributes := make(map[string]interface{})
						dynamicTables := []string{}
						var viewField nb.FieldResponse

						if relation.Type == "Many2Dynamic" {
							for _, dynamicTable := range relation.DynamicTables {
								if err != nil {
									return nil, err
								}
								viewFieldsOfDynamicRelation := dynamicTable.ViewFields
								var viewOfDynamicRelation nb.View
								err = conn.QueryRow(ctx, "SELECT id, relation_id, relation_table_slug FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, dynamicTable.TableSlug).Scan(&viewOfDynamicRelation.Id, &viewOfDynamicRelation.RelationId, &viewOfDynamicRelation.RelationTableSlug)
								if err != nil {
									if err != pgx.ErrNoRows {
										return nil, err
									}
								}
								if err != nil {
									return nil, err
								}
								if len(viewOfDynamicRelation.ViewFields) > 0 {
									viewFieldsOfDynamicRelation = viewOfDynamicRelation.ViewFields
								}
								viewFieldsInDynamicTable := []string{}
								for _, fieldID := range viewFieldsOfDynamicRelation {
									err := conn.QueryRow(ctx, "SELECT id, table_id, slug, label, attributes FROM field WHERE id = $1", fieldID).Scan(&viewField.Id, &viewField.TableId, &viewField.Slug, &viewField.Label, &viewField.Attributes)
									if err != nil {
										return nil, err
									}
									if viewField.Attributes != nil {
										attributesBytes, err := viewField.Attributes.MarshalJSON()
										if err != nil {
											return nil, err
										}
										err = json.Unmarshal(attributesBytes, &viewField.Attributes)
										if err != nil {
											return nil, err

										}
										if req.LanguageSetting != "" && viewField.EnableMultilanguage {
											if strings.HasSuffix(viewField.Slug, "_"+req.LanguageSetting) {
												viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, viewField.Slug)
											} else {
												continue
											}
										} else {
											viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, viewField.Slug)
										}
									}
									dynamicTableToAttribute := make(map[string]interface{})
									dynamicTableToAttribute["view_fields"] = viewFieldsInDynamicTable

									originalAttributes := make(map[string]interface{})
									if viewOfDynamicRelation.Attributes != nil {
										attributesBytes, err := viewOfDynamicRelation.Attributes.MarshalJSON()
										if err != nil {
											return nil, err
										}
										err = json.Unmarshal(attributesBytes, &originalAttributes)
										if err != nil {
											return nil, err
										}
									}
								}

							}
							originalAttributes = make(map[string]interface{})

							originalAttributes["autofill"] = autofillFields
							originalAttributes["view_fields"] = fieldAsAttribute
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
						} else {
							originalAttributes["autofill"] = autofillFields
							originalAttributes["view_fields"] = fieldAsAttribute
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
							for k, v := range viewOfRelation.Attributes.AsMap() {
								originalAttributes[k] = v
							}

						}
						originalAttributes["default_values"] = viewOfRelation.DefaultValues
						originalAttributes["creatable"] = viewOfRelation.Creatable

						originalAttributesJSON, err := json.Marshal(originalAttributes)
						if err != nil {
							return nil, err
						}
						var encodedAttributes []byte
						err = json.Unmarshal(originalAttributesJSON, &encodedAttributes)
						if err != nil {
							return nil, err
						}
						var attributes structpb.Struct
						err = protojson.Unmarshal(encodedAttributes, &attributes)
						if err != nil {
							return nil, err
						}
						field.Attributes = &attributes
						fieldRes = append(fieldRes, &nb.FieldResponse{
							Id:                  field.Id,
							TableId:             field.TableId,
							Required:            field.Required,
							Slug:                field.Slug,
							Label:               field.Label,
							Default:             field.Default,
							Type:                field.Type,
							Index:               field.Index,
							EnableMultilanguage: field.EnableMultilanguage,
							Attributes: &structpb.Struct{
								Fields: field.Attributes.Fields,
							},
						})
						if strings.Contains(fieldReq.Id, "@") {
							field.Id = fieldReq.Id
						} else {
							guid := fieldReq.Id
							fieldQuery := `
								SELECT
									"id",
									"table_id",
									"required",
									"slug",
									"label",
									"default",
									"type",
									"index",
									"attributes",
									"is_visible",
									"is_system",
									"is_search",
									"autofill_field",
									"autofill_table",
									"relation_id",
									"unique",
									"automatic",
									"enable_multilanguage",

								FROM "field"
								WHERE id = $1
							`
							err := conn.QueryRow(ctx, fieldQuery, guid).Scan(
								&field.Id,
								&field.TableId,
								&field.Required,
								&field.Slug,
								&field.Label,
								&field.Default,
								&field.Type,
								&field.Index,
								&field.Attributes,
								&field.IsVisible,
								&field.IsSystem,
								&field.IsSearch,
								&field.AutofillField,
								&field.AutofillTable,
								&field.RelationId,
								&field.Unique,
								&field.Automatic,
								&field.EnableMultilanguage,
							)
							if err != nil {
								return nil, err
							}
							if field != nil {
								viewField.IsVisibleLayout = fieldReq.IsVisibleLayout
								viewField.Order = fieldReq.Order
								viewField.Id = fieldReq.Id
								viewField.RelationType = fieldReq.RelationType
								viewField.Label = fieldReq.FieldName
								viewField.Attributes = field.Attributes
								fieldRes = append(fieldRes, &nb.FieldResponse{
									Id:              viewField.Id,
									Column:          viewField.Column,
									RelationType:    viewField.RelationType,
									Order:           viewField.Order,
									IsVisibleLayout: viewField.IsVisibleLayout,
									Label:           viewField.Label,
									Attributes:      viewField.Attributes,
								})
							}
							fieldRes = append(fieldRes, &nb.FieldResponse{
								Id:              viewField.Id,
								Column:          viewField.Column,
								RelationType:    viewField.RelationType,
								Order:           viewField.Order,
								IsVisibleLayout: viewField.IsVisibleLayout,
								Label:           viewField.Label,
								Attributes:      viewField.Attributes,
							})
						}

					}
				}
			}

			var sectionResponses []*nb.SectionResponse
			for _, section := range sections {
				var fieldsResponse []*nb.FieldResponse
				for _, field := range section.Fields {
					fieldsResponse = append(fieldsResponse, &nb.FieldResponse{
						Id:         field.Id,
						Order:      field.Order,
						Column:     field.Column,
						Label:      field.FieldName,
						Attributes: field.Attributes,
					})
				}

				fieldsWithPermissions, err := helper.AddPermissionToField(ctx, conn, fieldRes, req.RoleId, req.TableSlug, req.ProjectId)
				if err != nil {
					return nil, err
				}

				var fieldsForSection []*nb.FieldForSection
				for _, field := range fieldsWithPermissions {
					fieldsForSection = append(fieldsForSection, &nb.FieldForSection{
						Id:              field.Id,
						Column:          field.Column,
						RelationType:    field.RelationType,
						Order:           field.Order,
						IsVisibleLayout: field.IsVisibleLayout,
						FieldName:       field.Label,
						Attributes:      field.Attributes,
					})
				}
				section.Fields = fieldsForSection
				sectionResponses = append(sectionResponses, &nb.SectionResponse{
					Id:     section.Id,
					Label:  section.Label,
					Fields: fieldsResponse,
				})
			}

			resp = &nb.GetAllSectionsResponse{
				Sections: sectionResponses,
			}
		}
	}

	return resp, nil

}
