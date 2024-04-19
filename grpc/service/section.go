package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type sectionService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedSectionServiceServer
}

func NewSectionService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *sectionService { // ,
	return &sectionService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (s *tableService) GetViewRelation(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetViewRelationResponse, err error) {
	s.log.Info("---GetViewRelation--->>>", logger.Any("req", req))

	resp, err = s.strg.Section().GetViewRelation(ctx, req)
	if err != nil {
		s.log.Error("---GetViewRelation--->>>", logger.Error(err))
		return &nb.GetViewRelationResponse{}, err
	}

	return resp, nil
}
