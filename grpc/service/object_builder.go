package service

import (
	"context"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type objectBuilderService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedBuilderProjectServiceServer
}

func NewObjectBuilderService(strg storage.StorageI, cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *objectBuilderService { // strg storage.StorageI,
	return &objectBuilderService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (b *objectBuilderService) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!ObjectBuilderGetList--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetList(ctx, req)
	if err != nil {
		b.log.Error("!!!ObjectBuilderGetList--->", logger.Error(err))
		return resp, err

	}

	return resp, nil
}
