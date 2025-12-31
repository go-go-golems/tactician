#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./tactician/scripts/tactician-step.sh <tactician-subcommand> [args...]
#
# Runs the tactician command, snapshots the current project graph as Mermaid into:
#   .tactician/mermaid/project-<timestamp>.mmd
# and writes a timestamped markdown step report into:
#   .tactician/steps/step-<timestamp>.md
# It also prints the next available tactics + ready nodes (numbered, with short descriptions).

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <tactician-subcommand> [args...]" >&2
  exit 2
fi

TS="$(date +%Y-%m-%d_%H-%M-%S)"
OUT_DIR=".tactician/mermaid"
OUT_FILE="$OUT_DIR/project-$TS.mmd"
STEPS_DIR=".tactician/steps"
STEP_FILE="$STEPS_DIR/step-$TS.md"
PROJECT_YAML=".tactician/project.yaml"

if [[ ! -d "$OUT_DIR" ]]; then
  echo "Error: missing $OUT_DIR (expected it to exist). Create it once and re-run." >&2
  exit 1
fi
if [[ ! -d "$STEPS_DIR" ]]; then
  echo "Error: missing $STEPS_DIR (expected it to exist). Create it once and re-run." >&2
  exit 1
fi

echo "+ tactician $*"

# Prefer YAML output for commands that support glazed output, unless the caller already passed --output.
ARGS=("$@")
if [[ "${ARGS[0]}" == "goals" || "${ARGS[0]}" == "search" || ( "${ARGS[0]}" == "node" && "${#ARGS[@]}" -ge 2 && "${ARGS[1]}" == "show" ) || "${ARGS[0]}" == "graph" ]]; then
  if ! printf '%s\n' "${ARGS[@]}" | grep -qE '^--output$|^--output='; then
    ARGS+=("--output" "yaml")
  fi
fi

set +e
CMD_OUT="$(tactician "${ARGS[@]}" 2>&1)"
CMD_STATUS=$?
set -e

echo "$CMD_OUT"

echo "+ snapshot mermaid -> $OUT_FILE"
MERMAID="$(tactician graph --mermaid --select mermaid)"
printf "%s\n" "$MERMAID" > "$OUT_FILE"

echo "---"
echo "Next available tactics (ready):"
READY_TACTICS_YAML="$(tactician search --ready --verbose --output yaml 2>&1)"
echo "$READY_TACTICS_YAML"

echo "---"
echo "Next executable nodes (ready):"

# Full goals table as YAML for reference (includes blocked nodes too).
GOALS_YAML="$(tactician goals --output yaml 2>&1)"

# Print a numbered list of READY nodes with a short description pulled from the node's stored prompt
# in .tactician/project.yaml (since `tactician node show` doesn't include prompt text).
READY_NODES_MD="$(python3 - <<'PY'
import json, re, subprocess, sys
from pathlib import Path

raw = subprocess.check_output(["tactician", "goals", "--output", "json"])
rows = json.loads(raw.decode("utf-8"))
ready = [r for r in rows if r.get("status") == "ready"]

if not ready:
    print("(none)")
    raise SystemExit(0)

project = Path(".tactician/project.yaml").read_text(encoding="utf-8", errors="replace").splitlines()

def prompt_preview(node_id: str) -> str:
    # Minimal YAML-ish scan for:
    # - id: <node_id>
    #   ...
    #   data:
    #     prompt: | / prompt: "..."
    #       <indented prompt text...>
    id_pat = re.compile(r"^(\s*)-\s+id:\s*%s\s*$" % re.escape(node_id))
    any_id_pat = re.compile(r"^(\s*)-\s+id:\s*(\S+)\s*$")

    start = None
    start_indent = None
    for i, line in enumerate(project):
        m = id_pat.match(line)
        if m:
            start = i
            start_indent = len(m.group(1))
            break
    if start is None:
        return "(node not found in .tactician/project.yaml)"

    # Find end of this node (next "- id:" at same indent, or "edges:")
    end = len(project)
    for j in range(start + 1, len(project)):
        if project[j].startswith("edges:"):
            end = j
            break
        m2 = any_id_pat.match(project[j])
        if m2 and len(m2.group(1)) == start_indent:
            end = j
            break

    # Find prompt line within node block
    # Prefer data.description if present (higher signal for humans), otherwise fall back to prompt.
    for k in range(start, end):
        m = re.match(r"^\s*description:\s*(.*)\s*$", project[k])
        if m:
            s = m.group(1).strip()
            if not s:
                continue
            if len(s) >= 2 and ((s[0] == s[-1] == '"') or (s[0] == s[-1] == "'")):
                s = s[1:-1]
            return (s[:160] + "…") if len(s) > 160 else s

    prompt_line_idx = None
    prompt_indent = None
    prompt_inline = None
    for k in range(start, end):
        line = project[k]
        m = re.match(r"^(\s*)prompt:\s*(.*)\s*$", line)
        if m:
            prompt_line_idx = k
            prompt_indent = len(m.group(1))
            rhs = m.group(2)
            if rhs.startswith("|"):
                prompt_inline = None
            else:
                prompt_inline = rhs.strip()
            break
    if prompt_line_idx is None:
        return "(no prompt found)"

    if prompt_inline:
        # remove surrounding quotes if present (best-effort)
        s = prompt_inline
        if len(s) >= 2 and ((s[0] == s[-1] == '"') or (s[0] == s[-1] == "'")):
            s = s[1:-1]
        s = s.replace("\\n", " ").strip()
        return (s[:160] + "…") if len(s) > 160 else s

    # block scalar: look for first non-empty line indented more than prompt_indent
    for k in range(prompt_line_idx + 1, end):
        line = project[k]
        if not line.strip():
            continue
        if len(line) <= prompt_indent:
            break
        s = line.strip()
        return (s[:160] + "…") if len(s) > 160 else s

    return "(empty prompt)"

