---
Title: Creating Tactics (Schema + Semantics)
Slug: tactician-creating-tactics
Short: How to write tactics as one YAML file per tactic, and how tactician interprets match/premises/subtasks.
Topics:
- tactician
- tactics
- yaml
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Creating Tactics (Schema + Semantics)

## Overview

Tactics are the reusable building blocks of Tactician. Think of them as "project recipes": each tactic encodes **when it's applicable** (what artifacts must exist first) and **what it produces** (new nodes and their dependencies). The key insight is that many software project tasks follow patterns—"write technical spec" always depends on having requirements, "write unit tests" depends on having code to test—and tactics let you encode those patterns once and reuse them across projects.

In the Go port, each tactic is stored as a **single YAML file** in `.tactician/tactics/<tactic-id>.yaml`. When you run any command, tactician loads all tactic files into an **in-memory SQLite** database, which makes searching and ranking fast (SQL queries can check dependency status, compute critical path impact, and filter by tags in a single query). This design keeps the persistent state simple (just YAML files) while making the runtime queries powerful.

## File location and naming

Each tactic is stored as a YAML file whose filename matches the tactic id.

```
.tactician/
  tactics/
    gather_requirements.yaml
    write_technical_spec.yaml
```

Rules:
- The YAML must contain an `id`.
- The filename should match `id` to keep the library understandable and avoid collisions.

## Tactic schema (YAML)

This section defines the stable schema used by the Go port today.

### Minimal example (single output tactic)

```yaml
id: gather_requirements
type: team_activity
output: requirements_document
match: []
tags: [planning, requirements]
description: Meet with stakeholders to gather requirements
```

### Full example (dependencies + premises + subtasks)

```yaml
id: implement_crud_endpoints
type: llm_coding_strategy
output: api_code
match: [api_specification]
premises: [data_model]
tags: [backend, api, crud]
description: Implement CRUD endpoints with analysis and tests
subtasks:
  - id: endpoints_analysis
    type: analysis
    output: endpoints_analysis
    depends_on: []
    data:
      estimated_tokens: 2500
  - id: api_code
    type: implementation
    output: api_code
    depends_on: [endpoints_analysis]
data:
  model: claude-sonnet-4-5-20250929
```

Fields:
- **id (string, required)**: stable identifier of the tactic.
- **type (string, required)**: type of the main node created by the tactic.
- **output (string, required)**: output artifact produced by the tactic (also used as node id for single-output tactics).
- **description (string, optional)**: human explanation; used by search.
- **tags ([]string, optional)**: used by search filtering and ranking.
- **match ([]string, optional)**: required dependencies (see below).
- **premises ([]string, optional)**: optional dependencies that can be introduced (see below).
- **subtasks ([]subtask, optional)**: creates multiple nodes and internal edges (see below).
- **data (map, optional)**: freeform metadata; stored and preserved.

Subtask fields:
- **id (string, required)**: node id created for the subtask.
- **type (string, required)**: node type created for the subtask.
- **output (string, required)**: output artifact produced by the subtask.
- **depends_on ([]string, optional)**: subtask node ids this subtask depends on (creates edges).
- **data (map, optional)**: freeform metadata stored on the created node.

## Dependency semantics

Tactics declare two kinds of dependencies: `match` (required) and `premises` (optional, can be introduced). This distinction lets you express "I need X to exist and be complete" vs "I'd like Y to exist, but I can introduce it as a placeholder if not".

### `match`: required dependencies

`match` is a list of **output artifacts** that must already exist *and be complete* for the tactic to be "ready".

When searching for tactics (`tactician search`):
- Tactics with all `match` dependencies satisfied get a large ranking boost.
- Tactics with unsatisfied `match` dependencies still appear but rank lower.

When applying a tactic (`tactician apply <id>`):
- If a `match` dependency output is missing or incomplete, `apply` fails with a clear error unless `--force`.
- If all `match` dependencies are satisfied, edges are created from those nodes to the newly created nodes.

**Why this matters**: `match` is your contract. It says "don't apply this tactic until these artifacts exist and are complete." This prevents applying tactics out of order.

### `premises`: optional dependencies that can be introduced

`premises` are outputs that are "nice to have" but tactician will **introduce as new pending nodes** if they're missing entirely.

When applying a tactic:
- If a premise output is already complete, it is satisfied (edges created, just like `match`).
- If it exists but is not complete, it is treated as missing (blocks unless `--force`).
- If it does not exist at all, tactician creates a new pending node with:
  - `id: <premise-output>`
  - `type: document` (default)
  - `introduced_as: premise`
  - `created_by: tactic:<tactic-id>`

**Why this matters**: `premises` let you model "soft dependencies"—things that ideally exist but can be stubbed out. For example, a tactic for "implement authentication" might have `premises: [security_guidelines_doc]`. If that doc doesn't exist, tactician creates a placeholder so you remember to write it.

**Practical note**: Premises only auto-create when the output is **missing entirely**. If a premise node already exists but is pending, it blocks (just like `match`). This prevents accidentally introducing duplicate nodes.

## Application semantics (`apply`)

When you run `tactician apply <tactic-id> --yes`, tactician performs a sequence of node and edge creations. Understanding this flow helps you predict what your DAG will look like after applying a tactic.

### Nodes created

Applying a tactic can create several kinds of nodes:

1. **Premise nodes** (if premises are missing entirely):
   - One node per missing premise
   - `id: <premise-output>`, `type: document`, `introduced_as: premise`

