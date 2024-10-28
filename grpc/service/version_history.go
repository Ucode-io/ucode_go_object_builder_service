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
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_version_history.GetByID", req)
	defer dbSpan.Finish()

	v.log.Info("---GetByID version--->>>", logger.Any("req", req))

	resp, err := v.strg.VersionHistory().GetById(ctx, req)
	if err != nil {
		v.log.Error("---GetByID version--->>>", logger.Error(err))
		return &nb.VersionHistory{}, err
	}

	return resp, nil
}

func (v *versionHistoryService) GatAll(ctx context.Context, req *nb.GetAllRquest) (*nb.ListVersionHistory, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_version_history.GatAll", req)
	defer dbSpan.Finish()

	v.log.Info("---GatAll Version--->>>", logger.Any("req", req))

	resp, err := v.strg.VersionHistory().GetAll(ctx, req)
	if err != nil {
		v.log.Error("---GatAll Version--->>>", logger.Error(err))
		return &nb.ListVersionHistory{}, err
	}

	return resp, nil
}

func (v *versionHistoryService) Update(ctx context.Context, req *nb.UsedForEnvRequest) (*emptypb.Empty, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_version_history.Update", req)
	defer dbSpan.Finish()

	v.log.Info("---UpdateVersionHistory--->>>", logger.Any("req", req))

	err := v.strg.VersionHistory().Update(ctx, req)
	if err != nil {
		v.log.Error("---UpdateVersionHistory--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (v *versionHistoryService) Create(ctx context.Context, req *nb.CreateVersionHistoryRequest) (*emptypb.Empty, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_version_history.Create", req)
	defer dbSpan.Finish()

	err := v.strg.VersionHistory().Create(ctx, req)
	if err != nil {
		v.log.Error("---CreateVersionHistory--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
