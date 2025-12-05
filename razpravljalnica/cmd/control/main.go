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
				Name:     "host",
				Aliases:  []string{"h"},
				Usage:    "Host for the control node",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Usage:    "Port for the control node",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "server-nodes",
				Aliases:  []string{"s"},
				Usage:    "Number of server nodes",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			host := command.String("host")
			port := command.Int("port")
			nServerNodes := command.Int("server-nodes")

			control, err := control.NewControlNode(host, port, nServerNodes)
			if err != nil {
				return err
			}

			return control.Run()
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		shared.Logger.ErrorContext(ctx, "client node failed", "error", err)
		os.Exit(1)
	}
}
