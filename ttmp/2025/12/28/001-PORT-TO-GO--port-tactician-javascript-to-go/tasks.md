# Tasks

## TODO

- [x] Setup Go module structure + Cobra command tree skeleton (compiling)
- [x] Create initial shared DB-path schema section (**obsolete after YAML source-of-truth pivot**)
- [x] Implement initial ProjectDB SQLite wrapper (will be refactored to be *in-memory* runtime DB)
- [x] Create analysis doc for YAML source-of-truth + in-memory SQLite import/export model

### Pivot: YAML as source of truth + in-memory SQLite runtime (NO disk DB)

- [x] Replace DB-path flags with `.tactician/` storage settings:
  - [x] New section: `--tactician-dir` (default: `.tactician`)
  - [x] Remove/stop using `--project-db-path` and `--tactics-db-path`
- [x] Define YAML on-disk layout (initially):
  - [x] `.tactician/project.yaml`
  - [x] `.tactician/action-log.yaml` (currently regenerated on save, newest-first)
  - [x] `.tactician/tactics/<tactic-id>.yaml` (one file per tactic)
- [x] Implement `pkg/store` orchestrator:
  - [x] Load: read YAML from `.tactician/`, create in-memory sqlite, init schema, import YAML → sqlite
  - [x] Save (mutating commands only): export sqlite → YAML (stable/canonical output)
- [x] Refactor DB wrappers to accept a provided `*sql.DB` (in-memory) rather than owning `sql.Open` by file path (implemented via constructors that accept `*sql.DB`)
- [x] Implement TacticsDB wrapper + one-file-per-tactic import/export.

### Command implementations (using store)

- [x] `init`: create `.tactician/` YAML structure, import default tactics, write initial YAML (no sqlite on disk)
- [x] `node show`: read-only; load state → query → glazed output (batch node IDs)
- [x] `node add`: mutating; load state → insert node → log action → save YAML
- [x] `node edit`: mutating; load state → update statuses (batch) → log action(s) → save YAML
- [x] `node delete`: mutating; load state → enforce blocks unless --force → delete → log → save YAML
- [x] `graph`: read-only; load state → query edges/nodes → output (table + optional mermaid; mermaid currently basic)
- [x] `goals`: read-only; load state → compute pending/blocked → output (table + optional mermaid)
- [x] `history`: read-only; load state → query action log → output (with summary option, relative time)
- [x] `search`: read-only; load state → rank/filter tactics → output (LLM rerank not implemented)
- [x] `apply`: mutating; load state → dependency check → create nodes/edges → log → save YAML (requires --yes; non-interactive)

### Helpers + tests

- [x] Implement helper logic (inline in commands for now): status computation, relative time parsing, ranking, dependency checks
- [ ] Add unit tests:
  - [ ] YAML ↔ sqlite roundtrip idempotence
  - [ ] DB query helpers
  - [ ] command settings decoding
- [ ] Add integration tests:
  - [ ] init → add node → search → apply → graph
  - [ ] Mermaid outputs
  - [ ] flag combinations

### Optional

- [ ] LLM reranking (optional): port reranker + integrate with search
