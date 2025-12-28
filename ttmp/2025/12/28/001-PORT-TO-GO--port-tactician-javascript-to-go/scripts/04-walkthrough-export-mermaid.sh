#!/usr/bin/env bash
set -euo pipefail

# Ticket script: “reviewable” walkthrough + Mermaid export (Go port).
#
# Motivation:
# - The full-tactics script instantiates the whole tactics library, which creates a *lot* of nodes.
# - For reviewing Mermaid rendering and overall UX, a small “real project” walkthrough is more useful.
#
# This follows the JS reference idea in:
#   tactician/js-version/tactician/smoke-tests/walkthrough-real-project.sh
#
# Output:
# - Writes report to: ticket archive/walkthrough-mermaid-<timestamp>.md
# - Keeps WORK dir by default so you can inspect `.tactician/`
#
# Usage:
#   ./scripts/04-walkthrough-export-mermaid.sh
#
# Optional:
#   KEEP_WORK=0 ./scripts/04-walkthrough-export-mermaid.sh

ROOT="/home/manuel/workspaces/2025-12-28/port-tactician-go"
TACTICIAN_MODULE="$ROOT/tactician"

TICKET_DIR="$TACTICIAN_MODULE/ttmp/2025/12/28/001-PORT-TO-GO--port-tactician-javascript-to-go"
ARCHIVE_DIR="$TICKET_DIR/archive"

if [[ ! -d "$TACTICIAN_MODULE" ]]; then
  echo "ERROR: expected tactician module at: $TACTICIAN_MODULE" >&2
  exit 2
fi
if [[ ! -d "$ARCHIVE_DIR" ]]; then
  echo "ERROR: expected ticket archive dir at: $ARCHIVE_DIR" >&2
  exit 2
fi

ts="$(date -u +%Y%m%d-%H%M%S)"
WORK="$(mktemp -d)"
BIN="$WORK/tactician"
REPORT="$ARCHIVE_DIR/walkthrough-mermaid-$ts.md"

KEEP_WORK="${KEEP_WORK:-1}"
cleanup() {
  if [[ "$KEEP_WORK" == "0" ]]; then
    rm -rf "$WORK"
  else
    echo "NOTE: keeping WORK directory: $WORK"
  fi
}
trap cleanup EXIT

echo "WORK=$WORK"

echo "== preflight: go test ./... =="
(cd "$TACTICIAN_MODULE" && go test ./... -count=1)

echo "== build: tactician binary into WORK =="
(cd "$TACTICIAN_MODULE" && go build -o "$BIN" ./cmd/tactician)

cd "$WORK"

echo "== init =="
"$BIN" init

echo "== walkthrough: seed minimal project state =="
# Root goal is optional, but it makes the graph feel like a “project”.
"$BIN" node add task_management_saas complete_system --type product --status pending

echo "== apply + complete requirements =="
"$BIN" apply gather_requirements --yes
"$BIN" node edit requirements_document --status complete

echo "== apply + complete technical spec + api spec =="
"$BIN" apply write_technical_spec --yes --force
"$BIN" node edit technical_specification --status complete
"$BIN" apply design_api --yes --force
"$BIN" node edit api_specification --status complete

echo "== apply rich tactic and progress through subtasks =="
"$BIN" apply implement_crud_endpoints --yes --force
"$BIN" node edit endpoints_analysis --status complete
"$BIN" node edit api_code --status complete
"$BIN" node edit endpoint_tests --status complete

NODE_COUNT="$(grep -c '^    - id:' .tactician/project.yaml 2>/dev/null || true)"
EDGE_COUNT="$(grep -c '^    - source:' .tactician/project.yaml 2>/dev/null || true)"

graph_json="$("$BIN" graph --mermaid --output json)"
goals_json="$("$BIN" goals --mermaid --output json)"

graph_mermaid="$(python3 -c 'import json,sys; arr=json.load(sys.stdin); print(arr[0].get("mermaid","") if isinstance(arr,list) and arr else "")' <<<"$graph_json")"
goals_mermaid="$(python3 -c 'import json,sys; arr=json.load(sys.stdin); print(arr[0].get("mermaid","") if isinstance(arr,list) and arr else "")' <<<"$goals_json")"

TACTICIAN_GIT_SHA="$(cd "$TACTICIAN_MODULE" && git rev-parse HEAD 2>/dev/null || echo "unknown")"

{
  echo "# Walkthrough smoke test (Go) — Mermaid export"
  echo
  echo "- **Timestamp (UTC)**: $ts"
  echo "- **Repo**: $TACTICIAN_MODULE"
  echo "- **Commit**: $TACTICIAN_GIT_SHA"
  echo "- **Work dir**: \`$WORK\`"
  echo "- **Node count**: $NODE_COUNT"
  echo "- **Edge count**: $EDGE_COUNT"
  echo
  echo "## Graph (Mermaid)"
  echo
  echo '```mermaid'
  printf '%s\n' "$graph_mermaid"
  echo '```'
  echo
  echo "## Goals (Mermaid)"
  echo
  echo '```mermaid'
  printf '%s\n' "$goals_mermaid"
  echo '```'
  echo
  echo "## Scenario"
  echo
  echo "- init"
  echo "- add root goal"
  echo "- apply+complete: gather_requirements"
  echo "- apply+complete: write_technical_spec"
  echo "- apply+complete: design_api"
  echo "- apply+complete: implement_crud_endpoints (subtasks)"
} >"$REPORT"

echo "Wrote report: $REPORT"


