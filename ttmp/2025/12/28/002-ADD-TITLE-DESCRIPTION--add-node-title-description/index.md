---
Title: Add node title + description
Ticket: 002-ADD-TITLE-DESCRIPTION
Status: active
Topics:
    - tactician
    - go
    - cli
    - dag
    - ux
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/apply/apply.go
      Note: apply note + populate node metadata + richer action log
    - Path: pkg/commands/goals/goals.go
      Note: Mermaid goals labeling
    - Path: pkg/commands/graph/graph.go
      Note: Mermaid node label formatting
    - Path: pkg/commands/history/history.go
      Note: Verify action log rendering
    - Path: pkg/commands/node/add.go
      Note: node add flags + updating metadata
    - Path: pkg/commands/node/show.go
      Note: surface new metadata in output
    - Path: pkg/db/project.go
      Note: Nodes table schema + CRUD needs new columns
    - Path: pkg/db/types.go
      Note: Node struct; will add title/description/instantiation note
    - Path: pkg/integration/cli_integration_test.go
      Note: Extend integration assertions for metadata + mermaid
    - Path: pkg/store/disk_io.go
      Note: YAML import/export mapping
    - Path: pkg/store/disk_types.go
      Note: YAML persisted node schema
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:37:54.528374158-05:00
WhatFor: Add first-class node metadata (title/description + instantiation note) to make DAG nodes, action logs, and Mermaid exports more informative.
WhenToUse: When implementing or reviewing node metadata UX improvements (CLI output, apply logging, Mermaid rendering).
---


# Add node title + description

## Overview

Add human-friendly **title** and **description** fields to nodes (persisted in YAML), plus an optional **instantiation note** captured at `apply` time. The goal is to make Mermaid exports and history logs reviewable as the DAG grows.

Start with the design spec:
- `design/01-spec-node-title-description-instantiation-details.md`

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- tactician
- go
- cli
- dag
- ux

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
