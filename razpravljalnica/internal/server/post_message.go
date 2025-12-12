package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerNode) PostMessage(ctx context.Context, req *api.PostMessageRequest) (*api.Message, error) {
	if !s.isTail() {
		return s.upstreamClient.Public.PostMessage(ctx, req)
	}

	user, err := s.db.FindUserByID(req.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "user %v doesn't exist", req.UserId)
	}

	topic, err := s.db.FindTopicByID(req.TopicId)
	if err != nil {
		return nil, err
	}
	if topic == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "topic %v doesn't exist", req.TopicId)
	}

	messageModel := database.Message{
		TopicID: req.TopicId,
		UserID:  req.UserId,
		Text:    req.Text,
		Likes:   0,
	}

	if err := s.db.Save(ctx, &messageModel); err != nil {
		shared.Logger.ErrorContext(ctx, "failed to save message", "error", err)
		return nil, err
	}

	return database.ToApiMessage(&messageModel), nil
}
