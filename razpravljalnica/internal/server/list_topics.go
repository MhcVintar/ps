package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerNode) ListTopics(ctx context.Context, _ *emptypb.Empty) (*api.ListTopicsResponse, error) {
	if !s.isTail() {
		shared.Logger.WarnContext(ctx, "read operation on non-tail node", "rpc", "ListTopics")
		return nil, status.Error(codes.FailedPrecondition, "cannot read from non-tail node")
	}

	topicModels, err := s.db.FindAllTopics()
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to find all topics", "error", err)
		return nil, err
	}

	apiTopics := make([]*api.Topic, len(topicModels))
	for i, topicModel := range topicModels {
		apiTopics[i] = database.ToApiTopic(&topicModel)
	}

	shared.Logger.InfoContext(ctx, "listed topics")
	return &api.ListTopicsResponse{
		Topics: apiTopics,
	}, nil
}
