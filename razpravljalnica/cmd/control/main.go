package main

import (
	"context"
	"os"

	"razpravljalnica/internal/control"
	"razpravljalnica/internal/shared"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:  "control",
		Usage: "Start the control",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Aliases:  []string{"a"},
				Usage:    "Address for the control",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "head-address",
				Aliases:  []string{"ha"},
				Usage:    "Address of the chain's head server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "head-id",
				Aliases:  []string{"hi"},
				Usage:    "ID of the chain's head server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "tail-address",
				Aliases:  []string{"ta"},
				Usage:    "Address of the chain's tail server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "tail-id",
				Aliases:  []string{"ti"},
				Usage:    "ID of the chain's tail server",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			address := command.String("address")
			headAddress := command.String("head-address")
			headID := command.String("head-id")
			tailAddress := command.String("tail-address")
			tailID := command.String("tail-id")

			control := control.NewControl(address, headAddress, headID, tailAddress, tailID)

			return control.Run()
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		shared.Logger.ErrorContext(ctx, "server failed", "error", err)
		os.Exit(1)
	}
}
