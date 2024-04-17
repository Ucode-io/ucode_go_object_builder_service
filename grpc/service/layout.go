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
	nb.UnimplementedFileServiceServer
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

	f.log.Info("---CreateFile--->>>", logger.Any("req", req))

	resp, err = f.strg.Layout().Update(ctx, req)
	if err != nil {
		f.log.Error("---CreateFile--->>>", logger.Error(err))
		return &nb.LayoutResponse{}, err
	}

	return resp, nil
}

// CreateAll implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) CreateAll(context.Context, *nb.CreateLayoutRequest) (*nb.GetListLayoutResponse, error) {
	return nil, nil
}

// GetAll implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetAll(context.Context, *nb.GetListLayoutRequest) (*nb.GetListLayoutResponse, error) {
	return nil, nil

}

// GetByID implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetByID(context.Context, *nb.LayoutPrimaryKey) (*nb.LayoutResponse, error) {
	return nil, nil

}

// GetSingleLayout implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) GetSingleLayout(context.Context, *nb.GetSingleLayoutRequest) (*nb.LayoutResponse, error) {
	return nil, nil

}

// RemoveLayout implements new_object_builder_service.LayoutServiceServer.
func (f *layoutService) RemoveLayout(context.Context, *nb.LayoutPrimaryKey) (*emptypb.Empty, error) {
	return nil, nil

}
