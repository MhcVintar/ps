package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateTopic implements api.MessageBoardServer.
func (s *ServerNode) CreateTopic(context.Context, *api.CreateTopicRequest) (*api.Topic, error) {
	panic("unimplemented")
}

// DeleteMessage implements api.MessageBoardServer.
func (s *ServerNode) DeleteMessage(context.Context, *api.DeleteMessageRequest) (*emptypb.Empty, error) {
	panic("unimplemented")
}

// GetMessages implements api.MessageBoardServer.
func (s *ServerNode) GetMessages(context.Context, *api.GetMessagesRequest) (*api.GetMessagesResponse, error) {
	panic("unimplemented")
}

// GetSubscriptionNode implements api.MessageBoardServer.
func (s *ServerNode) GetSubscriptionNode(context.Context, *api.SubscriptionNodeRequest) (*api.SubscriptionNodeResponse, error) {
	panic("unimplemented")
}

// LikeMessage implements api.MessageBoardServer.
func (s *ServerNode) LikeMessage(context.Context, *api.LikeMessageRequest) (*api.Message, error) {
	panic("unimplemented")
}

// ListTopics implements api.MessageBoardServer.
func (s *ServerNode) ListTopics(context.Context, *emptypb.Empty) (*api.ListTopicsResponse, error) {
	panic("unimplemented")
}

// PostMessage implements api.MessageBoardServer.
func (s *ServerNode) PostMessage(context.Context, *api.PostMessageRequest) (*api.Message, error) {
	panic("unimplemented")
}

// SubscribeTopic implements api.MessageBoardServer.
func (s *ServerNode) SubscribeTopic(*api.SubscribeTopicRequest, grpc.ServerStreamingServer[api.MessageEvent]) error {
	panic("unimplemented")
}

// UpdateMessage implements api.MessageBoardServer.
func (s *ServerNode) UpdateMessage(context.Context, *api.UpdateMessageRequest) (*api.Message, error) {
	panic("unimplemented")
}

type ServerNode struct {
	api.UnimplementedInternalMessageBoardServiceServer
	api.UnimplementedMessageBoardServer
	id                   int
	address              string
	db                   *database.Database
	grpcServer           *grpc.Server
	healthServer         *health.Server
	controlClient        *shared.ControlNodeClient
	downstreamClient     *shared.ServerNodeClient
	upstreamClient       *shared.ServerNodeClient
	messageEventObserver *shared.Observable[api.MessageEvent]
}

var _ api.InternalMessageBoardServiceServer = (*ServerNode)(nil)
var _ api.MessageBoardServer = (*ServerNode)(nil)

func NewServerNode(id int, address, control string) (*ServerNode, error) {
	s := ServerNode{
		id:                   id,
		address:              address,
		grpcServer:           grpc.NewServer(),
		healthServer:         health.NewServer(),
		messageEventObserver: shared.NewObservable[api.MessageEvent](),
	}

	// Prepare database
	db, err := database.NewDatabase()
	if err != nil {
		return nil, err
	}
	s.db = db

	// Prepare health server
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Prepare control client
	conn, err := grpc.NewClient(control, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to open client connection to %q: %v", address, err)
	}
	s.controlClient = &shared.ControlNodeClient{
		Conn: conn,
	}

	// Register gRPC service
	api.RegisterInternalMessageBoardServiceServer(s.grpcServer, &s)
	api.RegisterMessageBoardServer(s.grpcServer, &s)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthServer)
	reflection.Register(s.grpcServer)

	// TODO Sync database

	return &s, nil
}

func (s *ServerNode) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.consumeWALQueue(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		shared.Logger.Info("received shutdown signal", "signal", sig.String())
		cancel()
		s.grpcServer.GracefulStop()

		s.controlClient.Conn.Close()
		if s.downstreamClient != nil {
			s.downstreamClient.Conn.Close()
		}
		if s.upstreamClient != nil {
			s.upstreamClient.Conn.Close()
		}
	}()

	shared.Logger.Info("server running", "address", s.address)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	shared.Logger.Info("server stopped gracefully")
	return nil
}

func (s *ServerNode) isHead() bool {
	return s.downstreamClient == nil
}

func (s *ServerNode) isTail() bool {
	return s.upstreamClient == nil
}

func (s *ServerNode) consumeWALQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			shared.Logger.InfoContext(ctx, "stopping WAL consumer")
			return

		case entry := <-s.db.WALQueue():
			if !s.isHead() {
				backoff := 100 * time.Millisecond
				maxBackoff := 10 * time.Second

				for {
					req := api.ApplyWALEntryRequest{
						Entry: database.ToApiWALEntry(entry),
					}
					_, err := s.downstreamClient.Internal.ApplyWALEntry(ctx, &req)
					if err == nil {
						shared.Logger.InfoContext(ctx, "WAL entry applied downstream")
						break
					}

					shared.Logger.ErrorContext(ctx, "failed processing WAL entry", "error", err)

					backoff = backoff * 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}

					jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
					sleep := backoff/2 + jitter

					select {
					case <-ctx.Done():
						shared.Logger.InfoContext(ctx, "stopping WAL consumer")
						return

					case <-time.After(sleep):
					}
				}
			}
		}
	}
}
