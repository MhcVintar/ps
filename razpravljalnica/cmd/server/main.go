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
				Name:  "downstream-id",
				Usage: "ID of the downstream server",
			},
			&cli.StringFlag{
				Name:  "downstream-address",
				Usage: "Address of the downstream server",
			},
			&cli.BoolFlag{
				Name: "automatic-setup",
				Usage: "Set up servers automatically",
				Value: false,
			},
			&cli.IntFlag{
				Name: "server-count",
				Usage: "This flag is used when automatic-setup flag is set to true. If this isn't set, it defaults to 1",
				Value: 1,
			},
		
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			address := command.String("address")
			if(!command.Bool("automatic-setup")){
				var downstreamID *string
				if command.IsSet("downstream-id") {
					downstreamID = shared.AnyPtr(command.String("downstream-id"))
				}
				var downstreamAddress *string
				if command.IsSet("downstream-address") {
					downstreamAddress = shared.AnyPtr(command.String("downstream-address"))
				}
			}else{
				
				for i = port; i-port <= command.Int("server-count"); i++{
					
				}
				var downstreamID *string
				if command.IsSet("downstream-id") {
					downstreamID = shared.AnyPtr(command.String("downstream-id"))
				}
				var downstreamAddress *string
				if command.IsSet("downstream-address") {
					downstreamAddress = shared.AnyPtr(command.String("downstream-address"))
				}
			}
			
			server, err := server.NewServerNode(address, downstreamID, downstreamAddress)
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
