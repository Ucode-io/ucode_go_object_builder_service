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

type layoutService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedLayoutServiceServer
}

// mustEmbedUnimplementedLayoutServiceServer implements new_object_builder_service.LayoutServiceServer.
// func (f *layoutService) mustEmbedUnimplementedLayoutServiceServer() {
// 	panic("unimplemented")
// }

func NewLayoutService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *layoutService { // ,
	return &layoutService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (f *layoutService) Update(ctx context.Context, req *nb.LayoutRequest) (resp *nb.LayoutResponse, err error) {

	f.log.Info("---UpdateLayout--->>>", logger.Any("req", req))

	resp, err = f.strg.Layout().Update(ctx, req)
	if err != nil {
		f.log.Error("---UpdateLayout--->>>", logger.Error(err))
		return &nb.LayoutResponse{}, err
	}

	return resp, nil
}

// CreateAll implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) CreateAll(context.Context, *nb.CreateLayoutRequest) (*nb.GetListLayoutResponse, error) {
	return nil, nil
}

// GetAll implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetAll(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error) {
	f.log.Info("---GetAll--->>>", logger.Any("req", req))

	resp, err = f.strg.Layout().GetAllV2(ctx, req)
	if err != nil {
		f.log.Error("---GetAll--->>>", logger.Error(err))
		return &nb.GetListLayoutResponse{}, err
	}

	return resp, nil

}

// GetByID implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetByID(ctx context.Context, req *nb.LayoutPrimaryKey) (resp *nb.LayoutResponse, err error) {

	f.log.Info("---GetSingleLayout--->>>", logger.Any("req", req))

	resp, err = f.strg.Layout().GetByID(ctx, req)
	if err != nil {
		f.log.Error("---GetSingleLayout--->>>", logger.Error(err))
		return &nb.LayoutResponse{}, err
	}

	return resp, nil

}

// GetSingleLayout implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetSingleLayout(ctx context.Context, req *nb.GetSingleLayoutRequest) (resp *nb.LayoutResponse, err error) {

	f.log.Info("---GetSingleLayout--->>>", logger.Any("req", req))

	resp, err = f.strg.Layout().GetSingleLayout(ctx, req)
	if err != nil {
		f.log.Error("---GetSingleLayout--->>>", logger.Error(err))
		return &nb.LayoutResponse{}, err
	}

	return resp, nil

}

func (f *layoutService) RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) (*emptypb.Empty, error) {
	f.log.Info("---RemvoeLayout--->>>", logger.Any("req", req))

	err := f.strg.Layout().RemoveLayout(ctx, req)
	if err != nil {
		f.log.Error("---RemoveLayout--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil

}
