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

type loginService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedLoginServiceServer
}

func NewLoginService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *loginService {
	return &loginService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (l *loginService) LoginData(ctx context.Context, req *nb.LoginDataReq) (resp *nb.LoginDataRes, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_login.LoginData", req)
	defer dbSpan.Finish()

	l.log.Info("---LoginData--->>>", logger.Any("req", req))

	resp, err = l.strg.Login().LoginData(ctx, req)
	if err != nil {
		l.log.Error("---LoginData--->>>", logger.Error(err))
		return &nb.LoginDataRes{}, err
	}

	return resp, nil
}

func (l *loginService) GetConnetionOptions(ctx context.Context, req *nb.GetConnetionOptionsRequest) (resp *nb.GetConnectionOptionsResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_login.GetConnetionOptions", req)
	defer dbSpan.Finish()

	l.log.Info("---GetConnetionOptions--->>>", logger.Any("req", req))

	resp, err = l.strg.Login().GetConnectionOptions(ctx, req)
	if err != nil {
		l.log.Error("---GetConnectionOptions--->>>", logger.Error(err))
		return
	}

	return
}
