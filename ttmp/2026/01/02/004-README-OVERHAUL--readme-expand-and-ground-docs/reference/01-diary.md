---
Title: Diary
Ticket: 004-README-OVERHAUL
Status: active
Topics:
    - docs
    - readme
    - tactician
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Expanded project README grounded in real CLI behavior
    - Path: pkg/doc/playbooks/smoke-test.md
      Note: Smoke test procedure used to validate README commands
    - Path: pkg/doc/topics/creating-tactics.md
      Note: Tactics schema + semantics (match vs premises)
    - Path: pkg/doc/topics/how-to-use.md
      Note: Canonical end-to-end usage + storage/runtime model
    - Path: scripts/tactician-step.sh
      Note: Repo wrapper script documented in README
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-02T09:59:18.943277826-05:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Create a grounded, expansive `README.md` for tactician: what it is, how it works internally (YAML + in-memory SQLite), how to install/run it, and how to use it in real workflows (including the `scripts/tactician-step.sh` wrapper).

---

## Step 1: Read docmgr + diary workflows; set up the documentation loop

This step established the workflow constraints for the task: we need to create the README via a docmgr ticket, and keep a diary with frequent, specific logging (commands, failures, and decisions). The goal is to make the README reflect real, runnable behavior rather than aspirational claims.

**Commit (code):** N/A

### What I did
- Read `~/.cursor/commands/docmgr.md` and `~/.cursor/commands/diary.md` to confirm ticket + diary expectations and command forms.
- Confirmed the repo already uses docmgr with `ttmp/` as the docs root.

### Why
- The request explicitly requires docmgr ticketing and a diary process.
- Getting the workflow right early prevents “docs drift” and missing context later.

### What worked
- `docmgr` is installed and recognizes this repo’s `ttmp/` root.

### What didn't work
- N/A

### What I learned
- The strongest guardrail is “capture exact commands + errors” and keep the diary step-based (not a wall of notes).

### What was tricky to build
- N/A (setup-only).

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- Keep diary steps short but frequent; record outputs that will be copied into the README (help text, smoke test behavior).

### Code review instructions
- N/A (docs-only step).

### Technical details
- Commands read:
  - `sed -n '1,260p' ~/.cursor/commands/docmgr.md`
  - `sed -n '1,260p' ~/.cursor/commands/diary.md`

---

## Step 2: Inspect the repo to understand “what tactician is” today

This step mapped what already exists in the codebase so the README can describe the actual CLI surface, internal model, and recommended workflows. The key discovery is that tactician already includes built-in docs in `pkg/doc/`, and the CLI advertises a specific storage/runtime model (YAML source-of-truth + in-memory SQLite).

**Commit (code):** N/A

### What I did
- Read the current root `README.md` (it’s a placeholder template).
- Located the command entrypoint (`cmd/tactician/main.go`) and enumerated command groups.
- Read built-in documentation pages under `pkg/doc/` (how-to, tactics schema, smoke test playbook, feature development playbook).
- Located the `scripts/tactician-step.sh` wrapper that generates step reports and Mermaid snapshots.
- Read Go module/tooling config (`go.mod`, `Makefile`, `.goreleaser.yaml`) to ground install/run instructions.
- Queried existing docmgr tickets to reuse prior “port” context where helpful (especially YAML + in-memory SQLite rationale and smoke test guidance).

### Why
- The README should be the front door: it needs to match real commands, flags, files on disk, and workflows used in the repo.

### What worked
- The repo already has a clear conceptual model and playbooks; the README can link to them and summarize them.

### What didn't work
- N/A

### What I learned
- There are two “docs systems” here: (1) built-in CLI help pages under `pkg/doc/` and (2) repo-local docmgr tickets under `ttmp/`. The README should explain both and how they fit together.

### What was tricky to build
- N/A (inspection-only).

### What warrants a second pair of eyes
- Confirm the README should recommend `scripts/tactician-step.sh` as the default workflow (the existing playbook says “always use this”).

### What should be done in the future
- After running the smoke test, update the README with any “gotchas” discovered in practice (directories that must exist, etc.).

### Code review instructions
- Start by reading `pkg/doc/topics/how-to-use.md` and `pkg/doc/playbooks/smoke-test.md`.

### Technical details
- Key files referenced:
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/README.md`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/cmd/tactician/main.go`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/pkg/doc/topics/how-to-use.md`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/pkg/doc/topics/creating-tactics.md`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/pkg/doc/playbooks/smoke-test.md`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/pkg/doc/playbooks/feature-development-with-tactician-and-docmgr.md`
  - `/home/manuel/code/wesen/corporate-headquarters/tactician/scripts/tactician-step.sh`

---

## Step 3: Create the ticket workspace and seed this diary

