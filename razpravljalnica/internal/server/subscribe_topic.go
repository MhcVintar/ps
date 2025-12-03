package server

import (
	"razpravljalnica/internal/api"
	"razpravljalnica/internal/shared"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SubscribeTopic(request *api.SubscribeTopicRequest, stream grpc.ServerStreamingServer[api.MessageEvent]) error {
	ctx := stream.Context()

	subscription := s.pubSub.Subscribe(ctx, request.SubscribeToken, request.TopicId...)
	defer s.pubSub.Unsubscribe(ctx, request.SubscribeToken, request.TopicId...)

	// TODO: implement the from message id part (send messages as if they were just created)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-subscription:
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
