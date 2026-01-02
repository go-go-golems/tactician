---
Title: 'README: expand and ground docs'
Ticket: 004-README-OVERHAUL
Status: active
Topics:
    - docs
    - readme
    - tactician
    - cli
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Root README rewritten in this ticket
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-02T09:59:14.607994718-05:00
WhatFor: ""
WhenToUse: ""
---


# README: expand and ground docs

## Overview

This ticket rewrites tactician’s root `README.md` so it reflects the *actual* CLI and implementation model:

- what tactician is for (project decomposition into a dependency-aware DAG),
- how it works internally (YAML source-of-truth + in-memory SQLite runtime),
- how to install/run it (go run/build/install),
- how to use it end-to-end (init → search → apply → goals/graph/history),
- how to use the repo’s helper wrapper (`scripts/tactician-step.sh`).

The README content is grounded by building and running the current binary and validating the command snippets.

## Key Links

- Root README (result): `../../../../../README.md`
- Diary (implementation log): `reference/01-diary.md`
- Ticket changelog: `changelog.md`

## Status

Current status: **active**

## Topics

- docs
- readme
- tactician
- cli

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
