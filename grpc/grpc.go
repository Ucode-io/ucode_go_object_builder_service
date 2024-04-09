package grpc

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/storage"

	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/grpc/service"
	"ucode/ucode_go_object_builder_service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func SetUpServer(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) (grpcServer *grpc.Server) { // ,
	grpcServer = grpc.NewServer()

	nb.RegisterBuilderProjectServiceServer(grpcServer, service.NewBuilderProjectService(cfg, log, svcs))
	nb.RegisterFieldServiceServer(grpcServer, service.NewFieldService(cfg, log, svcs, strg))
	nb.RegisterFunctionServiceV2Server(grpcServer, service.NewFunctionService(cfg, log, svcs, strg))
	nb.RegisterTableServiceServer(grpcServer, service.NewTableService(cfg, log, svcs, strg))

	reflection.Register(grpcServer)
	return
}
