package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type projectYAML struct {
	Nodes []struct {
		ID     string `yaml:"id"`
		Status string `yaml:"status"`
	} `yaml:"nodes"`
	Edges []struct {
		Source string `yaml:"source"`
		Target string `yaml:"target"`
	} `yaml:"edges"`
}

func decodeRowsJSON(t *testing.T, b []byte) []map[string]any {
	t.Helper()

	// Glazed typically emits JSON array for --output json. Be tolerant and also accept
	// newline-delimited JSON objects as a fallback.
	var arr []map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(b), &arr); err == nil {
		return arr
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	for dec.More() {
		var obj map[string]any
		if err := dec.Decode(&obj); err != nil {
			t.Fatalf("failed to decode json output:\n%s\nerr=%v", string(b), err)
		}
		arr = append(arr, obj)
	}
	if len(arr) == 0 {
		t.Fatalf("expected at least one json row, got 0\noutput=\n%s", string(b))
	}
	return arr
}

func TestCLI_InitAddApply(t *testing.T) {
	base := t.TempDir()

	// Build binary once.
	bin := filepath.Join(base, "tactician")
	build := exec.Command("go", "build", "-o", bin, "./cmd/tactician")
	build.Dir = filepath.Join(base, "..") // overwritten below

	// We need to run go build from the tactician module root.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	// This test file lives under tactician/pkg/integration, so module root is ../../..
	build.Dir = filepath.Clean(filepath.Join(wd, "..", ".."))

	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = base
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			t.Fatalf("command failed: %v\nargs=%v\noutput=\n%s", err, args, buf.String())
		}
	}

	runOut := func(args ...string) []byte {
		cmd := exec.Command(bin, args...)
		cmd.Dir = base
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nargs=%v\noutput=\n%s", err, args, string(out))
		}
		return out
	}

	runFail := func(args ...string) string {
		cmd := exec.Command(bin, args...)
		cmd.Dir = base
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected command to fail\nargs=%v\noutput=\n%s", args, buf.String())
		}
		return buf.String()
	}

	run("init")
	run("node", "add", "root", "README.md", "--type", "project_artifact", "--status", "pending")
	run("apply", "gather_requirements", "--yes")
	run("apply", "write_technical_spec", "--yes", "--force")
	run("search", "requirements")
	run("graph")
	run("goals")
	run("history")

	// Validate project.yaml contains expected nodes.
	b, err := os.ReadFile(filepath.Join(base, ".tactician", "project.yaml"))
	if err != nil {
		t.Fatalf("read project.yaml: %v", err)
	}
	var p projectYAML
	if err := yaml.Unmarshal(b, &p); err != nil {
		t.Fatalf("unmarshal project.yaml: %v", err)
	}
	seen := map[string]bool{}
	status := map[string]string{}
	for _, n := range p.Nodes {
		seen[n.ID] = true
		status[n.ID] = n.Status
	}
	if !seen["root"] {
		t.Fatalf("expected node root in project.yaml")
	}
	if !seen["requirements_document"] {
		t.Fatalf("expected node requirements_document in project.yaml (created by apply gather_requirements)")
	}
	if !seen["technical_specification"] {
		t.Fatalf("expected node technical_specification in project.yaml (created by apply write_technical_spec)")
	}

	// Root node should still be present and complete-ness isn't critical for this test,
	// but having a status field ensures YAML schema stability.
	if status["root"] == "" {
		t.Fatalf("expected node root to have a status in project.yaml")
	}

	// Ensure dependency edge exists: requirements_document -> technical_specification.
	edgeSeen := false
	for _, e := range p.Edges {
		if e.Source == "requirements_document" && e.Target == "technical_specification" {
			edgeSeen = true
			break
		}
	}
	if !edgeSeen {
		t.Fatalf("expected dependency edge requirements_document -> technical_specification in project.yaml")
	}

	// Unforced delete should fail when blocking; forced delete should succeed.
	outStr := runFail("node", "delete", "requirements_document")
	if !bytes.Contains([]byte(outStr), []byte("use --force")) {
		t.Fatalf("expected delete failure output to mention --force\noutput=\n%s", outStr)
	}
	run("node", "delete", "requirements_document", "--force")

	// Mermaid output contract: should emit a single JSON row with a `mermaid` field containing "graph TD".
	// (We use JSON output so this test doesn't depend on table formatting.)
	graphJSON := runOut("graph", "--mermaid", "--output", "json")
	rows := decodeRowsJSON(t, graphJSON)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 mermaid row from graph --mermaid, got %d\noutput=\n%s", len(rows), string(graphJSON))
	}
	m, _ := rows[0]["mermaid"].(string)
	if !bytes.Contains([]byte(m), []byte("graph TD")) {
		t.Fatalf("expected graph mermaid output to contain \"graph TD\"\nmermaid=\n%s", m)
	}

	goalsJSON := runOut("goals", "--mermaid", "--output", "json")
	rows2 := decodeRowsJSON(t, goalsJSON)
	if len(rows2) != 1 {
		t.Fatalf("expected exactly 1 mermaid row from goals --mermaid, got %d\noutput=\n%s", len(rows2), string(goalsJSON))
	}
	m2, _ := rows2[0]["mermaid"].(string)
	if !bytes.Contains([]byte(m2), []byte("graph TD")) {
		t.Fatalf("expected goals mermaid output to contain \"graph TD\"\nmermaid=\n%s", m2)
	}
}

func TestCLI_FlagCombinations(t *testing.T) {
	base := t.TempDir()

	// Build binary once.
	bin := filepath.Join(base, "tactician")
	build := exec.Command("go", "build", "-o", bin, "./cmd/tactician")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	build.Dir = filepath.Clean(filepath.Join(wd, "..", ".."))

	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	// Scenario A: `--tactician-dir` override should isolate state into a custom dir.
	dirA := t.TempDir()
	runInDir := func(dir string, args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\ndir=%s\nargs=%v\noutput=\n%s", err, dir, args, string(out))
		}
	}
	runFailInDir := func(dir string, args ...string) string {
		cmd := exec.Command(bin, args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected command to fail\ndir=%s\nargs=%v\noutput=\n%s", dir, args, string(out))
		}
		return string(out)
	}

	runInDir(dirA, "init", "--tactician-dir", "state-a")
	if _, err := os.Stat(filepath.Join(dirA, "state-a", "project.yaml")); err != nil {
		t.Fatalf("expected state-a/project.yaml to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dirA, ".tactician", "project.yaml")); err == nil {
		t.Fatalf("did not expect default .tactician/project.yaml when using --tactician-dir override")
	}

	// Scenario B: short flag `-f` should work for node delete and allow flags after args.
	dirB := t.TempDir()

	runInDir(dirB, "init")
	runInDir(dirB, "node", "add", "root", "README.md", "--type", "project_artifact", "--status", "pending")
	runInDir(dirB, "node", "edit", "root", "--status", "complete")
	runInDir(dirB, "apply", "gather_requirements", "--yes")
	runInDir(dirB, "apply", "write_technical_spec", "--yes", "--force")

	outStr := runFailInDir(dirB, "node", "delete", "requirements_document")
	if !bytes.Contains([]byte(outStr), []byte("use --force")) {
		t.Fatalf("expected delete failure output to mention --force\noutput=\n%s", outStr)
	}

	// Flags after args.
	runInDir(dirB, "node", "delete", "requirements_document", "-f")
}


