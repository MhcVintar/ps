package shared

import (
	"razpravljalnica/internal/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type ServerNodeClient struct {
	Id       int
	Conn     *grpc.ClientConn
	Health   grpc_health_v1.HealthClient
	Internal api.InternalMessageBoardServiceClient
}

type ControlNodeClient struct {
	Conn *grpc.ClientConn
}
