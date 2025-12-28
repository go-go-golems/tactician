#!/usr/bin/env bash
set -euo pipefail

# Ticket script: 001-PORT-TO-GO smoke test runner (from pkg/doc/playbooks/smoke-test.md).
# - Runs the full smoke test in a fresh temp directory.
# - Leaves the temp directory on disk so you can inspect `.tactician/` afterwards.
# - Exits non-zero if any step (except the first unforced delete) fails.
#
# Usage:
#   ./scripts/01-run-smoke-test.sh
#
# Optional:
#   KEEP_WORK=0 ./scripts/01-run-smoke-test.sh   # deletes temp dir on success

ROOT="/home/manuel/workspaces/2025-12-28/port-tactician-go"
TACTICIAN_CMD_DIR="$ROOT/tactician/cmd/tactician"

if [[ ! -d "$ROOT/tactician" ]]; then
  echo "ERROR: expected tactician module at: $ROOT/tactician" >&2
  exit 2
fi
if [[ ! -f "$TACTICIAN_CMD_DIR/main.go" ]]; then
  echo "ERROR: expected tactician main at: $TACTICIAN_CMD_DIR/main.go" >&2
  exit 2
fi

WORK="$(mktemp -d)"
echo "WORK=$WORK"

BIN="$WORK/tactician"

echo "== preflight: go test ./... (tactician module) =="
(
  cd "$ROOT/tactician" && go test ./... -count=1
)

echo "== build: tactician binary into WORK =="
(
  cd "$ROOT/tactician" && go build -o "$BIN" ./cmd/tactician
)

KEEP_WORK="${KEEP_WORK:-1}"
cleanup() {
  if [[ "$KEEP_WORK" == "0" ]]; then
    rm -rf "$WORK"
  else
    echo "NOTE: keeping WORK directory: $WORK"
  fi
}
trap cleanup EXIT

cd "$WORK"

FAILED=0

run_step() {
  local name="$1"
  shift
  echo
  echo "== $name =="
  if ! "$@"; then
    FAILED=1
  fi
}

run_step "1) init" "$BIN" init

echo
echo "== check .tactician tree (first 50 files) =="
find .tactician -maxdepth 2 -type f | sed 's#^#- #' | head -n 50

run_step "2) node add" "$BIN" node add root README.md --type project_artifact --status pending
run_step "3) node show" "$BIN" node show root
run_step "4) node edit" "$BIN" node edit root --status complete

run_step "5a) search requirements" "$BIN" search requirements
run_step "5b) search --ready" "$BIN" search --ready
run_step "5c) search --tags planning,requirements" "$BIN" search --tags planning,requirements

run_step "6a) apply gather_requirements --yes" "$BIN" apply gather_requirements --yes
run_step "6b) apply write_technical_spec --yes --force" "$BIN" apply write_technical_spec --yes --force

run_step "7a) goals" "$BIN" goals
run_step "7b) goals --mermaid" "$BIN" goals --mermaid

run_step "8a) graph" "$BIN" graph
run_step "8b) graph --mermaid" "$BIN" graph --mermaid

run_step "9a) history" "$BIN" history
run_step "9b) history --summary" "$BIN" history --summary
run_step "9c) history --since 1d" "$BIN" history --since 1d

echo
echo "== 10a) node delete requirements_document (may fail; expected sometimes) =="
"$BIN" node delete requirements_document || true

run_step "10b) node delete requirements_document --force" "$BIN" node delete requirements_document --force

echo
echo "== DONE =="
echo "FAILED=$FAILED"
exit "$FAILED"


