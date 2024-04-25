package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	pa "ucode/ucode_go_object_builder_service/genproto/auth_service"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/spf13/cast"
)

type itemsService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedItemsServiceServer
}

func NewItemsService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *itemsService {
	return &itemsService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (i *itemsService) Create(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	i.log.Info("---CreateItems--->>>", logger.Any("req", req))

	resp, err = i.strg.Items().Create(ctx, req)
	if err != nil {
		i.log.Error("---CreateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	data, err := helper.ConvertStructToMap(resp.Data)
	if err != nil {
		i.log.Error("---CreateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}
	authInfo := cast.ToStringMap(data["authInfo"])

	if cast.ToBool(data["create_user"]) {
		_, err = i.services.UserService().CreateUser(ctx, &pa.CreateUserRequest{
			ClientTypeId:          cast.ToString(data["client_type_id"]),
			RoleId:                cast.ToString(data["role_id"]),
			Login:                 cast.ToString(data[cast.ToString(authInfo["login"])]),
			Email:                 cast.ToString(data[cast.ToString(authInfo["email"])]),
			Phone:                 cast.ToString(data[cast.ToString(authInfo["phone"])]),
			ProjectId:             cast.ToString(data["company_service_project_id"]),
			CompanyId:             cast.ToString(data["company_service_company_id"]),
			Password:              cast.ToString(data[cast.ToString(authInfo["password"])]),
			ResourceEnvironmentId: req.ProjectId,
			Invite:                cast.ToBool(data["invite"]),
			EnvironmentId:         cast.ToString(data["company_service_environment_id"]),
		})
		if err != nil {
			i.log.Error("---CreateItems--->>>", logger.Error(err))
			return &nb.CommonMessage{}, err
		}
	}

	return resp, nil
}
