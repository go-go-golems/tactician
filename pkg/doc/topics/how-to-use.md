---
Title: How to Use Tactician (Go)
Slug: tactician-how-to-use
Short: How the Go Tactician CLI works end-to-end: YAML state, in-memory SQLite, and core workflows.
Topics:
- tactician
- cli
- yaml
- sqlite
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# How to Use Tactician (Go)

## Overview

Tactician helps you decompose software projects into actionable task graphs. Instead of keeping todos in your head or scattered across tools, Tactician lets you build a dependency-aware DAG (directed acyclic graph) where each node represents something you need to create, and edges capture what blocks what. The real power comes from **tactics**—reusable templates that know when they're applicable and what nodes/edges they introduce.

The Go port makes an intentional design choice: **all persistent state lives as YAML files in `.tactician/`**. This means your entire project graph is git-friendly, reviewable in diffs, and mergeable across branches. At runtime, every command loads that YAML into an **in-memory SQLite database** to run queries and updates efficiently, then saves the result back to YAML before exiting.

This architecture gives you the simplicity of flat files (no DB server, no installation) with the power of SQL queries (ranking tactics, computing critical paths, finding blocked nodes).

## Key concepts

Tactician is built around a few small concepts that compose cleanly.

- **Nodes**: "work items" or "artifacts" represented as entries in the project DAG. Each node has an id, type, output, status, and optional metadata (who created it, which tactic introduced it, when it completed).
- **Edges**: dependencies between nodes. An edge `source → target` means "target depends on source" (or equivalently, "source blocks target").
- **Tactics**: reusable templates that know when they're applicable (via `match`/`premises` dependencies) and what nodes/edges they introduce when applied.
- **Action log**: an append-only audit log capturing what changed and why, useful for history and debugging.

## Storage model (YAML source-of-truth)

The Go port persists everything to `.tactician/` as YAML so the full project state is reviewable in diffs and mergeable across branches. This design decision means you can version your project DAG alongside your code, making it easy to track what changed, when, and why. The trade-off is that each command does a small amount of YAML parsing on startup, but for typical projects (hundreds of nodes, dozens of tactics) this overhead is negligible.

### `.tactician/` layout

```
.tactician/
  project.yaml          # nodes + edges + project meta
  action-log.yaml       # append-only history (newest first)
  tactics/
    gather_requirements.yaml
    write_technical_spec.yaml
    design_architecture.yaml
    ...
```

### What is persisted where

- **Project graph**: `.tactician/project.yaml` contains all nodes (with their status, created timestamps, parent tactics), all edges (as a simple list of `source → target` pairs), and project metadata (name, root_goal).
- **Action log**: `.tactician/action-log.yaml` is regenerated on every save, sorted newest-first for easy reading.
- **Tactics**: `.tactician/tactics/*.yaml` is one file per tactic. The default library seeds ~80 tactics covering common software project phases (planning, backend, frontend, testing, devops, documentation).

## Runtime model (in-memory SQLite only)

Every command follows the same lifecycle: **load YAML → create in-memory SQLite → run → maybe save YAML**. This design gives you the expressiveness of SQL (ranking, graph queries, dependency checks) without requiring a running database server or persistent DB files.

```
┌─────────────────────────────────────────────────────────────┐
│ Command Start                                               │
├─────────────────────────────────────────────────────────────┤
│  1. Load YAML from .tactician/                              │
│     • project.yaml → nodes/edges/meta                       │
│     • action-log.yaml → action_log table                    │
│     • tactics/*.yaml → tactics/dependencies/subtasks tables │
│                                                             │
│  2. Import into in-memory SQLite (file::memory:?cache=...)│
│     • nodes, edges, action_log, project (meta) tables      │
│     • tactics, tactic_dependencies, tactic_subtasks tables │
│                                                             │
│  3. Execute command logic                                   │
│     • Read-only commands: query & output                    │
│     • Mutating commands: update tables, log action          │
│                                                             │
│  4. Save YAML (mutating commands only)                      │
│     • Export tables → YAML files                            │
│     • DB is discarded (memory freed)                        │
└─────────────────────────────────────────────────────────────┘
```

**Why in-memory only?** The in-memory approach means there's no stale DB file to get out of sync with YAML. The source-of-truth is always `.tactician/*.yaml`, and SQLite is just a query engine that starts fresh every time. This makes the system more predictable and eliminates an entire class of "DB corruption" or "schema migration" issues.

