package service

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type customErrorMessageService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedCustomErrorMessageServiceServer
}

func NewCustomErrorMessageService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *customErrorMessageService { // ,
	return &customErrorMessageService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}