This step created the required docmgr ticket that will “own” the README rewrite, and converted the generated reference template into a real step-based implementation diary per `~/.cursor/commands/diary.md`.

**Commit (code):** N/A

### What I did
- Created a new ticket:
  - `docmgr ticket create-ticket --ticket 004-README-OVERHAUL --title "README: expand and ground docs" --topics docs,readme,tactician,cli`
- Created the diary doc:
  - `docmgr doc add --ticket 004-README-OVERHAUL --doc-type reference --title "Diary"`
- Replaced the generated “reference template” body with an actual diary structure (this document).

### Why
- The request requires that the README is created “using a new ticket” and that work is logged as a diary.

### What worked
- docmgr created the workspace under `ttmp/2026/01/02/...` with the standard structure (index/tasks/changelog + docs directories).

### What didn't work
- N/A

### What I learned
- docmgr’s “Diary” doc is created as a `reference` doc-type by default; the important part is the *content format* (step-based) rather than doc-type semantics.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- Relate `README.md` (repo root) and the key source/docs files to this ticket once the README edit begins.

### Code review instructions
- N/A.

### Technical details
- Ticket path created by docmgr:
  - `ttmp/2026/01/02/004-README-OVERHAUL--readme-expand-and-ground-docs/`

---

## Step 4: Build + run tactician; execute a real smoke loop to ground the README

This step validated that tactician actually builds and runs end-to-end on this machine, and captured the “real” CLI contracts that the README must reflect (including one important gotcha: this directory is inside a parent `go.work` workspace, so you may need `GOWORK=off` when running Go tooling directly). I also ran a minimal smoke loop (init → node ops → search → apply → goals/graph/history) in a temporary directory and verified the `scripts/tactician-step.sh` wrapper produces Mermaid + step reports when the expected directories exist.

**Commit (code):** N/A

### What I did
- Attempted to run tests:
  - `go test ./... -count=1`
- Re-ran tests with workspace disabled:
  - `GOWORK=off go test ./... -count=1`
- Built a binary for repeatable smoke testing:
  - `GOWORK=off go build -o /tmp/tactician ./cmd/tactician`
- Ran help to confirm available commands + topics:
  - `/tmp/tactician --help`
  - `/tmp/tactician help`
  - `/tmp/tactician help how-to-use`
- Ran a real smoke loop in `WORK=$(mktemp -d)`:
  - `/tmp/tactician init`
  - `/tmp/tactician node add root README.md --type project_artifact --status pending`
  - `/tmp/tactician node edit root --status complete`
  - `/tmp/tactician search requirements`
  - `/tmp/tactician apply gather_requirements --yes`
  - `/tmp/tactician apply write_technical_spec --yes --force`
  - `/tmp/tactician goals --output yaml`
  - `/tmp/tactician graph --mermaid --select mermaid`
  - `/tmp/tactician history --summary --output yaml`
- Smoked the repo wrapper script (requires `tactician` on PATH and a couple directories):
  - `mkdir -p .tactician/mermaid .tactician/steps`
  - `PATH="/tmp:$PATH" /home/manuel/code/wesen/corporate-headquarters/tactician/scripts/tactician-step.sh goals`

### Why
- The request explicitly asks to build/run the system so the README is grounded in actual behavior.

### What worked
- With `GOWORK=off`, tests and builds succeed.
- The smoke loop produces the expected YAML state under `.tactician/` and the expected behavior:
  - `search --ready` shows dependency-free tactics as `ready: true`
  - applying `write_technical_spec --force` creates an edge `requirements_document --> technical_specification`
  - `history --summary` returns aggregate counters
- `scripts/tactician-step.sh` successfully:
  - runs tactician,
  - snapshots Mermaid to `.tactician/mermaid/`,
  - emits a step report to `.tactician/steps/`,
  - prints “ready tactics” and “ready nodes” in a consistent format.

### What didn't work
- `go test ./...` failed initially due to an enclosing `go.work` workspace:
  - `pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies`
- I also mistakenly tried `tactician goals --limit ...`; `goals` uses Glazed pagination flags (e.g. `--glazed-limit`), not a bespoke `--limit` flag like `search` does.

### What I learned
- The README should include a short troubleshooting note for `go.work` users (`GOWORK=off`).
- Some tactician commands expose pagination through Glazed generic flags (e.g. `--glazed-limit`), while others (like `search`) have their own dedicated flags (e.g. `--limit`).

### What was tricky to build
- Making wrapper usage reliable requires calling out implicit prerequisites (directories + PATH) in docs; otherwise it fails in a confusing way (“missing .tactician/mermaid”).

### What warrants a second pair of eyes
- Confirm what the README should recommend for pagination (`--glazed-limit` vs command-specific flags) to avoid confusing new users.

