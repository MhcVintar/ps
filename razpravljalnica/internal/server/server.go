package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
	"syscall"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	api.UnimplementedMessageBoardServer
	id      string
	address string
	server  *grpc.Server
	db      *gorm.DB
	pubSub  *shared.PubSub
}

var _ api.MessageBoardServer = (*Server)(nil)

func NewServer(address string) (*Server, error) {
	db, err := shared.NewDatabase()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate node id: %w", err)
	}

	server := &Server{
		id:      id.String(),
		address: address,
		db:      db,
		pubSub:  shared.NewPubSub(),
	}

	grpcServer := grpc.NewServer()
	api.RegisterMessageBoardServer(grpcServer, server)
	reflection.Register(grpcServer)

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
		s.server.GracefulStop()
	}()

	shared.Logger.Info("server running", "address", s.address)

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("server stopped gracefully")
	return nil
}
