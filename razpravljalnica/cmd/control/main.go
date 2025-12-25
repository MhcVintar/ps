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
				Name:     "grpc-address",
				Usage:    "Address for the server",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "raft-address",
				Usage:    "Address for the server",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "server-nodes",
				Aliases:  []string{"n"},
				Usage:    "Number of server nodes",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "server-nodes-start-port",
				Aliases:  []string{"p"},
				Usage:    "Starting port for server nodes",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "raft-dir",
				Aliases:  []string{"r"},
				Usage:    "Directory for raft to persist changes",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "peer",
				Usage: "Raft addresses of peer control nodes",
			},
			&cli.BoolFlag{
				Name:  "bootstrap",
				Usage: "Bootstrap this node as the initial cluster member",
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			grpcAddress := command.String("grpc-address")
			raftAddress := command.String("raft-address")
			nServerNodes := command.Int("server-nodes")
			serverNodesStartPort := command.Int("server-nodes-start-port")
			raftDir := command.String("raft-dir")
			peers := command.StringSlice("peer")
			bootstrap := command.Bool("bootstrap")

			control, err := control.NewControlNode(grpcAddress, raftAddress, nServerNodes, serverNodesStartPort, raftDir, peers, bootstrap)
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
