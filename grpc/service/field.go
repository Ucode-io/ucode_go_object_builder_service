package service

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
)

type fieldService struct {
	cfg config.Config
	log logger.LoggerI
	// strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedFieldServiceServer
}

func NewFieldService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *fieldService { // strg storage.StorageI,
	return &fieldService{
		cfg: cfg,
		log: log,
		// strg:     strg,
		services: svcs,
	}
}

// func (f *fieldService) Create(ctx context.Context, req *nb.CreateFieldRequest) (resp *nb.Field, err error)

// func (f *fieldService) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error)

// func (f *fieldService) GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error)

// func (f *fieldService) GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error)

// func (f *fieldService) Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error)

// func (f *fieldService) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) (resp *empty.Empty, err error)

// func (f *fieldService) Delete(ctx context.Context, req *nb.FieldPrimaryKey) (resp *empty.Empty, err error)
