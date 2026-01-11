package main

import (

	"context"
	"fmt"
	"razpravljalnica/internal/client"
	"github.com/urfave/cli/v3"
	"os"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:  "server",
		Usage: "Start the server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Aliases:  []string{"a"},
				Usage:    "Address for the server",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name: "port",
				Aliases: []string{"p"},
				Usage: "Port to connect to server",
				Value: 8080,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			address := c.String("address")

			port := c.Int("port")

			client.Bootstrap(address, port)

			return nil
		},
	}
	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Println("fffff")
	}
}

