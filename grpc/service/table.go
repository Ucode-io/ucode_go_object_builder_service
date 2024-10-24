package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	span "ucode/ucode_go_object_builder_service/pkg/jaeger"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/protobuf/types/known/emptypb"
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
