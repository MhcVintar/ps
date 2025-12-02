package server

import (
	"context"
	"errors"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func (s *Server) DeleteMessage(ctx context.Context, request *api.DeleteMessageRequest) (*emptypb.Empty, error) {
	var message shared.Message
	result := s.db.First(&message, request.MessageId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			shared.Logger.InfoContext(ctx, "message already deleted", "id", request.MessageId)
			return &emptypb.Empty{}, nil
		}

		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %w", result.Error)
	}

	if message.UserID != request.UserId {
		shared.Logger.WarnContext(ctx, "unauthorized to delete message", "id", message.ID)
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized to delete message %v", message.ID)
	}

	s.db.Delete(&message)
	shared.Logger.InfoContext(ctx, "message deleted", "message", message)

	s.pubSub.Publish(ctx, request.TopicId, &api.MessageEvent{
		Op: api.OpType_OP_DELETE,
		Message: &api.Message{
			Id:        message.ID,
			TopicId:   message.TopicID,
			UserId:    message.UserID,
			Text:      message.Text,
			CreatedAt: timestamppb.New(message.CreatedAt),
			Likes:     message.Likes,
		},
	})

	return &emptypb.Empty{}, nil
}
