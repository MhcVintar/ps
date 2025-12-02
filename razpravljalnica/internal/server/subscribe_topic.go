package server

import (
	"fmt"
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
)

func (s *Server) SubscribeTopic(request *api.SubscribeTopicRequest, stream grpc.ServerStreamingServer[api.MessageEvent]) error {
	ctx := stream.Context()

	subscription := s.pubSub.Subscribe(ctx, request.SubscribeToken, request.TopicId...)
	defer s.pubSub.Unsubscribe(ctx, request.SubscribeToken, request.TopicId...)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-subscription:
			if !ok {
				shared.Logger.ErrorContext(ctx, "subscription closed", "subscribe_token", request.SubscribeToken)
				return fmt.Errorf("subscription %v closed", request.SubscribeToken)
			}

			if err := stream.Send(event); err != nil {
				shared.Logger.ErrorContext(ctx, "failed to send event", "event", event, "subscribe_token", request.SubscribeToken)
				return err
			}
		}
	}
}
