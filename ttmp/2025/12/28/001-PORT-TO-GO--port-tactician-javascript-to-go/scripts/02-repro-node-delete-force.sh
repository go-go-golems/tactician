#!/usr/bin/env bash
set -euo pipefail

# Ticket script: repro + trace for `node delete --force` parsing/behavior.
# - Builds a tactician binary (compiling step).
# - Runs a minimal scenario that creates a blocking dependency edge.
# - Captures exit codes and (optionally) `--print-parsed-parameters` output to a log file.
#
# Usage:
#   ./scripts/02-repro-node-delete-force.sh
#
# Optional:
#   KEEP_WORK=0 ./scripts/02-repro-node-delete-force.sh   # deletes temp dir on success
#
# Outputs:
#   - Prints a short summary to stdout
#   - Writes a full log to: various/repro-node-delete-force-<timestamp>.log

ROOT="/home/manuel/workspaces/2025-12-28/port-tactician-go"
TACTICIAN_MODULE="$ROOT/tactician"
TACTICIAN_PKG="./cmd/tactician"

TICKET_DIR="$TACTICIAN_MODULE/ttmp/2025/12/28/001-PORT-TO-GO--port-tactician-javascript-to-go"
SCRIPTS_DIR="$TICKET_DIR/scripts"
VARIOUS_DIR="$TICKET_DIR/various"

if [[ ! -d "$TACTICIAN_MODULE" ]]; then
  echo "ERROR: expected tactician module at: $TACTICIAN_MODULE" >&2
  exit 2
fi
if [[ ! -d "$SCRIPTS_DIR" ]]; then
  echo "ERROR: expected ticket scripts dir at: $SCRIPTS_DIR" >&2
  exit 2
fi
if [[ ! -d "$VARIOUS_DIR" ]]; then
  echo "ERROR: expected ticket various dir at: $VARIOUS_DIR" >&2
  exit 2
fi

ts="$(date -u +%Y%m%d-%H%M%S)"
LOG="$VARIOUS_DIR/repro-node-delete-force-$ts.log"

WORK="$(mktemp -d)"
BIN="$WORK/tactician"

KEEP_WORK="${KEEP_WORK:-1}"
cleanup() {
  if [[ "$KEEP_WORK" == "0" ]]; then
    rm -rf "$WORK"
  else
    echo "NOTE: keeping WORK directory: $WORK"
  fi
}
trap cleanup EXIT

{
  echo "timestamp_utc=$ts"
  echo "WORK=$WORK"
  echo

  echo "== preflight: go test ./... =="
  (cd "$TACTICIAN_MODULE" && go test ./... -count=1)
  echo

  echo "== build: tactician binary into WORK =="
  (cd "$TACTICIAN_MODULE" && go build -o "$BIN" "$TACTICIAN_PKG")
  echo

  echo "== setup: init + root node complete =="
  cd "$WORK"
  set +e
  "$BIN" init; echo "init_exit=$?"
  "$BIN" node add root README.md --type project_artifact --status pending; echo "add_exit=$?"
  "$BIN" node edit root --status complete; echo "edit_exit=$?"
  "$BIN" apply gather_requirements --yes; echo "apply1_exit=$?"
  "$BIN" apply write_technical_spec --yes --force; echo "apply2_exit=$?"
  set -e
  echo

  echo "== sanity: goals + graph (should show edge requirements_document --> technical_specification) =="
  set +e
  "$BIN" goals; echo "goals_exit=$?"
  "$BIN" graph --mermaid; echo "graph_mermaid_exit=$?"
  set -e
  echo

  echo "== debug: parsed parameters (flags AFTER args) =="
  set +e
  "$BIN" node delete requirements_document --force --print-parsed-parameters
  echo "pp_after_exit=$?"
  set -e
  echo

  echo "== debug: parsed parameters (flags BEFORE args) =="
  set +e
  "$BIN" node delete --force requirements_document --print-parsed-parameters
  echo "pp_before_exit=$?"
  set -e
  echo

  echo "== delete attempts (capture exit codes) =="
  set +e
  "$BIN" node delete requirements_document
  echo "del_no_force_exit=$?"

  "$BIN" node delete requirements_document --force
  echo "del_force_after_exit=$?"

  "$BIN" node delete --force requirements_document
  echo "del_force_before_exit=$?"

  "$BIN" node delete requirements_document -f
  echo "del_f_after_exit=$?"

  "$BIN" node delete -f requirements_document
  echo "del_f_before_exit=$?"
  set -e
  echo

  echo "== final: project.yaml =="
  sed -n '1,220p' .tactician/project.yaml
  echo
} >"$LOG" 2>&1

echo "Wrote repro log: $LOG"
echo "WORK=$WORK"
echo "Tip: tail -n 200 \"$LOG\""


