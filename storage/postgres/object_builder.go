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
}

func (o *objectBuilderRepo) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	conn := o.db

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
	jsonBytes = []byte(fmt.Sprintf(`{"data": %s}`, jsonBytes))

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
