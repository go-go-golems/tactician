---
Title: Feature Development Playbook (Tactician + docmgr)
Slug: feature-development-playbook
Short: "A practical workflow for developing features using Tactician (tactics/nodes/DAG), docmgr tickets, and the tactician step scripts."
Topics:
- tactician
- playbook
- docmgr
- workflow
- dag
- cli
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Feature Development Playbook (Tactician + docmgr)

## Overview

This playbook describes the project workflow we use to implement features: we model work as a dependency-aware DAG in **Tactician**, keep the “human record” in **docmgr** (ticket + diary + changelog + tasks), and after every graph mutation we generate a timestamped **step report** (markdown + Mermaid) so the project state is reviewable and continuation-friendly.

## Prerequisites

You need the CLI tools available in your environment:

- `tactician` (Tactician CLI)
- `docmgr` (ticket/docs/tasks/changelog)
- `plz-confirm` (optional, for interactive clarifying questions via browser UI)

Useful references:

- Tactician basics:
  - `tactician help how-to-use`
  - `tactician help creating-tactics`
- docmgr basics:
  - `docmgr help how-to-use`
- If you are executing “implementation” nodes:
  - `~/.cursor/commands/diary.md`
  - `~/.cursor/commands/git-commit-instructions.md`

## What lives where (files + conventions)

This workflow relies on a few stable locations.

### Tactician state (the DAG)

Tactician persists project state as YAML:

- `.tactician/project.yaml`
  - nodes + edges
  - node metadata (we use `data.description`, `data.ticket`, `data.related-files`)
- `.tactician/tactics/*.yaml`
  - the tactic library (“templates” that create nodes/edges)
- `.tactician/action-log.yaml`
  - append-only action history

### Step reports (one per tactician command)

After each tactician step, we snapshot the graph and write a report:

- `.tactician/mermaid/project-<timestamp>.mmd`
  - raw Mermaid graph string
- `.tactician/steps/step-<timestamp>.md`
  - a “one page” snapshot containing:
    - the command you ran
    - Mermaid graph
    - next ready tactics
    - current nodes/goals
    - numbered ready nodes with descriptions + linked tickets/files

### docmgr tickets (the “human record”)

docmgr stores the ticket workspace under the configured docs root (from `.ttmp.yaml`).
In this repo, confirm it with:

```bash
docmgr status --summary-only
```

Within a ticket, we expect:

- `index.md`: short overview + links
- `tasks.md`: actionable checklist (managed via `docmgr task ...`)
- `changelog.md`: dated history (managed via `docmgr changelog ...`)
- `reference/01-diary.md`: step-by-step diary (created once, updated continuously)

## The “tactician step” wrapper (always use this)

We don’t run `tactician` directly for normal work. Use:

```bash
./tactician/scripts/tactician-step.sh <tactician-subcommand> [args...]
```

This wrapper:

- runs the tactician command (defaulting to `--output yaml` for commands that support it)
- snapshots Mermaid to `.tactician/mermaid/project-<timestamp>.mmd`
- writes a markdown step report to `.tactician/steps/step-<timestamp>.md`
- prints “what’s next” (ready tactics + ready nodes) in a consistent format

Examples:

```bash
# What’s ready?
./tactician/scripts/tactician-step.sh search --ready --verbose
./tactician/scripts/tactician-step.sh goals

# Apply a tactic (materialize nodes/edges)
./tactician/scripts/tactician-step.sh apply <tactic-id> --yes

# Mark a node complete after you’ve done the work it describes
./tactician/scripts/tactician-step.sh node edit <node-id> --status complete
```

## Core workflow (end-to-end)

This is the repeatable loop we use for feature work.

### 1) Define/confirm the goal(s)

A goal is represented as a node in `.tactician/project.yaml` with `type: goal_definition`.
We keep the goal statement stable and treat “complete” as “the goal statement is agreed upon”, not “work is done”.

We also use **side goals** for parallel threads of work that should not replace the main goal.

### 2) Apply a tactic to create the next set of nodes

