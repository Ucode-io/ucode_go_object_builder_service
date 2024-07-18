package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type folderGroupRepo struct {
	db *pgxpool.Pool
}

func NewFolderGroupRepo(db *pgxpool.Pool) storage.FolderGroupRepoI {
	return &folderGroupRepo{
		db: db,
	}
}

func (f *folderGroupRepo) Create(ctx context.Context, req *nb.CreateFolderGroupRequest) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.GetProjectId())

	folderGroupId := uuid.NewString()

	var parentId interface{} = req.ParentId
	if req.ParentId == "" {
		parentId = nil
	}

	query := `INSERT INTO "folder_group" (
		id,
		table_id,
		name,
		comment,
		code,
		parent_id
	) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := conn.Exec(ctx, query, folderGroupId, req.TableId, req.Name, req.Comment, req.Code, parentId)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return f.GetByID(ctx, &nb.FolderGroupPrimaryKey{Id: folderGroupId, ProjectId: req.ProjectId})
}

func (f *folderGroupRepo) GetByID(ctx context.Context, req *nb.FolderGroupPrimaryKey) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.ProjectId)

	var (
		id       sql.NullString
		tableId  sql.NullString
		name     sql.NullString
		comment  sql.NullString
		code     sql.NullString
		parentId sql.NullString
	)

	query := `
		SELECT
			id,
			table_id,
			name,
			comment,
			code,
			parent_id
		FROM folder_group fg
		WHERE fg.id = $1
	`

	err := conn.QueryRow(ctx, query, req.Id).Scan(
		&id,
		&tableId,
		&name,
		&comment,
		&code,
		&parentId,
	)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return &nb.FolderGroup{
		Id:       id.String,
		TableId:  tableId.String,
		Name:     name.String,
		Comment:  comment.String,
		Code:     code.String,
		ParentId: parentId.String,
	}, nil
}

func (f *folderGroupRepo) GetAll(ctx context.Context, req *nb.GetAllFolderGroupRequest) (*nb.GetAllFolderGroupResponse, error) {
	var (
		conn = psqlpool.Get(req.GetProjectId())
		resp = &nb.GetAllFolderGroupResponse{}

		query, adds                                          string
		folderGroupCount, itemCount, queryLimit, queryOffset int32
	)

	if len(req.ParentId) == 0 {
		adds = ` AND parent_id IS NULL`
	} else {
		adds = fmt.Sprintf(" AND parent_id = '%s'", req.ParentId)
	}
	query = fmt.Sprintf(`SELECT COUNT(*) FROM "folder_group" WHERE table_id = $1 %s`, adds)

	err := conn.QueryRow(ctx, query, req.TableId).Scan(&folderGroupCount)
	if err != nil {
		return &nb.GetAllFolderGroupResponse{}, err
	}

	var (
		searchFields = []string{}
	)

	var tableSlug string
	err = conn.QueryRow(ctx, `SELECT slug FROM "table" WHERE id = $1`, req.TableId).Scan(&tableSlug)
	if err != nil {
		return &nb.GetAllFolderGroupResponse{}, err
	}

	query = fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE folder_id IS NULL`, tableSlug)
	err = conn.QueryRow(ctx, query).Scan(&itemCount)
	if err != nil {
		return &nb.GetAllFolderGroupResponse{}, err
	}

	query = `
		SELECT 
			f.type, 
			f.slug, 
			f.attributes,
			f.is_search
		FROM "field" f 
		JOIN "table" t ON t.id = f.table_id 
		WHERE t.slug = $1`
	fieldRows, err := conn.Query(ctx, query, tableSlug)
	if err != nil {
		return &nb.GetAllFolderGroupResponse{}, err
	}
	defer fieldRows.Close()

	fields := make(map[string]models.Field)

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
			&fBody.IsSearch,
		)
		if err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}

		if fBody.IsSearch && helper.FIELD_TYPES[fBody.Type] == "VARCHAR" {
			searchFields = append(searchFields, fBody.Slug)
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}

		fields[fBody.Slug] = fBody
	}

	queryX := req.Offset - folderGroupCount
	if queryX > 0 {
		getItemsReq := models.GetItemsBody{
			TableSlug:    tableSlug,
			FieldsMap:    fields,
			SearchFields: searchFields,
			Params: map[string]interface{}{
				"folder_id": nil,
				"limit":     req.Limit,
				"offset":    queryX,
			},
		}
		items, count, err := helper.GetItems(ctx, conn, getItemsReq)
		if err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}

		response := map[string]interface{}{
			"count":    count,
			"response": items,
		}

		itemsStruct, err := helper.ConvertMapToStruct(response)
		if err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}
		resp.FolderGroups = append(resp.FolderGroups, &nb.FolderGroup{
			Id:      "",
			TableId: req.TableId,
			Items:   itemsStruct,
		})
	} else {
		queryLimit = req.Limit
		queryOffset = req.Offset

		query = `
			SELECT
				id,
				table_id,
				name,
				comment,
				code,
				parent_id
			FROM folder_group fg
			WHERE table_id = $1 AND
		`

		args := []interface{}{req.TableId, queryOffset, queryLimit}
		if req.ParentId == "" {
			query += ` parent_id is NULL OFFSET $2 LIMIT $3`
		} else {
			query += ` parent_id = $4 OFFSET $2 LIMIT $3`
			args = append(args, req.ParentId)
		}

		rows, err := conn.Query(ctx, query, args...)
		if err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id       sql.NullString
				tableId  sql.NullString
				name     sql.NullString
				comment  sql.NullString
				code     sql.NullString
				parentId sql.NullString
			)

			err := rows.Scan(
				&id,
				&tableId,
				&name,
				&comment,
				&code,
				&parentId,
			)
			if err != nil {
				return &nb.GetAllFolderGroupResponse{}, err
			}

			getItemsReq := models.GetItemsBody{
				TableSlug:    tableSlug,
				FieldsMap:    fields,
				SearchFields: searchFields,
				Params: map[string]interface{}{
					"folder_id": id.String,
				},
			}
			items, count, err := helper.GetItems(ctx, conn, getItemsReq)
			if err != nil {
				return &nb.GetAllFolderGroupResponse{}, err
			}

			response := map[string]interface{}{
				"count":    count,
				"response": items,
			}

			itemsStruct, err := helper.ConvertMapToStruct(response)
			if err != nil {
				return &nb.GetAllFolderGroupResponse{}, err
			}

			resp.FolderGroups = append(resp.FolderGroups, &nb.FolderGroup{
				Id:       id.String,
				TableId:  tableId.String,
				Name:     name.String,
				Comment:  comment.String,
				Code:     code.String,
				Items:    itemsStruct,
				ParentId: parentId.String,
				Type:     "FOLDER",
			})
		}

		queryOffset = 0
		queryLimit = req.Limit + queryX
		if queryLimit > 0 {
			if len(req.ParentId) == 0 {
				getItemsReq := models.GetItemsBody{
					TableSlug:    tableSlug,
					FieldsMap:    fields,
					SearchFields: searchFields,
					Params: map[string]interface{}{
						"folder_id": nil,
						"limit":     queryLimit,
						"offset":    queryOffset,
					},
				}
				items, count, err := helper.GetItems(ctx, conn, getItemsReq)
				if err != nil {
					return &nb.GetAllFolderGroupResponse{}, err
				}

				response := map[string]interface{}{
					"count":    count,
					"response": items,
				}

				itemsStruct, err := helper.ConvertMapToStruct(response)
				if err != nil {
					return &nb.GetAllFolderGroupResponse{}, err
				}
				resp.FolderGroups = append(resp.FolderGroups, &nb.FolderGroup{
					Id:      "",
					TableId: req.TableId,
					Items:   itemsStruct,
				})
			}
		}
	}

	resp.Count = folderGroupCount + itemCount
	return resp, nil
}

func (f *folderGroupRepo) Update(ctx context.Context, req *nb.UpdateFolderGroupRequest) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.GetProjectId())

	var parentId interface{} = req.ParentId

	if parentId == "" {
		parentId = nil
	}

	query := `
		UPDATE folder_group SET
			table_id = $1,
			name = $2,
			comment = $3,
			code = $4,
			parent_id = $6,
			updated_at = now()
		WHERE id = $5
	`

	_, err := conn.Exec(ctx, query, req.TableId, req.Name, req.Comment, req.Code, req.Id, parentId)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return &nb.FolderGroup{
		Id:       req.Id,
		TableId:  req.TableId,
		Name:     req.Name,
		Comment:  req.Comment,
		Code:     req.Code,
		ParentId: req.ParentId,
	}, nil
}

func (f *folderGroupRepo) Delete(ctx context.Context, req *nb.FolderGroupPrimaryKey) error {
	conn := psqlpool.Get(req.GetProjectId())

	query := `DELETE FROM folder_group WHERE id = $1`

	_, err := conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}
