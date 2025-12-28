---
Title: 'Spec: Node title/description + instantiation details'
Ticket: 002-ADD-TITLE-DESCRIPTION
Status: active
Topics:
    - tactician
    - go
    - cli
    - dag
    - ux
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/apply/apply.go
      Note: Apply note and instantiation metadata
    - Path: pkg/commands/node/add.go
      Note: Node add metadata flags + update behavior
    - Path: pkg/db/project.go
      Note: Node storage schema updates
    - Path: pkg/store/disk_types.go
      Note: YAML disk schema updates
    - Path: ttmp/2025/12/28/001-PORT-TO-GO--port-tactician-javascript-to-go/scripts/04-walkthrough-export-mermaid.sh
      Note: Existing reviewable Mermaid walkthrough script to compare outputs
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:37:59.789307947-05:00
WhatFor: Make DAG nodes more informative by adding title/description + instantiation note fields, and surface them in CLI output, Mermaid graphs, and action logs.
WhenToUse: When implementing node metadata improvements and improving reviewability of Mermaid exports and history logs.
---


## Overview

We want instantiated nodes (especially those created by `apply`) to be **human-readable** and to carry enough context to be useful in:
- Mermaid exports (`graph --mermaid`, `goals --mermaid`)
- `node show`
- `history` / action log entries

Right now nodes only have `id`, `output`, and a few provenance fields. Many nodes end up with `id == output`, and Mermaid currently renders `id` and `output`, which is often redundant and not informative.

This ticket introduces:
- **Node title**: short, human-friendly label
- **Node description**: longer human explanation
- **Node instantiation note** (name TBD): “why/when we created this node”, usually provided at `apply` time

## Goals

- Persist **title + description** for nodes in the YAML source-of-truth and runtime DB.
- Allow `node add` to set (and update) these fields so manual graph building can be descriptive.
- When applying a tactic, populate node title/description from the tactic (and/or derived defaults), and accept an optional **user-provided note** that:
  - enriches the action log entry, and
  - is stored on the created node(s) as an instantiation note.
- Improve Mermaid labeling so graphs become reviewable:
  - show title (and avoid repeating `id == output`)
  - include status

## Non-goals

- Changing tactic schemas to add titles for subtasks (we’ll derive defaults for subtask nodes for now).
- Reworking ranking model or adding LLM reranking.
- Changing the persisted file layout under `.tactician/` beyond adding fields.

## Proposed data model changes

### Node fields

Add three optional strings to nodes:

- **title** (string, optional)
- **description** (string, optional)
- **instantiation_note** (string, optional) — name TBD

Naming in code/YAML should be consistent. Proposed YAML keys:

```yaml
- id: requirements_document
  type: team_activity
  output: requirements_document
  status: pending
  title: Gather requirements
  description: Meet with stakeholders to gather and document project requirements
  instantiation_note: "Kickoff: stakeholder interviews planned for next week"
```

### Where to store it

We want this to be **first-class** data, not only nested under `data`, so:
- SQLite `nodes` table gets new columns
- `db.Node` gets new fields
- YAML disk schema (`store/disk_types.go`) gets new fields

### Defaults

When a node is created by `apply`:
- **title**:
  - For single-output tactic nodes: derive from `tactic-id` (e.g. `write_technical_spec` → `Write technical spec`) OR from `tactic.description` if you prefer “sentence titles”.
  - For subtask nodes: derive from `subtask.id` (e.g. `endpoint_tests` → `Endpoint tests`).
- **description**:
  - For single-output tactic nodes: `tactic.description`
  - For subtask nodes: empty for now (unless we later extend tactic YAML with subtask descriptions).
- **instantiation_note**:
  - From a new `apply` CLI flag (see below). Applied to every node created by that apply invocation.

When a node is created by `node add`:
- **title/description** come from flags (optional)
- **instantiation_note** is empty unless we later add a flag for it

## CLI behavior changes

### `node add`

Add two optional flags:
- `--title <string>`
- `--description <string>`

Change semantics so `node add` can also *update* these strings:
- If a node with the same id already exists:
  - do **not** fail if `--title` and/or `--description` are provided
  - update only the provided fields (leave others untouched)
  - log an action (either `node_updated` with details, or a dedicated `node_described` action — TBD)
- If the node exists and no title/description is provided: keep current behavior (error).

### `node show`

Include these fields in output:
- `title`
- `description`
- `instantiation_note` (if present)

### `apply`

Add one optional flag (name TBD; suggestion: `--note`):
- `--note <string>`: an operator-provided note describing why we’re applying the tactic now.

Behavior:
- Include this note in the `tactic_applied` log details.
- Store it on created nodes as `instantiation_note`.

## Action log changes

Currently `apply` logs:
- action: `tactic_applied`
- details: `Applied tactic: <id>`

Change details to include more context:
- `Applied tactic: <id> — <tactic.description>` (always)
- If note provided: append ` — <note>`

## Mermaid rendering changes

Mermaid should become human-readable by preferring `title`:
- Node label should start with **title** when present, otherwise fall back to `id` / `output`.
- Avoid duplicate lines when `id == output`.
- Include status (ready/blocked/complete).

### Recommended verification output

Use the walkthrough Mermaid report (small graph) to review:
- node labels show titles
- there’s no `id`/`output` duplication
- the apply note shows up in `history` and (optionally) in node show output

## Verification scripts / playbooks

### Primary: walkthrough Mermaid export (small graph)

Use a walkthrough script that produces a *reviewable* graph and exports Mermaid into a markdown report.

- Script (to create in this ticket): `scripts/01-walkthrough-title-description-export-mermaid.sh`
- Output markdown: `archive/walkthrough-title-description-mermaid-<timestamp>.md`

Expected:
- Mermaid nodes include title (not id twice)
- `history` shows `tactic_applied` entries with tactic description + note

### Secondary: stress test (large graph)

Use the “apply all tactics” export from ticket 001 to ensure titles scale:
- `ttmp/.../001-PORT-TO-GO.../scripts/03-run-all-tactics-export-mermaid.sh`

## Relevant files (pointers)

- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/db/types.go`: `db.Node` struct (add fields)
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/db/project.go`: SQLite schema + node CRUD (add columns, scan/insert/update)
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/store/disk_types.go`: YAML node schema (add fields)
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/store/disk_io.go`: YAML import/export mapping
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/node/add.go`: add flags + “update metadata if exists”
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/node/show.go`: display new fields
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/apply/apply.go`: apply note flag + node field population + log details
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/history/history.go`: verify action log rendering
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/graph/graph.go`: Mermaid label building (prefer title)
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/commands/goals/goals.go`: Mermaid label building (prefer title)
- `/home/manuel/workspaces/2025-12-28/port-tactician-go/tactician/pkg/integration/cli_integration_test.go`: extend tests to assert title/description show up and mermaid label includes them

