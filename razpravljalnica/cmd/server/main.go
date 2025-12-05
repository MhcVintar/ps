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
			&cli.IntFlag{
				Name:     "id",
				Aliases:  []string{"i"},
				Usage:    "ID for the server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "address",
				Aliases:  []string{"a"},
				Usage:    "Address for the server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "control",
				Aliases:  []string{"c"},
				Usage:    "Address of the control node",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			id := command.Int("id")
			address := command.String("address")
			control := command.String("control")

			server, err := server.NewServerNode(id, address, control)
			if err != nil {
				return err
			}

			return server.Run()
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		shared.Logger.ErrorContext(ctx, "server node failed", "error", err)
		os.Exit(1)
	}
}
