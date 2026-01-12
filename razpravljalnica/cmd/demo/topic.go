package main

import (
	"context"
	"fmt"
	"razpravljalnica/internal/api"

	"google.golang.org/protobuf/types/known/emptypb"
)

func topicDemo() {
	// Connect to the control plane
	ctlClient, ctlConn, err := connectToControlPlane("localhost:8000", "localhost:8002", "localhost:8004")
	if err != nil {
		panic(err)
	}
	defer ctlConn.Close()

	// Connect to the data plane
	ctx := context.Background()
	headClient, headConn, tailClient, tailConn, err := connectToDataPlane(ctx, ctlClient)
	if err != nil {
		panic(err)
	}
	defer headConn.Close()
	defer tailConn.Close()

	// Create topics using headClient
	names := []string{"Gorsko kolesarstvo", "Računalništvo"}

	for _, name := range names {
		req := &api.CreateTopicRequest{
			Name: name,
		}
		_, err := headClient.CreateTopic(ctx, req)
		if err != nil {
			panic(err)
		}
	}

	// List all res using tailClient
	res, err := tailClient.ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		panic(err)
	}

	fmt.Println("List of topics:")
	for _, topic := range res.Topics {
		fmt.Printf("  - %+v\n", topic)
	}
}
