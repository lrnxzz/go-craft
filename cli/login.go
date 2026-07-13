package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lrnxzz/go-craft/mojang"
	"github.com/spf13/cobra"
)

func loginCommand() *cobra.Command {
	var clientID string

	command := &cobra.Command{
		Use:   "login",
		Short: "Authenticate a Microsoft account for online-mode servers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if clientID == "" {
				return fmt.Errorf("a Microsoft application client id is required (--client-id or GOCRAFT_CLIENT_ID)")
			}

			provider := mojang.Microsoft{
				ClientID: clientID,
				Prompt: func(code mojang.DeviceCode) {
					fmt.Fprintf(os.Stderr, "visit %s and enter code %s\n", code.VerificationURI, code.UserCode)
				},
			}

			session, err := provider.Authenticate(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "logged in as %s (%s)\n", session.Profile.Name, session.Profile.ID)

			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			return encoder.Encode(session)
		},
	}

	command.Flags().StringVar(&clientID, "client-id", os.Getenv("GOCRAFT_CLIENT_ID"), "Microsoft application (Azure) client id")

	return command
}
