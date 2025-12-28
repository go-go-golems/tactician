package defaults

import _ "embed"

// DefaultTacticsYAML is the built-in tactics library used by `tactician init`.
//
// The persistent format is one-file-per-tactic under `.tactician/tactics/`, but we
// embed the library as a single YAML list so the binary is self-contained.
//
//go:embed default-tactics.yaml
var DefaultTacticsYAML []byte
