package client

import (
	"fmt"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/auth_service"
	"ucode/ucode_go_object_builder_service/genproto/company_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceManagerI interface {
	UserService() auth_service.UserServiceClient
	ResourceService() company_service.ResourceServiceClient
	SyncUserService() auth_service.SyncUserServiceClient
}

type grpcClients struct {
	userService     auth_service.UserServiceClient
	resourceService company_service.ResourceServiceClient
	syncUserService auth_service.SyncUserServiceClient
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {

	connAuthService, err := grpc.Dial(
		fmt.Sprintf("%s%s", cfg.AuthServiceHost, cfg.AuthGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
	)
	if err != nil {
		return nil, err
	}

	connCompanyService, err := grpc.Dial(
		fmt.Sprintf("%s%s", cfg.CompanyServiceHost, cfg.CompanyServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClients{
		userService:     auth_service.NewUserServiceClient(connAuthService),
		resourceService: company_service.NewResourceServiceClient(connCompanyService),
		syncUserService: auth_service.NewSyncUserServiceClient(connAuthService),
	}, nil
}

func (g *grpcClients) UserService() auth_service.UserServiceClient {
	return g.userService
}

func (g *grpcClients) ResourceService() company_service.ResourceServiceClient {
	return g.resourceService
}

func (g *grpcClients) SyncUserService() auth_service.SyncUserServiceClient {
	return g.syncUserService
}
