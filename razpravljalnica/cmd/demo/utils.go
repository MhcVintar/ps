package main

import (
	"context"
	"fmt"
	"razpravljalnica/internal/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func connectToControlPlane(
	addresses ...string,
) (
	client api.ControlPlaneClient,
	conn *grpc.ClientConn,
	err error,
) {
	for _, address := range addresses {
		conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}

		fmt.Printf("connected to control plane at %v\n", conn.Target())
		return api.NewControlPlaneClient(conn), conn, nil
	}

	return nil, nil, err
}

func connectToDataPlane(
	ctx context.Context,
	client api.ControlPlaneClient,
) (
	head api.MessageBoardClient,
	headConn *grpc.ClientConn,
	tail api.MessageBoardClient,
	tailConn *grpc.ClientConn,
	err error,
) {
	res, err := client.GetClusterState(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	headConn, err = grpc.NewClient(res.Head.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fmt.Printf("connected to head at %v\n", headConn.Target())
	head = api.NewMessageBoardClient(headConn)

	tailConn, err = grpc.NewClient(res.Tail.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		headConn.Close()
		return nil, nil, nil, nil, err
	}
	fmt.Printf("connected to tail at %v\n", tailConn.Target())
	tail = api.NewMessageBoardClient(tailConn)

	return head, headConn, tail, tailConn, nil
}

func connectToSubscriptionNode(
	address string,
) (
	client api.MessageBoardClient,
	conn *grpc.ClientConn,
	err error,
) {
	conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("connected to subscription node at %v\n", conn.Target())
	return api.NewMessageBoardClient(conn), conn, nil
}
