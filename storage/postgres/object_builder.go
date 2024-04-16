package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/structpb"
)

type objectBuilderRepo struct {
	db *pgxpool.Pool
}

func NewObjectBuilder(db *pgxpool.Pool) storage.ObjectBuilderRepoI {
	return &objectBuilderRepo{
		db: db,
	}
}

type ClientType struct {
	Guid              string   `json:"guid"`
	ProjectId         string   `json:"project_id"`
	Name              string   `json:"name"`
	SelfRegister      bool     `json:"self_register"`
	SelfRecover       bool     `json:"self_recover"`
	ClientPlatformIds []string `json:"client_platform_ids"`
	ConfirmBy         string   `json:"confirm_by"`
	IsSystem          bool     `json:"is_system"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	TableSlug         string   `json:"table_slug"`
	DefaultPage       string   `json:"default_page"`
}

type Connection struct {
	Guid          string `json:"guid"`
	TableSlug     string `json:"table_slug"`
	ViewSlug      string `json:"view_slug"`
	ViewLabel     string `json:"view_label"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Icon          string `json:"icon"`
	MainTableSlug string `json:"main_table_slug"`
	FieldSlug     string `json:"field_slug"`
	ClientTypeId  string `json:"client_type_id"`
}

type Role struct {
	Guid             string `json:"guid"`
	Name             string `json:"name"`
	ProjectId        string `json:"project_id"`
	ClientPlatformId string `json:"client_platform_id"`
	ClientTypeId     string `json:"client_type_id"`
	IsSystem         bool   `json:"is_system"`
}

type ClientPlatform struct {
	Guid      string `json:"guid"`
	ProjectId string `json:"project_id"`
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
}

func (o *objectBuilderRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.GetProjectId())

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:new@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			"guid",
			"project_id",
			"name",
			"self_register",
			"self_recover",
			"client_platform_ids",
			"confirm_by",
			"is_system"
		FROM client_type
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	data := make([]ClientType, 0)
	for rows.Next() {
		var clientType ClientType

		err = rows.Scan(
			&clientType.Guid,
			&clientType.ProjectId,
			&clientType.Name,
			&clientType.SelfRegister,
			&clientType.SelfRecover,
			&clientType.ClientPlatformIds,
			&clientType.ConfirmBy,
			&clientType.IsSystem,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		data = append(data, clientType)
	}
	fmt.Println(data)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	var dataStruct structpb.Struct
	jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

	err = json.Unmarshal(jsonBytes, &dataStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          &dataStruct,
		VersionId:     req.VersionId,
		CustomMessage: req.CustomMessage,
		IsCached:      req.IsCached,
	}, err
}

func (o *objectBuilderRepo) GetListConnection(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	// conn := psqlpool.Get(req.GetProjectId())

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:new@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			"guid",
			"table_slug",
			"view_slug",
			"view_label",
			"name",
			"type",
			"icon",
			"main_table_slug",
			"field_slug",
			"client_type_id"
		FROM "connection"
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	data := make([]Connection, 0)
	for rows.Next() {
		var connection Connection

		err = rows.Scan(
			&connection.Guid,
			&connection.TableSlug,
			&connection.ViewSlug,
			&connection.ViewLabel,
			&connection.Name,
			&connection.Type,
			&connection.Icon,
			&connection.MainTableSlug,
			&connection.FieldSlug,
			&connection.ClientTypeId,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		data = append(data, connection)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	var dataStruct structpb.Struct
	jsonBytes = []byte(fmt.Sprintf(`{"response": %s}`, jsonBytes))

	err = json.Unmarshal(jsonBytes, &dataStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{
		TableSlug:     req.TableSlug,
		ProjectId:     req.ProjectId,
		Data:          &dataStruct,
		VersionId:     req.VersionId,
		CustomMessage: req.CustomMessage,
		IsCached:      req.IsCached,
	}, err
}
