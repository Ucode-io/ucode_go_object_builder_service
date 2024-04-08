package storage

import (
	"context"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
)

type StorageI interface {
	BuilderProject() BuilderProjectRepoI
	Field() FieldRepoI
}

type BuilderProjectRepoI interface {
	Register(ctx context.Context, req *nb.RegisterProjectRequest) error
	RegisterProjects(ctx context.Context, req *nb.RegisterProjectRequest) error
	Deregister(ctx context.Context, req *nb.DeregisterProjectRequest) error
	Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) error
	RegisterMany(ctx context.Context, req *nb.RegisterManyProjectsRequest) (resp *nb.RegisterManyProjectsResponse, err error)
	DeregisterMany(ctx context.Context, req *nb.DeregisterManyProjectsRequest) (resp *nb.DeregisterManyProjectsResponse, err error)
}

type FieldRepoI interface {
	Create(ctx context.Context, req *nb.CreateFieldRequest) error
	GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error)
	GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error)
	GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error)
	Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error)
	UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error
	Delete(ctx context.Context, req *nb.FieldPrimaryKey) error
}

type FunctionRepoI interface {
	Create(ctx context.Context, req *nb.CreateFunctionRequest) (resp *nb.Function, err error)
	GetList(ctx context.Context, req *nb.GetAllFunctionsRequest) (resp *nb.GetAllFunctionsResponse, err error)
	GetSingle(ctx context.Context, req *nb.FunctionPrimaryKey) (resp *nb.Function, err error)
	Update(ctx context.Context, req *nb.Function) error
	Delete(ctx context.Context, req *nb.FunctionPrimaryKey) error
}

type TableRepoI interface{}
