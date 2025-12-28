package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/tactician/pkg/commands/apply"
	"github.com/go-go-golems/tactician/pkg/commands/goals"
	"github.com/go-go-golems/tactician/pkg/commands/graph"
	"github.com/go-go-golems/tactician/pkg/commands/history"
	"github.com/go-go-golems/tactician/pkg/commands/initcmd"
	"github.com/go-go-golems/tactician/pkg/commands/node"
	"github.com/go-go-golems/tactician/pkg/commands/search"
	"github.com/spf13/cobra"
)

func buildRoot() *cobra.Command {
	return &cobra.Command{
		Use:   "tactician",
		Short: "Decompose software projects into task DAGs using reusable tactics",
	}
}

func main() {
	root := buildRoot()

	if err := initcmd.RegisterInitCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering init commands: %v\n", err)
		os.Exit(1)
	}
	if err := node.RegisterNodeCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering node commands: %v\n", err)
		os.Exit(1)
	}
	if err := graph.RegisterGraphCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering graph commands: %v\n", err)
		os.Exit(1)
	}
	if err := goals.RegisterGoalsCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering goals commands: %v\n", err)
		os.Exit(1)
	}
	if err := history.RegisterHistoryCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering history commands: %v\n", err)
		os.Exit(1)
	}
	if err := search.RegisterSearchCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering search commands: %v\n", err)
		os.Exit(1)
	}
	if err := apply.RegisterApplyCommands(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering apply commands: %v\n", err)
		os.Exit(1)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
