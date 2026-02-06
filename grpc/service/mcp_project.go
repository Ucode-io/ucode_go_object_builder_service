package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	span "ucode/ucode_go_object_builder_service/pkg/jaeger"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type mcpProjectService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedMcpProjectServiceServer
}

func NewMcpProjectService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *mcpProjectService {
	return &mcpProjectService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (v *mcpProjectService) CreateMcpProject(ctx context.Context, req *nb.CreateMcpProjectReqeust) (*nb.McpProject, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_mcp_project.CreateMcpProject", req)
	defer dbSpan.Finish()

	v.log.Info("--- CreateMcpProject --->>>", logger.Any("req", req))

	resp, err := v.strg.McpProject().CreateMcpProject(ctx, req)
	if err != nil {
		v.log.Error("---CreateMcpProject--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (v *mcpProjectService) UpdateMcpProject(ctx context.Context, req *nb.McpProject) (*nb.McpProject, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_mcp_project.UpdateMcpProject", req)
	defer dbSpan.Finish()

	v.log.Info("---UpdateMcpProject--->>>", logger.Any("req", req))

	resp, err := v.strg.McpProject().UpdateMcpProject(ctx, req)
	if err != nil {
		v.log.Error("---UpdateMcpProject--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (v *mcpProjectService) GetAllMcpProject(ctx context.Context, req *nb.GetMcpProjectListReq) (*nb.McpProjectList, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_mcp_project.GetAllMcpProject", req)
	defer dbSpan.Finish()

	v.log.Info("---GetAllMcpProject--->>>", logger.Any("req", req))

	resp, err := v.strg.McpProject().GetAllMcpProject(ctx, req)
	if err != nil {
		v.log.Error("--- GetAllMcpProject --->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (v *mcpProjectService) GetMcpProjectFiles(ctx context.Context, req *nb.McpProjectId) (*nb.McpProject, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_mcp_project.GetMcpProjectFiles", req)
	defer dbSpan.Finish()

	resp, err := v.strg.McpProject().GetMcpProjectFiles(ctx, req)
	if err != nil {
		v.log.Error("---GetMcpProjectFiles--->>>", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (v *mcpProjectService) DeleteMcpProject(ctx context.Context, req *nb.McpProjectId) (*nb.McpProject, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_mcp_project.DeleteMcpProject", req)
	defer dbSpan.Finish()

	err := v.strg.McpProject().DeleteMcpProject(ctx, req)
	if err != nil {
		v.log.Error("---DeleteMcpProject--->>>", logger.Error(err))
		return nil, err
	}

	return nil, nil
}
