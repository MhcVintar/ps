package control

import (
	"context"
	"razpravljalnica/internal/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *ControlNode) GetClusterState(ctx context.Context, _ *emptypb.Empty) (*api.GetClusterStateResponse, error) {
	if !c.isLeader() {
		return nil, status.Error(codes.PermissionDenied, "reads are only allowed from leader")
	}

	nHealthy := len(c.healthyClients)
	if nHealthy == 0 {
		return nil, status.Error(codes.Unavailable, "chain is dead")
	}

	head := c.healthyClients[0]
	tail := c.healthyClients[nHealthy-1]
	return &api.GetClusterStateResponse{
		Head: &api.NodeInfo{
			NodeId:  head.Id,
			Address: head.Conn.Target(),
		},
		Tail: &api.NodeInfo{
			NodeId:  tail.Id,
			Address: tail.Conn.Target(),
		},
	}, nil
}
