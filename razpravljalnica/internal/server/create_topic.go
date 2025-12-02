package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"
)

func (s *Server) CreateTopic(ctx context.Context, request *api.CreateTopicRequest) (*api.Topic, error) {
	topic := shared.Topic{
		Name: request.Name,
	}
	s.db.Create(&topic)
	shared.Logger.InfoContext(ctx, "topic created", "topic", topic)

	return &api.Topic{
		Id:   topic.ID,
		Name: topic.Name,
	}, nil
}
