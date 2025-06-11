package storage

import (
	"context"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/jackc/pgx/v5"
)

type StorageI interface {
	CloseDB()
	BuilderProject() BuilderProjectRepoI
	Field() FieldRepoI
	Function() FunctionRepoI
	File() FileRepoI
	Table() TableRepoI
	ObjectBuilder() ObjectBuilderRepoI
	View() ViewRepoI
	Menu() MenuRepoI
	Login() LoginRepoI
	Layout() LayoutRepoI
	Section() SectionRepoI
	Items() ItemsRepoI
	Relation() RelationRepoI
	Permission() PermissionRepoI
	Excel() ExcelRepoI
	Version() VersionRepoI
	CustomEvent() CustomEventRepoI
	VersionHistory() VersionHistoryRepoI
	FolderGroup() FolderGroupRepoI
	CSV() CSVRepoI
	DocxTemplate() DocxTemplateRepoI
	Language() LanguageRepoI
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
	CreateWithTx(ctx context.Context, req *nb.CreateFieldRequest, tableSlug string, tx pgx.Tx) (resp *nb.Field, err error)
	GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error)
	GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error)
	Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error)
	UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error
	Delete(ctx context.Context, req *nb.FieldPrimaryKey) error
	FieldsWithPermissions(ctx context.Context, req *nb.FieldsWithRelationRequest) (resp *nb.FieldsWithRelationsResponse, err error)
}

type FunctionRepoI interface {
	Create(ctx context.Context, req *nb.CreateFunctionRequest) (resp *nb.Function, err error)
	GetList(ctx context.Context, req *nb.GetAllFunctionsRequest) (resp *nb.GetAllFunctionsResponse, err error)
	GetSingle(ctx context.Context, req *nb.FunctionPrimaryKey) (resp *nb.Function, err error)
	Update(ctx context.Context, req *nb.Function) error
	Delete(ctx context.Context, req *nb.FunctionPrimaryKey) error
	GetCountByType(ctx context.Context, req *nb.GetCountByTypeRequest) (*nb.GetCountByTypeResponse, error)
}

