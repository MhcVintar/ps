package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerNode) Rewire(ctx context.Context, req *api.RewireRequest) (*emptypb.Empty, error) {
	if err := s.rewireHelper(ctx, &s.downstreamClient, req.Downstream); err != nil {
		return nil, err
	}

	if err := s.rewireHelper(ctx, &s.upstreamClient, req.Upstream); err != nil {
		return nil, err
	}

	s.upstreamCount = req.UpstreamCount

	return &emptypb.Empty{}, nil
}

func (s *ServerNode) rewireHelper(ctx context.Context, client **shared.ServerNodeClient, target *api.RewireRequest_Target) error {
	if target == nil {
		return nil
	}

	if target.Drop {
		if *client != nil {
			shared.Logger.InfoContext(ctx, "dropping connection", "client", (*client).Id)
			(*client).Conn.Close()
		}
		*client = nil
		return nil
	}

	if *client != nil && (*client).Conn.Target() == target.Address {
		shared.Logger.InfoContext(ctx, "already connected", "client", target.Address)
		return nil
	}

	if *client != nil {
		(*client).Conn.Close()
	}

	conn, err := grpc.NewClient(target.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	*client = &shared.ServerNodeClient{
		Id:       target.Id,
		Conn:     conn,
		Public:   api.NewMessageBoardClient(conn),
		Internal: api.NewInternalMessageBoardServiceClient(conn),
	}
	shared.Logger.InfoContext(ctx, "connected", "client", target.Address)

	return nil
}
