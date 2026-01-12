package main

import (
	"context"
	"fmt"
	"razpravljalnica/internal/api"
)

func subscribeDemo() {
	// Connect to the control plane
	ctlClient, ctlConn, err := connectToControlPlane("localhost:8000", "localhost:8002", "localhost:8004")
	if err != nil {
		panic(err)
	}
	defer ctlConn.Close()

	// Connect to the data plane
	ctx := context.Background()
	headClient, headConn, _, tailConn, err := connectToDataPlane(ctx, ctlClient)
	if err != nil {
		panic(err)
	}
	defer headConn.Close()
	defer tailConn.Close()

	// Get subscription information using headClient
	res, err := headClient.GetSubscriptionNode(ctx, &api.SubscriptionNodeRequest{
		UserId:  1,
		TopicId: []int64{1, 2},
	})
	if err != nil {
		panic(err)
	}

	// Subscribe to updates using new client
	subscribeClient, subscribeConn, err := connectToSubscriptionNode(res.Node.Address)
	if err != nil {
		panic(err)
	}
	defer subscribeConn.Close()

	stream, err := subscribeClient.SubscribeTopic(ctx, &api.SubscribeTopicRequest{
		TopicId:        []int64{1, 2},
		UserId:         1,
		FromMessageId:  0,
		SubscribeToken: res.SubscribeToken,
	})

	// Receive messages from the stream
	for {
		msg, err := stream.Recv()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Received message: %+v\n", msg)
	}
}
