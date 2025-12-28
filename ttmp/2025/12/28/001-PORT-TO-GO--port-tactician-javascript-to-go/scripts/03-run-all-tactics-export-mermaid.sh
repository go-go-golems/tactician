#!/usr/bin/env bash
set -euo pipefail

# Ticket script: full-tactics smoke run + Mermaid export (Go port).
#
# Purpose:
# - Create a fresh project (temp dir)
# - Seed default tactics via `init`
# - Attempt to apply the full set of tactics (one by one, with --force to keep going)
# - Export the resulting project graph/goals as Mermaid into a markdown report for review
#
# This is inspired by the JS reference scripts under:
#   tactician/js-version/tactician/smoke-tests/ (especially visual-walkthrough.sh + test-all.sh)
#
# Usage (run from anywhere):
#   ./scripts/03-run-all-tactics-export-mermaid.sh
#
# Optional:
#   KEEP_WORK=0 ./scripts/03-run-all-tactics-export-mermaid.sh   # deletes temp dir on success
#
# Output:
# - Writes report to: ticket archive/full-tactics-mermaid-<timestamp>.md
# - Prints WORK dir path (kept by default) so you can inspect `.tactician/`

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
REPORT="$ARCHIVE_DIR/full-tactics-mermaid-$ts.md"

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

echo "== init project =="
cd "$WORK"
"$BIN" init

echo "== enumerate tactics =="
mapfile -t tactic_files < <(ls -1 .tactician/tactics/*.yaml 2>/dev/null | sort || true)
if [[ "${#tactic_files[@]}" -eq 0 ]]; then
  echo "ERROR: no tactics found under .tactician/tactics/ after init" >&2
  exit 3
fi

declare -a tactic_ids=()
for f in "${tactic_files[@]}"; do
  id="$(awk -F': ' '/^id:/{print $2; exit}' "$f" | tr -d '\r' | tr -d '"' | tr -d "'")"
  if [[ -n "$id" ]]; then
    tactic_ids+=("$id")
  fi
done

echo "tactics_total=${#tactic_ids[@]}"

APPLIED=0
SKIPPED_EXISTS=0
FAILED=0
declare -a FAILED_IDS=()

echo "== apply all tactics (best-effort) =="
for id in "${tactic_ids[@]}"; do
  # Best-effort: keep going even if a particular apply fails.
  # `--force` is used so we can run through the full set without satisfying every dependency first.
  out="$("$BIN" apply "$id" --yes --force 2>&1)" || ec=$? || true
  ec="${ec:-0}"

  if [[ "$ec" -eq 0 ]]; then
    APPLIED=$((APPLIED + 1))
  elif echo "$out" | grep -q "node already exists:"; then
    SKIPPED_EXISTS=$((SKIPPED_EXISTS + 1))
  else
    FAILED=$((FAILED + 1))
    FAILED_IDS+=("$id")
  fi

  unset ec
done

PROJECT_YAML=".tactician/project.yaml"
ACTION_YAML=".tactician/action-log.yaml"

NODE_COUNT="$(grep -c '^    - id:' "$PROJECT_YAML" 2>/dev/null || true)"
EDGE_COUNT="$(grep -c '^    - source:' "$PROJECT_YAML" 2>/dev/null || true)"
ACTION_COUNT="$(grep -c '^-' "$ACTION_YAML" 2>/dev/null || true)"

echo "== export mermaid (graph/goals) =="
graph_json="$("$BIN" graph --mermaid --output json 2>/dev/null || true)"
goals_json="$("$BIN" goals --mermaid --output json 2>/dev/null || true)"

graph_mermaid="$(python3 - <<'PY'
import json, sys
raw = sys.stdin.read().strip()
if not raw:
    print("")
    raise SystemExit(0)
try:
    arr = json.loads(raw)
    if isinstance(arr, list) and arr:
        print(arr[0].get("mermaid","") or "")
    else:
        print("")
except Exception:
    # If output is not JSON for some reason, just print the raw string.
    print(raw)
PY
<<<"$graph_json")"

goals_mermaid="$(python3 - <<'PY'
import json, sys
raw = sys.stdin.read().strip()
if not raw:
    print("")
    raise SystemExit(0)
try:
    arr = json.loads(raw)
    if isinstance(arr, list) and arr:
        print(arr[0].get("mermaid","") or "")
    else:
        print("")
except Exception:
    print(raw)
PY
<<<"$goals_json")"

TACTICIAN_GIT_SHA="$(cd "$TACTICIAN_MODULE" && git rev-parse HEAD 2>/dev/null || echo "unknown")"

{
  echo "# Full tactics smoke test (Go) — Mermaid export"
  echo
  echo "- **Timestamp (UTC)**: $ts"
  echo "- **Repo**: $TACTICIAN_MODULE"
  echo "- **Commit**: $TACTICIAN_GIT_SHA"
  echo "- **Work dir**: \`$WORK\`"
  echo "- **Tactics total**: ${#tactic_ids[@]}"
  echo "- **Applied (exit 0)**: $APPLIED"
  echo "- **Skipped (node already exists)**: $SKIPPED_EXISTS"
  echo "- **Failed (other errors)**: $FAILED"
  echo "- **Node count**: $NODE_COUNT"
  echo "- **Edge count**: $EDGE_COUNT"
  echo "- **Action count**: $ACTION_COUNT"
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
  echo "## Failed tactics (if any)"
  echo
  if [[ "${#FAILED_IDS[@]}" -eq 0 ]]; then
    echo "N/A"
  else
    for id in "${FAILED_IDS[@]}"; do
      echo "- \`$id\`"
    done
  fi
  echo
  echo "## Notes"
  echo
  echo "- This run uses \`apply --yes --force\` to maximize coverage; failures are still counted and listed."
  echo "- Tactics that would create already-existing nodes are counted as skips."
  echo "- Inspect the full state under \`$WORK/.tactician/\`."
} >"$REPORT"

echo "Wrote report: $REPORT"


