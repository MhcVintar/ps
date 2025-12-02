package main

import (
	"context"
	"os"

	"razpravljalnica/internal/server"
	"razpravljalnica/internal/shared"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:  "server",
		Usage: "Start the gRPC server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Aliases:  []string{"a"},
				Usage:    "Address for the gRPC server",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			address := command.String("address")

			server, err := server.NewServer(address)
			if err != nil {
				return err
			}

			return server.Run()
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		shared.Logger.ErrorContext(ctx, "server failed", "error", err)
		os.Exit(1)
	}
}
