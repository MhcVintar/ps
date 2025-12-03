package server

import (
	"context"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) GetMessages(ctx context.Context, request *api.GetMessagesRequest) (*api.GetMessagesResponse, error) {
	var messages []shared.Message
	result := s.db.Where("id >= ? AND topic_id = ?", request.FromMessageId, request.TopicId).
		Order("id ASC").
		Limit(int(request.Limit)).
		Find(&messages)
	if result.Error != nil {
		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return nil, status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	apiMessages := make([]*api.Message, len(messages))
	for i, message := range messages {
		apiMessages[i] = &api.Message{
			Id:        message.ID,
			TopicId:   message.TopicID,
			UserId:    message.UserID,
			Text:      message.Text,
			CreatedAt: timestamppb.New(message.CreatedAt),
			Likes:     message.Likes,
		}
	}

	shared.Logger.InfoContext(ctx, "messages listed", "messages", apiMessages)

	return &api.GetMessagesResponse{
		Messages: apiMessages,
	}, nil
}
