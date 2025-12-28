---
Title: YAML Source-of-Truth with In-Memory SQLite
Ticket: 001-PORT-TO-GO
Status: active
Topics:
    - port
    - go
    - cli
    - tactician
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/db/project.go:Current ProjectDB implementation (disk-backed today)
    - /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/sections/project.go:Current DB path schema (to be replaced/repurposed)
    - /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/js-version/tactician/src/db/project.js:Reference behavior for ProjectDB
    - /home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/js-version/tactician/src/db/tactics.js:Reference behavior for TacticsDB
ExternalSources: []
Summary: "Design for storing all persistent state as YAML files under .tactician/, importing into an in-memory SQLite database on command start, then exporting changes back to YAML on command exit (no disk DB)."
LastUpdated: 2025-12-28T00:00:00Z
WhatFor: "Switch Tactician persistence model to YAML on disk while still leveraging SQL queries during command execution."
WhenToUse: "When implementing/refactoring the Go port to remove disk-backed SQLite and introduce YAML persistence + in-memory SQLite runtime."
---

# YAML Source-of-Truth with In-Memory SQLite

## Goal

Make **YAML files** in `.tactician/` the **only persistent state** (tactics, project graph/goals, action log, etc.), while still using **SQLite queries** by loading that YAML into an **in-memory SQLite DB** at command start. There is **no disk DB**.

Key outcomes:
- The repo/project directory contains the *entire state* as human-editable YAML.
- Commands execute against a transient in-memory SQLite DB for fast/expressive queries.
- At the end of a mutating command, changes are **exported back to YAML** deterministically.

## Core idea (runtime model)

### On every command start
- Locate `.tactician/` in the current project directory (or via `--tactician-dir`).
- Create a **fresh in-memory SQLite database**.
- Run `InitSchema()` on the in-memory DB.
- Load YAML files from `.tactician/` and insert into the in-memory DB tables.

### During the command
- All reads/writes are SQL operations against the in-memory DB.

### On command exit
- If the command is **read-only**: do nothing.
- If the command **mutated state**:
  - Export the in-memory DB to YAML files in `.tactician/`.
  - Optionally also append to an action log YAML stream (or regenerate it from DB).

## Proposed `.tactician/` layout

The request explicitly wants **one file per tactic**, but the rest can be structured in a way that is stable + merge-friendly.

Recommended layout:

```
.tactician/
  project.yaml                 # meta + graph (nodes + edges) in canonical form
  action-log.yaml              # persistent log (append-only list) OR regenerated
  tactics/
    <tactic-id>.yaml           # one file per tactic (canonical serialization)
```

Notes:
- Keeping project state in a single `project.yaml` makes “graph consistency” (nodes+edges) easier.
- If merge-friendliness becomes important, we can evolve this to `nodes/<id>.yaml` and `edges.yaml`, but start simple.

## YAML schema (disk format)

### `project.yaml`

Keep it close to the existing JS `exportToYAML()` structure (but stable ordering/canonical formatting in Go):

```yaml
project:
  name: my-project
  root_goal: goal-1
nodes:
  node-1:
    type: project_artifact
    output: README.md
    status: pending
    created_at: "2025-12-28T12:00:00Z"
    completed_at: null
    created_by: null
    parent_tactic: null
    introduced_as: null
    dependencies:
      match: [node-0]          # matches edges.source -> edges.target
    blocks: [node-2]           # optional convenience (redundant; derivable)
    data: {}                   # optional freeform map
```

Important: `blocks` is **redundant** (derivable by reversing edges). We can:
- either continue writing it for parity/human readability, or
- omit it from disk and compute it on display.

### `tactics/<tactic-id>.yaml`

Keep it close to JS `TacticsDB.exportToYAML()` per tactic:

```yaml
id: write-unit-tests
type: project_artifact
output: "tests/"
description: "Add unit tests for critical packages"
tags: [testing, quality]
match: ["go.mod"]
premises: ["README.md"]
subtasks:
  - id: add-test-skeleton
    output: "tests/skeleton"
    type: project_artifact
    depends_on: []
    data: {}
data: {}
```

### `action-log.yaml`

Two viable options:

- **Option A (append-only)**: Keep an append-only list. Mutating commands append one entry.
- **Option B (regenerate)**: Treat DB as the authoritative log during runtime and regenerate the full file from DB on export.

Append-only is attractive for “human auditing” and reduces churn.

Example append-only structure:

```yaml
- timestamp: "2025-12-28T12:34:56Z"
  action: node_created
  details: "Created node: node-123"
  node_id: node-123
  tactic_id: null
```

## In-memory SQLite schema (runtime format)

Keep the existing schema (mirrors the JS DB schema) because it’s already designed to support the CLI queries:

- `project(key,value)`
- `nodes(...)`
- `edges(source_node_id,target_node_id, UNIQUE(...))`
- `action_log(timestamp, action, details, node_id, tactic_id)`

