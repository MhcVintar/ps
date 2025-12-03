package control

import (
	"context"
	"razpravljalnica/internal/api"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (c *Control) GetClusterState(ctx context.Context, _ *emptypb.Empty) (*api.GetClusterStateResponse, error) {
	return &api.GetClusterStateResponse{
		Head: &api.NodeInfo{
			NodeId:  c.headID,
			Address: c.headAddress,
		},
		Tail: &api.NodeInfo{
			NodeId:  c.tailID,
			Address: c.tailAddress,
		},
	}, nil
}
