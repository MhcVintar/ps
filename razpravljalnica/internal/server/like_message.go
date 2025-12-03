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

// TODO: implement chain
func (s *Server) LikeMessage(ctx context.Context, request *api.LikeMessageRequest) (*api.Message, error) {
	var message shared.Message
	result := s.db.First(&message, request.MessageId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.Logger.InfoContext(ctx, "message not found", "id", request.MessageId)
			return nil, status.Errorf(codes.NotFound, "message %v not found", request.MessageId)
		}

		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	var like shared.Like
	result = s.db.First(&like, request.UserId, request.MessageId)

	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
			return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
		}

		like := shared.Like{
			TopicID:   request.TopicId,
			MessageID: request.MessageId,
			UserID:    request.UserId,
		}
		s.db.Create(&like)

		message.Likes++
		s.db.Save(&message)

		shared.Logger.InfoContext(ctx, "message liked", "message", message)
	} else {
		shared.Logger.InfoContext(ctx, "message already liked", "message", message)
	}

	apiMessage := &api.Message{
		Id:        message.ID,
		TopicId:   message.TopicID,
		UserId:    message.UserID,
		Text:      message.Text,
		CreatedAt: timestamppb.New(message.CreatedAt),
		Likes:     message.Likes,
	}

	s.pubSub.Publish(ctx, request.TopicId, &api.MessageEvent{
		Op:      api.OpType_OP_LIKE,
		Message: apiMessage,
	})

	return apiMessage, nil
}
