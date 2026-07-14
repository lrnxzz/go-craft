package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "gen",
		Short:         "Code generators for go-craft",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(blocksCommand())
	root.AddCommand(biomesCommand())

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func ident(name string) string {
	var b strings.Builder
	for _, part := range strings.Split(name, "_") {
		if part != "" {
			b.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}

	return b.String()
}
