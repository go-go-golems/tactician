# Tactician (Go)

Decompose software projects into dependency-aware task DAGs using reusable tactics.

Tactician is a CLI that helps you turn “we should build X” into a concrete, queryable graph of work:

- A project is a directed acyclic graph (DAG).
- Each node is a work item / artifact (a document, a feature, a refactor, a test suite…).
- Each edge is a dependency (“B is blocked by A”).
- “Tactics” are reusable templates that can add nodes + edges when they’re applicable.

The Go port is intentionally simple operationally:

- Persistent state is plain YAML on disk under `.tactician/` (git-friendly, diffable, mergeable).
- At runtime, tactician loads YAML into an in-memory SQLite database (fast queries), runs the command, then saves back to YAML (mutating commands).

---

## Contents

- Why tactician
- Core concepts
- Getting started
- A quick end-to-end session
- How state is stored
- Writing tactics
- Output formats (Glazed)
- Repo workflow: step reports + Mermaid snapshots
- Built-in documentation
- Development (build/test/lint)
- Troubleshooting

---

## Why tactician

Many project trackers are optimized for teams and calendars; they’re not optimized for dependency reasoning. Tactician is for the kind of work where you frequently ask:

- “What can I do next that is actually unblocked?”
- “What do I need to create before I can do that?”
- “If I do this now, what does it unlock?”

Instead of a flat todo list, you get a DAG you can query:

- List “ready” nodes vs “blocked” nodes.
- Visualize the dependency graph.
- Search for tactics that are applicable given the current state of the project.
- Apply a tactic to materialize the next slice of work.

---

## Core concepts (short version)

- Nodes: work items / artifacts (id, type, output, status, metadata).
- Edges: dependencies between nodes (source → target means “target depends on source”).
- Tactics: reusable templates that encode:
  - when they’re applicable (dependencies),
  - what they produce (nodes/edges),
  - and how they should rank in search (tags + text).
- Status:
  - `complete`: done
  - `pending`: not done yet
  - `ready` vs `blocked`: computed by `goals` based on dependency completion

---

## Getting started

### Install

Options:

1) Build from source (recommended for contributors)

```bash
git clone https://github.com/go-go-golems/tactician.git
cd tactician

# Go toolchain: this module uses Go 1.24.x (see `go.mod`).

# If you’re inside a parent go.work workspace that doesn’t include this module:
GOWORK=off go test ./... -count=1
GOWORK=off go build -o ./dist/tactician ./cmd/tactician
```

2) Run without installing (development-friendly)

```bash
GOWORK=off go run ./cmd/tactician --help
```

3) Install via Go toolchain

```bash
go install github.com/go-go-golems/tactician/cmd/tactician@latest
```

After that, ensure `tactician` is on your `PATH`.

### Initialize a project

From a project directory where you want to track work:

```bash
tactician init
```

This creates `.tactician/` (if missing), a minimal `project.yaml`, an `action-log.yaml`, and seeds a default tactics library into `.tactician/tactics/`.

### Use a different state directory (`--tactician-dir`)

By default, tactician reads/writes `.tactician/` in your current working directory. You can point it elsewhere:

```bash
tactician --tactician-dir .my-state init
tactician --tactician-dir .my-state goals
```

---

## A quick end-to-end session

This is the shortest loop that demonstrates how tactician is meant to be used.

```bash
# 1) Initialize state
tactician init

# 2) See tactics you could apply immediately (dependency-free tactics)
tactician search --ready

# 3) Apply a tactic (non-interactive; you must pass --yes)
tactician apply gather_requirements --yes

# 4) See what nodes exist and which ones are ready vs blocked
tactician goals

# 5) Mark a node complete once you’ve done the work it represents
tactician node edit requirements_document --status complete

# 6) Search again (new tactics may now be ready)
tactician search --ready

# 7) Visualize the graph
tactician graph
tactician graph --mermaid --select mermaid

# 8) Inspect change history
tactician history
tactician history --summary
```

