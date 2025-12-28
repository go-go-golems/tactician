package node

import (
	"context"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type NodeEditCommand struct {
	*cmds.CommandDefinition
}

type NodeEditSettings struct {
	NodeIDs []string `glazed.parameter:"node-ids"`
	Status  string   `glazed.parameter:"status"`
}

func NewNodeEditCommand() (*NodeEditCommand, error) {
	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("node-ids", fields.TypeStringList,
				fields.WithHelp("Node ID(s) to edit (supports multiple: node edit id1 id2 --status complete)"),
				fields.WithRequired(true),
			),
		),
		schema.WithFields(
			fields.New("status", fields.TypeChoice,
				fields.WithHelp("Update status (applied to all specified nodes)"),
				fields.WithChoices("pending", "complete"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"edit",
		cmds.WithShort("Edit one or more nodes"),
		cmds.WithLong("Update status for nodes. Supports batch operations: 'node edit id1 id2 --status complete'"),
		cmds.WithSchema(s),
	)

	return &NodeEditCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.BareCommand = &NodeEditCommand{}

func (c *NodeEditCommand) Run(ctx context.Context, vals *values.Values) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &NodeEditSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode node edit settings")
	}
	if len(settings.NodeIDs) == 0 {
		return errors.New("at least one node id is required")
	}
	if settings.Status == "" {
		return errors.New("--status is required")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	var completedAt *time.Time
	if settings.Status == "complete" {
		now := time.Now().UTC()
		completedAt = &now
	}

	for _, id := range settings.NodeIDs {
		n, err := st.Project.GetNode(ctx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return errors.Errorf("node not found: %s", id)
		}

		if err := st.Project.UpdateNodeStatus(ctx, id, settings.Status, completedAt); err != nil {
			return err
		}

		action := "node_updated"
		details := "Updated " + id + " status to " + settings.Status
		if settings.Status == "complete" {
			action = "node_completed"
		}
		nodeID := id
		if err := st.Project.LogAction(ctx, action, &details, &nodeID, nil); err != nil {
			return err
		}
	}

	st.Dirty = true
	return st.Save(ctx)
}
