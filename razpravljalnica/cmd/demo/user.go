package main

import (
	"context"
	"fmt"
	"razpravljalnica/internal/api"
)

func userDemo() {
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

	// Create users using headClient
	names := []string{"Miha", "Arne", "Uroš", "Davor"}
	users := make([]*api.User, 0, len(names))

	for _, name := range names {
		req := &api.CreateUserRequest{
			Name: name,
		}
		res, err := headClient.CreateUser(ctx, req)
		if err != nil {
			panic(err)
		}

		users = append(users, res)
	}

	fmt.Println("Created users:")
	for _, user := range users {
		fmt.Printf("  - %+v\n", user)
	}
}
