package service

import (
	"context"
	"fmt"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type csvService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedCSVServiceServer
}

func NewCSVService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *csvService {
	return &csvService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (b *csvService) GetListInCSV(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetListInCSV--->", logger.Any("req", req))

	fmt.Println(b.strg.CSV())
	resp, err = b.strg.CSV().GetListInCSV(ctx, req)
	if err != nil {
		b.log.Error("!!!GetListInCSV--->GetList", logger.Error(err))
		return resp, err
	}

	return resp, nil
}
