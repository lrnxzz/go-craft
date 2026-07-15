package main

import (
	"encoding/json"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// generator turns one embedded JSON asset into one Go source file of named constants.
type generator[T any] struct {
	name     string
	asset    string
	output   string
	template *template.Template
}

type versionedEntries[T any] struct {
	Version string
	Entries T
}

func (g generator[T]) command() *cobra.Command {
	return &cobra.Command{
		Use:   g.name + " <version>",
		Short: "Generate " + g.output + " for a codec version from its embedded assets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return g.render(args[0])
		},
	}
}

func (g generator[T]) render(version string) error {
	raw, err := os.ReadFile("../assets/" + g.asset)
	if err != nil {
		return err
	}

	var entries T
	if err := json.Unmarshal(raw, &entries); err != nil {
		return err
	}

	input := versionedEntries[T]{
		Version: version,
		Entries: entries,
	}

	var rendered strings.Builder
	if err := g.template.Execute(&rendered, input); err != nil {
		return err
	}

	formatted, err := format.Source([]byte(rendered.String()))
	if err != nil {
		return err
	}

	return os.WriteFile(g.output, formatted, 0o644)
}
