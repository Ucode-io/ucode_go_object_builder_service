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
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.GetAllMenuPermissions", req)
	defer dbSpan.Finish()

	p.log.Info("---GetAllMenuPermissions--->", logger.Any("req", req))

	resp, err = p.strg.Permission().GetAllMenuPermissions(ctx, req)
	if err != nil {
		p.log.Error("---GetAllMenuPermissions--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) CreateDefaultPermission(ctx context.Context, req *nb.CreateDefaultPermissionRequest) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.CreateDefaultPermission", req)
	defer dbSpan.Finish()

	p.log.Info("---CreateDefaultPermission--->", logger.Any("req", req))

	err = p.strg.Permission().CreateDefaultPermission(ctx, req)
	if err != nil {
		p.log.Error("---CreateDefaultPermission--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) GetListWithRoleAppTablePermissions(ctx context.Context, req *nb.GetListWithRoleAppTablePermissionsRequest) (resp *nb.GetListWithRoleAppTablePermissionsResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.GetListWithRoleAppTablePermissions", req)
	defer dbSpan.Finish()

	p.log.Info("---GetListWithRoleAppTablePermissions--->", logger.Any("req", req))

	resp, err = p.strg.Permission().GetListWithRoleAppTablePermissions(ctx, req)
	if err != nil {
		p.log.Error("---GetListWithRoleAppTablePermissions--->", logger.Error(err))
	}
	return resp, nil
}

func (p *permissionService) UpdateMenuPermissions(ctx context.Context, req *nb.UpdateMenuPermissionsRequest) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.UpdateMenuPermissions", req)
	defer dbSpan.Finish()

	p.log.Info("---UpdateMenuPermission--->", logger.Any("req", req))

	err = p.strg.Permission().UpdateMenuPermissions(ctx, req)
	if err != nil {
		p.log.Error("---UpdateMenuPermission--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) UpdateRoleAppTablePermissions(ctx context.Context, req *nb.UpdateRoleAppTablePermissionsRequest) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.UpdateRoleAppTablePermissions", req)
	defer dbSpan.Finish()

	p.log.Info("---UpdateRoleAppTablePermissions--->", logger.Any("req", req))

	err = p.strg.Permission().UpdateRoleAppTablePermissions(ctx, req)
	if err != nil {
		p.log.Error("---UpdateRoleAppTablePermissions--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) GetPermissionsByTableSlug(ctx context.Context, req *nb.GetPermissionsByTableSlugRequest) (resp *nb.GetPermissionsByTableSlugResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.GetPermissionsByTableSlug", req)
	defer dbSpan.Finish()

	p.log.Info("---GetPermissionsByTableSlug--->", logger.Any("req", req))

	resp, err = p.strg.Permission().GetPermissionsByTableSlug(ctx, req)
	if err != nil {
		p.log.Error("---GetPermissionsByTableSlug--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (p *permissionService) UpdatePermissionsByTableSlug(ctx context.Context, req *nb.UpdatePermissionsRequest) (resp *emptypb.Empty, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_permission.UpdatePermissionsByTableSlug", req)
	defer dbSpan.Finish()

	p.log.Info("---UpdatePermissionsByTableSlug--->", logger.Any("req", req))

	err = p.strg.Permission().UpdatePermissionsByTableSlug(ctx, req)
	if err != nil {
		p.log.Error("---UpdatePermissionsByTableSlug--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}
