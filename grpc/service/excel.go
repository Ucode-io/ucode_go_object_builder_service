package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	span "ucode/ucode_go_object_builder_service/pkg/jaeger"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type excelService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedExcelServiceServer
}

func NewExcelService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *excelService { // ,
	return &excelService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (e *excelService) ExcelRead(ctx context.Context, req *nb.ExcelReadRequest) (resp *nb.ExcelReadResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_excel.ExcelRead", req)
	defer dbSpan.Finish()
	e.log.Info("---ExcelRead--->>>", logger.Any("req", req))

	resp, err = e.strg.Excel().ExcelRead(ctx, req)
	if err != nil {
		e.log.Error("---ExcelRead--->>>", logger.Error(err))
		return &nb.ExcelReadResponse{}, err
	}

	return resp, nil
}

func (e *excelService) ExcelToDb(ctx context.Context, req *nb.ExcelToDbRequest) (resp *nb.ExcelToDbResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_excel.ExcelToDb", req)
	defer dbSpan.Finish()

	e.log.Info("---ExcelToDb--->>>", logger.Any("req", req))

	resp, err = e.strg.Excel().ExcelToDb(ctx, req)
	if err != nil {
		e.log.Error("---ExcelToDb--->>>", logger.Error(err))
		return &nb.ExcelToDbResponse{}, err
	}

	return resp, nil
}
