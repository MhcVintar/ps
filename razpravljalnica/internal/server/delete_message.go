package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ServerNode) DeleteMessage(ctx context.Context, req *api.DeleteMessageRequest) (*emptypb.Empty, error) {
	if !s.isTail() {
		empty, err := s.upstreamClient.Public.DeleteMessage(ctx, req)
		if err == nil {
			s.messageEventObserver.Notify(ctx, req.TopicId, &api.MessageEvent{
				EventAt: timestamppb.Now(),
				Op:      api.OpType_OP_DELETE,
				Message: &api.Message{
					Id: req.MessageId,
				},
			})
		}

		return empty, err
	}

	message, err := s.db.FindMessageByID(req.MessageId)
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to find message", "error", err, "id", req.MessageId)
		return nil, err
	}

	if message != nil {
		if message.UserID != req.UserId {
			shared.Logger.InfoContext(ctx, "not allowed to delete message", "message", message)
			return nil, status.Errorf(codes.PermissionDenied, "not allowed to delete message %v", message.ID)
		}

		if err := s.db.Delete(ctx, message); err != nil {
			shared.Logger.ErrorContext(ctx, "failed to delete message", "error", err, "message", message)
			return nil, err
		}
	}

	shared.Logger.InfoContext(ctx, "deleted message", "id", req.MessageId)
	return &emptypb.Empty{}, nil
}
