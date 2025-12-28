package integration

import (
	"bytes"
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
}


