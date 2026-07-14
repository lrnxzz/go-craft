package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
	"github.com/lrnxzz/go-craft/codec/v765/blocks"
	"github.com/spf13/cobra"
)

func joinCommand() *cobra.Command {
	var (
		username string
		timeout  time.Duration
		observe  bool
	)

	command := &cobra.Command{
		Use:   "join <host[:port]>",
		Short: "Connect a bot to a server and disconnect once it spawns",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, port := resolveAddress(args[0])
			address := net.JoinHostPort(host, strconv.Itoa(int(port)))

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			conn, err := gocraft.Dial(ctx, address)
			if err != nil {
				return err
			}

			client := gocraft.NewClient(conn, v765.Protocol())

			slog.Info("connecting", "server", address, "as", username)

			ready := func(c *gocraft.Client, join *v765.JoinGame) error {
				slog.Info("joined", "entity", join.EntityID, "dimension", join.DimensionName)
				if !observe {
					return c.Close()
				}

				slog.Info("observing", "for", timeout)

				return nil
			}

			session, err := v765.Join(client, host, port, username, ready)
			if err != nil {
				return err
			}

			if err := client.Run(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			if observe {
				perceived(session)
			}

			slog.Info("disconnected")

			return nil
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_bot", "bot username (offline mode)")
	command.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "connection timeout")
	command.Flags().BoolVar(&observe, "observe", false, "stay connected until timeout and report the perceived world")

	return command
}

func perceived(session *v765.Session) {
	player := session.Player()
	feet := player.Position.Floor().Add(0, -1, 0)

	state, _ := session.World().Block(feet.X, feet.Y, feet.Z)

	name := "air"
	if block, ok := blocks.Of(state); ok {
		name = string(block.Name)
	}

	slog.Info("perceived",
		"position", player.Position,
		"health", player.Health,
		"loaded_chunks", session.World().Loaded(),
		"standing_on", name)
}

func resolveAddress(arg string) (string, uint16) {
	host, port, err := net.SplitHostPort(arg)
	if err != nil {
		return arg, 25565
	}

	parsed, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return host, 25565
	}

	return host, uint16(parsed)
}
