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

type permissionService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedPermissionServiceServer
}

func NewPermissionService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *permissionService {
	return &permissionService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (p *permissionService) GetAllMenuPermissions(ctx context.Context, req *nb.GetAllMenuPermissionsRequest) (resp *nb.GetAllMenuPermissionsResponse, err error) {
	resp = &nb.GetAllMenuPermissionsResponse{}

	p.log.Info("---GetAllMenuPermissions--->", logger.Any("req", req))

	resp, err = p.strg.Permission().GetAllMenuPermissions(ctx, req)
	if err != nil {
		p.log.Error("---GetAllMenuPermissions--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) CreateDefaultPermission(ctx context.Context, req *nb.CreateDefaultPermissionRequest) (resp *emptypb.Empty, err error) {
	p.log.Info("---CreateDefaultPermission--->", logger.Any("req", req))

	err = p.strg.Permission().CreateDefaultPermission(ctx, req)
	if err != nil {
		p.log.Error("---CreateDefaultPermission--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) UpdateMenuPermissions(ctx context.Context, req *nb.UpdateMenuPermissionsRequest) (resp *emptypb.Empty, err error) {
	p.log.Info("---UpdateMenuPermission--->", logger.Any("req", req))

	err = p.strg.Permission().UpdateMenuPermissions(ctx, req)
	if err != nil {
		p.log.Error("---UpdateMenuPermission--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}
