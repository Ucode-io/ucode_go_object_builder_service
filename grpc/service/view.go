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

type viewService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedViewServiceServer
}

func NewViewService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *viewService { // ,
	return &viewService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (f *viewService) Create(ctx context.Context, req *nb.CreateViewRequest) (resp *nb.View, err error) {

	f.log.Info("---CreateView--->>>", logger.Any("req", req))

	resp, err = f.strg.View().Create(ctx, req)
	if err != nil {
		f.log.Error("---CreateView--->>>", logger.Error(err))
		return &nb.View{}, err
	}

	return resp, nil
}

func (f *viewService) GetSingle(ctx context.Context, req *nb.ViewPrimaryKey) (resp *nb.View, err error) {

	f.log.Info("---GetByIDView--->>>", logger.Any("req", req))

	resp, err = f.strg.View().GetSingle(ctx, req)
	if err != nil {
		f.log.Error("---GetByIDView--->>>", logger.Error(err))
		return &nb.View{}, err
	}

	return resp, nil
}

func (f *viewService) GetList(ctx context.Context, req *nb.GetAllViewsRequest) (resp *nb.GetAllViewsResponse, err error) {

	f.log.Info("---GetAllView--->>>", logger.Any("req", req))

	resp, err = f.strg.View().GetList(ctx, req)
	if err != nil {
		f.log.Error("---GetAllView--->>>", logger.Error(err))
		return &nb.GetAllViewsResponse{}, err
	}

	return resp, nil
}

func (f *viewService) Update(ctx context.Context, req *nb.View) (resp *nb.View, err error) {
	f.log.Info("---UpdateView--->>>", logger.Any("req", req))

	resp, err = f.strg.View().Update(ctx, req)
	if err != nil {
		f.log.Error("---UpdateView--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (f *viewService) Delete(ctx context.Context, req *nb.ViewPrimaryKey) (resp *emptypb.Empty, err error) {
	f.log.Info("---DeleteView--->>>", logger.Any("req", req))

	err = f.strg.View().Delete(ctx, req)
	if err != nil {
		f.log.Error("---DeleteView--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
func (f *viewService) UpdateViewOrder(ctx context.Context, req *nb.UpdateViewOrderRequest) (resp *emptypb.Empty, err error) {
	f.log.Info("---UpdateViewOrder--->>>", logger.Any("req", req))

	err = f.strg.View().UpdateViewOrder(ctx, req)
	if err != nil {
		f.log.Error("---UpdateViewOrder--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
