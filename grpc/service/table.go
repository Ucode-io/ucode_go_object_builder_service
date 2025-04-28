package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	span "ucode/ucode_go_object_builder_service/pkg/jaeger"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type tableService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedTableServiceServer
}

func NewTableService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *tableService { // ,
	return &tableService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (t *tableService) Create(ctx context.Context, req *nb.CreateTableRequest) (resp *nb.CreateTableResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.Create", req)
	defer dbSpan.Finish()

	t.log.Info("---CreateTable--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().Create(ctx, req)
	if err != nil {
		t.log.Error("---CreateTable--->>>", logger.Error(err))
		return &nb.CreateTableResponse{}, err
	}

	return resp, nil
}

func (t *tableService) GetByID(ctx context.Context, req *nb.TablePrimaryKey) (resp *nb.Table, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetByID", req)
	defer dbSpan.Finish()

	t.log.Info("---GetByIDTable--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().GetByID(ctx, req)
	if err != nil {
		t.log.Error("---GetByIDTable--->>>", logger.Error(err))
		return &nb.Table{}, err
	}

	return resp, nil
}

func (t *tableService) GetAll(ctx context.Context, req *nb.GetAllTablesRequest) (resp *nb.GetAllTablesResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetAll", req)
	defer dbSpan.Finish()

	t.log.Info("---GetAllTable--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().GetAll(ctx, req)
	if err != nil {
		t.log.Error("---GetAllTable--->>>", logger.Error(err))
		return &nb.GetAllTablesResponse{}, err
	}

	return resp, nil
}

func (t *tableService) Update(ctx context.Context, req *nb.UpdateTableRequest) (resp *nb.Table, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.Update", req)
	defer dbSpan.Finish()

	t.log.Info("---UpdateTable--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().Update(ctx, req)
	if err != nil {
		t.log.Error("---UpdateTable--->>>", logger.Error(err))
		return &nb.Table{}, err
	}

	return resp, nil
}

func (t *tableService) Delete(ctx context.Context, req *nb.TablePrimaryKey) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.Delete", req)
	defer dbSpan.Finish()

	t.log.Info("---DeleteTable--->>>", logger.Any("req", req))

	err = t.strg.Table().Delete(ctx, req)
	if err != nil {
		t.log.Error("---DeleteTable--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (t *tableService) GetTablesByLabel(ctx context.Context, req *nb.GetTablesByLabelReq) (resp *nb.GetAllTablesResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetTablesByLabel", req)
	defer dbSpan.Finish()

	t.log.Info("---UpdateLabel--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().GetTablesByLabel(ctx, req)
	if err != nil {
		t.log.Error("---UpdateLabel--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (t *tableService) GetChart(ctx context.Context, req *nb.ChartPrimaryKey) (resp *nb.GetChartResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetChart", req)
	defer dbSpan.Finish()

	t.log.Info("---GetChart--->>>", logger.Any("req", req))

	resp, err = t.strg.Table().GetChart(ctx, req)
	if err != nil {
		t.log.Error("---GetChart--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (t *tableService) CreateConnectionAndSchema(ctx context.Context, req *nb.CreateConnectionAndSchemaReq) (*nb.GetTrackedUntrackedTableResp, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.CreateConnectionAndSchema", req)
	defer dbSpan.Finish()

	t.log.Info("---CreateConnectionAndSchema--->>>", logger.Any("req", req))

	resp, err := t.strg.Table().CreateConnectionAndSchema(ctx, req)
	if err != nil {
		t.log.Error("---CreateConnectionAndSchema--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (t *tableService) GetTrackedUntrackedTables(ctx context.Context, req *nb.GetTrackedUntrackedTablesReq) (*nb.GetTrackedUntrackedTableResp, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetTrackedUntrackedTables", req)
	defer dbSpan.Finish()

	t.log.Info("---GetTrackedUntrackedTables--->>>", logger.Any("req", req))

	resp, err := t.strg.Table().GetTrackedUntrackedTables(ctx, req)
	if err != nil {
		t.log.Error("---GetTrackedUntrackedTables--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (t *tableService) GetTrackedConnections(ctx context.Context, req *nb.GetTrackedConnectionsReq) (*nb.GetTrackedConnectionsResp, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.GetTrackedConnections", req)
	defer dbSpan.Finish()

	t.log.Info("---GetTrackedConnections--->>>", logger.Any("req", req))

	resp, err := t.strg.Table().GetTrackedConnections(ctx, req)
	if err != nil {
		t.log.Error("---GetTrackedConnections--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (t *tableService) TrackedTablesByIds(ctx context.Context, req *nb.TrackedTablesByIdsReq) (*nb.TrackedTablesByIdsResp, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.TrackedTablesByIds", req)
	defer dbSpan.Finish()

	t.log.Info("---TrackedTablesByIds--->>>", logger.Any("req", req))

	resp, err := t.strg.Table().TrackedTablesByIds(ctx, req)
	if err != nil {
		t.log.Error("---TrackedTablesByIds--->>>", logger.Error(err))
		return nil, err
	}

	for _, table := range resp.Tables {
		tableResp, err := t.strg.Table().Create(ctx, &nb.CreateTableRequest{
			Label:      table.TableName,
			Slug:       table.TableName,
			ShowInMenu: true,
			ViewId:     uuid.NewString(),
			LayoutId:   uuid.NewString(),
			Attributes: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"label_en": structpb.NewStringValue(table.TableName),
				},
			},
			ProjectId: req.ProjectId,
		})
		if err != nil {
			t.log.Error("---TrackedTablesByIds create table--->>>", logger.Error(err))
			return nil, err
		}

		_, err = t.strg.Menu().Create(ctx, &nb.CreateMenuRequest{
			Label:    table.TableName,
			TableId:  tableResp.Id,
			Type:     "TABLE",
			ParentId: "c57eedc3-a954-4262-a0af-376c65b5a284",
			Attributes: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"label_en": structpb.NewStringValue(table.TableName),
				},
			},
			ProjectId: req.ProjectId,
		})
		if err != nil {
			t.log.Error("---TrackedTablesByIds create menu--->>>", logger.Error(err))
			return nil, err
		}

		for _, field := range table.Fields {
			if field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
				continue
			}
			_, err := t.strg.Field().Create(ctx, &nb.CreateFieldRequest{
				Id:      uuid.NewString(),
				TableId: tableResp.Id,
				Type:    helper.GetCustomToPostgres(field.Type),
				Label:   field.Name,
				Slug:    field.Name,
				Attributes: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"label_en": structpb.NewStringValue(field.Name),
					},
				},
				ProjectId: req.ProjectId,
			})
			if err != nil {
				t.log.Error("---TrackedTablesByIds create field--->>>", logger.Error(err))
				return nil, err
			}
		}
	}

	return resp, nil
}

func (t *tableService) UntrackTableById(ctx context.Context, req *nb.UntrackTableByIdReq) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_table.UntrackTableById", req)
	defer dbSpan.Finish()

	t.log.Info("---UntrackTableById--->>>", logger.Any("req", req))

	err = t.strg.Table().UntrackTableById(ctx, req)
	if err != nil {
		t.log.Error("---UntrackTableById--->>>", logger.Error(err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
