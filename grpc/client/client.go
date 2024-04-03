package client

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"google.golang.org/grpc"
)

type ServiceManagerI interface {
	BuilderProject() nb.BuilderProjectServiceClient
	Field() nb.FieldServiceClient
}

type grpcClients struct {
	builderProjectService nb.BuilderProjectServiceClient
	fieldService          nb.FieldServiceClient
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
		fieldService:          nb.NewFieldServiceClient(connNewObjectBuilderService),
	}, nil
}

func (g *grpcClients) BuilderProject() nb.BuilderProjectServiceClient {
	return g.builderProjectService
}

func (g *grpcClients) Field() nb.FieldServiceClient {
	return g.fieldService
}
