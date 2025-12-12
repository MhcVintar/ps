package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerNode) GetMessages(ctx context.Context, req *api.GetMessagesRequest) (*api.GetMessagesResponse, error) {
	if !s.isTail() {
		shared.Logger.WarnContext(ctx, "read operation on non-tail node", "rpc", "GetMessages")
		return nil, status.Error(codes.FailedPrecondition, "cannot read from non-tail node")
	}

	messageModels, err := s.db.FindAllMessages(req.FromMessageId, req.TopicId, int(req.Limit))
	if err != nil {
		shared.Logger.ErrorContext(ctx, "failed to find messages", "error", err)
		return nil, err
	}

	apiMessages := make([]*api.Message, len(messageModels))
	for i, messageModel := range messageModels {
		apiMessages[i] = database.ToApiMessage(&messageModel)
	}

	shared.Logger.InfoContext(ctx, "listed messages")
	return &api.GetMessagesResponse{
		Messages: apiMessages,
	}, nil
}
