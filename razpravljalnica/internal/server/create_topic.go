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

	if s.nextServer != nil {
		shared.Logger.InfoContext(ctx, "forwarding request to next server in chain", "request", request)
		apiTopic, err := s.nextServer.CreateTopic(ctx, request)
		if err != nil {
			return nil, err
		}

		topic.ID = apiTopic.Id
	}

	s.db.Save(&topic)
	shared.Logger.InfoContext(ctx, "topic created", "topic", topic)

	return &api.Topic{
		Id:   topic.ID,
		Name: topic.Name,
	}, nil
}
