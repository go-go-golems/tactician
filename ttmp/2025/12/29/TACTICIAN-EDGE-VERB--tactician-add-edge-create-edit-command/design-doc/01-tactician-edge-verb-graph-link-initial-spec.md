---
Title: Tactician edge verb (graph link) - initial spec
Ticket: TACTICIAN-EDGE-VERB
Status: active
Topics:
    - tactician
    - cli
    - dag
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-29T11:43:40.119678025-05:00
WhatFor: ""
WhenToUse: ""
---

# Tactician edge verb (graph link) - initial spec

## Executive Summary

Add a first-class **tactician CLI verb** to create edges (dependencies) between existing nodes in the project DAG.
This removes the need to manually edit `.tactician/project.yaml` when adjusting the graph, while preserving the
current “YAML is source-of-truth” model.

## Problem Statement

Today, edges are created implicitly by applying tactics (and occasionally by manual edits to `.tactician/project.yaml`).
For iterative planning, we need a safe, explicit way to add/update/remove edges after nodes exist.

Pain points:
- Manual YAML edits are error-prone (formatting mistakes, accidental duplicate edges, hard-to-review diffs).
- No CLI feedback (cycle creation, missing node ids, “did I already add this edge?”).
- Hard to script / automate.

## Proposed Solution

Introduce a new command group:

- `tactician edge add <source-id> <target-id>`: add edge `source -> target` (meaning “target depends on source”)
- `tactician edge delete <source-id> <target-id>`: remove edge if it exists
- `tactician edge list [--from <id>] [--to <id>]`: list edges (optionally filtered)

Behavior:
- Validate both nodes exist.
- Default: prevent cycles (DAG must remain acyclic). If it would introduce a cycle, error out with an explanation.
- Idempotency:
  - `edge add` on an existing edge should be a no-op (or emit a warning row) rather than failing hard.
  - `edge delete` on a missing edge should be a no-op (or warning) rather than failing hard.
- Persist changes back to `.tactician/project.yaml` (same persistence model as other mutating commands).

Flags (initial):
- `--force` (optional): allow adding an edge even if it would introduce a cycle (defaults to false).
- `--yes` (optional): match the apply command style for non-interactive operation if a confirmation would otherwise be shown.
- Standard `--tactician-dir` support.

Output:
- Human default: table output describing what changed.
- Scriptable: `--output json|yaml` via glazed flags (like other commands).

## Design Decisions

- **CLI-first edge editing**: edges are fundamental graph structure; they deserve a dedicated verb rather than piggybacking on `node edit`.
- **DAG safety by default**: most workflows rely on acyclicity; cycles should be opt-in.
- **Idempotent UX**: makes scripting easier and avoids brittle “already exists” failures.

## Alternatives Considered

- **Manual editing of `.tactician/project.yaml`**: too error-prone and non-discoverable.
- **Overloading `tactician node edit`**: node edit currently only controls status; mixing concerns is confusing.
- **A tactic-only workflow**: tactics are great for templated expansions, but manual graph correction is still needed.

## Implementation Plan

1. Add cobra command group `edge` with subcommands: `add`, `delete`, `list`.
2. Implement store operations in the YAML-backed graph store:
   - add/remove edge with deterministic ordering and stable serialization.
3. Implement cycle detection for `edge add`:
   - simplest: DFS from `target` to see if `source` is reachable (adding `source -> target` would create cycle if yes).
4. Add tests:
   - add edge success
   - add edge duplicate is no-op
   - add edge creates cycle errors (unless `--force`)
   - delete edge success/no-op
5. Update help docs (`tactician help how-to-use`) and optionally add a new help topic for edge editing.

## Open Questions

- Should `edge add` also allow specifying by **output artifact** (like tactic `match`), or only by node id?
- Should we have `edge add --before/--after` convenience helpers for ordering inside a tactic-generated subgraph?
- What’s the preferred no-op behavior: silent success vs explicit “noop” rows?

## References

- Current project graph is persisted in `.tactician/project.yaml`
- This ticket was created via `docmgr` (see `docmgr help how-to-use`)