Tactics are templates that create nodes and edges. Find what’s ready:

```bash
./tactician/scripts/tactician-step.sh search --ready --verbose
```

Then apply:

```bash
./tactician/scripts/tactician-step.sh apply <tactic-id> --yes
```

### 3) Execute ready nodes (do the real work)

Most of our nodes are `type: prompt`. Their `data.prompt` is the “what to do” instructions.

Find ready nodes:

```bash
./tactician/scripts/tactician-step.sh goals
```

Pick the next ready node and execute it:

- If it's a **clarification node** (e.g., `*__draft_questions_from_analysis` or `*__ask_via_plz_confirm`):
  - For `*__draft_questions_from_analysis`: Draft questions from analysis doc, save to ticket reference
  - For `*__ask_via_plz_confirm`: **Must execute this step** - Ask questions via plz-confirm widgets, capture structured answers (JSON), save answers document
  - Update diary and changelog after each clarification step
- If it's a docmgr node:
  - create/update the ticket, docs, tasks, and changelog
  - relate key files to the relevant doc(s)
- If it's an implementation node:
  - implement code changes
  - commit changes following:
    - `~/.cursor/commands/git-commit-instructions.md`
  - update the ticket diary and changelog following:
    - `~/.cursor/commands/diary.md`

When done, mark the node complete:

```bash
./tactician/scripts/tactician-step.sh node edit <node-id> --status complete
```

### 4) Convert analysis → clarifications → tasks → execution

Our typical feature flow is:

- **analysis**: create a file/symbol map and record it in the docmgr ticket
- **clarify**: ask targeted questions to remove ambiguity (optionally via `plz-confirm`)
  - **Step 4a**: Draft clarifying questions from the analysis document
    - Create a questions document in the ticket (e.g., `reference/03-clarifying-questions.md`)
    - Each question should include: title, why it matters, recommended plz-confirm widget type
  - **Step 4b**: Ask the questions via plz-confirm and capture structured answers
    - Use appropriate widgets: `confirm` for yes/no, `select` for choices, `form` for structured multi-field inputs
    - Capture all answers in a structured format (e.g., JSON) for later translation into tasks
    - Save answers document in the ticket (e.g., `reference/04-clarification-answers.json`)
    - **Important**: This step must be completed before task planning can proceed
- **tasks**: convert analysis + clarifications into a task list in `docmgr tasks.md`
- **execute**: implement tasks one-by-one, keeping a diary and committing incrementally

If using plz-confirm, read:

```bash
plz-confirm help how-to-use
```

**Note**: When executing clarification nodes, you must:
1. First draft the questions (if not already done)
2. Then ask them via plz-confirm widgets
3. Capture the structured answers in a document
4. Update the diary and changelog
5. Mark the clarification node(s) as complete

### 5) Keep nodes informative (metadata)

We rely on node metadata so “what is this?” is obvious when scanning step reports:

- `data.description`: short explanation of the node’s purpose in this project
- `data.ticket`: docmgr ticket id (if the node’s work belongs to a ticket)
- `data.related-files`: list of:
  - `name`: absolute path to relevant file/doc
  - `note`: why it matters

This metadata is stored in `.tactician/project.yaml` and shown by the step reports.

## Troubleshooting

### “tactician node show doesn’t show prompt/description”

That’s expected in the current CLI output surface. The canonical node data lives in `.tactician/project.yaml`.
Use the step report output under `.tactician/steps/` for the “human view”.

### “I need to add edges manually”

Today, there is no first-class `tactician edge ...` command. Edges are created by tactics or by editing `.tactician/project.yaml`.
If you must do manual edge edits, do them in a small diff and immediately run:

```bash
./tactician/scripts/tactician-step.sh goals
```

so the Mermaid + step report captures the new graph state.

## References

- Tactician docs:
  - `tactician help how-to-use`
  - `tactician help creating-tactics`
  - `tactician help smoke-test-playbook`
- docmgr docs:
  - `docmgr help how-to-use`
- Documentation style guide used for this page:
  - `glaze help how-to-write-good-documentation-pages`


