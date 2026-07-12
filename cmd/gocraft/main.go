package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "gocraft",
		Short:         "Minecraft protocol toolbox",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(pingCommand())
	root.AddCommand(loginCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
