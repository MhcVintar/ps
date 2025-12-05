package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerNode) Rewire(ctx context.Context, req *api.RewireRequest) (*emptypb.Empty, error) {
	newDownstream := s.downstreamClient
	if req.DownstreamAddress == nil {
		newDownstream = nil
	} else if req.GetDownstreamAddress() != s.downstreamClient.Conn.Target() {
		conn, err := grpc.NewClient(req.GetDownstreamAddress())
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
	s.downstreamClient.Conn.Close()
	s.downstreamClient = newDownstream

	newUpstream := s.upstreamClient
	if req.UpstreamAddress == nil {
		newUpstream = nil
	} else if req.GetUpstreamAddress() != s.upstreamClient.Conn.Target() {
		conn, err := grpc.NewClient(req.GetUpstreamAddress())
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
	s.upstreamClient.Conn.Close()
	s.upstreamClient = newUpstream

	return &emptypb.Empty{}, nil
}
