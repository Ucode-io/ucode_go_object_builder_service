package client

import (
	"ucode/ucode_go_object_builder_service/config"
)

type ServiceManagerI interface {
}

type grpcClients struct {
}

func NewGrpcClients(cfg config.Config) (ServiceManagerI, error) {

	return &grpcClients{}, nil
}
