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

Tactician is a CLI for turning “things we need to do” into a dependency graph (a DAG) and applying reusable “tactics” to grow that graph. In the Go port, **all persistent state lives as YAML files in `.tactician/`**, and every command loads that YAML into an **in-memory SQLite database** to run queries and updates efficiently.

## Key concepts

Tactician is built around a few small concepts that compose cleanly.

- **Nodes**: “work items” or “artifacts” represented as entries in the project DAG (id/type/output/status + metadata).
- **Edges**: dependencies between nodes (source → target means “target depends on source”).
- **Tactics**: reusable templates that introduce one or more nodes (and sometimes edges) into the project graph.
- **Action log**: an audit log of what changed and why.

## Storage model (YAML source-of-truth)

The Go port persists everything to `.tactician/` so the full project state is reviewable and mergeable.

### `.tactician/` layout

```
.tactician/
  project.yaml
  action-log.yaml
  tactics/
    <tactic-id>.yaml
```

### What is persisted where

- **Project graph**: `.tactician/project.yaml` (nodes + edges + meta).
- **Action log**: `.tactician/action-log.yaml`.
- **Tactics**: `.tactician/tactics/*.yaml` (one file per tactic).

## Runtime model (in-memory SQLite only)

Every command follows the same lifecycle: **load YAML → create in-memory SQLite → run → maybe save YAML**.

- **Load**:
  - read `.tactician/project.yaml`, `.tactician/action-log.yaml`, `.tactician/tactics/*.yaml`
  - import into in-memory sqlite tables
- **Execute**:
  - read-only commands query sqlite and emit output
  - mutating commands update sqlite tables and log actions
- **Save** (mutating commands only):
  - export sqlite tables back to YAML files

Important: there is **no disk SQLite database**; sqlite is a transient query engine.

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

This section gives copy-paste workflows that match how the system is designed to be used.

### Start a new project

```bash
go run ./cmd/tactician init
go run ./cmd/tactician search --ready
go run ./cmd/tactician apply gather_requirements --yes
go run ./cmd/tactician goals
```

### Keep the DAG moving

```bash
go run ./cmd/tactician goals
go run ./cmd/tactician node edit <node-id> --status complete
go run ./cmd/tactician search --ready
go run ./cmd/tactician apply <tactic-id> --yes
```

## Where to continue development

This section highlights the “next developer” continuation points.

- **Mermaid output**: currently minimal strings; define a stable output contract (plain text vs structured row) and add styling classes.
- **LLM reranking**: `--llm-rerank` is currently rejected with a clear error; implement when needed.
- **Canonicalization**: YAML writers are deterministic for lists; consider adding a canonical field ordering and stable formatting rules for maps if needed.


