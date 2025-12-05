package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

// CreateUser implements api.MessageBoardServer.
func (s *ServerNode) CreateUser(context.Context, *api.CreateUserRequest) (*api.User, error) {
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

// ExecuteDatabaseOperation implements api.InternalMessageBoardServiceServer.
func (s *ServerNode) ExecuteDatabaseOperation(context.Context, *api.ExecuteDatabaseOperationRequest) (*emptypb.Empty, error) {
	panic("unimplemented")
}

// Rewire implements api.InternalMessageBoardServiceServer.
func (s *ServerNode) Rewire(context.Context, *api.RewireRequest) (*emptypb.Empty, error) {
	panic("unimplemented")
}

// TransferDatabase implements api.InternalMessageBoardServiceServer.
func (s *ServerNode) TransferDatabase(*emptypb.Empty, grpc.ServerStreamingServer[api.TransferDatabaseEvent]) error {
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

func NewServerNode(id int, address, controlAddress string) (*ServerNode, error) {
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
	conn, err := grpc.NewClient(controlAddress)
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

// type Server struct {
// 	api.UnimplementedMessageBoardServer
// 	id           string
// 	address      string
// 	grpcServer   *grpc.Server
// 	healthServer *health.Server
// 	db           *gorm.DB
// 	pubSub       *shared.Observer
// 	nextConn     *grpc.ClientConn
// 	nextServer   api.MessageBoardClient
// }

// var _ api.MessageBoardServer = (*Server)(nil)

// func NewServer(address, id, nextAddress string) (*Server, error) {
// 	db, err := shared.NewDatabase()
// 	if err != nil {
// 		return nil, err
// 	}

// 	server := &Server{
// 		id:           id,
// 		address:      address,
// 		grpcServer:   grpc.NewServer(),
// 		healthServer: health.NewServer(),
// 		db:           db,
// 		pubSub:       shared.NewObserver(),
// 	}

// 	if nextAddress != "" {
// 		server.nextConn, err = grpc.NewClient(nextAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			return nil, status.Errorf(codes.InvalidArgument, "failed to connect to next server: %v", err)
// 		}

// 		shared.Logger.Info("connected to next server in chain")
// 		server.nextServer = api.NewMessageBoardClient(server.nextConn)
// 	}

// 	api.RegisterMessageBoardServer(server.grpcServer, server)
// 	grpc_health_v1.RegisterHealthServer(server.grpcServer, server.healthServer)
// 	reflection.Register(server.grpcServer)

// 	return server, nil
// }

// func (s *Server) Run() error {
// 	listener, err := net.Listen("tcp", s.address)
// 	if err != nil {
// 		return fmt.Errorf("failed to listen: %w", err)
// 	}

// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

// 	go func() {
// 		sig := <-signalChan
// 		shared.Logger.Info("received shutdown signal", "signal", sig.String())
// 		s.grpcServer.GracefulStop()
// 	}()

// 	shared.Logger.Info("server running", "address", s.address)

// 	if err := s.grpcServer.Serve(listener); err != nil {
// 		return fmt.Errorf("failed to serve: %w", err)
// 	}

// 	shared.Logger.Info("server stopped gracefully")
// 	return nil
// }

// func (s *Server) Close() {
// 	if s.nextConn != nil {
// 		s.nextConn.Close()
// 	}
// }
