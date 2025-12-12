package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerNode) CreateTopic(ctx context.Context, req *api.CreateTopicRequest) (*api.Topic, error) {
	if !s.isTail() {
		return s.upstreamClient.Public.CreateTopic(ctx, req)
	}

	topicModel := database.Topic{
		Name: req.Name,
	}
	if err := s.db.Save(ctx, &topicModel); err != nil {
		shared.Logger.ErrorContext(ctx, "failed to save topic", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to save topic: %v", err)
	}

	shared.Logger.InfoContext(ctx, "created topic", "topic", topicModel)

	return database.ToApiTopic(&topicModel), nil
}
