package service

import (
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type customErrorMessageService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedCustomErrorMessageServiceServer
}

func NewCustomErrorMessageService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *customErrorMessageService { // ,
	return &customErrorMessageService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

// func (f *customErrorMessageService) Create(ctx context.Context, req *nb.CreateCustomErrorMessage) (resp *nb.CustomErrorMessage, err error) {

// 	f.log.Info("---CreateCustomErrorMessage--->>>", logger.Any("req", req))

// 	resp, err = f.strg.CustomErrorMessage().Create(ctx, req)
// 	if err != nil {
// 		f.log.Error("---CreateCustomErrorMessage--->>>", logger.Error(err))
// 		return &nb.CustomErrorMessage{}, err
// 	}

// 	return resp, nil
// }

// func (f *customErrorMessageService) GetSingle(ctx context.Context, req *nb.CustomErrorMessagePK) (resp *nb.CustomErrorMessage, err error) {

// 	f.log.Info("---GetByIDCustomErrorMessage--->>>", logger.Any("req", req))

// 	resp, err = f.strg.CustomErrorMessage().GetSingle(ctx, req)
// 	if err != nil {
// 		f.log.Error("---GetByIDCustomErrorMessage--->>>", logger.Error(err))
// 		return &nb.CustomErrorMessage{}, err
// 	}

// 	return resp, nil
// }

// func (f *customErrorMessageService) GetList(ctx context.Context, req *nb.GetCustomErrorMessageListRequest) (resp *nb.GetCustomErrorMessageListResponse, err error) {

// 	f.log.Info("---GetAllCustomErrorMessage--->>>", logger.Any("req", req))

// 	resp, err = f.strg.CustomErrorMessage().GetList(ctx, req)
// 	if err != nil {
// 		f.log.Error("---GetAllCustomErrorMessage--->>>", logger.Error(err))
// 		return &nb.GetCustomErrorMessageListResponse{}, err
// 	}

// 	return resp, nil
// }

// func (f *customErrorMessageService) Update(ctx context.Context, req *nb.CustomErrorMessage) (resp *emptypb.Empty, err error) {
// 	f.log.Info("---UpdateCustomErrorMessage--->>>", logger.Any("req", req))

// 	err = f.strg.CustomErrorMessage().Update(ctx, req)
// 	if err != nil {
// 		f.log.Error("---UpdateCustomErrorMessage--->>>", logger.Error(err))
// 		return &emptypb.Empty{}, err
// 	}

// 	return resp, nil
// }

// func (f *customErrorMessageService) Delete(ctx context.Context, req *nb.CustomErrorMessagePK) (resp *emptypb.Empty, err error) {
// 	f.log.Info("---DeleteCustomErrorMessage--->>>", logger.Any("req", req))

// 	err = f.strg.CustomErrorMessage().Delete(ctx, req)
// 	if err != nil {
// 		f.log.Error("---DeleteCustomErrorMessage--->>>", logger.Error(err))
// 		return &emptypb.Empty{}, err
// 	}

// 	return &emptypb.Empty{}, nil
// }

// func (f *customErrorMessageService) GetListForObject(ctx context.Context, req *nb.GetListForObjectRequest) (resp *nb.GetCustomErrorMessageListResponse, err error) {

// 	f.log.Info("---GetListForObject Cus Err Message--->>>", logger.Any("req", req))

// 	resp, err = f.strg.CustomErrorMessage().GetListForObject(ctx, req)
// 	if err != nil {
// 		f.log.Error("---GetListForObject Cus Err Message--->>>", logger.Error(err))
// 		return &nb.GetCustomErrorMessageListResponse{}, err
// 	}

// 	return resp, nil
// }