def extract_ticket_and_related_files(node_id: str):
    id_pat = re.compile(r"^(\s*)-\s+id:\s*%s\s*$" % re.escape(node_id))
    any_id_pat = re.compile(r"^(\s*)-\s+id:\s*(\S+)\s*$")
    start = None
    start_indent = None
    for i, line in enumerate(project):
        m = id_pat.match(line)
        if m:
            start = i
            start_indent = len(m.group(1))
            break
    if start is None:
        return None, []
    end = len(project)
    for j in range(start + 1, len(project)):
        if project[j].startswith("edges:"):
            end = j
            break
        m2 = any_id_pat.match(project[j])
        if m2 and len(m2.group(1)) == start_indent:
            end = j
            break

    ticket = None
    related = []

    # ticket: <value>
    for k in range(start, end):
        m = re.match(r"^\s*ticket:\s*(.*)\s*$", project[k])
        if m:
            v = m.group(1).strip()
            if v:
                ticket = v.strip('"').strip("'")
            break

    # related-files:
    rel_idx = None
    rel_indent = None
    for k in range(start, end):
        m = re.match(r"^(\s*)related-files:\s*$", project[k])
        if m:
            rel_idx = k
            rel_indent = len(m.group(1))
            break
    if rel_idx is None:
        return ticket, related

    cur = {}
    for k in range(rel_idx + 1, end):
        line = project[k]
        if not line.strip():
            continue
        if len(line) <= rel_indent:
            break
        m_name = re.match(r"^\s*-\s+name:\s*(.*)\s*$", line)
        if m_name:
            if cur:
                related.append(cur)
                cur = {}
            cur["name"] = m_name.group(1).strip().strip('"').strip("'")
            continue
        m_note = re.match(r"^\s*note:\s*(.*)\s*$", line)
        if m_note and cur is not None:
            cur["note"] = m_note.group(1).strip().strip('"').strip("'")
            continue
    if cur:
        related.append(cur)
    return ticket, related

for i, r in enumerate(ready, 1):
    node_id = r.get("id", "")
    output = r.get("output", "")
    parent = r.get("parent_tactic", "")
    deps = r.get("dependencies", "")
    desc = prompt_preview(node_id)
    ticket, related = extract_ticket_and_related_files(node_id)
    print(f"{i}. `{node_id}`")
    print(f"   - output: `{output}`")
    if parent:
        print(f"   - parent_tactic: `{parent}`")
    if deps:
        print(f"   - depends_on: `{deps}`")
    if ticket:
        print(f"   - ticket: `{ticket}`")
    print(f"   - description: {desc}")
    if related:
        print("   - related-files:")
        for rf in related:
            name = rf.get("name", "").strip()
            note = rf.get("note", "").strip()
            if note:
                print(f"     - `{name}` — {note}")
            else:
                print(f"     - `{name}`")
PY
)"

echo "$READY_NODES_MD"

echo "---"
echo "+ wrote step report -> $STEP_FILE"

cat > "$STEP_FILE" <<EOF
# Tactician step report

- Timestamp: \`$TS\`
- Command: \`tactician $*\`
- Mermaid snapshot: \`$OUT_FILE\`

## Command output

\`\`\`text
$CMD_OUT
\`\`\`

## Project graph (Mermaid)

\`\`\`mermaid
$MERMAID
\`\`\`

## Next available tactics (ready)

\`\`\`yaml
$READY_TACTICS_YAML
\`\`\`

## Current goals (all nodes)

\`\`\`yaml
$GOALS_YAML
\`\`\`

## Next executable nodes (ready, numbered)

$READY_NODES_MD
EOF

if [[ $CMD_STATUS -ne 0 ]]; then
  echo "Command exit code: $CMD_STATUS" >&2
  exit "$CMD_STATUS"
fi


