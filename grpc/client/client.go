package client

import (
	"fmt"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/auth_service"

	"google.golang.org/grpc"
)

type ServiceManagerI interface {
	UserService() auth_service.UserServiceClient
}

type grpcClients struct {
	userService auth_service.UserServiceClient
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {

	connAuthService, err := grpc.Dial(
		fmt.Sprintf("%s%s", cfg.AuthServiceHost, cfg.AuthGRPCPort),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClients{
		userService: auth_service.NewUserServiceClient(connAuthService),
	}, nil
}

func (g *grpcClients) UserService() auth_service.UserServiceClient {
	return g.userService
}
