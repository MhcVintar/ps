package server

import (
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/database"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ServerNode) SubscribeTopic(req *api.SubscribeTopicRequest, stream grpc.ServerStreamingServer[api.MessageEvent]) error {
	ctx := stream.Context()
	observable, cancel := s.messageEventObserver.Observe(ctx, req.SubscribeToken, req.TopicId...)
	defer cancel()

	sequenceNumber := int64(1)

	for _, topicID := range req.TopicId {
		messages, err := s.db.FindAllMessages(req.FromMessageId, topicID, 1000)
		if err != nil {
			shared.Logger.ErrorContext(ctx, "failed to find messages", "error", err)
			return err
		}

		for _, message := range messages {
			stream.Send(&api.MessageEvent{
				SequenceNumber: sequenceNumber,
				EventAt:        timestamppb.New(message.CreatedAt),
				Op:             api.OpType_OP_POST,
				Message:        database.ToApiMessage(&message),
			})
			sequenceNumber++
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case messageEvent := <-observable:
			messageEvent.SequenceNumber = sequenceNumber
			sequenceNumber++

			if err := stream.Send(messageEvent); err != nil {
				return err
			}
		}
	}
}
