package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	span "ucode/ucode_go_object_builder_service/pkg/jaeger"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/protobuf/types/known/emptypb"
)

type languageService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedLanguageServiceServer
}

func NewLanguageService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *languageService {
	return &languageService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (l *languageService) Create(ctx context.Context, req *nb.CreateLanguageRequest) (resp *nb.Language, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_language.Create", req)
	defer dbSpan.Finish()

	l.log.Info("---Create Language--->>>", logger.Any("req", req))

	resp, err = l.strg.Language().Create(ctx, req)
	if err != nil {
		l.log.Error("---Create Language--->>>", logger.Error(err))
		return &nb.Language{}, err
	}

	return resp, nil
}

func (l *languageService) GetById(ctx context.Context, req *nb.PrimaryKey) (resp *nb.Language, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_language.GetById", req)
	defer dbSpan.Finish()

	l.log.Info("---GetById Language--->>>", logger.Any("req", req))

	resp, err = l.strg.Language().GetById(ctx, req)
	if err != nil {
		l.log.Error("---GetById Language--->>>", logger.Error(err))
		return &nb.Language{}, err
	}

	return resp, nil
}

func (l *languageService) GetList(ctx context.Context, req *nb.GetListLanguagesRequest) (resp *nb.GetListLanguagesResponse, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_language.GetList", req)
	defer dbSpan.Finish()

	l.log.Info("---GetList Language--->>>", logger.Any("req", req))

	resp, err = l.strg.Language().GetList(ctx, req)
	if err != nil {
		l.log.Error("---GetList Language--->>>", logger.Error(err))
		return &nb.GetListLanguagesResponse{}, err
	}

	return resp, nil
}

func (l *languageService) Update(ctx context.Context, req *nb.UpdateLanguageRequest) (resp *nb.Language, err error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_language.Update", req)
	defer dbSpan.Finish()

	l.log.Info("---Update Language--->>>", logger.Any("req", req))

	resp, err = l.strg.Language().UpdateLanguage(ctx, req)
	if err != nil {
		l.log.Error("---Update Language--->>>", logger.Error(err))
		return &nb.Language{}, err
	}

	return resp, nil
}

func (l *languageService) Delete(ctx context.Context, req *nb.PrimaryKey) (*emptypb.Empty, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_language.Delete", req)
	defer dbSpan.Finish()

	l.log.Info("---Delete Language--->>>", logger.Any("req", req))

	err := l.strg.Language().Delete(ctx, req)
	if err != nil {
		l.log.Error("---Delete Language--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
