package service

import (
	"context"
	"fmt"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/protobuf/types/known/emptypb"
)

type builderProjectService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedBuilderProjectServiceServer
}

func NewBuilderProjectService(strg storage.StorageI, cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *builderProjectService { // strg storage.StorageI,
	return &builderProjectService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (b *builderProjectService) Register(ctx context.Context, req *nb.RegisterProjectRequest) (resp *emptypb.Empty, err error) {
	b.log.Info("!!!RegisterProject--->", logger.Any("req", req))

	if req.UserId == "" {
		err = fmt.Errorf("error user_id is required")
		b.log.Error("!!!RegisterProjectError--->", logger.Error(err))
		return resp, err
	}

	if req.ProjectId == "" {
		err = fmt.Errorf("error project_id is required")
		b.log.Error("!!!RegisterProjectError--->", logger.Error(err))
		return resp, err
	}

	err = b.strg.BuilderProject().Register(ctx, req)
	if err != nil {
		b.log.Error("!!!RegisterProjectError--->", logger.Error(err))
		return resp, err

	}

	return resp, nil
}

// func (b *builderProjectService) RegisterProjects(ctx context.Context, req *nb.RegisterProjectRequest) (resp *empty.Empty, err error)

// func (b *builderProjectService) Deregister(ctx context.Context, req *nb.DeregisterProjectRequest) (resp *empty.Empty, err error)

// func (b *builderProjectService) Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) (resp *empty.Empty, err error)

// func (b *builderProjectService) RegisterMany(ctx context.Context, req *nb.RegisterManyProjectsRequest) (resp *nb.RegisterManyProjectsResponse, err error)

// func (b *builderProjectService) DeregisterMany(ctx context.Context, req *nb.DeregisterManyProjectsRequest) (resp *nb.DeregisterManyProjectsResponse, err error)
