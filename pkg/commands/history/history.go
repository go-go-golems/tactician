package history

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type HistoryCommand struct {
	*cmds.CommandDefinition
}

type HistorySettings struct {
	Limit   int    `glazed.parameter:"limit"`
	Since   string `glazed.parameter:"since"`
	Summary bool   `glazed.parameter:"summary"`
}

func NewHistoryCommand() (*HistoryCommand, error) {
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	tacticianSection, err := sections.NewTacticianSection()
	if err != nil {
		return nil, err
	}

	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithFields(
			fields.New("limit", fields.TypeInteger,
				fields.WithHelp("Limit number of entries"),
				fields.WithShortFlag("l"),
			),
			fields.New("since", fields.TypeString,
				fields.WithHelp("Show actions since (e.g., 1h, 2d, 30m)"),
				fields.WithShortFlag("s"),
			),
			fields.New("summary", fields.TypeBool,
				fields.WithHelp("Show session summary instead of detailed log"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(glazedSection, tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"history",
		cmds.WithShort("View action history and session summary"),
		cmds.WithSchema(s),
	)

	return &HistoryCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &HistoryCommand{}

func (c *HistoryCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &HistorySettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode history settings")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	var since *time.Time
	if strings.TrimSpace(settings.Since) != "" {
		t, err := parseRelativeTime(settings.Since)
		if err != nil {
			return err
		}
		since = &t
	}

	if settings.Summary {
		summary, err := st.Project.GetSessionSummary(ctx, since)
		if err != nil {
			return err
		}
		row := types.NewRow(
			types.MRP("total_actions", summary.TotalActions),
			types.MRP("nodes_created", summary.NodesCreated),
			types.MRP("nodes_completed", summary.NodesCompleted),
			types.MRP("tactics_applied", summary.TacticsApplied),
			types.MRP("nodes_modified", summary.NodesModified),
		)
		return gp.AddRow(ctx, row)
	}

	var limit *int
	if settings.Limit > 0 {
		limit = &settings.Limit
	}

	logs, err := st.Project.GetActionLog(ctx, limit, since)
	if err != nil {
		return err
	}
	for _, l := range logs {
		var details any
		if l.Details != nil {
			details = *l.Details
		}
		var nodeID any
		if l.NodeID != nil {
			nodeID = *l.NodeID
		}
		var tacticID any
		if l.TacticID != nil {
			tacticID = *l.TacticID
		}
		row := types.NewRow(
			types.MRP("timestamp", l.Timestamp),
			types.MRP("action", l.Action),
			types.MRP("details", details),
			types.MRP("node_id", nodeID),
			types.MRP("tactic_id", tacticID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func parseRelativeTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty relative time")
	}

	// Support suffixes: s, m, h (time.ParseDuration) and d, w (custom).
	last := s[len(s)-1]
	numStr := s[:len(s)-1]
	n, err := strconv.Atoi(numStr)
	if err != nil {
		// As a fallback, accept Go duration strings like "90m" or "1h30m".
		d, err2 := time.ParseDuration(s)
		if err2 != nil {
			return time.Time{}, errors.Wrap(err, "parse relative time")
		}
		return time.Now().Add(-d), nil
	}

	switch last {
	case 's', 'm', 'h':
		d, err := time.ParseDuration(s)
		if err != nil {
			return time.Time{}, errors.Wrap(err, "parse duration")
		}
		return time.Now().Add(-d), nil
	case 'd':
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
	case 'w':
		return time.Now().Add(-time.Duration(n) * 7 * 24 * time.Hour), nil
	default:
		return time.Time{}, errors.Errorf("unsupported relative time suffix: %q (expected s/m/h/d/w)", string(last))
	}
}
