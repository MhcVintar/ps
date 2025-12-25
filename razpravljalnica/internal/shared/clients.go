package shared

import (
	"razpravljalnica/internal/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type ServerNodeClient struct {
	Id       string
	Conn     *grpc.ClientConn
	Health   grpc_health_v1.HealthClient
	Public   api.MessageBoardClient
	Internal api.InternalMessageBoardServiceClient
}
