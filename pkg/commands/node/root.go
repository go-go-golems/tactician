package node

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/tactician/pkg/commands/common"
	"github.com/spf13/cobra"
)

func RegisterNodeCommands(root *cobra.Command) error {
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Manage nodes in the project graph",
	}

	showCmd, err := NewNodeShowCommand()
	if err != nil {
		return err
	}
	cobraShowCmd, err := cli.BuildCobraCommandFromCommand(
		showCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}
	nodeCmd.AddCommand(cobraShowCmd)

	addCmd, err := NewNodeAddCommand()
	if err != nil {
		return err
	}
	cobraAddCmd, err := cli.BuildCobraCommandFromCommand(
		addCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}
	nodeCmd.AddCommand(cobraAddCmd)

	editCmd, err := NewNodeEditCommand()
	if err != nil {
		return err
	}
	cobraEditCmd, err := cli.BuildCobraCommandFromCommand(
		editCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}
	nodeCmd.AddCommand(cobraEditCmd)

	deleteCmd, err := NewNodeDeleteCommand()
	if err != nil {
		return err
	}
	cobraDeleteCmd, err := cli.BuildCobraCommandFromCommand(
		deleteCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: common.AppName}),
	)
	if err != nil {
		return err
	}
	nodeCmd.AddCommand(cobraDeleteCmd)

	root.AddCommand(nodeCmd)
	return nil
}
