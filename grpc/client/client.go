package client

import (
	"fmt"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/auth_service"
	"ucode/ucode_go_object_builder_service/genproto/company_service"
	"ucode/ucode_go_object_builder_service/genproto/transcoder_service"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServiceManagerI interface {
	UserService() auth_service.UserServiceClient
	SyncUserService() auth_service.SyncUserServiceClient
	ResourceService() company_service.ResourceServiceClient
	TranscoderService() transcoder_service.PipelineServiceClient
}

type grpcClients struct {
	userService       auth_service.UserServiceClient
	syncUserService   auth_service.SyncUserServiceClient
	resourceService   company_service.ResourceServiceClient
	transcoderService transcoder_service.PipelineServiceClient
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {
	connAuthService, err := grpc.NewClient(
		fmt.Sprintf("%s%s", cfg.AuthServiceHost, cfg.AuthGRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		return nil, err
	}

	connCompanyService, err := grpc.NewClient(
		fmt.Sprintf("%s%s", cfg.CompanyServiceHost, cfg.CompanyServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(52428800), grpc.MaxCallSendMsgSize(52428800)),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		return nil, err
	}

	fmt.Println("Transcoder->", cfg.TranscoderServiceHost, cfg.TranscoderServicePort)
	connTranscoderService, err := grpc.NewClient(
		fmt.Sprintf("%s%s", cfg.TranscoderServiceHost, cfg.TranscoderServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClients{
		userService:       auth_service.NewUserServiceClient(connAuthService),
		syncUserService:   auth_service.NewSyncUserServiceClient(connAuthService),
		resourceService:   company_service.NewResourceServiceClient(connCompanyService),
		transcoderService: transcoder_service.NewPipelineServiceClient(connTranscoderService),
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

func (g *grpcClients) TranscoderService() transcoder_service.PipelineServiceClient {
	return g.transcoderService
}
