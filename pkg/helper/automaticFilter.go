package helper

import (
	"context"
	"database/sql"
	"strings"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

func GetAutomaticFilter(ctx context.Context, req models.GetAutomaticFilterRequest) (map[string]interface{}, error) {
	var (
		tableSlug         sql.NullString
		customField       sql.NullString
		objectField       sql.NullString
		notUseInTab       sql.NullBool
		autofilter        nb.RoleWithAppTablePermissions_Table_AutomaticFilter
		many2ManyRelation bool
	)

	automaticFilterQuery := `
	SELECT
		table_slug,
		custom_field,
		object_field,
		not_use_in_tab
	FROM automatic_filter
	WHERE method = 'read' AND role_id = $1 AND table_slug = $2 AND deleted_at IS NULL
	`
	err := req.Conn.QueryRow(ctx, automaticFilterQuery, req.RoleIdFromToken, req.TableSlug).Scan(
		&tableSlug,
		&customField,
		&objectField,
		&notUseInTab,
	)
	if err != nil {
		if err.Error() != config.ErrNoRows {
			return req.Params, errors.Wrap(err, "when scan automaticFilter resp")
		}
	}
	autofilter.CustomField = customField.String
	autofilter.TableSlug = tableSlug.String
	autofilter.ObjectField = objectField.String
	autofilter.NotUseInTab = notUseInTab.Bool

	if len(autofilter.TableSlug) != 0 {
		if !autofilter.NotUseInTab {
			if strings.Contains(autofilter.ObjectField, "#") {
				var (
					splitedElement = strings.Split(autofilter.ObjectField, "#")
					reltype        sql.NullString
					fieldFrom      sql.NullString
				)
				autofilter.ObjectField = splitedElement[0]

				relquery := `SELECT type, field_from FROM "relation" WHERE id = $1`
				if err := req.Conn.QueryRow(ctx, relquery, splitedElement[1]).Scan(&reltype, &fieldFrom); err != nil {
					if err.Error() != config.ErrNoRows {
						return req.Params, errors.Wrap(err, "when get automaticFilter relation")
					}
				}

				switch reltype.String {
				case "Many2One":
					autofilter.CustomField = fieldFrom.String
				}
			}
			if autofilter.CustomField == "user_id" {
				if autofilter.ObjectField != req.TableSlug {
					if !many2ManyRelation {
						req.Params[autofilter.ObjectField+"_id"] = req.Params["user_id_from_token"]
					} else {
						req.Params[autofilter.ObjectField+"ids"] = req.Params["user_id_from_token"]
					}
				}
			} else {
				var connectionTableSlug = autofilter.CustomField[:len(autofilter.CustomField)-3]
				var objFromAuth = FindOneTableFromParams(cast.ToSlice(req.Params["tables"]), autofilter.ObjectField)
				if objFromAuth != nil {
					if connectionTableSlug != req.TableSlug {
						if !many2ManyRelation {
							req.Params[autofilter.CustomField] = objFromAuth["object_id"]
						}
					}
				} else {
					req.Params["guid"] = objFromAuth["object_id"]
				}

			}
		}
	}

	return req.Params, nil
}
