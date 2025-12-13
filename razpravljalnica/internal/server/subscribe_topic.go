package server

import (
	"razpravljalnica/internal/api"

	"google.golang.org/grpc"
)

func (s *ServerNode) SubscribeTopic(req *api.SubscribeTopicRequest, stream grpc.ServerStreamingServer[api.MessageEvent]) error {
	ctx := stream.Context()
	observable, cancel := s.messageEventObserver.Observe(ctx, req.SubscribeToken, req.TopicId...)
	defer cancel()

	var sequenceNumber int64
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
