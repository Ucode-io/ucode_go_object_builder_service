package storage

import (
	"context"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
)

type StorageI interface {
	CloseDB()
	BuilderProject() BuilderProjectRepoI
	Field() FieldRepoI
	Function() FunctionRepoI
	File() FileRepoI
	// CustomErrorMessage() CustomErrorMessageRepoI
	Table() TableRepoI
	ObjectBuilder() ObjectBuilderRepoI
	View() ViewRepoI
	Menu() MenuRepoI
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
	Create(ctx context.Context, req *nb.CreateFieldRequest) (resp *nb.Field, err error)
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

type TableRepoI interface {
	Create(ctx context.Context, req *nb.CreateTableRequest) (resp *nb.CreateTableResponse, err error)
	GetByID(ctx context.Context, req *nb.TablePrimaryKey) (resp *nb.Table, err error)
	GetAll(ctx context.Context, req *nb.GetAllTablesRequest) (resp *nb.GetAllTablesResponse, err error)
	Update(ctx context.Context, req *nb.UpdateTableRequest) (resp *nb.Table, err error)
	Delete(ctx context.Context, req *nb.TablePrimaryKey) error

	// GetListTableHistory(ctx context.Context, req *nb.GetTableHistoryRequest) (resp *nb.GetTableHistoryResponse, err error)
	// GetTableHistoryById(ctx context.Context, req *nb.TableHistoryPrimaryKey) (resp *nb.Table, err error)
	// RevertTableHistory(ctx context.Context, req *nb.RevertTableHistoryRequest) (resp *nb.TableHistory, err error)
	// InsertVersionsToCommit(ctx context.Context, req *nb.InsertVersionsToCommitRequest) (resp *nb.TableHistory, err error)
}

type FileRepoI interface {
	Create(ctx context.Context, req *nb.CreateFileRequest) (resp *nb.File, err error)
	GetList(ctx context.Context, req *nb.GetAllFilesRequest) (resp *nb.GetAllFilesResponse, err error)
	GetSingle(ctx context.Context, req *nb.FilePrimaryKey) (resp *nb.File, err error)
	Update(ctx context.Context, req *nb.File) error
	Delete(ctx context.Context, req *nb.FileDeleteRequest) error
}

type ObjectBuilderRepoI interface {
	GetList(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
}

// type CustomErrorMessageRepoI interface {
// 	Create(ctx context.Context, req *nb.CreateCustomErrorMessage) (resp *nb.CustomErrorMessage, err error)
// 	GetList(ctx context.Context, req *nb.GetCustomErrorMessageListRequest) (resp *nb.GetCustomErrorMessageListResponse, err error)
// 	GetListForObject(ctx context.Context, req *nb.GetListForObjectRequest) (resp *nb.GetCustomErrorMessageListResponse, err error)
// 	GetSingle(ctx context.Context, req *nb.CustomErrorMessagePK) (resp *nb.CustomErrorMessage, err error)
// 	Update(ctx context.Context, req *nb.CustomErrorMessage) error
// 	Delete(ctx context.Context, req *nb.CustomErrorMessagePK) error
// }

type ViewRepoI interface {
	Create(ctx context.Context, req *nb.CreateViewRequest) (resp *nb.View, err error)
	GetList(ctx context.Context, req *nb.GetAllViewsRequest) (resp *nb.GetAllViewsResponse, err error)
	GetSingle(ctx context.Context, req *nb.ViewPrimaryKey) (resp *nb.View, err error)
	Update(ctx context.Context, req *nb.View) (resp *nb.View, err error)
	Delete(ctx context.Context, req *nb.ViewPrimaryKey) error
	// ConvertHtmlToPdf(ctx, req *nb.HtmlBody) (resp *nb.PdfBody, err error)
	// ConvertTemplateToHtml(ctx, req *nb.HtmlBody) (resp *nb.HtmlBody, err error)
	// UpdateViewOrder(ctx, req *nb.UpdateViewOrderRequest) error
}

type MenuRepoI interface {
	Create(ctx context.Context, req *nb.CreateMenuRequest) (*nb.Menu, error)
	GetById(ctx context.Context, req *nb.MenuPrimaryKey) (*nb.Menu, error)
	GetAll(ctx context.Context, req *nb.GetAllMenusRequest) (*nb.GetAllMenusResponse, error)
	Update(ctx context.Context, req *nb.Menu) (*nb.Menu, error)
	Delete(ctx context.Context, req *nb.MenuPrimaryKey) error
	UpdateMenuOrder(ctx context.Context, req *nb.UpdateMenuOrderRequest) error
}
