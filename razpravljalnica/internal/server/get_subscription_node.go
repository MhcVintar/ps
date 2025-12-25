package server

import (
	"context"
	"math/rand"
	"razpravljalnica/internal/api"

	"github.com/google/uuid"
)

func (s *ServerNode) GetSubscriptionNode(ctx context.Context, req *api.SubscriptionNodeRequest) (*api.SubscriptionNodeResponse, error) {
	shouldPassUpstream := rand.Float64() >= 1.0/float64(s.upstreamCount+1)
	if !s.isTail() && shouldPassUpstream {
		return s.upstreamClient.Public.GetSubscriptionNode(ctx, req)
	}

	return &api.SubscriptionNodeResponse{
		SubscribeToken: uuid.NewString(),
		Node: &api.NodeInfo{
			NodeId:  s.id,
			Address: s.address,
		},
	}, nil
}
