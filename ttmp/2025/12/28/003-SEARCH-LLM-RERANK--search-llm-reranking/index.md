---
Title: 'Search: LLM reranking'
Ticket: 003-SEARCH-LLM-RERANK
Status: active
Topics:
    - tactician
    - go
    - cli
    - search
    - llm
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: js-version/tactician/README.md
      Note: JS docs for env vars and feature description
    - Path: js-version/tactician/USER_GUIDE.md
      Note: JS user guide text for --llm-rerank
    - Path: js-version/tactician/smoke-tests/test-all.sh
      Note: JS smoke test includes rerank step
    - Path: js-version/tactician/src/commands/search.js
      Note: JS wiring for --llm-rerank
    - Path: js-version/tactician/src/llm/reranker.js
      Note: JS reranker reference (prompt + reorder + env vars)
    - Path: pkg/commands/search/search.go
      Note: Go search command; flag exists but currently errors out
    - Path: pkg/integration/cli_integration_test.go
      Note: Integration harness to extend (conditional on OPENAI_API_KEY)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:53:46.141423524-05:00
WhatFor: ""
WhenToUse: ""
---




# Search: LLM reranking

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- tactician
- go
- cli
- search
- llm

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
