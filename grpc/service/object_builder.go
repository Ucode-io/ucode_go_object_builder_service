package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"
)

type objectBuilderService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedObjectBuilderServiceServer
}

func NewObjectBuilderService(strg storage.StorageI, cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *objectBuilderService {
	return &objectBuilderService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (b *objectBuilderService) GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!ObjectBuilderGetList--->", logger.Any("req", req))

	if req.TableSlug == "client_type" {
		resp, err = b.strg.ObjectBuilder().GetList(ctx, req)
		if err != nil {
			b.log.Error("!!!ObjectBuilderGetList--->", logger.Error(err))
			return resp, err
		}
	} else if req.TableSlug == "connections" {
		resp, err = b.strg.ObjectBuilder().GetListConnection(ctx, req)
		if err != nil {
			b.log.Error("!!!ObjectBuilderGetList--->", logger.Error(err))
			return resp, err
		}
	}

	return resp, nil
}

func (b *objectBuilderService) GetTableDetails(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetTableDetails--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetTableDetails(ctx, req)
	if err != nil {
		b.log.Error("!!!GetTableDetails--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (b *objectBuilderService) GetAll(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetAll--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetAll(ctx, req)
	if err != nil {
		b.log.Error("!!!GetAll--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) GetList2(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetList2--->", logger.Any("req", req))

	if req.TableSlug == "client_type" || req.TableSlug == "role" || req.TableSlug == "template" {
		resp, err = b.strg.ObjectBuilder().GetList2(ctx, req)
		if err != nil {
			b.log.Error("!!!GetList2--->", logger.Error(err))
			return resp, err
		}
	} else {
		resp, err = b.strg.ObjectBuilder().GetListV2(ctx, req)
		if err != nil {
			b.log.Error("!!!GetList2--->", logger.Error(err))
			return resp, err
		}
	}

	return resp, nil
}

func (b *objectBuilderService) GetListInExcel(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetListInExcel--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetListInExcel(ctx, req)
	if err != nil {
		b.log.Error("!!!GetListInExcel--->GetList", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (b *objectBuilderService) GetListSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetListSlim--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetListSlim(ctx, req)
	if err != nil {
		b.log.Error("!!!GetListSlim--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) TestApi(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!TestApi--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().TestApi(ctx, req)
	if err != nil {
		b.log.Error("!!!TestApi--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) UpdateWithQuery(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!UpdateWithQuery--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().UpdateWithQuery(ctx, req)
	if err != nil {
		b.log.Error("!!!UpdateWithQuery--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) GetGroupByField(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GroupByColumns--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GroupByColumns(ctx, req)
	if err != nil {
		b.log.Error("!!!GroupByColumns--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) UpdateWithParams(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!UpdateWithParams--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().UpdateWithParams(ctx, req)
	if err != nil {
		b.log.Error("!!!UpdateWithParams--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) GetListForDocx(ctx context.Context, req *nb.CommonForDocxMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetListForDocx--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetListForDocxMultiTables(ctx, req)
	if err != nil {
		b.log.Error("!!!GetListForDocx--->", logger.Error(err))
		return resp, err
	}

	return resp, nil
}

func (b *objectBuilderService) GetSingleSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetSingleSlim--->", logger.Any("req", req))

	resp, err = b.strg.ObjectBuilder().GetSingleSlim(ctx, req)
	if err != nil {
		b.log.Error("!!!GetSingleSlim--->", logger.Error(err))
		return resp, err
	}
	return resp, nil
}

func (b *objectBuilderService) GetAllForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetAllForDocx--->", logger.Any("req", req))

	params := make(map[string]interface{})
	res := make(map[string]interface{})
	mainTableSlug := req.TableSlug

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err = json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	tableSlugs := cast.ToStringSlice(params["table_slugs"])

	response, err := b.strg.ObjectBuilder().GetAllForDocx(ctx, req)
	if err != nil {
		b.log.Error(fmt.Sprintf("!!!GetAllForDocx---> %d", 1), logger.Error(err))
		return resp, err
	}

	res[req.TableSlug] = response["response"]

	for i, tableSlug := range tableSlugs {
		req.TableSlug = tableSlug
		response, err = b.strg.ObjectBuilder().GetAllForDocx(ctx, req)
		if err != nil {
			b.log.Error(fmt.Sprintf("!!!GetAllForDocx---> %d", i+1), logger.Error(err))
			return resp, err
		}

		res[tableSlug] = response["response"]
	}

	newResp, err := helper.ConvertMapToStruct(res)
	if err != nil {
		b.log.Error("!!!GetAllForDocx--->", logger.Error(err))
		return resp, err
	}

	return &nb.CommonMessage{
		TableSlug: mainTableSlug,
		Data:      newResp,
	}, nil
}

func (b *objectBuilderService) GetAllFieldsForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {
	b.log.Info("!!!GetAllFieldsForDocx--->", logger.Any("req", req))

	//params := make(map[string]interface{})
	//res := make(map[string]interface{})
	//mainTableSlug := req.TableSlug
	//
	//paramBody, err := json.Marshal(req.Data)
	//if err != nil {
	//	return &nb.CommonMessage{}, err
	//}
	//if err = json.Unmarshal(paramBody, &params); err != nil {
	//	return &nb.CommonMessage{}, err
	//}

	//tableSlugs := cast.ToStringSlice(params["table_slugs"])

	response, err := b.strg.ObjectBuilder().GetAllFieldsForDocx(ctx, req)
	if err != nil {
		b.log.Error(fmt.Sprintf("!!!GetAllForDocx---> %d", 1), logger.Error(err))
		return resp, err
	}

	return response, nil
}