---

## CLI overview

Tactician’s CLI surface is intentionally small:

- `init`: create `.tactician/` and seed tactics
- `search`: find tactics (with readiness + ranking)
- `apply`: apply one tactic (creates nodes/edges)
- `goals`: list incomplete nodes and show which are `ready` vs `blocked`
- `graph`: print a traversal of the graph (or Mermaid)
- `node`: CRUD for project nodes
- `history`: inspect the action log (or show a summary)

### `node` commands (batch-friendly)

The `node` subcommands support batch operations:

```bash
# Add one node
tactician node add root README.md --type project_artifact --status pending

# Show one or many
tactician node show root other-node-id --output yaml

# Mark one or many complete
tactician node edit root other-node-id --status complete

# Delete one or many (refuses if it blocks others unless --force)
tactician node delete requirements_document
tactician node delete requirements_document --force
```

### `search` tips

```bash
# Keyword search
tactician search requirements

# Only tactics whose required deps are satisfied
tactician search --ready

# Filter by tags / type (as defined in the tactic YAML)
tactician search --tags planning,documentation
tactician search --type document

# Return more/less output
tactician search --verbose
tactician search --limit 5
```

### `graph` Mermaid output

The `--mermaid` flag makes `graph` emit a Mermaid diagram field. If you want only the Mermaid string, use Glazed’s `--select`:

```bash
tactician graph --mermaid --select mermaid
```

---

## How state is stored (YAML source-of-truth)

Tactician persists everything under `.tactician/`:

```text
.tactician/
  project.yaml          # nodes + edges + project meta
  action-log.yaml       # action history (newest first)
  tactics/
    gather_requirements.yaml
    write_technical_spec.yaml
    ...
```

Key properties:

- Git-friendly: YAML changes are reviewable as normal diffs.
- Merge-friendly: the graph can evolve across branches.
- Deterministic output: nodes and edges are written in stable order to reduce diff churn.

At runtime, every command:

1) Loads YAML from disk
2) Imports into an in-memory SQLite database
3) Runs queries/updates
4) Exports back to YAML (mutating commands only)

This means you never have to manage a database file, migrations, or a DB server; SQLite exists only as a runtime query engine.

---

## Writing tactics

A “tactic” is one YAML file under `.tactician/tactics/<id>.yaml`.

Tactics answer two questions:

1) When is this tactic applicable?
2) What nodes/edges should be created when it’s applied?

### Minimal tactic

```yaml
id: gather_requirements
type: team_activity
output: requirements_document
match: []
tags: [planning, requirements, documentation]
description: Meet with stakeholders to gather and document project requirements
```

### Dependencies: `match` vs `premises`

- `match`: required dependencies. For a tactic to be “ready”, all `match` outputs must already exist and be `complete`.
- `premises`: optional dependencies that tactician may introduce as pending placeholder nodes *only when the output is missing entirely*.

Practical rule:

- Put real prerequisites in `match`.
- Put “nice to have, but we can create it if missing” in `premises`.

### Multi-node tactics via `subtasks`

If a tactic contains `subtasks`, it can create multiple nodes with internal dependencies:

```yaml
id: implement_crud_endpoints
type: llm_coding_strategy
output: api_code
match: [api_specification]
premises: [data_model]
tags: [backend, api, crud]
description: Implement CRUD endpoints with analysis and tests
subtasks:
  - id: crud_endpoints_analysis
    type: analysis
    output: crud_endpoints_analysis
    depends_on: []
  - id: crud_endpoints_implementation
    type: implementation
    output: crud_endpoints_implementation
    depends_on: [crud_endpoints_analysis]
```

Important gotchas:

- For single-output tactics (no `subtasks`), `output` becomes the node id that is created.
- Subtask ids become node ids too and must be globally unique across the project.

---

