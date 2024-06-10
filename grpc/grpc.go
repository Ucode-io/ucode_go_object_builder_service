package grpc

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/grpc/service"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func SetUpServer(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) (grpcServer *grpc.Server) { // ,
	grpcServer = grpc.NewServer()

	project := service.NewBuilderProjectService(strg, cfg, log, svcs)
	err := project.AutoConnect(context.Background())
	if err != nil {
		logger.Any("project.AutoConnect", logger.Error(err))
	}
	nb.RegisterBuilderProjectServiceServer(grpcServer, project)
	nb.RegisterFieldServiceServer(grpcServer, service.NewFieldService(cfg, log, svcs, strg))
	nb.RegisterFunctionServiceV2Server(grpcServer, service.NewFunctionService(cfg, log, svcs, strg))
	nb.RegisterTableServiceServer(grpcServer, service.NewTableService(cfg, log, svcs, strg))
	nb.RegisterFileServiceServer(grpcServer, service.NewFileService(cfg, log, svcs, strg))
	nb.RegisterViewServiceServer(grpcServer, service.NewViewService(cfg, log, svcs, strg))
	nb.RegisterCustomErrorMessageServiceServer(grpcServer, service.NewCustomErrorMessageService(cfg, log, svcs, strg))
	nb.RegisterObjectBuilderServiceServer(grpcServer, service.NewObjectBuilderService(strg, cfg, log, svcs))
	nb.RegisterLoginServiceServer(grpcServer, service.NewLoginService(cfg, log, svcs, strg))
	nb.RegisterMenuServiceServer(grpcServer, service.NewMenuService(cfg, log, svcs, strg))
	nb.RegisterLayoutServiceServer(grpcServer, service.NewLayoutService(cfg, log, svcs, strg))
	nb.RegisterSectionServiceServer(grpcServer, service.NewSectionService(cfg, log, svcs, strg))
	nb.RegisterItemsServiceServer(grpcServer, service.NewItemsService(cfg, log, svcs, strg))
	nb.RegisterRelationServiceServer(grpcServer, service.NewRelationService(cfg, log, svcs, strg))
	nb.RegisterPermissionServiceServer(grpcServer, service.NewPermissionService(cfg, log, svcs, strg))
	nb.RegisterExcelServiceServer(grpcServer, service.NewExcelService(cfg, log, svcs, strg))
	nb.RegisterVersionServiceServer(grpcServer, service.NewVersionService(cfg, log, svcs, strg))
	nb.RegisterCustomEventServiceServer(grpcServer, service.NewCustomEventService(cfg, log, svcs, strg))
	nb.RegisterVersionHistoryServiceServer(grpcServer, service.NewVersionHistoryService(cfg, log, svcs, strg))

	reflection.Register(grpcServer)
	return
}
