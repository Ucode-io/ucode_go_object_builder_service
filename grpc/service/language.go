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
		l.log.Error("---Update--->>>", logger.Error(err))
		return &nb.Language{}, err
	}

	return resp, nil
}
