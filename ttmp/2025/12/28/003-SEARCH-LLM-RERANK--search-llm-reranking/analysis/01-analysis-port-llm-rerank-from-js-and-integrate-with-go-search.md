---
Title: 'Analysis: Port LLM rerank from JS and integrate with Go search'
Ticket: 003-SEARCH-LLM-RERANK
Status: active
Topics:
    - tactician
    - go
    - cli
    - search
    - llm
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: js-version/tactician/src/llm/reranker.js
      Note: Exact prompt and reorder semantics to port
    - Path: pkg/commands/search/search.go
      Note: Insertion point for rerank step
    - Path: pkg/db/project.go
      Note: Queries for nodes/status used by reranker context
    - Path: pkg/db/tactics.go
      Note: Tactic fields (id/type/output/description/tags/match/subtasks) used in prompt
    - Path: pkg/store/state.go
      Note: Project context source (nodes/status)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-28T15:43:29.023037904-05:00
WhatFor: Document the JS reference implementation of LLM reranking and lay out an implementation path to add it to the Go `search` command with tests and a reproducible validation script.
WhenToUse: When implementing `tactician search --llm-rerank` in Go and wanting a ready-to-execute checklist, code pointers, and behavioral parity notes.
---


## Goal

Implement **optional semantic reranking** for `tactician search` in the Go port:

- keep existing heuristic ranking as the baseline
- when `--llm-rerank` is set, call an LLM to reorder the *top N* candidate tactics
- on any error (missing key, API error, JSON parse error), **fallback** to the original heuristic order and continue

This mirrors the JavaScript reference behavior.

## JS reference behavior (source of truth)

### Where it lives

- Reranker module: `tactician/js-version/tactician/src/llm/reranker.js`
- Search wiring: `tactician/js-version/tactician/src/commands/search.js`
- CLI flag: `tactician/js-version/tactician/src/index.js` (`--llm-rerank`)
- Docs: `tactician/js-version/tactician/README.md`, `tactician/js-version/tactician/USER_GUIDE.md`
- Smoke test: `tactician/js-version/tactician/smoke-tests/test-all.sh` (LLM rerank step)

### Env vars (JS)

From JS README:
- `OPENAI_API_KEY`: required
- `TACTICIAN_LLM_MODEL`: default `gpt-4.1-mini`
- `TACTICIAN_RERANK_LIMIT`: default `10`

### Prompt strategy (JS)

The JS reranker:
- builds **project context** from nodes:
  - counts completed vs pending
  - lists completed outputs (full list)
  - lists up to 10 pending goals (id/output/type)
- builds a prompt that includes the current heuristic-ranked candidates (top N) with:
  - id, type, output, description, tags, match deps, subtask count
- asks the model to return **ONLY** a JSON array of tactic IDs in new order
- parses the JSON array and reorders the top N accordingly
- appends any missing IDs (from top N) and then appends the remaining tactics (beyond limit)
- falls back to heuristic order on error and prints a warning

See:
- `buildProjectContext(projectDB)`
- `buildRerankPrompt(query, tactics, projectContext)`
- `rerank(query, rankedTactics, projectDB, options)`

## Go port current state

### Current behavior

In Go, the flag exists but is currently rejected:
- `tactician/pkg/commands/search/search.go`: returns `errors.New("--llm-rerank not implemented yet")`

### Where to integrate

Add a rerank step after heuristic ranking and filtering, before limiting/output:
- `tactician/pkg/commands/search/search.go`
  - near `rankTactics(...)` / `if settings.Ready { ... }` / `limit := settings.Limit`

## Proposed Go design

### Package layout

Add a small package for reranking, similar to the JS module:

- `tactician/pkg/llm/reranker.go`
  - `type Reranker struct { ... }`
  - `func (r *Reranker) Rerank(ctx context.Context, query string, ranked []rankedTactic, project *db.ProjectDB, opts Options) ([]rankedTactic, error)`

Keep the interface narrow so `search` can call it cleanly.

### Provider choice

JS uses OpenAI chat completions. In Go:
- start with OpenAI-compatible API (or pick an existing Go client)
- keep provider specifics behind a small interface so we can swap later

### Config surface

Match JS env vars for compatibility:
- `OPENAI_API_KEY`
- `TACTICIAN_LLM_MODEL` (default `gpt-4.1-mini`)
- `TACTICIAN_RERANK_LIMIT` (default `10`)

Optionally add CLI overrides later (but start with env parity).

### Failure semantics

Strictly follow JS fallback behavior:
- any rerank error → log warning (or verbose output) → return original heuristic ordering
- do **not** fail the command unless the user explicitly opts into “strict”

## Testing strategy

### Unit tests (fast)

- Prompt formatting test:
  - stable output, contains required sections (query, project counts, candidates)
- Response parsing test:
  - handles valid JSON list
  - handles missing IDs (append remaining)
  - handles unknown IDs (ignore)

Use a fake client that returns a canned response.

### Integration test (CLI-level)

Extend `pkg/integration/cli_integration_test.go`:
- set `OPENAI_API_KEY` to a dummy value but inject a fake server/client (or skip if not available)

If mocking the client is hard, keep integration tests as “skipped unless env set”:
- just like JS smoke test checks `OPENAI_API_KEY`

## Validation playbook / script

Add a ticket script (in this ticket) that:
- builds the tactician binary
- initializes a temp project
- runs `search "database" --llm-rerank --limit 5` with:
  - `OPENAI_API_KEY` set
  - `TACTICIAN_LLM_MODEL` optionally set
- captures output to a markdown file for review

Use the existing walkthrough scripts from ticket 001 as a base:
- `ttmp/.../001-PORT-TO-GO--.../scripts/04-walkthrough-export-mermaid.sh`

## Pointers to relevant Go files

- `tactician/pkg/commands/search/search.go`: flag wiring + insertion point
- `tactician/pkg/store/state.go`: project state load (source of project context)
- `tactician/pkg/db/project.go`: node queries used to construct context
- `tactician/pkg/db/tactics.go`: tactic fields used in prompt
- `tactician/pkg/integration/cli_integration_test.go`: integration harness to extend

