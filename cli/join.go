package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	gocraft "github.com/lrnxzz/go-craft"
	"github.com/lrnxzz/go-craft/codec/v765"
	"github.com/spf13/cobra"
)

func joinCommand() *cobra.Command {
	var (
		username string
		timeout  time.Duration
	)

	command := &cobra.Command{
		Use:   "join <host[:port]>",
		Short: "Connect a bot to a server and disconnect once it spawns",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			port := uint16(25565)
			if h, p, err := net.SplitHostPort(args[0]); err == nil {
				host = h
				if parsed, err := strconv.ParseUint(p, 10, 16); err == nil {
					port = uint16(parsed)
				}
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			conn, err := gocraft.Dial(ctx, net.JoinHostPort(host, strconv.Itoa(int(port))))
			if err != nil {
				return err
			}

			client := gocraft.NewClient(conn, v765.Protocol())

			fmt.Fprintf(cmd.OutOrStdout(), "connecting to %s:%d as %s...\n", host, port, username)

			ready := func(c *gocraft.Client, join *v765.JoinGame) error {
				fmt.Fprintf(cmd.OutOrStdout(), "joined! entity %d, dimension %s, gamemode %d — disconnecting\n",
					join.EntityID, join.DimensionName, join.GameMode)

				return c.Close()
			}

			if err := v765.Join(client, host, port, username, ready); err != nil {
				return err
			}

			if err := client.Run(ctx); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "disconnected cleanly")

			return nil
		},
	}

	command.Flags().StringVar(&username, "username", "gocraft_bot", "bot username (offline mode)")
	command.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "connection timeout")

	return command
}
