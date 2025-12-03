package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) ListTopics(ctx context.Context, request *emptypb.Empty) (*api.ListTopicsResponse, error) {
	var topics []shared.Topic
	result := s.db.Find(&topics)
	if result.Error != nil {
		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	apiTopis := make([]*api.Topic, len(topics))
	for i, topic := range topics {
		apiTopis[i] = &api.Topic{
			Id:   topic.ID,
			Name: topic.Name,
		}
	}

	shared.Logger.InfoContext(ctx, "topics listed", "topics", apiTopis)

	return &api.ListTopicsResponse{
		Topics: apiTopis,
	}, nil
}
