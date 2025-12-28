package store

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	projectFileName   = "project.yaml"
	actionLogFileName = "action-log.yaml"
	tacticsDirName    = "tactics"
)

func projectFilePath(tacticianDir string) string {
	return filepath.Join(tacticianDir, projectFileName)
}

func actionLogFilePath(tacticianDir string) string {
	return filepath.Join(tacticianDir, actionLogFileName)
}

func tacticsDirPath(tacticianDir string) string {
	return filepath.Join(tacticianDir, tacticsDirName)
}

func readProjectFile(tacticianDir string) (*diskProjectFile, error) {
	p := projectFilePath(tacticianDir)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errors.Wrap(err, "read project.yaml")
	}

	var f diskProjectFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, errors.Wrap(err, "unmarshal project.yaml")
	}

	// Normalize defaults.
	for i := range f.Nodes {
		if f.Nodes[i].Status == "" {
			f.Nodes[i].Status = "pending"
		}
	}

	return &f, nil
}

func writeProjectFile(tacticianDir string, f *diskProjectFile) error {
	if f == nil {
		return errors.New("nil project file")
	}

	// Deterministic ordering for stable diffs.
	sort.Slice(f.Nodes, func(i, j int) bool { return f.Nodes[i].ID < f.Nodes[j].ID })
	sort.Slice(f.Edges, func(i, j int) bool {
		if f.Edges[i].Source == f.Edges[j].Source {
			return f.Edges[i].Target < f.Edges[j].Target
		}
		return f.Edges[i].Source < f.Edges[j].Source
	})

	b, err := yaml.Marshal(f)
	if err != nil {
		return errors.Wrap(err, "marshal project.yaml")
	}
	p := projectFilePath(tacticianDir)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return errors.Wrap(err, "write project.yaml")
	}
	return nil
}

func readActionLogFile(tacticianDir string) (diskActionLogFile, error) {
	p := actionLogFilePath(tacticianDir)
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return diskActionLogFile{}, nil
		}
		return nil, errors.Wrap(err, "stat action-log.yaml")
	}

	b, err := os.ReadFile(p)
	if err != nil {
		return nil, errors.Wrap(err, "read action-log.yaml")
	}
	if len(b) == 0 {
		return diskActionLogFile{}, nil
	}

	var f diskActionLogFile
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, errors.Wrap(err, "unmarshal action-log.yaml")
	}
	return f, nil
}

func writeActionLogFile(tacticianDir string, f diskActionLogFile) error {
	// Deterministic ordering: newest first (matches most CLI displays).
	sort.Slice(f, func(i, j int) bool {
		return f[i].Timestamp.After(f[j].Timestamp)
	})

	b, err := yaml.Marshal(f)
	if err != nil {
		return errors.Wrap(err, "marshal action-log.yaml")
	}
	p := actionLogFilePath(tacticianDir)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return errors.Wrap(err, "write action-log.yaml")
	}
	return nil
}

// Ensure initial on-disk structure exists.
func ensureTacticianDir(tacticianDir string) error {
	if err := os.MkdirAll(tacticianDir, 0o755); err != nil {
		return errors.Wrap(err, "mkdir tactician dir")
	}
	if err := os.MkdirAll(tacticsDirPath(tacticianDir), 0o755); err != nil {
		return errors.Wrap(err, "mkdir tactics dir")
	}

	// If files don't exist yet, create minimal defaults.
	if _, err := os.Stat(projectFilePath(tacticianDir)); os.IsNotExist(err) {
		now := time.Now().UTC()
		_ = now
		f := &diskProjectFile{
			Project: diskProjectMeta{Name: "untitled", RootGoal: ""},
			Nodes:   []diskNode{},
			Edges:   []diskEdge{},
		}
		if err := writeProjectFile(tacticianDir, f); err != nil {
			return err
		}
	}
	if _, err := os.Stat(actionLogFilePath(tacticianDir)); os.IsNotExist(err) {
		if err := writeActionLogFile(tacticianDir, diskActionLogFile{}); err != nil {
			return err
		}
	}
	return nil
}
