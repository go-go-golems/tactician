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
		ID string `yaml:"id"`
	} `yaml:"nodes"`
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

	run("init")
	run("node", "add", "root", "README.md", "--type", "project_artifact", "--status", "pending")
	run("apply", "gather_requirements", "--yes")

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
	for _, n := range p.Nodes {
		seen[n.ID] = true
	}
	if !seen["root"] {
		t.Fatalf("expected node root in project.yaml")
	}
	if !seen["requirements_document"] {
		t.Fatalf("expected node requirements_document in project.yaml (created by apply gather_requirements)")
	}
}


