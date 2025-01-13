package postgres

import (
	"context"
	"fmt"
	"log"

	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/grpc/client"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db             *psqlpool.Pool
	grpcClient     client.ServiceManagerI
	builderProject storage.BuilderProjectRepoI
	field          storage.FieldRepoI
	function       storage.FunctionRepoI
	file           storage.FileRepoI
	table          storage.TableRepoI
	object_builder storage.ObjectBuilderRepoI
	view           storage.ViewRepoI
	menu           storage.MenuRepoI
	login          storage.LoginRepoI
	layout         storage.LayoutRepoI
	relation       storage.RelationRepoI
	section        storage.SectionRepoI
	permission     storage.PermissionRepoI
	items          storage.ItemsRepoI
	excel          storage.ExcelRepoI
	version        storage.VersionRepoI
	customEvent    storage.CustomEventRepoI
	versionHistory storage.VersionHistoryRepoI
	folderGroup    storage.FolderGroupRepoI
	csv            storage.CSVRepoI
	docxTemplate   storage.DocxTemplateRepoI
}

func NewPostgres(ctx context.Context, cfg config.Config, grpcClient client.ServiceManagerI) (storage.StorageI, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
	))
	if err != nil {
		return nil, err
	}

	config.MaxConns = cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	dbPool := &psqlpool.Pool{
		Db: pool,
	}

	return &Store{
		db:         dbPool,
		grpcClient: grpcClient,
	}, err
}

func (s *Store) CloseDB() {
	s.db.Db.Close()
}

func (s *Store) Log(ctx context.Context, msg string, data map[string]interface{}) {
	args := make([]interface{}, 0, len(data)+2) // making space for arguments + msg
	args = append(args, msg)
	for k, v := range data {
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	log.Println(args...)
}

func (s *Store) BuilderProject() storage.BuilderProjectRepoI {
	if s.builderProject == nil {
		s.builderProject = NewBuilderProjectRepo(s.db)
	}

	return s.builderProject
}

func (s *Store) Field() storage.FieldRepoI {
	if s.field == nil {
		s.field = NewFieldRepo(s.db, NewRelationRepo(s.db))
	}

	return s.field
}

func (s *Store) Function() storage.FunctionRepoI {
	if s.function == nil {
		s.function = NewFunctionRepo(s.db)
	}

	return s.function
}

func (s *Store) File() storage.FileRepoI {
	if s.file == nil {
		s.file = NewFileRepo(s.db)
	}

	return s.file
}

func (s *Store) Table() storage.TableRepoI {
	if s.table == nil {
		s.table = NewTableRepo(s.db)
	}

	return s.table
}

func (s *Store) ObjectBuilder() storage.ObjectBuilderRepoI {
	if s.object_builder == nil {
		s.object_builder = NewObjectBuilder(s.db)
	}
	return s.object_builder
}

func (s *Store) View() storage.ViewRepoI {
	if s.view == nil {
		s.view = NewViewRepo(s.db)
	}

	return s.view
}

func (s *Store) Menu() storage.MenuRepoI {
	if s.menu == nil {
		s.menu = NewMenuRepo(s.db)
	}
	return s.menu
}

func (s *Store) Login() storage.LoginRepoI {
	if s.login == nil {
		s.login = NewLoginRepo(s.db)
	}

	return s.login
}

func (s *Store) Layout() storage.LayoutRepoI {
	if s.layout == nil {
		s.layout = NewLayoutRepo(s.db)
	}

	return s.layout
}

func (s *Store) Relation() storage.RelationRepoI {
	if s.relation == nil {
		s.relation = NewRelationRepo(s.db)
	}

	return s.relation
}

func (s *Store) Section() storage.SectionRepoI {
	if s.section == nil {
		s.section = NewSectionRepo(s.db)
	}
	return s.section
}

func (s *Store) Permission() storage.PermissionRepoI {
	if s.permission == nil {
		s.permission = NewPermissionRepo(s.db)
	}
	return s.permission
}

func (s *Store) Items() storage.ItemsRepoI {
	if s.items == nil {
		s.items = NewItemsRepo(s.db, s.grpcClient)
	}
	return s.items
}

func (s *Store) Excel() storage.ExcelRepoI {
	if s.excel == nil {
		s.excel = NewExcelRepo(s.db)
	}

	return s.excel
}

func (s *Store) Version() storage.VersionRepoI {
	if s.version == nil {
		s.version = NewVersionRepo(s.db)
	}

	return s.version
}

func (s *Store) CustomEvent() storage.CustomEventRepoI {
	if s.customEvent == nil {
		s.customEvent = NewCustomEventRepo(s.db)
	}

	return s.customEvent
}

func (s *Store) VersionHistory() storage.VersionHistoryRepoI {
	if s.versionHistory == nil {
		s.versionHistory = NewVersionHistoryRepo(s.db)
	}

	return s.versionHistory
}

func (s *Store) FolderGroup() storage.FolderGroupRepoI {
	if s.folderGroup == nil {
		s.folderGroup = NewFolderGroupRepo(s.db)
	}
	return s.folderGroup
}

func (s *Store) CSV() storage.CSVRepoI {
	if s.csv == nil {
		s.csv = NewCSVRepo(s.db)
	}
	return s.csv
}

func (s *Store) DocxTemplate() storage.DocxTemplateRepoI {
	if s.docxTemplate == nil {
		s.docxTemplate = NewDocxTemplateRepo(s.db)
	}

	return s.docxTemplate
}
