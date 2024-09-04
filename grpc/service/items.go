package service

import (
	"context"
	"ucode/ucode_go_object_builder_service/config"
	pa "ucode/ucode_go_object_builder_service/genproto/auth_service"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/models"
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
		i.log.Error("---CreateItems--->>> !!!", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	data, err := helper.ConvertStructToMap(resp.Data)
	if err != nil {
		i.log.Error("---CreateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	reqData, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		i.log.Error("---CreateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	authInfo := cast.ToStringMap(data["authInfo"])

	if cast.ToBool(data["create_user"]) {

		user, err := i.services.SyncUserService().CreateUser(ctx, &pa.CreateSyncUserRequest{
			ClientTypeId:          cast.ToString(data["client_type_id"]),
			RoleId:                cast.ToString(data["role_id"]),
			Login:                 cast.ToString(data[cast.ToString(authInfo["login"])]),
			Email:                 cast.ToString(data[cast.ToString(authInfo["email"])]),
			Phone:                 cast.ToString(data[cast.ToString(authInfo["phone"])]),
			ProjectId:             cast.ToString(reqData["company_service_project_id"]),
			Password:              cast.ToString(data[cast.ToString(authInfo["password"])]),
			ResourceEnvironmentId: req.ProjectId,
			Invite:                cast.ToBool(data["invite"]),
			EnvironmentId:         cast.ToString(reqData["company_service_environment_id"]),
		})
		if err != nil {
			i.log.Error("---CreateItems--->>>", logger.Error(err))
			return &nb.CommonMessage{}, err
		}

		err = i.strg.Items().UpdateGuid(ctx, &models.ItemsChangeGuid{
			TableSlug: req.TableSlug,
			ProjectId: req.ProjectId,
			OldId:     cast.ToString(data["guid"]),
			NewId:     user.UserId,
		})
		if err != nil {
			i.log.Error("---UpdateGuid--->>>", logger.Error(err))
			return &nb.CommonMessage{}, err
		}
	}

	return resp, nil
}

func (i *itemsService) GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	i.log.Info("---GetSingleItems--->>>", logger.Any("req", req))

	resp, err = i.strg.Items().GetSingle(ctx, req)
	if err != nil {
		i.log.Error("---GetSingleItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	return resp, nil
}

func (i *itemsService) Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	i.log.Info("---UpdateItems--->>>", logger.Any("req", req))

	resp, err = i.strg.Items().Update(ctx, req)
	if err != nil {
		i.log.Error("---UpdateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	return resp, nil
}

func (i *itemsService) Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	i.log.Info("---DeleteItems--->>>", logger.Any("req", req))

	resp, err = i.strg.Items().Delete(ctx, req)
	if err != nil {
		i.log.Error("---DeleteItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		i.log.Error("---DeleteItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	_, delete := data["delete_user"]
	_, login := data["login_data"]

	if delete && login {

		_, err = i.services.UserService().DeleteUser(ctx, &pa.UserPrimaryKey{
			Id:                    cast.ToString(data["guid"]),
			ClientTypeId:          cast.ToString(data["client_type_id"]),
			ProjectId:             cast.ToString(data["company_service_project_id"]),
			CompanyId:             cast.ToString(data["company_service_company_id"]),
			ResourceEnvironmentId: req.ProjectId,
			// EnvironmentId:         cast.ToString(data["company_service_environment_id"]),
		})
		if err != nil {
			i.log.Error("---CreateItems--->>>", logger.Error(err))
			return &nb.CommonMessage{}, err
		}
	}

	return resp, nil
}

func (i *itemsService) DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	i.log.Info("---DeleteItems--->>>", logger.Any("req", req))

	delete, err := i.strg.Items().DeleteMany(ctx, req)
	if err != nil {
		i.log.Error("---DeleteItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	if delete.IsDelete {
		_, err = i.services.SyncUserService().DeleteManyUser(ctx, &pa.DeleteManyUserRequest{
			Users:         delete.Users,
			ProjectId:     delete.ProjectId,
			EnvironmentId: delete.EnvironmentId,
		})
		if err != nil {
			i.log.Error("---CreateItems--->>>", logger.Error(err))
			return &nb.CommonMessage{}, err
		}
	}

	return &nb.CommonMessage{}, err
}

func (i *itemsService) MultipleUpdate(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	i.log.Info("---MultipleUpdateItems--->>>", logger.Any("req", req))

	resp, err = i.strg.Items().MultipleUpdate(ctx, req)
	if err != nil {
		i.log.Error("---MultipleUpdateItems--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	return resp, nil
}

func (i *itemsService) UpsertMany(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	i.log.Info("---UpsertMany--->>>", logger.Any("req", req))

	if err = i.strg.Items().UpsertMany(ctx, req); err != nil {
		i.log.Error("---UpsertMany--->>>", logger.Error(err))
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{}, nil
}
