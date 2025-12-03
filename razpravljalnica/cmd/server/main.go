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
		Usage: "Start the server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Aliases:  []string{"a"},
				Usage:    "Address for the server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "id",
				Aliases:  []string{"i"},
				Usage:    "ID for the server",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "next",
				Aliases: []string{"n"},
				Usage:   "Address of the next server in the chain",
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			address := command.String("address")
			id := command.String("id")
			nextAddress := command.String("next")

			server, err := server.NewServer(address, id, nextAddress)
			if err != nil {
				return err
			}

			defer server.Close()

			return server.Run()
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		shared.Logger.ErrorContext(ctx, "server failed", "error", err)
		os.Exit(1)
	}
}
