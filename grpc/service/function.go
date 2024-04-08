package service

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
)

type functionService struct {
	cfg config.Config
	log logger.LoggerI
	// strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedFunctionServiceV2Server
}

func NewFunctionService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *functionService { // strg storage.StorageI,
	return &functionService{
		cfg: cfg,
		log: log,
		// strg:     strg,
		services: svcs,
	}
}

// func (f *functionService) Create(ctx context.Context, req *nb.CreateFunctionRequest) (resp *nb.Function, err error)

// func (f *functionService) GetList(ctx context.Context, req *nb.GetAllFunctionsRequest) (resp *nb.GetAllFunctionsResponse, err error)

// func (f *functionService) GetSingle(ctx context.Context, req *nb.FunctionPrimaryKey) (resp *nb.Function, err error)

// func (f *functionService) Update(ctx context.Context, req *nb.Function) (resp *empty.Empty, err error)

// func (f *functionService) Delete(ctx context.Context, req *nb.FunctionPrimaryKey) (resp *empty.Empty, err error)
