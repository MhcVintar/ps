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

func (s *ServerNode) UpdateMessage(ctx context.Context, req *api.UpdateMessageRequest) (*api.Message, error) {
	if !s.isTail() {
		message, err := s.upstreamClient.Public.UpdateMessage(ctx, req)
		if err == nil {
			s.messageEventObserver.Notify(ctx, req.TopicId, &api.MessageEvent{
				EventAt: timestamppb.Now(),
				Op:      api.OpType_OP_UPDATE,
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

	if message.UserID != req.UserId {
		shared.Logger.InfoContext(ctx, "not allowed to update message", "message", message)
		return nil, status.Errorf(codes.PermissionDenied, "not allowed to update message %v", message.ID)
	}

	message.Text = req.Text
	if err := s.db.Save(ctx, message); err != nil {
		shared.Logger.ErrorContext(ctx, "failed to update message", "error", err, "message", message)
		return nil, err
	}

	shared.Logger.InfoContext(ctx, "updated message", "message", message)
	return database.ToApiMessage(message), nil
}
