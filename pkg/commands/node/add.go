package node

import (
	"context"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type NodeAddCommand struct {
	*cmds.CommandDefinition
}

type NodeAddSettings struct {
	NodeID string `glazed.parameter:"node-id"`
	Output string `glazed.parameter:"output"`
	Type   string `glazed.parameter:"type"`
	Status string `glazed.parameter:"status"`
}

func NewNodeAddCommand() (*NodeAddCommand, error) {
	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("node-id", fields.TypeString,
				fields.WithHelp("Node ID"),
				fields.WithRequired(true),
			),
			fields.New("output", fields.TypeString,
				fields.WithHelp("Node output artifact"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("type", fields.TypeString,
				fields.WithHelp("Node type"),
				fields.WithDefault("project_artifact"),
			),
			fields.New("status", fields.TypeChoice,
				fields.WithHelp("Initial status"),
				fields.WithChoices("pending", "complete"),
				fields.WithDefault("pending"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"add",
		cmds.WithShort("Add a new node"),
		cmds.WithSchema(s),
	)

	return &NodeAddCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &NodeAddCommand{}

func (c *NodeAddCommand) Run(ctx context.Context, vals *values.Values) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &NodeAddSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode node add settings")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	existing, err := st.Project.GetNode(ctx, settings.NodeID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.Errorf("node already exists: %s", settings.NodeID)
	}

	node := &db.Node{
		ID:     settings.NodeID,
		Type:   settings.Type,
		Output: settings.Output,
		Status: settings.Status,
	}
	if node.Status == "complete" {
		now := time.Now().UTC()
		node.CompletedAt = &now
	}

	if err := st.Project.AddNode(ctx, node); err != nil {
		return err
	}

	details := "Created node: " + settings.NodeID
	nodeID := settings.NodeID
	if err := st.Project.LogAction(ctx, "node_created", &details, &nodeID, nil); err != nil {
		return err
	}

	st.Dirty = true
	return st.Save(ctx)
}
