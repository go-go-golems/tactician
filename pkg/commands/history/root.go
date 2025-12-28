package history

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/tactician/pkg/commands/common"
	"github.com/spf13/cobra"
)

func RegisterHistoryCommands(root *cobra.Command) error {
	historyCmd, err := NewHistoryCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		historyCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}

	root.AddCommand(cobraCmd)
	return nil
}
