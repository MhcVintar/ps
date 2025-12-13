package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerNode) Rewire(ctx context.Context, req *api.RewireRequest) (*emptypb.Empty, error) {
	newDownstream := s.downstreamClient
	if req.DownstreamAddress == nil {
		newDownstream = nil
	} else if s.downstreamClient == nil || s.downstreamClient.Conn.Target() != req.GetDownstreamAddress() {
		conn, err := grpc.NewClient(req.GetDownstreamAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to open downstream client connection to %q: %v", req.GetDownstreamAddress(), err)
		}
		newDownstream = &shared.ServerNodeClient{
			Id:       int(req.GetDownstreamId()),
			Conn:     conn,
			Public:   api.NewMessageBoardClient(conn),
			Internal: api.NewInternalMessageBoardServiceClient(conn),
		}
	}

	if s.downstreamClient != nil {
		s.downstreamClient.Conn.Close()
	}
	s.downstreamClient = newDownstream
	shared.Logger.InfoContext(ctx, "rewired downstream", "downstream", req.GetDownstreamAddress())

	newUpstream := s.upstreamClient
	if req.UpstreamAddress == nil {
		newUpstream = nil
	} else if s.upstreamClient == nil || s.upstreamClient.Conn.Target() != req.GetUpstreamAddress() {
		conn, err := grpc.NewClient(req.GetUpstreamAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to open upstream client connection to %q: %v", req.GetUpstreamAddress(), err)
		}
		newUpstream = &shared.ServerNodeClient{
			Id:       int(req.GetUpstreamId()),
			Conn:     conn,
			Public:   api.NewMessageBoardClient(conn),
			Internal: api.NewInternalMessageBoardServiceClient(conn),
		}
	}

	if s.upstreamClient != nil {
		s.upstreamClient.Conn.Close()
	}
	s.upstreamClient = newUpstream
	s.upstreamCount = req.UpstreamCount
	shared.Logger.InfoContext(ctx, "rewired upstream", "upstream", req.GetUpstreamAddress())

	return &emptypb.Empty{}, nil
}
