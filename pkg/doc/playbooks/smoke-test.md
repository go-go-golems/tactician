---
Title: Smoke Test Playbook (Tactician Go)
Slug: tactician-smoke-test-playbook
Short: A step-by-step smoke test procedure covering init, node ops, search, apply, graph, goals, and history.
Topics:
- tactician
- testing
- playbook
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Smoke Test Playbook (Tactician Go)

## Overview

This playbook is a pragmatic “does everything basically work?” procedure for the Go Tactician CLI. It exercises the full feature surface (YAML persistence + in-memory sqlite runtime) in a fresh directory, and provides concrete expected outcomes so the next developer can validate changes quickly.

## Preconditions

You need a working Go toolchain and access to the `tactician/` module.

```bash
cd tactician && go test ./...
go build -o /tmp/tactician ./cmd/tactician
```

## Test environment setup

This section creates a clean working directory so the test is deterministic.

```bash
WORK="$(mktemp -d)"
cd "$WORK"

# Use the prebuilt binary from the preconditions step.
TACTICIAN="/tmp/tactician"
```

## 1) Initialize a project (`init`)

This step validates `.tactician/` scaffolding and default tactics seeding.

```bash
"$TACTICIAN" init
```

Expected:
- `.tactician/project.yaml` exists
- `.tactician/action-log.yaml` exists
- `.tactician/tactics/` exists and contains many `*.yaml` files (one per tactic)

## 2) Add a root node (`node add`)

This step validates YAML → in-memory sqlite load, mutation, and YAML save.

```bash
"$TACTICIAN" node add root README.md --type project_artifact --status pending
```

Expected:
- `.tactician/project.yaml` includes a node with `id: root`
- `.tactician/action-log.yaml` includes a `node_created` entry

## 3) Show nodes (`node show`, batch)

This step validates batch positional arguments and Glazed structured output.

```bash
"$TACTICIAN" node show root
```

Expected:
- Output includes at least the fields `id`, `type`, `output`, `status`

## 4) Edit node status (`node edit`, batch)

This step validates status updates and completion timestamps.

```bash
"$TACTICIAN" node edit root --status complete
```

Expected:
- `.tactician/project.yaml` shows `root` as `status: complete`
- `.tactician/action-log.yaml` includes a `node_completed` entry

## 5) Search tactics (`search`)

This step validates tactics import (file-per-tactic), keyword ranking, and readiness filters.

```bash
"$TACTICIAN" search requirements
"$TACTICIAN" search --ready
"$TACTICIAN" search --tags planning,requirements
```

Expected:
- Output is a table/rows that includes `id`, `output`, and `ready`
- `--ready` only shows tactics whose `match` deps are satisfied by complete project outputs

## 6) Apply a tactic (`apply`)

This step validates end-to-end: tactic selection → dependency checks → node/edge creation → YAML save.

Use a dependency-free tactic first:

```bash
"$TACTICIAN" apply gather_requirements --yes
```

Expected:
- `.tactician/project.yaml` contains a new node with `id: requirements_document` (single-output tactic behavior)
- `.tactician/action-log.yaml` includes a `tactic_applied` entry

Now apply a tactic with dependencies (should succeed if dependencies are complete; otherwise use `--force` to validate behavior):

```bash
"$TACTICIAN" apply write_technical_spec --yes --force
```

Expected:
- Creates one node per tactic behavior (single output or subtasks)
- Adds edges from satisfied match deps to created nodes

## 7) View goals (`goals`)

This step validates “actual status” computation (ready vs blocked) and pending node listing.

```bash
"$TACTICIAN" goals
"$TACTICIAN" goals --mermaid
```

Expected:
- Lists pending nodes and marks them `ready` or `blocked`
- Mermaid mode outputs a single row containing a Mermaid graph string

## 8) View graph (`graph`)

This step validates node/edge traversal and root selection.

```bash
"$TACTICIAN" graph
"$TACTICIAN" graph --mermaid
```

Expected:
- Non-mermaid output emits a traversal of nodes with a `depth` field
- Mermaid output returns a single row containing a Mermaid graph string

## 9) View history (`history`)

This step validates action log queries and time filters.

```bash
"$TACTICIAN" history
"$TACTICIAN" history --summary
"$TACTICIAN" history --since 1d
```

Expected:
- `history` lists action entries
- `--summary` outputs aggregate counters

## 10) Delete node behavior (`node delete`)

This step validates blocked-node protection.

```bash
# This may fail if the node blocks others; that is expected.
"$TACTICIAN" node delete requirements_document

# If the unforced delete failed due to blocking, force delete should succeed.
"$TACTICIAN" node delete requirements_document --force
```

Expected:
- Without `--force`, deletion fails when the node blocks others
- With `--force`, deletion succeeds and logs `node_deleted`

## Troubleshooting checklist

This section lists common failure modes and the fastest way to narrow them down.

- **Command says project not initialized**:
  - verify `.tactician/project.yaml` exists
  - re-run `init`
- **Search returns zero tactics**:
  - verify `.tactician/tactics/*.yaml` exist
  - re-run `init`
- **Apply fails with missing deps**:
  - that’s expected when `match` outputs are not complete
  - either complete the required nodes or re-run with `--force` to validate the error path


