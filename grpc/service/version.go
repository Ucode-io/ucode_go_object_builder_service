package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/protobuf/types/known/emptypb"
)

type versionService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedVersionServiceServer
}

func NewVersionService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *versionService { // ,
	return &versionService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (f *versionService) Create(ctx context.Context, req *nb.CreateVersionRequest) (resp *nb.Version, err error) {
	f.log.Info("---CreateFunction--->>>", logger.Any("req", req))

	resp, err = f.strg.Version().Create(ctx, req)
	if err != nil {
		f.log.Error("---CreateFunction--->>>", logger.Error(err))
		return &nb.Version{}, err
	}

	return resp, nil
}

func (f *versionService) GetList(ctx context.Context, req *nb.GetVersionListRequest) (resp *nb.GetVersionListResponse, err error) {
	f.log.Info("---GetListFunction--->>>", logger.Any("req", req))

	resp, err = f.strg.Version().GetList(ctx, req)
	if err != nil {
		f.log.Error("---GetListFunction--->>>", logger.Error(err))
		return &nb.GetVersionListResponse{}, err
	}

	return resp, nil
}

func (f *versionService) GetSingle(ctx context.Context, req *nb.VersionPrimaryKey) (resp *nb.Version, err error) {
	f.log.Info("---GetSingleFunction--->>>", logger.Any("req", req))

	resp, err = f.strg.Version().GetSingle(ctx, req)
	if err != nil {
		f.log.Error("---GetSingleFunction--->>>", logger.Error(err))
		return &nb.Version{}, err
	}

	return resp, nil
}

func (f *versionService) Update(ctx context.Context, req *nb.Version) (resp *emptypb.Empty, err error) {
	f.log.Info("---UpdateFunction--->>>", logger.Any("req", req))

	err = f.strg.Version().Update(ctx, req)
	if err != nil {
		f.log.Error("---UpdateFunction--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (f *versionService) Delete(ctx context.Context, req *nb.VersionPrimaryKey) (resp *emptypb.Empty, err error) {
	f.log.Info("---DeleteFunction--->>>", logger.Any("req", req))

	err = f.strg.Version().Delete(ctx, req)
	if err != nil {
		f.log.Error("---DeleteFunction--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
