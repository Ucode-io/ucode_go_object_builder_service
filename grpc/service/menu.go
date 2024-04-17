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

type menuService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedMenuServiceServer
}

func NewMenuService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *menuService {
	return &menuService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (f *menuService) Create(ctx context.Context, req *nb.CreateMenuRequest) (resp *nb.Menu, err error) {

	f.log.Info("---CreateMenu--->>>", logger.Any("req", req))

	resp, err = f.strg.Menu().Create(ctx, req)
	if err != nil {
		f.log.Error("---CreateMenu--->>>", logger.Error(err))
		return &nb.Menu{}, err
	}

	return resp, nil
}

func (f *menuService) GetByID(ctx context.Context, req *nb.MenuPrimaryKey) (resp *nb.Menu, err error) {

	f.log.Info("---GetByIDMenu--->>>", logger.Any("req", req))

	resp, err = f.strg.Menu().GetById(ctx, req)
	if err != nil {
		f.log.Error("---GetByIDMenu--->>>", logger.Error(err))
		return &nb.Menu{}, err
	}

	return resp, nil
}

func (f *menuService) GetAll(ctx context.Context, req *nb.GetAllMenusRequest) (resp *nb.GetAllMenusResponse, err error) {

	f.log.Info("---GetAllView--->>>", logger.Any("req", req))

	resp, err = f.strg.Menu().GetAll(ctx, req)
	if err != nil {
		f.log.Error("---GetAllMenusResponse--->>>", logger.Error(err))
		return &nb.GetAllMenusResponse{}, err
	}

	return resp, nil
}

func (f *menuService) Update(ctx context.Context, req *nb.Menu) (resp *nb.Menu, err error) {
	f.log.Info("---UpdateMenu--->>>", logger.Any("req", req))

	resp, err = f.strg.Menu().Update(ctx, req)
	if err != nil {
		f.log.Error("---UpdateMenu--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (f *menuService) UpdateMenuOrder(ctx context.Context, req *nb.UpdateMenuOrderRequest) (resp *emptypb.Empty, err error) {
	f.log.Info("---UpdateMenuOrder--->>>", logger.Any("req", req))

	err = f.strg.Menu().UpdateMenuOrder(ctx, req)
	if err != nil {
		f.log.Error("---UpdateMenuOrder--->>>", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (f *menuService) GetAllMenuSettings(ctx context.Context, req *nb.GetAllMenuSettingsRequest) (resp *nb.GetAllMenuSettingsResponse, err error) {

	f.log.Info("---GetAllView--->>>", logger.Any("req", req))

	resp, err = f.strg.Menu().GetAllMenuSettings(ctx, req)
	if err != nil {
		f.log.Error("---GetAllMenusResponse--->>>", logger.Error(err))
		return &nb.GetAllMenuSettingsResponse{}, err
	}

	return resp, nil
}

func (f *menuService) Delete(ctx context.Context, req *nb.MenuPrimaryKey) (resp *emptypb.Empty, err error) {
	f.log.Info("---DeleteMenu--->>>", logger.Any("req", req))

	err = f.strg.Menu().Delete(ctx, req)
	if err != nil {
		f.log.Error("---DeleteMenu--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