### What should be done in the future
- Add a tiny “copy/paste smoke script” section to the README (adapted from the built-in smoke playbook) so new contributors can validate quickly.

### Code review instructions
- Validate the smoke loop by following `pkg/doc/playbooks/smoke-test.md` or by building `/tmp/tactician` and running the commands above in a fresh temp directory.

### Technical details
- Observed state layout on disk after `init`:
  - `.tactician/project.yaml`
  - `.tactician/action-log.yaml`
  - `.tactician/tactics/*.yaml`

---

## Step 5: Rewrite the root README.md (grounded, copy/paste workflows)

This step replaced the placeholder root `README.md` with a real, expansive README that matches the current CLI and implementation model. The README is grounded in the smoke run from Step 4 (real commands, flags, and gotchas like `GOWORK=off`), and it intentionally points readers to the embedded docs (`tactician help …`) for deeper dives.

**Commit (code):** N/A

### What I did
- Replaced the repository root `README.md` with:
  - a clear “what is tactician?” description,
  - the YAML + in-memory SQLite architecture explanation,
  - installation / running options (`go run`, `go build`, `go install`),
  - a copy/paste end-to-end session,
  - a CLI overview (including batch node ops),
  - tactics schema examples + gotchas,
  - Glazed output examples,
  - a “repo workflow” section documenting `scripts/tactician-step.sh`,
  - pointers to built-in docs and smoke testing,
  - troubleshooting (notably `go.work` → `GOWORK=off`).
- Related the updated README + key supporting files to this diary document via `docmgr doc relate`.

### Why
- The existing README was a template placeholder and did not reflect the current state of the project.
- A grounded README reduces onboarding time and prevents “docs promise behavior that doesn’t exist”.

### What worked
- The README examples align with the observed CLI help surfaces (flags/commands).
- The README calls out the directory prerequisites for `scripts/tactician-step.sh`, which otherwise fails with a hard error.

### What didn't work
- N/A (docs-only change).

### What I learned
- It’s important to separate:
  - persisted node status (`pending` / `complete`) vs
  - computed “actual status” (`ready` / `blocked`) shown by `goals`.

### What was tricky to build
- Balancing “expansive and verbose” with “still copy/paste-safe”: I focused on concrete workflows and avoided inventing features (e.g. `--llm-rerank` is called out as not implemented).

### What warrants a second pair of eyes
- The install instructions:
  - confirm that `go install github.com/go-go-golems/tactician/cmd/tactician@latest` is the desired path (vs recommending brew packages).
- The semantics of “append-only” for `action-log.yaml`:
  - it behaves append-only from a user perspective, but the file may be rewritten in sorted order on save; confirm wording is acceptable.

### What should be done in the future
- Consider adding a tiny note to the wrapper script or `init` to create `.tactician/mermaid` and `.tactician/steps` automatically if we want the wrapper to be “no surprises”.

### Code review instructions
- Read `README.md` top-to-bottom and sanity-check every command.
- Optionally validate the “quick end-to-end session” section by running it in a fresh temp directory.

---

## Step 6: Validate README claims against the live binary

This step sanity-checked that the README examples and claims match what the current binary actually does (flags, command forms, and known “not implemented” surfaces). The goal is to avoid a README that looks plausible but fails on copy/paste.

**Commit (code):** N/A

### What I did
- Verified `--tactician-dir` works both before and after the subcommand:
  - `/tmp/tactician --tactician-dir .my-state init`
  - `/tmp/tactician init --tactician-dir .my-state`
- Verified `search --llm-rerank` behavior:
  - `/tmp/tactician search --llm-rerank` → `Error: --llm-rerank not implemented yet`
- Confirmed the “go.work gotcha” and the workaround are real:
  - `go test ./...` failed with a go.work workspace error
  - `GOWORK=off go test ./... -count=1` succeeds
- Ran `docmgr doctor --ticket 004-README-OVERHAUL` and added minimal vocabulary entries (topics/docTypes/intent/status) to clear warnings for this ticket.

### What worked
- The README command snippets match real help output for:
  - `init`, `node add/show/edit/delete`, `search`, `apply`, `goals`, `graph`, `history`, and `help`.

### What didn't work
- N/A

### What I learned
- Keeping the README precise sometimes means leaning on “built-in docs” (`tactician help …`) for the long tail of flags rather than duplicating them.

### What was tricky to build
- N/A (verification-only).

### What warrants a second pair of eyes
- Confirm the “install” section recommendations match how this project is distributed in practice (Go install vs brew vs release assets).

### What should be done in the future
- If the CLI flag surface changes, update README snippets and consider adding a tiny CI check that runs `tactician --help` and the smoke loop.

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
