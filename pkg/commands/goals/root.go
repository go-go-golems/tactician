package goals

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/tactician/pkg/commands/common"
	"github.com/spf13/cobra"
)

func RegisterGoalsCommands(root *cobra.Command) error {
	goalsCmd, err := NewGoalsCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		goalsCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}

	root.AddCommand(cobraCmd)
	return nil
}
