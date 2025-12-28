# Tasks

## TODO

- [x] Setup Go module structure + Cobra command tree skeleton (compiling)
- [x] Create initial shared DB-path schema section (**obsolete after YAML source-of-truth pivot**)
- [x] Implement initial ProjectDB SQLite wrapper (will be refactored to be *in-memory* runtime DB)
- [x] Create analysis doc for YAML source-of-truth + in-memory SQLite import/export model

### Pivot: YAML as source of truth + in-memory SQLite runtime (NO disk DB)

- [ ] Replace DB-path flags with `.tactician/` storage settings:
  - [ ] New section: `--tactician-dir` (default: `.tactician`)
  - [ ] Remove/stop using `--project-db-path` and `--tactics-db-path`
- [ ] Define YAML on-disk layout (initially):
  - [ ] `.tactician/project.yaml`
  - [ ] `.tactician/action-log.yaml` (append-only or regenerated)
  - [ ] `.tactician/tactics/<tactic-id>.yaml` (one file per tactic)
- [ ] Implement `pkg/store` (or `pkg/state`) orchestrator:
  - [ ] Load: read YAML from `.tactician/`, create in-memory sqlite, init schema, import YAML → sqlite
  - [ ] Save (mutating commands only): export sqlite → YAML (stable/canonical output)
- [ ] Refactor DB wrappers to accept a provided `*sql.DB` (in-memory) rather than owning `sql.Open` by file path.
- [ ] Implement TacticsDB wrapper + one-file-per-tactic import/export.

### Command implementations (using store)

- [ ] `init`: create `.tactician/` YAML structure, import default tactics, write initial YAML (no sqlite on disk)
- [ ] `node show`: read-only; load state → query → glazed output (batch node IDs)
- [ ] `node add`: mutating; load state → insert node → log action → save YAML
- [ ] `node edit`: mutating; load state → update statuses (batch) → log action(s) → save YAML
- [ ] `node delete`: mutating; load state → enforce blocks unless --force → delete → log → save YAML
- [ ] `graph`: read-only; load state → query edges/nodes → output (table + optional mermaid)
- [ ] `goals`: read-only; load state → compute pending/blocked → output (table + optional mermaid)
- [ ] `history`: read-only; load state → query action log → output (with summary option)
- [ ] `search`: read-only; load state → rank/filter tactics → output
- [ ] `apply`: mutating; load state → dependency check → create nodes/edges → log → save YAML

### Helpers + tests

- [ ] Implement helper functions: `computeNodeStatus()`, `getPendingNodes()`, `parseRelativeTime()`, `rankTactics()`, `checkDependencies()` matching JavaScript logic
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
