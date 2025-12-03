package server

import (
	"context"
	"razpravljalnica/internal/api"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: decide which node ges this by calling control plane
func (s *Server) GetSubscriptionNode(ctx context.Context, request *api.SubscriptionNodeRequest) (*api.SubscriptionNodeResponse, error) {
	token, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate subscribe token: %v", err)
	}

	return &api.SubscriptionNodeResponse{
		SubscribeToken: token.String(),
		Node: &api.NodeInfo{
			NodeId:  s.id,
			Address: s.address,
		},
	}, nil
}
