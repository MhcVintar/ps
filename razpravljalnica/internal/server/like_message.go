package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ServerNode) LikeMessage(ctx context.Context, req *api.LikeMessageRequest) (*api.Message, error) {
	if !s.isTail() {
		message, err := s.upstreamClient.Public.LikeMessage(ctx, req)
		if err == nil {
			s.messageEventObserver.Notify(ctx, req.TopicId, &api.MessageEvent{
				EventAt: timestamppb.Now(),
				Op:      api.OpType_OP_LIKE,
				Message: message,
			})
		}

		return message, err
	}

	message, err := s.db.FindMessageByID(req.MessageId)
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to find message", "error", err, "id", req.MessageId)
		return nil, err
	}
	if message == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "message %v doesn't exist", req.MessageId)
	}

	isLiked, err := s.db.HasUserLikedMessage(req.UserId, req.MessageId)
	if err != nil {
		return nil, err
	}

	if !isLiked {
		message.Likes++
		like := database.Like{
			TopicID:   req.TopicId,
			MessageID: req.MessageId,
			UserID:    req.UserId,
		}

		if err := s.db.Save(ctx, &like); err != nil {
			shared.Logger.ErrorContext(ctx, "failed to save like", "error", err)
			return nil, err
		}

		if err := s.db.Save(ctx, &message); err != nil {
			shared.Logger.ErrorContext(ctx, "failed to save message", "error", err)
			return nil, err
		}
	}

	return database.ToApiMessage(message), nil
}