## Output formats (Glazed)

Most tactician commands support Glazed output formatting, which lets you emit structured data instead of just a human table.

Examples:

```bash
tactician goals --output yaml
tactician goals --output json
tactician search --ready --output yaml
tactician graph --mermaid --select mermaid
```

Notes:

- Some commands have their own pagination flags (for example, `search --limit`).
- Other commands use Glazed’s generic pagination flags (for example, `--glazed-limit` and `--glazed-skip`).

---

## Smoke testing

There’s a built-in smoke test playbook in the CLI docs:

```bash
tactician help smoke-test-playbook
```

In this repo, a quick “does everything basically work?” loop is:

```bash
GOWORK=off go build -o /tmp/tactician ./cmd/tactician
WORK="$(mktemp -d)"
cd "$WORK"
/tmp/tactician init
/tmp/tactician apply gather_requirements --yes
/tmp/tactician goals --output yaml
```

---

## Repo workflow: step reports + Mermaid snapshots

This repo includes a helper script to make “what changed in the DAG?” reviewable after every tactician command:

- Mermaid snapshots: `.tactician/mermaid/project-<timestamp>.mmd`
- Step reports: `.tactician/steps/step-<timestamp>.md`

Script: `scripts/tactician-step.sh`

Prerequisites:

- You need a `tactician` binary on your `PATH`.
- You must create the output directories once:

```bash
mkdir -p .tactician/mermaid .tactician/steps
```

Example:

```bash
tactician init
mkdir -p .tactician/mermaid .tactician/steps

./scripts/tactician-step.sh search --ready --verbose
./scripts/tactician-step.sh apply gather_requirements --yes
./scripts/tactician-step.sh goals
```

If you’re using tactician outside this repo, treat this script as a pattern you can copy into your own project (it’s just Bash + calls to the CLI).

---

## Built-in documentation

Tactician ships documentation pages inside the binary. Start here:

```bash
tactician help
```

Useful topics:

- `tactician help how-to-use`
- `tactician help creating-tactics`
- `tactician help smoke-test-playbook`
- `tactician help feature-development-playbook`

There’s also an interactive help TUI:

```bash
tactician help --ui
```

---

## Development

### Build

```bash
GOWORK=off go build ./...
GOWORK=off go build -o ./dist/tactician ./cmd/tactician
```

### Test

```bash
GOWORK=off go test ./... -count=1
```

### Lint

```bash
golangci-lint run -v
```

or:

```bash
make lint
```

### Release tooling

This repo is set up for GoReleaser (`.goreleaser.yaml`) and has a `Makefile` with common targets (`make test`, `make build`, `make goreleaser`).

---

## Troubleshooting

### `go test ./...` fails with a go.work error

If you’re in a bigger workspace with a `go.work` file that doesn’t include this module, Go tooling can fail with:

```text
pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies
```

Fix:

```bash
GOWORK=off go test ./... -count=1
```

### `scripts/tactician-step.sh` says “missing .tactician/mermaid”

Create the directories once:

```bash
mkdir -p .tactician/mermaid .tactician/steps
```

Also ensure `tactician` is on your `PATH` (the script calls `tactician`, not `go run ...`).

### `tactician apply ...` refuses to run

`apply` is non-interactive; pass `--yes`:

```bash
tactician apply gather_requirements --yes
```

### `search --llm-rerank` exists but doesn’t do anything

The flag is defined but not implemented yet.

---

## Repo layout

- `cmd/tactician/`: CLI entrypoint
- `pkg/commands/`: command implementations
- `pkg/store/`: YAML load/save + in-memory runtime orchestration
- `pkg/defaults/`: embedded default tactics library
- `pkg/doc/`: documentation pages embedded into the CLI
- `js-version/`: JavaScript version (reference / parity checks)
- `ttmp/`: docmgr ticket workspaces used for design/diaries

---

## License

MIT (see `LICENSE`).
