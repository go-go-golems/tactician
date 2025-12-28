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
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:37:54.528374158-05:00
WhatFor: "Add first-class node metadata (title/description + instantiation note) to make DAG nodes, action logs, and Mermaid exports more informative."
WhenToUse: "When implementing or reviewing node metadata UX improvements (CLI output, apply logging, Mermaid rendering)."
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
