package sections

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

// This is a small “settings decoding” unit test to ensure:
// - the `tactician` section correctly registers `--tactician-dir`,
// - defaults are applied when the flag is missing,
// - and the value decodes into the expected settings struct.
func TestTacticianDir_DefaultAndOverride(t *testing.T) {
	tacticianSection, err := NewTacticianSection()
	if err != nil {
		t.Fatalf("NewTacticianSection: %v", err)
	}

	// Minimal command schema: tactician section + one positional arg, so we also validate
	// that the parser can handle args while decoding the section.
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithArguments(
			fields.New("dummy", fields.TypeString, fields.WithRequired(true)),
		),
	)
	if err != nil {
		t.Fatalf("NewSection(default): %v", err)
	}

	s := schema.NewSchema(schema.WithSections(tacticianSection, defaultSection))
	cmdDef := cmds.NewCommandDefinition("dummy-cmd", cmds.WithSchema(s))

	cobraCmd := cli.NewCobraCommandFromCommandDescription(cmdDef.Description())
	parser, err := cli.NewCobraParserFromLayers(cmdDef.Description().Layers, &cli.CobraParserConfig{AppName: "tactician"})
	if err != nil {
		t.Fatalf("NewCobraParserFromLayers: %v", err)
	}
	if err := parser.AddToCobraCommand(cobraCmd); err != nil {
		t.Fatalf("AddToCobraCommand: %v", err)
	}
	// Ensure Cobra behaves like real execution: parse flags from args.
	cobraCmd.SetArgs([]string{"foo"})
	if err := cobraCmd.ParseFlags([]string{"foo"}); err != nil {
		t.Fatalf("ParseFlags(default): %v", err)
	}

	parsed, err := parser.Parse(cobraCmd, cobraCmd.Flags().Args())
	if err != nil {
		t.Fatalf("Parse(default): %v", err)
	}

	// Decode and verify default.
	tSettings := &TacticianSettings{}
	if err := values.DecodeSectionInto(parsed, TacticianSlug, tSettings); err != nil {
		t.Fatalf("DecodeSectionInto(default): %v", err)
	}
	if tSettings.Dir != ".tactician" {
		t.Fatalf("expected default Dir .tactician, got %q", tSettings.Dir)
	}

	// Override via flag.
	cobraCmd2 := &cobra.Command{Use: "dummy-cmd"}
	if err := parser.AddToCobraCommand(cobraCmd2); err != nil {
		t.Fatalf("AddToCobraCommand(override): %v", err)
	}
	cobraCmd2.SetArgs([]string{"foo", "--tactician-dir", "x/.tactician"})
	if err := cobraCmd2.ParseFlags([]string{"foo", "--tactician-dir", "x/.tactician"}); err != nil {
		t.Fatalf("ParseFlags(override): %v", err)
	}
	parsed2, err := parser.Parse(cobraCmd2, cobraCmd2.Flags().Args())
	if err != nil {
		t.Fatalf("Parse(override): %v", err)
	}

	tSettings2 := &TacticianSettings{}
	if err := values.DecodeSectionInto(parsed2, TacticianSlug, tSettings2); err != nil {
		t.Fatalf("DecodeSectionInto(override): %v", err)
	}
	if tSettings2.Dir != "x/.tactician" {
		t.Fatalf("expected override Dir x/.tactician, got %q", tSettings2.Dir)
	}
}