2. **Subtask nodes** (if tactic defines `subtasks`):
   - One node per subtask entry
   - `id: subtask.id`, `type: subtask.type`, `output: subtask.output`
   - `parent_tactic: <tactic-id>` (links back to the tactic that created it)

3. **Single-output node** (if no subtasks):
   - One node with `id: output`, `type: type`, `output: output`

All nodes start with `status: pending` and get a `created_at` timestamp. The `created_by` field records `tactic:<tactic-id>` so you can trace where nodes came from.

### Edges created

Edges capture dependencies and ordering constraints:

1. **Subtask internal dependencies**:
   - For each `subtask.depends_on` entry, creates edge `depends_on → subtask.id`
   - This encodes the order in which subtasks should be completed

2. **External dependencies (from satisfied `match` entries)**:
   - For each satisfied `match` dependency, finds the node producing that output
   - Creates edges from that node to each newly created node (or all subtask nodes)
   - This ensures the new work "depends on" the completed prerequisite

### Visualizing tactic application

Here's what happens when you apply a tactic with `match: [X]` and two subtasks:

```
Before:
  [X: complete]

After applying tactic "foo":
  [X: complete] ──┬──> [foo-subtask-1: pending]
                  └──> [foo-subtask-2: pending]
  
  (foo-subtask-2 might also depend on foo-subtask-1 via subtask.depends_on)
```

The edges from `X` ensure that if you were to delete `X` or mark it incomplete, the new subtasks would become blocked.

## How search ranks tactics

The `search` command ranks tactics using a multi-factor scoring model that balances "what's ready now" with "what would have the biggest impact". Understanding the ranking helps you write tactics that surface at the right time.

### Ranking factors (in order of weight)

1. **Readiness** (highest weight): Tactics whose `match` dependencies are all satisfied get a +1000 boost; tactics with missing dependencies get a -500 penalty. This ensures that "ready to apply now" tactics always bubble to the top.

2. **Critical path impact**: The search command counts how many pending nodes would be unblocked if this tactic's output completed. Tactics that unblock multiple nodes rank higher because they're on the critical path.

3. **Keyword relevance**: If you provide a search query, tactician matches keywords against the tactic's `id` (highest score), `tags` (medium score), and `description` (lowest score).

4. **Goal alignment**: If you pass `--goals node1,node2`, tactics whose outputs match those nodes (or their dependencies) get a boost. This helps you find "what will help me reach this specific goal".

### Tips for tactic authors

- **Choose meaningful ids**: Use lowercase-with-underscores that describe the work (e.g., `write_technical_spec`, not `task_002`).
- **Tag consistently**: Use tags you'll actually filter on. Common patterns: `planning`, `backend`, `frontend`, `testing`, `documentation`, `devops`.
- **Write descriptions for humans**: The description is indexed by keyword search and displayed in search results, so make it clear and concise.
- **Model dependencies accurately**: If a tactic genuinely requires X to be complete, put it in `match`. If it's optional, use `premises`. Mismatched dependencies cause tactics to appear ready when they're not (or vice versa).

## Common pitfalls

These are the mistakes that trip up new tactic authors (and sometimes experienced ones).

### Output strings become node IDs in single-output tactics

For tactics without `subtasks`, the `output` field becomes the node id when applied. This means:
- Pick outputs that are stable, unique, and meaningful (`requirements_document`, not `doc1`).
- Avoid punctuation or special characters (they work, but can make CLI usage awkward).
- Think about whether two different tactics might produce the same output (if so, coordinate on the output name).

### Premise auto-creation only happens when the output is missing entirely

This is a common surprise: if you have a tactic with `premises: [security_doc]` and apply it:
- If `security_doc` doesn't exist → tactician creates a new pending node ✓
- If `security_doc` exists but is pending → tactician treats it as "missing" and blocks unless `--force` ✗

**Why this design**: Auto-creating a node that already exists would either duplicate it or silently do nothing, both of which are confusing. The current behavior is conservative: "if a premise exists but isn't complete, that's a real blocker."

**Workaround**: If you want "apply even if premises are pending," use `--force`.

### Subtask IDs must be globally unique

When a tactic has `subtasks`, each `subtask.id` becomes a node id in the project graph. This means:
- Subtask ids must be unique **across the entire project**, not just within the tactic.
- If you apply a tactic that would create a subtask node that already exists, `apply` fails immediately.

**Common mistake**: Using generic subtask ids like `analysis` or `implementation`. Prefer scoped ids like `<tactic-id>_analysis` or use descriptive names like `crud_endpoints_analysis`.

### Example: match vs premises in practice

Consider a tactic for "implement authentication":

```yaml
id: implement_authentication
type: software_implementation
output: authentication_system
match: [api_specification]      # MUST be complete
premises: [security_guidelines_doc]  # nice to have
```

**Scenario 1**: You have a complete `api_specification` but no `security_guidelines_doc`.
- ✓ Tactic is "ready" (all `match` satisfied)
- `apply` creates: `security_guidelines_doc` (pending, premise) + `authentication_system` (pending, main node)
- Edges: `api_specification → authentication_system`, `security_guidelines_doc → authentication_system`

**Scenario 2**: You have a complete `api_specification` and a pending `security_guidelines_doc`.
- ✗ Tactic is "ready" but `apply` fails: "premise exists but is not complete"
- Solution: either complete `security_guidelines_doc` first, or use `--force` to apply anyway.


