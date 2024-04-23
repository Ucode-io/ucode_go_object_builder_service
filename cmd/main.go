package main

import (
	"context"
	"net"
	"ucode/ucode_go_object_builder_service/config"
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

	// resp, err := pgStore.ObjectBuilder().GetAll(context.Background(), &new_object_builder_service.CommonMessage{
	// 	TableSlug: "get_list_test",
	// 	Data: &structpb.Struct{
	// 		Fields: map[string]*structpb.Value{
	// 			"offset":                    {Kind: &structpb.Value_NumberValue{NumberValue: 0}},
	// 			"limit":                     {Kind: &structpb.Value_NumberValue{NumberValue: 20}},
	// 			"role_id_from_token":        {Kind: &structpb.Value_StringValue{StringValue: "9a31fec6-1cd3-477a-ab7d-11a4281222bb"}},
	// 			"client_type_id_from_token": {Kind: &structpb.Value_StringValue{StringValue: "2e19339c-15b6-43ca-ade8-8286dde7c65d"}},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	fmt.Println("Err->", err)
	// 	return
	// }
	// fmt.Println("Resp->", resp)
	// return

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