**Performance note**: For projects with thousands of nodes, the YAML load/save adds ~100ms overhead per command. If this becomes an issue, the store layer can be extended to cache the in-memory DB and only reload when YAML changes (detected via mtime), but for typical usage the simplicity of "always fresh" is more valuable.

## Command overview

This section summarizes each command and its role.

### `init`

`init` creates `.tactician/` and seeds default tactics (one file per tactic).

```bash
# from your project root
go run ./cmd/tactician init
```

### `node`

The `node` command group manages project nodes.

```bash
# create a node
go run ./cmd/tactician node add root README.md --type project_artifact --status pending

# show one or more nodes (batch)
go run ./cmd/tactician node show root other-node

# mark one or more nodes complete
go run ./cmd/tactician node edit root --status complete

# delete nodes (refuses if it blocks others unless --force)
go run ./cmd/tactician node delete root --force
```

### `graph`

`graph` prints the project graph starting from a root (explicit `goal-id`, else `root_goal`, else first root with no incoming edges).

```bash
go run ./cmd/tactician graph
go run ./cmd/tactician graph my-goal-id
go run ./cmd/tactician graph --mermaid
```

### `goals`

`goals` lists pending nodes and their computed “actual status” (`ready` vs `blocked`).

```bash
go run ./cmd/tactician goals
go run ./cmd/tactician goals --mermaid
```

### `history`

`history` lists action log entries and can also show a summary.

```bash
go run ./cmd/tactician history
go run ./cmd/tactician history --limit 50
go run ./cmd/tactician history --since 2d
go run ./cmd/tactician history --summary
```

### `search`

`search` ranks tactics for the current project state using dependency readiness + critical path impact + keyword relevance + goal alignment.

```bash
go run ./cmd/tactician search "requirements"
go run ./cmd/tactician search --tags planning,documentation
go run ./cmd/tactician search --type document
go run ./cmd/tactician search --ready
go run ./cmd/tactician search --verbose
```

Note: `--llm-rerank` is defined but not implemented yet.

### `apply`

`apply` materializes a tactic into nodes/edges and saves the updated YAML state. It is **non-interactive**: you must pass `--yes`.

```bash
go run ./cmd/tactician apply gather_requirements --yes
go run ./cmd/tactician apply write_technical_spec --yes --force
```

## Configuration: `--tactician-dir`

Commands accept `--tactician-dir` to point at a different state directory.

```bash
go run ./cmd/tactician --tactician-dir .tactician init
```

## Recommended workflows

This section gives copy-paste workflows that match how the system is designed to be used. The core loop is: **check goals → complete work → mark complete → search for next tactic → apply**.

### Start a new project

```bash
# Initialize .tactician/ and seed default tactics
go run ./cmd/tactician init

# See what tactics are immediately applicable
go run ./cmd/tactician search --ready

# Apply a root tactic (e.g., planning phase)
go run ./cmd/tactician apply gather_requirements --yes

# View the new nodes
go run ./cmd/tactician goals
```

**Why this works**: The default tactics include several with `match: []` (no dependencies), so after `init` you'll have a handful of "ready" tactics to choose from. Applying one creates nodes and sets you up for subsequent tactics.

### Keep the DAG moving

```bash
# See what's currently pending and what's ready
go run ./cmd/tactician goals

# When you finish a task, mark it complete
go run ./cmd/tactician node edit <node-id> --status complete

# Search for what's newly unblocked
go run ./cmd/tactician search --ready

# Apply the next tactic
go run ./cmd/tactician apply <tactic-id> --yes

# Visualize the updated graph
go run ./cmd/tactician graph
```

**Why this works**: Marking nodes complete unblocks downstream nodes (dependencies are satisfied), which makes previously-blocked tactics "ready". The search ranking prioritizes tactics that unblock the most work (critical path impact).

## Where to continue development

This section highlights the “next developer” continuation points.

- **Mermaid output**: currently minimal strings; define a stable output contract (plain text vs structured row) and add styling classes.
- **LLM reranking**: `--llm-rerank` is currently rejected with a clear error; implement when needed.
- **Canonicalization**: YAML writers are deterministic for lists; consider adding a canonical field ordering and stable formatting rules for maps if needed.


