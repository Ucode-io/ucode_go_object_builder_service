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

type folderGroupService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedFolderGroupServiceServer
}

func NewFolderGroupService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *folderGroupService {
	return &folderGroupService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (fg *folderGroupService) Create(ctx context.Context, req *nb.CreateFolderGroupRequest) (resp *nb.FolderGroup, err error) {
	fg.log.Info("---CreateFolderGroup--->>>", logger.Any("req", req))

	resp, err = fg.strg.FolderGroup().Create(ctx, req)
	if err != nil {
		fg.log.Error("---CreateFolderGroup--->>>", logger.Error(err))
		return &nb.FolderGroup{}, err
	}

	return resp, nil
}

func (fg *folderGroupService) GetByID(ctx context.Context, req *nb.FolderGroupPrimaryKey) (resp *nb.FolderGroup, err error) {

	fg.log.Info("---GetByIDFolderGroup--->>>", logger.Any("req", req))

	resp, err = fg.strg.FolderGroup().GetByID(ctx, req)
	if err != nil {
		fg.log.Error("---GetByIDFolderGroup--->>>", logger.Error(err))
		return &nb.FolderGroup{}, err
	}

	return resp, nil
}

func (fg *folderGroupService) GetAll(ctx context.Context, req *nb.GetAllFolderGroupRequest) (resp *nb.GetAllFolderGroupResponse, err error) {

	fg.log.Info("---GetAllFolderGroup--->>>", logger.Any("req", req))

	resp, err = fg.strg.FolderGroup().GetAll(ctx, req)
	if err != nil {
		fg.log.Error("---GetAllFolderGroup--->>>", logger.Error(err))
		return &nb.GetAllFolderGroupResponse{}, err
	}

	return resp, nil
}

func (fg *folderGroupService) Update(ctx context.Context, req *nb.UpdateFolderGroupRequest) (resp *nb.FolderGroup, err error) {
	fg.log.Info("---UpdateFolderGroup--->>>", logger.Any("req", req))

	resp, err = fg.strg.FolderGroup().Update(ctx, req)
	if err != nil {
		fg.log.Error("---UpdateFolderGroup--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (fg *folderGroupService) Delete(ctx context.Context, req *nb.FolderGroupPrimaryKey) (resp *emptypb.Empty, err error) {
	fg.log.Info("---DeleteFolderGroup--->>>", logger.Any("req", req))

	err = fg.strg.FolderGroup().Delete(ctx, req)
	if err != nil {
		fg.log.Error("---DeleteFolderGroup--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
