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
	_ "github.com/lib/pq"
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

	svcs, err := client.NewGrpcClients(cfg)
	if err != nil {
		log.Panic("client.NewGrpcClients", logger.Error(err))
	}

	pgStore, err := postgres.NewPostgres(context.Background(), cfg, svcs)
	if err != nil {
		log.Panic("postgres.NewPostgres", logger.Error(err))
	}
	// defer pgStore.CloseDB()

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
