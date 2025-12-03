package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	api.UnimplementedMessageBoardServer
	id           string
	address      string
	grpcServer   *grpc.Server
	healthServer *health.Server
	db           *gorm.DB
	pubSub       *shared.PubSub
	nextConn     *grpc.ClientConn
	nextServer   api.MessageBoardClient
}

var _ api.MessageBoardServer = (*Server)(nil)

func NewServer(address, id, nextAddress string) (*Server, error) {
	db, err := shared.NewDatabase()
	if err != nil {
		return nil, err
	}

	server := &Server{
		id:           id,
		address:      address,
		grpcServer:   grpc.NewServer(),
		healthServer: health.NewServer(),
		db:           db,
		pubSub:       shared.NewPubSub(),
	}

	if nextAddress != "" {
		server.nextConn, err = grpc.NewClient(nextAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to connect to next server: %v", err)
		}

		shared.Logger.Info("connected to next server in chain")
		server.nextServer = api.NewMessageBoardClient(server.nextConn)
	}

	api.RegisterMessageBoardServer(server.grpcServer, server)
	grpc_health_v1.RegisterHealthServer(server.grpcServer, server.healthServer)
	reflection.Register(server.grpcServer)

	return server, nil
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		shared.Logger.Info("received shutdown signal", "signal", sig.String())
		s.grpcServer.GracefulStop()
	}()

	shared.Logger.Info("server running", "address", s.address)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("server stopped gracefully")
	return nil
}

func (s *Server) Close() {
	if s.nextConn != nil {
		s.nextConn.Close()
	}
}
