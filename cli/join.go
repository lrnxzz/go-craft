package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/lrnxzz/go-craft/agent"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/spf13/cobra"
)

func joinCommand() *cobra.Command {
	var (
		username string
		timeout  time.Duration
	)

	command := &cobra.Command{
		Use:   "join <host[:port]>",
		Short: "Connect a bot to a server and simulate it until the timeout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			slog.Info("connecting", "server", args[0], "as", username)

			bot, err := agent.Join(ctx, args[0], username)
			if err != nil {
				return err
			}

			if err := bot.Run(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			report(bot)
			slog.Info("disconnected")

			return nil
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_bot", "bot username (offline mode)")
	command.Flags().DurationVar(&timeout, "timeout", 15*time.Second, "how long to stay connected")

	return command
}

func report(bot *agent.Agent) {
	player := bot.Player()
	feet := player.Position.Floor().Add(0, -1, 0)

	state, _ := bot.World().Block(feet.X, feet.Y, feet.Z)

	name := "air"
	if block, ok := blocks.Of(state); ok {
		name = string(block.Name)
	}

	slog.Info("perceived",
		"position", player.Position,
		"health", player.Health,
		"on_ground", player.OnGround,
		"loaded_chunks", bot.World().Loaded(),
		"standing_on", name)
}
