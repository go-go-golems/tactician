#!/usr/bin/env bash
set -euo pipefail

# 002 ticket verification script (to be run AFTER implementing the ticket):
# - Uses new node metadata fields (title/description/instantiation note)
# - Generates a small walkthrough graph
# - Exports Mermaid into a markdown report for review
#
# This script is intentionally “self-checking”: it will refuse to run until
# the required flags exist in the tactician binary.
#
# Usage:
#   ./scripts/01-walkthrough-title-description-export-mermaid.sh

ROOT="/home/manuel/workspaces/2025-12-28/port-tactician-go"
TACTICIAN_MODULE="$ROOT/tactician"

TICKET_DIR="$TACTICIAN_MODULE/ttmp/2025/12/28/002-ADD-TITLE-DESCRIPTION--add-node-title-description"
ARCHIVE_DIR="$TICKET_DIR/archive"

ts="$(date -u +%Y%m%d-%H%M%S)"
WORK="$(mktemp -d)"
BIN="$WORK/tactician"
REPORT="$ARCHIVE_DIR/walkthrough-title-description-mermaid-$ts.md"

cleanup() {
  echo "NOTE: keeping WORK directory: $WORK"
}
trap cleanup EXIT

echo "WORK=$WORK"

echo "== build binary =="
(cd "$TACTICIAN_MODULE" && go test ./... -count=1 && go build -o "$BIN" ./cmd/tactician)

echo "== check required flags exist =="
if ! "$BIN" node add --help | grep -q -- "--title"; then
  echo "ERROR: node add --title not found. Implement ticket 002 first." >&2
  exit 2
fi
if ! "$BIN" node add --help | grep -q -- "--description"; then
  echo "ERROR: node add --description not found. Implement ticket 002 first." >&2
  exit 2
fi
if ! "$BIN" apply --help | grep -q -- "--note"; then
  echo "ERROR: apply --note not found. Implement ticket 002 first (or adjust this script if flag name differs)." >&2
  exit 2
fi

cd "$WORK"

echo "== init =="
"$BIN" init

echo "== add root goal with metadata =="
"$BIN" node add task_management_saas complete_system --type product --status pending \
  --title "Task management SaaS" \
  --description "Deliver a complete task management system across web + mobile"

echo "== apply tactic with note (instantiation context) =="
"$BIN" apply gather_requirements --yes --note "Kickoff: stakeholder interviews scheduled; capture requirements first"

echo "== export mermaid =="
graph_json="$("$BIN" graph --mermaid --output json)"
goals_json="$("$BIN" goals --mermaid --output json)"

graph_mermaid="$(python3 -c 'import json,sys; arr=json.load(sys.stdin); print(arr[0].get("mermaid","") if isinstance(arr,list) and arr else "")' <<<"$graph_json")"
goals_mermaid="$(python3 -c 'import json,sys; arr=json.load(sys.stdin); print(arr[0].get("mermaid","") if isinstance(arr,list) and arr else "")' <<<"$goals_json")"

echo "== write report =="
{
  echo "# Ticket 002 walkthrough — verify node title/description + instantiation note"
  echo
  echo "- **Timestamp (UTC)**: $ts"
  echo "- **Work dir**: \`$WORK\`"
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
  echo "## Spot checks"
  echo
  echo "- Run: \`$BIN node show task_management_saas\` and confirm title/description fields are present."
  echo "- Run: \`$BIN history\` and confirm the `tactic_applied` entry includes tactic description + note."
} >"$REPORT"

echo "Wrote report: $REPORT"


