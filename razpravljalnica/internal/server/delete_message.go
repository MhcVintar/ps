package server

import (
	"context"
	"errors"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Server) DeleteMessage(ctx context.Context, request *api.DeleteMessageRequest) (*emptypb.Empty, error) {
	event := &api.MessageEvent{
		Op: api.OpType_OP_DELETE,
		Message: &api.Message{
			Id: request.MessageId,
		},
	}

	if s.nextServer != nil {
		shared.Logger.InfoContext(ctx, "forwarding request to next server in chain", "request", request)
		_, err := s.nextServer.DeleteMessage(ctx, request)
		if err != nil {
			return nil, err
		}
	} else {
		var message shared.Message
		result := s.db.First(&message, request.MessageId)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
				return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
			}
		} else if message.UserID != request.UserId {
			shared.Logger.WarnContext(ctx, "unauthorized to delete message", "id", message.ID)
			return nil, status.Errorf(codes.PermissionDenied, "unauthorized to delete message %v", message.ID)
		}
	}

	s.db.Delete(event.Message)
	shared.Logger.InfoContext(ctx, "message deleted", "message", event.Message)

	s.pubSub.Publish(ctx, request.TopicId, event)

	return &emptypb.Empty{}, nil
}