For tactics:
- `tactics(id,type,output,description,tags,data)`
- `tactic_dependencies(tactic_id,dependency_type,artifact_type)`
- `tactic_subtasks(tactic_id,subtask_id,output,type,depends_on,data)`

## Implementation approach in Go

### New package: `pkg/store` (or `pkg/state`)

Introduce a single orchestration layer that commands use:

- `LoadState(ctx, tacticianDir) (*State, error)`
  - opens **in-memory** SQLite
  - initializes schema
  - imports YAML into DB
  - exposes DB handles + helpers

- `(*State) Save(ctx) error`
  - exports DB back to YAML files (project.yaml, tactics/*.yaml, action-log.yaml)

`State` contains:
- `Project *db.ProjectDB` (but backed by in-memory `*sql.DB`)
- `Tactics *db.TacticsDB` (same)
- `Dir string` (where YAML lives)
- possibly `Dirty bool` (set by mutating commands)

### Refactor DB wrappers to accept an existing `*sql.DB`

Currently `pkg/db/project.go` opens a SQLite database by path.
For in-memory mode, we want:
- `NewProjectDBFromSQL(db *sql.DB) *ProjectDB` (no open/close responsibility), or
- `ProjectDB.Open(ctx)` becomes internal and we instead expose `NewProjectDB(dbPath)` only for tests.

Same for `TacticsDB`.

### In-memory connection details

Use a pure-Go driver (already in use): `modernc.org/sqlite`.

Recommended DSN patterns:
- `":memory:"` creates a private in-memory DB per connection.
- If you need multiple connections sharing the same memory DB, use:
  - `"file::memory:?cache=shared"`

Given the CLI is single-process and can keep a single `*sql.DB`, either works.

## Command changes (what needs to change)

### New/changed global settings

Replace the current “db path” flags with a project directory flag:
- **Add**: `--tactician-dir` (default: `.tactician`)
- **Remove/Deprecated**:
  - `--project-db-path`
  - `--tactics-db-path`

Reason: there are no disk DBs anymore; the persistent paths are YAML paths inside `.tactician/`.

### Every command: startup wiring

All commands that currently “open DBs” should instead:
- call `state := store.LoadState(ctx, tacticianDir)`
- read from `state.Project` / `state.Tactics`
- if mutating, mark dirty and call `state.Save(ctx)` before returning

### Read-only commands (no save)

- `node show`, `graph`, `goals`, `history`, `search`
  - Load state → query → output → exit
  - No YAML writes (unless we decide to “compact/canonicalize” on every run; not recommended initially)

### Mutating commands (save on success)

- `init`
  - Create `.tactician/` structure and initial YAML files
  - No disk DB creation
  - May still use in-memory DB to validate/normalize before writing YAML

- `node add`, `node edit`, `node delete`
  - Load state → apply SQL mutations → append action log entry → save YAML

- `apply`
  - Load state → create nodes/edges/action log in SQL → save YAML

## Impact on existing Go port (concrete checklist)

### DB layer
- [ ] Split DB wrappers so they can run on a provided `*sql.DB` (in-memory) without owning open/close.
- [ ] Add YAML import/export helpers:
  - `ProjectDB.ImportFromYAMLFiles(...)` and `ExportToYAMLFiles(...)`
  - `TacticsDB.ImportFromDir(...)` and `ExportToDir(...)` (one file per tactic)
- [ ] Decide action-log persistence model (append-only vs regenerate).

### Command schemas
- [ ] Replace `pkg/commands/sections/project.go` with a new section:
  - `tactician-dir` (and possibly `--project-root` resolution rules)

### Command implementations
- [ ] Replace direct DB open-by-path with `store.LoadState`.
- [ ] Ensure every mutating command calls `Save()` only on success.
- [ ] Keep debug logging; add logs around:
  - YAML load
  - number of tactics/nodes/edges loaded
  - save targets written

### Tests
- [ ] Add “roundtrip” tests: YAML → in-memory DB → YAML is stable (idempotent).
- [ ] Add command integration tests using a temp directory with `.tactician/` YAML fixtures.

## Open questions / decisions to settle early

- **Canonicalization**: do we rewrite YAML on every command (even read-only) or only on mutations?
- **Log persistence**: append-only file vs regenerated file.
- **Conflicts/merges**: do we need a more granular disk format for project graph (nodes per file)?
- **Ordering**: ensure deterministic YAML output for stable diffs (sorting maps, stable lists).

## Migration notes (from current disk-DB model)

The current port started implementing disk-backed DB paths (`project-db-path`, `tactics-db-path`) and a disk-backed `ProjectDB`. With this new architecture:
- the schema section for DB paths becomes obsolete,
- the DB wrappers remain useful as *runtime query engines* but must be decoupled from disk.


