package graph

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/tactician/pkg/commands/common"
	"github.com/spf13/cobra"
)

func RegisterGraphCommands(root *cobra.Command) error {
	graphCmd, err := NewGraphCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		graphCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}

	root.AddCommand(cobraCmd)
	return nil
}
