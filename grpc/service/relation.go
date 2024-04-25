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

type relationService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedRelationServiceServer
}

func NewRelationService(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI, strg storage.StorageI) *relationService {
	return &relationService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (r *relationService) Create(ctx context.Context, req *nb.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error) {
	resp = &nb.CreateRelationRequest{}

	r.log.Info("---CreateRelation--->", logger.Any("req", req))

	resp, err = r.strg.Relation().Create(ctx, req)
	if err != nil {
		r.log.Error("---CreateRelation--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (r *relationService) GetByID(ctx context.Context, req *nb.RelationPrimaryKey) (resp *nb.RelationForGetAll, err error) {

	r.log.Info("---GetSingleRelation--->", logger.Any("req", req))

	resp, err = r.strg.Relation().GetByID(ctx, req)
	if err != nil {
		r.log.Error("---GetSingleRelation--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (r *relationService) GetAll(ctx context.Context, req *nb.GetAllRelationsRequest) (resp *nb.GetAllRelationsResponse, err error) {
	resp = &nb.GetAllRelationsResponse{}

	r.log.Info("---ListRelations--->", logger.Any("req", req))

	resp, err = r.strg.Relation().GetList(ctx, req)
	if err != nil {
		r.log.Error("---ListRelations--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (r *relationService) Update(ctx context.Context, req *nb.UpdateRelationRequest) (resp *nb.RelationForGetAll, err error) {
	r.log.Info("---UpdateRelation--->", logger.Any("req", req))

	resp, err = r.strg.Relation().Update(ctx, req)
	if err != nil {
		r.log.Error("---UpdateRelation--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (r *relationService) Delete(ctx context.Context, req *nb.RelationPrimaryKey) (resp *emptypb.Empty, err error) {
	r.log.Info("---DeleteRelation--->", logger.Any("req", req))

	err = r.strg.Relation().Delete(ctx, req)
	if err != nil {
		r.log.Error("---DeleteRelation--->>>", logger.Error(err))
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}
