package client

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"google.golang.org/grpc"
)

type ServiceManagerI interface {
	BuilderProject() nb.BuilderProjectServiceClient
}

type grpcClients struct {
	builderProjectService nb.BuilderProjectServiceClient
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {
	connNewObjectBuilderService, err := grpc.Dial(
		cfg.ServiceHost+cfg.ServicePort,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClients{
		builderProjectService: nb.NewBuilderProjectServiceClient(connNewObjectBuilderService),
	}, nil
}

func (g *grpcClients) BuilderProject() nb.BuilderProjectServiceClient {
	return g.builderProjectService
}