type TableRepoI interface {
	Create(ctx context.Context, req *nb.CreateTableRequest) (resp *nb.CreateTableResponse, err error)
	GetByID(ctx context.Context, req *nb.TablePrimaryKey) (resp *nb.Table, err error)
	GetAll(ctx context.Context, req *nb.GetAllTablesRequest) (resp *nb.GetAllTablesResponse, err error)
	Update(ctx context.Context, req *nb.UpdateTableRequest) (resp *nb.Table, err error)
	Delete(ctx context.Context, req *nb.TablePrimaryKey) error
	GetTablesByLabel(ctx context.Context, req *nb.GetTablesByLabelReq) (resp *nb.GetAllTablesResponse, err error)
	GetChart(ctx context.Context, req *nb.ChartPrimaryKey) (resp *nb.GetChartResponse, err error)
	CreateConnectionAndSchema(ctx context.Context, req *nb.CreateConnectionAndSchemaReq) error
	GetTrackedUntrackedTables(ctx context.Context, req *nb.GetTrackedUntrackedTablesReq) (resp *nb.GetTrackedUntrackedTableResp, err error)
	GetTrackedConnections(ctx context.Context, req *nb.GetTrackedConnectionsReq) (resp *nb.GetTrackedConnectionsResp, err error)
	TrackTables(ctx context.Context, req *nb.TrackedTablesByIdsReq) error
	UntrackTableById(ctx context.Context, req *nb.UntrackTableByIdReq) error
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
	GetList2(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListInExcel(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListConnection(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetTableDetails(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetAll(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetSingleSlim(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GroupByColumns(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error)
	UpdateWithParams(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListV2(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListForDocxMultiTables(ctx context.Context, req *nb.CommonForDocxMessage) (resp *nb.CommonMessage, err error)
	GetAllForDocx(ctx context.Context, req *nb.CommonMessage) (resp map[string]any, err error)
	GetAllFieldsForDocx(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetListAggregation(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	AgGridTree(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetBoardStructure(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetBoardData(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
}

type ViewRepoI interface {
	Create(ctx context.Context, req *nb.CreateViewRequest) (resp *nb.View, err error)
	GetList(ctx context.Context, req *nb.GetAllViewsRequest) (resp *nb.GetAllViewsResponse, err error)
	GetSingle(ctx context.Context, req *nb.ViewPrimaryKey) (resp *nb.View, err error)
	Update(ctx context.Context, req *nb.View) (resp *nb.View, err error)
	Delete(ctx context.Context, req *nb.ViewPrimaryKey) error
	UpdateViewOrder(ctx context.Context, req *nb.UpdateViewOrderRequest) error
}

type MenuRepoI interface {
	Create(ctx context.Context, req *nb.CreateMenuRequest) (*nb.Menu, error)
	CreateWithTx(ctx context.Context, req *nb.CreateMenuRequest, tx pgx.Tx) (*nb.Menu, error)
	GetById(ctx context.Context, req *nb.MenuPrimaryKey) (*nb.Menu, error)
	GetByLabel(ctx context.Context, req *nb.MenuPrimaryKey) (*nb.GetAllMenusResponse, error)
	GetAll(ctx context.Context, req *nb.GetAllMenusRequest) (*nb.GetAllMenusResponse, error)
	Update(ctx context.Context, req *nb.Menu) (*nb.Menu, error)
	Delete(ctx context.Context, req *nb.MenuPrimaryKey) error
	UpdateMenuOrder(ctx context.Context, req *nb.UpdateMenuOrderRequest) error

	GetAllMenuSettings(ctx context.Context, req *nb.GetAllMenuSettingsRequest) (resp *nb.GetAllMenuSettingsResponse, err error)
	GetByIDMenuSettings(ctx context.Context, req *nb.MenuSettingPrimaryKey) (resp *nb.MenuSettings, err error)
	GetAllMenuTemplate(ctx context.Context, req *nb.GetAllMenuSettingsRequest) (resp *nb.GatAllMenuTemplateResponse, err error)
	GetMenuTemplateWithEntities(ctx context.Context, req *nb.GetMenuTemplateRequest) (resp *nb.MenuTemplateWithEntities, err error)
}
type LoginRepoI interface {
	LoginData(ctx context.Context, req *nb.LoginDataReq) (resp *nb.LoginDataRes, err error)
	GetConnectionOptions(ctx context.Context, req *nb.GetConnetionOptionsRequest) (resp *nb.GetConnectionOptionsResponse, err error)
}

type LayoutRepoI interface {
	Update(ctx context.Context, req *nb.LayoutRequest) (resp *nb.LayoutResponse, err error)
	GetSingleLayout(ctx context.Context, req *nb.GetSingleLayoutRequest) (resp *nb.LayoutResponse, err error)
	GetAll(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error)
	GetAllV2(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error)
	RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) error
	GetByID(ctx context.Context, req *nb.LayoutPrimaryKey) (resp *nb.LayoutResponse, err error)
}

type RelationRepoI interface {
	Create(ctx context.Context, req *nb.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error)
	CreateWithTx(ctx context.Context, req *nb.CreateRelationRequest, tx pgx.Tx) (resp *nb.CreateRelationRequest, err error)
	GetList(ctx context.Context, req *nb.GetAllRelationsRequest) (resp *nb.GetAllRelationsResponse, err error)
	GetByID(ctx context.Context, req *nb.RelationPrimaryKey) (resp *nb.RelationForGetAll, err error)
	Update(ctx context.Context, req *nb.UpdateRelationRequest) (resp *nb.RelationForGetAll, err error)
	Delete(ctx context.Context, req *nb.RelationPrimaryKey) error
	GetSingleViewForRelation(ctx context.Context, req models.ReqForViewRelation) (resp *nb.RelationForGetAll, err error)
}

type SectionRepoI interface {
	GetViewRelation(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetViewRelationResponse, err error)
	GetAll(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetAllSectionsResponse, err error)
}

type PermissionRepoI interface {
	GetAllMenuPermissions(ctx context.Context, req *nb.GetAllMenuPermissionsRequest) (resp *nb.GetAllMenuPermissionsResponse, err error)
	CreateDefaultPermission(ctx context.Context, req *nb.CreateDefaultPermissionRequest) error
	GetListWithRoleAppTablePermissions(ctx context.Context, req *nb.GetListWithRoleAppTablePermissionsRequest) (resp *nb.GetListWithRoleAppTablePermissionsResponse, err error)
	UpdateMenuPermissions(ctx context.Context, req *nb.UpdateMenuPermissionsRequest) error
	UpdateRoleAppTablePermissions(ctx context.Context, req *nb.UpdateRoleAppTablePermissionsRequest) error
	GetPermissionsByTableSlug(ctx context.Context, req *nb.GetPermissionsByTableSlugRequest) (resp *nb.GetPermissionsByTableSlugResponse, err error)
	UpdatePermissionsByTableSlug(ctx context.Context, req *nb.UpdatePermissionsRequest) (err error)
	GetTablePermission(ctx context.Context, req *nb.GetTablePermissionRequest) (resp *nb.GetTablePermissionResponse, err error)
}

type ItemsRepoI interface {
	Create(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	GetSingle(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	Update(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	Delete(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	UpdateUserIdAuth(ctx context.Context, req *models.ItemsChangeGuid) error
	DeleteMany(ctx context.Context, req *nb.CommonMessage) (resp *models.DeleteUsers, err error)
	MultipleUpdate(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
	UpsertMany(ctx context.Context, req *nb.CommonMessage) error
	UpdateByUserIdAuth(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
}

type ExcelRepoI interface {
	ExcelRead(ctx context.Context, req *nb.ExcelReadRequest) (resp *nb.ExcelReadResponse, err error)
	ExcelToDb(ctx context.Context, req *nb.ExcelToDbRequest) (resp *nb.ExcelToDbResponse, err error)
}

type CSVRepoI interface {
	GetListInCSV(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error)
}

type VersionRepoI interface {
	Create(ctx context.Context, req *nb.CreateVersionRequest) (resp *nb.Version, err error)
	GetList(ctx context.Context, req *nb.GetVersionListRequest) (resp *nb.GetVersionListResponse, err error)
	Update(ctx context.Context, req *nb.Version) error
	CreateMany(ctx context.Context, req *nb.CreateManyVersionRequest) error
	GetSingle(ctx context.Context, req *nb.VersionPrimaryKey) (resp *nb.Version, err error)
	Delete(ctx context.Context, req *nb.VersionPrimaryKey) error
	UpdateLive(ctx context.Context, req *nb.VersionPrimaryKey) error
}
type CustomEventRepoI interface {
	Create(ctx context.Context, req *nb.CreateCustomEventRequest) (resp *nb.CustomEvent, err error)
	Update(ctx context.Context, req *nb.CustomEvent) (err error)
	GetList(ctx context.Context, req *nb.GetCustomEventsListRequest) (resp *nb.GetCustomEventsListResponse, err error)
	GetSingle(ctx context.Context, req *nb.CustomEventPrimaryKey) (resp *nb.CustomEvent, err error)
	Delete(ctx context.Context, req *nb.CustomEventPrimaryKey) (err error)
	UpdateByFunctionId(ctx context.Context, req *nb.UpdateByFunctionIdRequest) (err error)
}

type VersionHistoryRepoI interface {
	GetById(ctx context.Context, req *nb.VersionHistoryPrimaryKey) (*nb.VersionHistory, error)
	GetAll(ctx context.Context, req *nb.GetAllRquest) (resp *nb.ListVersionHistory, err error)
	Update(ctx context.Context, req *nb.UsedForEnvRequest) (err error)
	Create(ctx context.Context, req *nb.CreateVersionHistoryRequest) (err error)
}

type FolderGroupRepoI interface {
	Create(ctx context.Context, req *nb.CreateFolderGroupRequest) (*nb.FolderGroup, error)
	GetByID(ctx context.Context, req *nb.FolderGroupPrimaryKey) (*nb.FolderGroup, error)
	GetAll(ctx context.Context, req *nb.GetAllFolderGroupRequest) (*nb.GetAllFolderGroupResponse, error)
	Update(ctx context.Context, req *nb.UpdateFolderGroupRequest) (*nb.FolderGroup, error)
	Delete(ctx context.Context, req *nb.FolderGroupPrimaryKey) error
}

type DocxTemplateRepoI interface {
	Create(ctx context.Context, req *nb.CreateDocxTemplateRequest) (*nb.DocxTemplate, error)
	GetById(ctx context.Context, req *nb.DocxTemplatePrimaryKey) (*nb.DocxTemplate, error)
	GetAll(ctx context.Context, req *nb.GetAllDocxTemplateRequest) (*nb.GetAllDocxTemplateResponse, error)
	Update(ctx context.Context, req *nb.DocxTemplate) (*nb.DocxTemplate, error)
	Delete(ctx context.Context, req *nb.DocxTemplatePrimaryKey) error
}

type LanguageRepoI interface {
	Create(ctx context.Context, req *nb.CreateLanguageRequest) (resp *nb.Language, err error)
	GetById(ctx context.Context, req *nb.PrimaryKey) (resp *nb.Language, err error)
	GetList(ctx context.Context, req *nb.GetListLanguagesRequest) (resp *nb.GetListLanguagesResponse, err error)
	UpdateLanguage(ctx context.Context, req *nb.UpdateLanguageRequest) (resp *nb.Language, err error)
	Delete(ctx context.Context, req *nb.PrimaryKey) error
}
