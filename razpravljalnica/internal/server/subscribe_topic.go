package server

import (
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) SubscribeTopic(request *api.SubscribeTopicRequest, stream grpc.ServerStreamingServer[api.MessageEvent]) error {
	ctx := stream.Context()

	subscription := s.pubSub.Subscribe(ctx, request.SubscribeToken, request.TopicId...)
	defer s.pubSub.Unsubscribe(ctx, request.SubscribeToken, request.TopicId...)

	var messages []shared.Message
	result := s.db.Where("id > ? AND topic_id IN ?", request.FromMessageId, request.TopicId).
		Order("id ASC").
		Find(&messages)
	if result.Error != nil {
		shared.Logger.ErrorContext(ctx, "internal database error", "error", result.Error)
		return status.Errorf(codes.Internal, "internal database error: %v", result.Error)
	}

	var sequence_number int64
	for _, message := range messages {
		stream.Send(&api.MessageEvent{
			SequenceNumber: sequence_number,
			Op:             api.OpType_OP_POST,
			EventAt:        timestamppb.Now(),
			Message: &api.Message{
				Id:        message.ID,
				TopicId:   message.TopicID,
				UserId:    message.UserID,
				Text:      message.Text,
				CreatedAt: timestamppb.New(message.CreatedAt),
				Likes:     message.Likes,
			},
		})
		sequence_number++
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-subscription:
			event.SequenceNumber = sequence_number
			sequence_number++

			if !ok {
				shared.Logger.ErrorContext(ctx, "subscription closed", "subscribe_token", request.SubscribeToken)
				return status.Errorf(codes.Internal, "subscription %v closed", request.SubscribeToken)
			}

			if err := stream.Send(event); err != nil {
				shared.Logger.ErrorContext(ctx, "failed to send event", "event", event, "subscribe_token", request.SubscribeToken)
				return err
			}
		}
	}
}
