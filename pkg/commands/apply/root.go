package apply

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/tactician/pkg/commands/common"
	"github.com/spf13/cobra"
)

func RegisterApplyCommands(root *cobra.Command) error {
	applyCmd, err := NewApplyCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		applyCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}

	root.AddCommand(cobraCmd)
	return nil
}
