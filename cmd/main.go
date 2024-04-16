package main

import (
	"context"
	"fmt"
	"net"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	loggerLevel := logger.LevelDebug

	switch cfg.Environment {
	case config.DebugMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.DebugMode)
	case config.TestMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.TestMode)
	default:
		loggerLevel = logger.LevelInfo
		gin.SetMode(gin.ReleaseMode)
	}

	log := logger.NewLogger(cfg.ServiceName, loggerLevel)
	defer logger.Cleanup(log)
	log.Info("Service env", logger.Any("cfg", cfg))

	pgStore, err := postgres.NewPostgres(context.Background(), cfg)
	if err != nil {
		log.Panic("postgres.NewPostgres", logger.Error(err))
	}
	defer pgStore.CloseDB()

	// menu, err := pgStore.Menu().Create(
	// 	context.Background(),
	// 	&new_object_builder_service.CreateMenuRequest{
	// 		Label:    "Hello World",
	// 		Icon:     "apple-whole.svg",
	// 		ParentId: "c57eedc3-a954-4262-a0af-376c65b5a284",
	// 		Type:     "FOLDER",
	// 		Attributes: &structpb.Struct{
	// 			Fields: map[string]*structpb.Value{
	// 				"label":    {Kind: &structpb.Value_StringValue{StringValue: ""}},
	// 				"label_en": {Kind: &structpb.Value_StringValue{StringValue: "Hello World"}},
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	fmt.Println("Err->", err)
	// 	return
	// }
	// fmt.Println("Menu->", menu)
	// return
	// menu, err := pgStore.Menu().GetById(
	// 	context.Background(),
	// 	&new_object_builder_service.MenuPrimaryKey{
	// 		Id: "bb6602da-5679-4f5f-ba0a-2b6956bc928a",
	// 	},
	// )
	// if err != nil {
	// 	fmt.Println("Error->", err)
	// 	return
	// }
	// fmt.Println("Menu->", menu)
	// return

	menu, err := pgStore.Menu().GetAll(
		context.Background(),
		&new_object_builder_service.GetAllMenusRequest{
			Offset: 0,
			Limit:  5,
		},
	)
	if err != nil {
		fmt.Println("err->", err)
		return
	}
	fmt.Println("Menu->", menu)
	return

	svcs, err := client.NewGrpcClients(cfg)
	if err != nil {
		log.Panic("client.NewGrpcClients", logger.Error(err))
	}

	grpcServer := grpc.SetUpServer(cfg, log, svcs, pgStore) // pgStore

	lis, err := net.Listen("tcp", cfg.ServicePort)
	if err != nil {
		log.Panic("net.Listen", logger.Error(err))
	}

	log.Info("GRPC: Server being started...", logger.String("port", cfg.ServicePort))

	if err := grpcServer.Serve(lis); err != nil {
		log.Panic("grpcServer.Serve", logger.Error(err))
	}
}
