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
			&cli.Int64Flag{
				Name:  "downstream-id",
				Usage: "ID of the downstream server",
			},
			&cli.StringFlag{
				Name:  "downstream-address",
				Usage: "Address of the downstream server",
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			id := command.Int("id")
			address := command.String("address")
			control := command.String("control")

			var downstreamID *int64
			if command.IsSet("downstream-id") {
				downstreamID = shared.AnyPtr(command.Int64("downstream-id"))
			}
			var downstreamAddress *string
			if command.IsSet("downstream-address") {
				downstreamAddress = shared.AnyPtr(command.String("downstream-address"))
			}

			server, err := server.NewServerNode(id, address, control, downstreamID, downstreamAddress)
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
