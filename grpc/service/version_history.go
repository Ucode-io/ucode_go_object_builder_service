package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type versionHistoryService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedVersionHistoryServiceServer
}

func NewVersionHistoryService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *versionHistoryService {
	return &versionHistoryService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (v *versionHistoryService) GetByID(ctx context.Context, req *nb.VersionHistoryPrimaryKey) (*nb.VersionHistory, error) {
	v.log.Info("---GetByID version--->>>", logger.Any("req", req))

	resp, err := v.strg.VersionHistory().GetById(ctx, req)
	if err != nil {
		v.log.Error("---GetByID version--->>>", logger.Error(err))
		return &nb.VersionHistory{}, err
	}

	return resp, nil
}

func (v *versionHistoryService) GatAll(ctx context.Context, req *nb.GetAllRquest) (*nb.ListVersionHistory, error) {
	v.log.Info("---GatAll Version--->>>", logger.Any("req", req))

	resp, err := v.strg.VersionHistory().GetAll(ctx, req)
	if err != nil {
		v.log.Error("---GatAll Version--->>>", logger.Error(err))
		return &nb.ListVersionHistory{}, err
	}

	return resp, nil
}

func (v *versionHistoryService) Update(ctx context.Context, req *nb.UsedForEnvRequest) error {
	v.log.Info("---UpdateVersionHistory--->>>", logger.Any("req", req))

	err := v.strg.VersionHistory().Update(ctx, req)
	if err != nil {
		v.log.Error("---UpdateVersionHistory--->>>", logger.Error(err))
		return err
	}

	return nil
}

func (v *versionHistoryService) Create(ctx context.Context, req *nb.CreateVersionHistoryRequest) error {
	v.log.Info("---CreateVersionHistory--->>>", logger.Any("req", req))

	err := v.strg.VersionHistory().Create(ctx, req)
	if err != nil {
		v.log.Error("---CreateVersionHistory--->>>", logger.Error(err))
		return err
	}

	return nil
}
