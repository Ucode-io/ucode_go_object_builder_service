package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"google.golang.org/protobuf/types/known/emptypb"
)

type fieldService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedFieldServiceServer
}

func NewFieldService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *fieldService { // ,
	return &fieldService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (f *fieldService) Create(ctx context.Context, req *nb.CreateFieldRequest) (resp *nb.Field, err error) {

	f.log.Info("---CreateField--->>>", logger.Any("req", req))

	resp, err = f.strg.Field().Create(ctx, req)
	if err != nil {
		f.log.Error("---CreateField--->>>", logger.Error(err))
		return &nb.Field{}, err
	}

	return resp, nil
}

func (f *fieldService) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error) {

	f.log.Info("---GetByIDField--->>>", logger.Any("req", req))

	resp, err = f.strg.Field().GetByID(ctx, req)
	if err != nil {
		f.log.Error("---GetByIDField--->>>", logger.Error(err))
		return &nb.Field{}, err
	}

	return resp, nil
}

func (f *fieldService) GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error) {

	f.log.Info("---GetAllField--->>>", logger.Any("req", req))

	resp, err = f.strg.Field().GetAll(ctx, req)
	if err != nil {
		f.log.Error("---GetAllField--->>>", logger.Error(err))
		return &nb.GetAllFieldsResponse{}, err
	}

	return resp, nil
}

func (f *fieldService) GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error) {
	return &nb.AllFields{}, nil
}

func (f *fieldService) Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error) {
	f.log.Info("---UpdateField--->>>", logger.Any("req", req))

	resp, err = f.strg.Field().Update(ctx, req)
	if err != nil {
		f.log.Error("---UpdateField--->>>", logger.Error(err))
		return &nb.Field{}, err
	}

	return resp, nil
}

func (f *fieldService) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) (resp *emptypb.Empty, err error) {
	f.log.Info("---UpdateSearchField--->>>", logger.Any("req", req))

	err = f.strg.Field().UpdateSearch(ctx, req)
	if err != nil {
		f.log.Error("---UpdateSearchField--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (f *fieldService) Delete(ctx context.Context, req *nb.FieldPrimaryKey) (resp *emptypb.Empty, err error) {
	f.log.Info("---DeleteField--->>>", logger.Any("req", req))

	err = f.strg.Field().Delete(ctx, req)
	if err != nil {
		f.log.Error("---DeleteField--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
