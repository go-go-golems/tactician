package store

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/tactician/pkg/db"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func readTacticsDir(tacticianDir string) ([]*db.Tactic, error) {
	dir := tacticsDirPath(tacticianDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "read tactics dir")
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)

	var tactics []*db.Tactic
	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, errors.Wrap(err, "read tactic file")
		}
		var t db.Tactic
		if err := yaml.Unmarshal(b, &t); err != nil {
			return nil, errors.Wrap(err, "unmarshal tactic file")
		}
		if t.ID == "" {
			return nil, errors.Errorf("tactic file missing id: %s", p)
		}
		tactics = append(tactics, &t)
	}

	return tactics, nil
}

func writeTacticFile(tacticianDir string, tactic *db.Tactic) error {
	if tactic == nil {
		return errors.New("nil tactic")
	}
	if tactic.ID == "" {
		return errors.New("tactic has empty id")
	}

	b, err := yaml.Marshal(tactic)
	if err != nil {
		return errors.Wrap(err, "marshal tactic")
	}

	p := filepath.Join(tacticsDirPath(tacticianDir), tactic.ID+".yaml")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return errors.Wrap(err, "write tactic file")
	}
	return nil
}

func writeTacticsDir(tacticianDir string, tactics []*db.Tactic) error {
	dir := tacticsDirPath(tacticianDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "mkdir tactics dir")
	}

	// Build desired set.
	want := map[string]struct{}{}
	for _, t := range tactics {
		if t == nil || t.ID == "" {
			continue
		}
		want[t.ID] = struct{}{}
	}

	// Remove stale files.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.Wrap(err, "read tactics dir")
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		if _, ok := want[id]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return errors.Wrap(err, "remove stale tactic file")
		}
	}

	// Write current tactics.
	for _, t := range tactics {
		if err := writeTacticFile(tacticianDir, t); err != nil {
			return err
		}
	}

	return nil
}

// SeedTacticsIfMissing writes tactics to `.tactician/tactics/<id>.yaml` only if the file doesn't exist yet.
func SeedTacticsIfMissing(tacticianDir string, tactics []*db.Tactic) error {
	dir := tacticsDirPath(tacticianDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "mkdir tactics dir")
	}

	for _, t := range tactics {
		if t == nil || t.ID == "" {
			continue
		}
		p := filepath.Join(dir, t.ID+".yaml")
		if _, err := os.Stat(p); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return errors.Wrap(err, "stat tactic file")
		}
		if err := writeTacticFile(tacticianDir, t); err != nil {
			return err
		}
	}

	return nil
}
