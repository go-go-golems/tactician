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

Tactics are the reusable building blocks of Tactician: a tactic encodes “when this is applicable” (dependencies) and “what it introduces” (nodes and edges). In the Go port, each tactic is stored as a **single YAML file** in `.tactician/tactics/<tactic-id>.yaml` and is imported into an **in-memory SQLite** runtime DB on each command start.

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

This section explains how `match` and `premises` are interpreted when searching and applying.

### `match`: required dependencies

`match` is a list of **outputs** that must already exist *and be complete* for the tactic to be “ready”.

- If a `match` dependency output is missing, `apply` fails unless `--force`.
- If a `match` dependency output exists but is not complete, `apply` also fails unless `--force`.

### `premises`: optional dependencies that can be introduced

`premises` are outputs that are “nice to have” but can be created as new pending nodes if missing.

When applying a tactic:
- if a premise output is already complete, it is satisfied
- if it exists but is not complete, it is treated as missing (blocks unless `--force`)
- if it does not exist at all, tactician will create a new pending node with:
  - `introduced_as: premise`
  - `created_by: tactic:<tactic-id>`

## Application semantics (`apply`)

This section describes what `tactician apply <tactic-id>` does to the project graph.

### Nodes created

- **Premise nodes**: one node per missing premise (as described above).
- **If subtasks exist**:
  - creates one node per subtask
  - sets `parent_tactic: <tactic-id>`
- **If no subtasks exist**:
  - creates one node with id = `output`, type = `type`, output = `output`

### Edges created

- **Subtask edges**: for each `subtask.depends_on`, creates an edge `depends_on → subtask.id`.
- **Match edges**: for each satisfied match dependency, creates edges from the node producing that output to each newly created node.

## How search ranks tactics

This section explains the ranking model so authors can write tactics that surface correctly.

Ranking factors:
- **Readiness**: tactics whose match dependencies are satisfied get a large boost.
- **Critical path impact**: tactics that unblock more pending nodes rank higher.
- **Keyword relevance**: matches against id/tags/description.
- **Goal alignment**: boosts tactics whose outputs match a goal output or a goal dependency.

Tips for authors:
- Choose concise, meaningful ids.
- Use tags that you actually filter on (`--tags`).
- Put important words in `description` for keyword searches.

## Common pitfalls

This section lists mistakes that usually cause confusion.

- **Using ids and outputs inconsistently**: in single-output tactics, the output string becomes the node id. Pick outputs that are stable and unique.
- **Premise behavior surprise**: premises only “auto-create” when the output is missing; if a premise exists but is pending, it blocks unless `--force`.
- **Subtask ids must be unique** within the project: `apply` fails if it would create a node that already exists.


