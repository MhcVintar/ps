package main

import (
	"context"
	"fmt"
	"razpravljalnica/internal/api"
)

func messageDemo() {
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

	// Post a few messages using headClient
	reqs := []*api.PostMessageRequest{
		{UserId: 1, TopicId: 1, Text: "Pozdravljeni vsi skupaj!"},
		{UserId: 2, TopicId: 1, Text: "Živjo! Kako ste?"},
		{UserId: 3, TopicId: 2, Text: "Kateri programski jezik je najboljši?"},
		{UserId: 4, TopicId: 2, Text: "Mislim, da je Go super!"},
	}
	for _, req := range reqs {
		_, err := headClient.PostMessage(ctx, req)
		if err != nil {
			panic(err)
		}
	}

	// Retrieve and print all messages using tailClient
	for _, topicID := range []int64{1, 2} {
		req := &api.GetMessagesRequest{
			TopicId:       topicID,
			FromMessageId: 0,
			Limit:         10,
		}
		res, err := tailClient.GetMessages(ctx, req)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Messages in topic %v:\n", topicID)
		for _, msg := range res.Messages {
			fmt.Printf("  - %+v\n", msg)
		}
	}

	// Like a few messages using headClient
	likeReqs := []*api.LikeMessageRequest{
		{UserId: 2, TopicId: 1, MessageId: 1},
		{UserId: 3, TopicId: 2, MessageId: 3},
		{UserId: 4, TopicId: 2, MessageId: 3},
	}
	for _, req := range likeReqs {
		_, err := headClient.LikeMessage(ctx, req)
		if err != nil {
			panic(err)
		}
	}

	// Update a message using headClient
	updateReq := &api.UpdateMessageRequest{
		UserId:    1,
		TopicId:   1,
		MessageId: 1,
		Text:      "Pozdravljeni vsi skupaj! Upam, da ste dobro.",
	}
	_, err = headClient.UpdateMessage(ctx, updateReq)
	if err != nil {
		panic(err)
	}

	// Delete a message using headClient
	deleteReq := &api.DeleteMessageRequest{
		UserId:    4,
		TopicId:   2,
		MessageId: 4,
	}
	_, err = headClient.DeleteMessage(ctx, deleteReq)
	if err != nil {
		panic(err)
	}

	// Retrieve and print all messages again to see the updates using tailClient
	for _, topicID := range []int64{1, 2} {
		req := &api.GetMessagesRequest{
			TopicId:       topicID,
			FromMessageId: 0,
			Limit:         10,
		}
		res, err := tailClient.GetMessages(ctx, req)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Updated messages in topic %v:\n", topicID)
		for _, msg := range res.Messages {
			fmt.Printf("  - %+v\n", msg)
		}
	}
}
