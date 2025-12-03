package server

import (
	"context"
	"errors"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func (s *Server) PostMessage(ctx context.Context, request *api.PostMessageRequest) (*api.Message, error) {
	var user shared.User
	result := s.db.First(&user, request.UserId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.Logger.InfoContext(ctx, "user not found", "id", request.UserId)
			return nil, status.Errorf(codes.NotFound, "user %v not found", request.UserId)
		}

		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	var topic shared.Topic
	result = s.db.First(&topic, request.TopicId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.Logger.InfoContext(ctx, "topic not found", "id", request.TopicId)
			return nil, status.Errorf(codes.NotFound, "topic %v not found", request.TopicId)
		}

		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	message := shared.Message{
		TopicID: request.TopicId,
		UserID:  request.UserId,
		Text:    request.Text,
		Likes:   0,
	}
	s.db.Create(&message)

	shared.Logger.InfoContext(ctx, "message posted", "message", message)

	apiMessage := &api.Message{
		Id:        message.ID,
		TopicId:   message.TopicID,
		UserId:    message.UserID,
		Text:      message.Text,
		CreatedAt: timestamppb.New(message.CreatedAt),
		Likes:     message.Likes,
	}

	s.pubSub.Publish(ctx, request.TopicId, &api.MessageEvent{
		Op:      api.OpType_OP_POST,
		Message: apiMessage,
	})

	return apiMessage, nil
}
