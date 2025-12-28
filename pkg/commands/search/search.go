package search

import (
	"context"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/tactician/pkg/commands/sections"
	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/go-go-golems/tactician/pkg/store"
	"github.com/pkg/errors"
)

type SearchCommand struct {
	*cmds.CommandDefinition
}

type SearchSettings struct {
	Query     string `glazed.parameter:"query"`
	Ready     bool   `glazed.parameter:"ready"`
	Type      string `glazed.parameter:"type"`
	Tags      string `glazed.parameter:"tags"`
	Goals     string `glazed.parameter:"goals"`
	LLMRerank bool   `glazed.parameter:"llm-rerank"`
	Limit     int    `glazed.parameter:"limit"`
	Verbose   bool   `glazed.parameter:"verbose"`
}

func NewSearchCommand() (*SearchCommand, error) {
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
		schema.WithArguments(
			fields.New("query", fields.TypeString,
				fields.WithHelp("Search query (keywords)"),
			),
		),
		schema.WithFields(
			fields.New("ready", fields.TypeBool,
				fields.WithHelp("Show only ready tactics (all dependencies satisfied)"),
				fields.WithDefault(false),
			),
			fields.New("type", fields.TypeString,
				fields.WithHelp("Filter by tactic type"),
			),
			fields.New("tags", fields.TypeString,
				fields.WithHelp("Filter by tags (comma-separated)"),
			),
			fields.New("goals", fields.TypeString,
				fields.WithHelp("Align with specific goal nodes (comma-separated)"),
			),
			fields.New("llm-rerank", fields.TypeBool,
				fields.WithHelp("Use LLM to semantically rerank results"),
				fields.WithDefault(false),
			),
			fields.New("limit", fields.TypeInteger,
				fields.WithHelp("Limit number of results"),
				fields.WithDefault(20),
				fields.WithShortFlag("l"),
			),
			fields.New("verbose", fields.TypeBool,
				fields.WithHelp("Show detailed scoring information"),
				fields.WithDefault(false),
				fields.WithShortFlag("v"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(glazedSection, tacticianSection, defaultSection))

	cmdDef := cmds.NewCommandDefinition(
		"search",
		cmds.WithShort("Search for applicable tactics"),
		cmds.WithSchema(s),
	)

	return &SearchCommand{CommandDefinition: cmdDef}, nil
}

var _ cmds.GlazeCommand = &SearchCommand{}

func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	tSettings := &sections.TacticianSettings{}
	if err := values.DecodeSectionInto(vals, sections.TacticianSlug, tSettings); err != nil {
		return errors.Wrap(err, "decode tactician settings")
	}

	settings := &SearchSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "decode search settings")
	}

	if settings.LLMRerank {
		return errors.New("--llm-rerank not implemented yet")
	}

	st, err := store.Load(ctx, tSettings.Dir)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	keywords := splitWords(settings.Query)
	goalIDs := splitCSV(settings.Goals)
	tagFilters := splitCSV(settings.Tags)

	tactics, err := st.Tactics.SearchTactics(ctx, strings.TrimSpace(settings.Type), tagFilters, keywords)
	if err != nil {
		return err
	}

	allNodes, err := st.Project.GetAllNodes(ctx)
	if err != nil {
		return err
	}

	ranked, err := rankTactics(ctx, st, tactics, allNodes, keywords, goalIDs)
	if err != nil {
		return err
	}

	if settings.Ready {
		tmp := ranked[:0]
		for _, r := range ranked {
			if r.DepStatus.Ready {
				tmp = append(tmp, r)
			}
		}
		ranked = tmp
	}

	limit := settings.Limit
	if limit <= 0 {
		limit = 20
	}
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	for _, r := range ranked {
		row := types.NewRow(
			types.MRP("id", r.Tactic.ID),
			types.MRP("type", r.Tactic.Type),
			types.MRP("output", r.Tactic.Output),
			types.MRP("ready", r.DepStatus.Ready),
			types.MRP("satisfied", strings.Join(r.DepStatus.Satisfied, ",")),
			types.MRP("missing", strings.Join(r.DepStatus.Missing, ",")),
			types.MRP("can_introduce", strings.Join(r.DepStatus.CanIntroduce, ",")),
			types.MRP("tags", strings.Join(r.Tactic.Tags, ",")),
			types.MRP("description", r.Tactic.Description),
		)
		if settings.Verbose {
			row.Set("score_total", r.Scores.Total)
			row.Set("score_critical_path", r.Scores.CriticalPath)
			row.Set("score_keyword", r.Scores.Keyword)
			row.Set("score_goal", r.Scores.Goal)
		}
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

type depStatus struct {
	Ready        bool
	Satisfied    []string
	Missing      []string
	CanIntroduce []string
}

type scores struct {
	Total        int
	CriticalPath int
	Keyword      int
	Goal         int
}

type rankedTactic struct {
	Tactic    *db.Tactic
	DepStatus depStatus
	Scores    scores
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

func computeTacticDependencyStatus(tactic *db.Tactic, allNodes []*db.Node) depStatus {
	nodeOutputs := map[string]bool{}
	completeOutputs := map[string]bool{}
	for _, n := range allNodes {
		nodeOutputs[n.Output] = true
		if n.Status == "complete" {
			completeOutputs[n.Output] = true
		}
	}

	var satisfied, missing, canIntroduce []string

	for _, dep := range tactic.Match {
		if completeOutputs[dep] {
			satisfied = append(satisfied, dep)
		} else {
			missing = append(missing, dep)
		}
	}

	for _, dep := range tactic.Premises {
		// Skip if also in match
		inMatch := false
		for _, m := range tactic.Match {
			if m == dep {
				inMatch = true
				break
			}
		}
		if inMatch {
			continue
		}

		if completeOutputs[dep] {
			satisfied = append(satisfied, dep)
		} else if nodeOutputs[dep] {
			missing = append(missing, dep)
		} else {
			canIntroduce = append(canIntroduce, dep)
		}
	}

	allMatchSatisfied := true
	for _, dep := range tactic.Match {
		if !completeOutputs[dep] {
			allMatchSatisfied = false
			break
		}
	}

	return depStatus{
		Ready:        allMatchSatisfied,
		Satisfied:    satisfied,
		Missing:      missing,
		CanIntroduce: canIntroduce,
	}
}

func computeCriticalPathScore(ctx context.Context, st *store.State, tactic *db.Tactic, allNodes []*db.Node) (int, error) {
	score := 0
	for _, n := range allNodes {
		if n.Status != "pending" {
			continue
		}
		deps, err := st.Project.GetDependencies(ctx, n.ID)
		if err != nil {
			return 0, err
		}
		var blockers []*db.Node
		for _, d := range deps {
			if d.Status != "complete" {
				blockers = append(blockers, d)
			}
		}
		if len(blockers) == 0 {
			continue
		}
		blocksByThis := false
		for _, b := range blockers {
			if b.Output == tactic.Output {
				blocksByThis = true
				break
			}
		}
		if !blocksByThis {
			continue
		}
		if len(blockers) == 1 {
			score += 2
		} else {
			score += 1
		}
	}
	return score, nil
}

func computeKeywordScore(tactic *db.Tactic, keywords []string) int {
	if len(keywords) == 0 {
		return 0
	}
	score := 0
	idLower := strings.ToLower(tactic.ID)
	descLower := strings.ToLower(tactic.Description)
	var tagsLower []string
	for _, t := range tactic.Tags {
		tagsLower = append(tagsLower, strings.ToLower(t))
	}
	for _, kw := range keywords {
		kw = strings.ToLower(kw)
		if strings.Contains(idLower, kw) {
			score += 10
		}
		for _, tag := range tagsLower {
			if strings.Contains(tag, kw) {
				score += 5
				break
			}
		}
		if strings.Contains(descLower, kw) {
			score += 2
		}
	}
	return score
}

func computeGoalAlignmentScore(ctx context.Context, st *store.State, tactic *db.Tactic, goalIDs []string) (int, error) {
	if len(goalIDs) == 0 {
		return 0, nil
	}
	score := 0
	for _, goalID := range goalIDs {
		goal, err := st.Project.GetNode(ctx, goalID)
		if err != nil {
			return 0, err
		}
		if goal == nil {
			continue
		}
		if tactic.Output == goal.Output {
			score += 20
		}
		deps, err := st.Project.GetDependencies(ctx, goalID)
		if err != nil {
			return 0, err
		}
		for _, d := range deps {
			if d.Output == tactic.Output {
				score += 10
				break
			}
		}
	}
	return score, nil
}

func rankTactics(
	ctx context.Context,
	st *store.State,
	tactics []*db.Tactic,
	allNodes []*db.Node,
	keywords []string,
	goalIDs []string,
) ([]rankedTactic, error) {
	var ranked []rankedTactic
	for _, t := range tactics {
		deps := computeTacticDependencyStatus(t, allNodes)
		cp, err := computeCriticalPathScore(ctx, st, t, allNodes)
		if err != nil {
			return nil, err
		}
		kw := computeKeywordScore(t, keywords)
		gs, err := computeGoalAlignmentScore(ctx, st, t, goalIDs)
		if err != nil {
			return nil, err
		}

		total := 0
		if deps.Ready {
			total += 1000
		} else {
			total -= 500
		}
		total += cp * 50
		total += kw * 10
		total += gs * 5

		ranked = append(ranked, rankedTactic{
			Tactic:    t,
			DepStatus: deps,
			Scores: scores{
				Total:        total,
				CriticalPath: cp,
				Keyword:      kw,
				Goal:         gs,
			},
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Scores.Total == ranked[j].Scores.Total {
			return ranked[i].Tactic.ID < ranked[j].Tactic.ID
		}
		return ranked[i].Scores.Total > ranked[j].Scores.Total
	})

	return ranked, nil
}
