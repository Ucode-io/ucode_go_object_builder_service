package client

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"google.golang.org/grpc"
)

type ServiceManagerI interface {
	BuilderProject() nb.BuilderProjectServiceClient
	Field() nb.FieldServiceClient
	Funciton() nb.FunctionServiceV2Client
}

type grpcClients struct {
	builderProjectService nb.BuilderProjectServiceClient
	fieldService          nb.FieldServiceClient
	functionService       nb.FunctionServiceV2Client
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
		functionService:       nb.NewFunctionServiceV2Client(connNewObjectBuilderService),
	}, nil
}

func (g *grpcClients) BuilderProject() nb.BuilderProjectServiceClient {
	return g.builderProjectService
}

func (g *grpcClients) Field() nb.FieldServiceClient {
	return g.fieldService
}

func (g *grpcClients) Funciton() nb.FunctionServiceV2Client {
	return g.functionService
}
