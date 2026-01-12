package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "demo",
		Usage: "Run various demo functions",
		Commands: []*cli.Command{
			{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "Run user demo",
				Action: func(ctx context.Context, command *cli.Command) error {
					userDemo()
					return nil
				},
			},
			{
				Name:    "topic",
				Aliases: []string{"t"},
				Usage:   "Run topic demo",
				Action: func(ctx context.Context, command *cli.Command) error {
					topicDemo()
					return nil
				},
			},
			{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "Run message demo",
				Action: func(ctx context.Context, command *cli.Command) error {
					messageDemo()
					return nil
				},
			},
			{
				Name:    "subscribe",
				Aliases: []string{"s"},
				Usage:   "Run subscribe demo",
				Action: func(ctx context.Context, command *cli.Command) error {
					subscribeDemo()
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
