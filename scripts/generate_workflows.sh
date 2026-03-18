#!/usr/bin/env bash
# generate_workflows.sh
# Auto-generates ci-MODULE.yml and release-MODULE.yml for each Go module in the repo.
# Usage: ./scripts/generate_workflows.sh [--bump-type patch|minor|major] [--dry-run] [--module <name>...] [--yes]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORKFLOWS_DIR="${REPO_ROOT}/.github/workflows"

BUMP_TYPE="patch"
DRY_RUN=false
AUTO_YES=false
SELECTED_MODULES=()

# Modules to skip (legacy/non-active directories)
IGNORED_MODULES=("archive")

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --bump-type)
      BUMP_TYPE="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --module)
      SELECTED_MODULES+=("$2")
      shift 2
      ;;
    --yes|-y)
      AUTO_YES=true
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      echo "Usage: $0 [--bump-type patch|minor|major] [--dry-run] [--module <name>...] [--yes]" >&2
      exit 1
      ;;
  esac
done

if [[ ! "$BUMP_TYPE" =~ ^(patch|minor|major)$ ]]; then
  echo "Invalid bump_type: '$BUMP_TYPE'. Must be patch, minor, or major." >&2
  exit 1
fi

# Discover all modules: directories containing a go.mod (but not the repo root itself)
MODULE_DIRS=()
_tmpfile=$(mktemp)
find "${REPO_ROOT}" -mindepth 2 -maxdepth 2 -name "go.mod" -exec dirname {} \; | sort > "$_tmpfile"
while IFS= read -r dir; do
  MODULE_DIRS+=("$dir")
done < "$_tmpfile"
rm -f "$_tmpfile"

if [[ ${#MODULE_DIRS[@]} -eq 0 ]]; then
  echo "No Go modules found." >&2
  exit 1
fi

generate_ci() {
  local module="$1"
  cat <<EOF
name: CI ${module}

on:
  push:
    branches:
      - main
    paths:
      - "${module}/**"
  pull_request:
    paths:
      - "${module}/**"

jobs:
  ci:
    uses: ./.github/workflows/ci-go-module.yml
    with:
      module_path: ${module}
EOF
}

generate_release() {
  local module="$1"
  local bump_type="$2"
  cat <<EOF
name: Release ${module}

on:
  push:
    branches:
      - main
    paths:
      - "${module}/**"

permissions:
  contents: write

jobs:
  release:
    uses: ./.github/workflows/deploy-go-module.yml
    with:
      module_path: ${module}
      bump_type: ${bump_type}
    secrets: inherit
EOF
}

write_or_print() {
  local file="$1"
  local content="$2"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "=== [dry-run] ${file} ==="
    echo "$content"
    echo ""
    return
  fi
  if [[ -f "$file" ]] && [[ "$AUTO_YES" == "false" ]]; then
    read -r -p "File $(basename "$file") already exists. Overwrite? [y/N] " answer </dev/tty
    if [[ ! "$answer" =~ ^[Yy]$ ]]; then
      echo "Skipped: ${file}"
      return
    fi
  fi
  echo "$content" > "$file"
  echo "Written: ${file}"
}

mkdir -p "${WORKFLOWS_DIR}"

# Build list of non-ignored modules
AVAILABLE_MODULES=()
for module_dir in "${MODULE_DIRS[@]}"; do
  module="$(basename "${module_dir}")"
  skip=false
  for ignored in "${IGNORED_MODULES[@]}"; do
    [[ "$module" == "$ignored" ]] && skip=true && break
  done
  [[ "$skip" == "false" ]] && AVAILABLE_MODULES+=("$module")
done

# Determine target modules
TARGET_MODULES=()
if [[ ${#SELECTED_MODULES[@]} -gt 0 ]]; then
  # Validate provided module names
  for sel in "${SELECTED_MODULES[@]}"; do
    found=false
    for avail in "${AVAILABLE_MODULES[@]}"; do
      [[ "$sel" == "$avail" ]] && found=true && break
    done
    if [[ "$found" == "false" ]]; then
      echo "Unknown module: '${sel}'. Available modules: $(IFS=', '; echo "${AVAILABLE_MODULES[*]}")" >&2
      exit 1
    fi
    TARGET_MODULES+=("$sel")
  done
elif [[ -t 0 ]]; then
  # Interactive multi-select
  echo "Available modules:"
  for i in "${!AVAILABLE_MODULES[@]}"; do
    printf "  %d) %s\n" "$((i+1))" "${AVAILABLE_MODULES[$i]}"
  done
  echo ""
  read -r -p "Select modules (space-separated numbers, or 'a' for all): " selection </dev/tty
  if [[ "$selection" == "a" || "$selection" == "A" ]]; then
    TARGET_MODULES=("${AVAILABLE_MODULES[@]}")
  else
    for num in $selection; do
      if ! [[ "$num" =~ ^[0-9]+$ ]] || (( num < 1 || num > ${#AVAILABLE_MODULES[@]} )); then
        echo "Invalid selection: '${num}'" >&2
        exit 1
      fi
      TARGET_MODULES+=("${AVAILABLE_MODULES[$((num-1))]}")
    done
  fi
  if [[ ${#TARGET_MODULES[@]} -eq 0 ]]; then
    echo "No modules selected." >&2
    exit 1
  fi
  echo ""
  echo "Selected: $(IFS=', '; echo "${TARGET_MODULES[*]}")"
  read -r -p "Proceed? [y/N] " confirm </dev/tty
  [[ "$confirm" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }
  echo ""
else
  TARGET_MODULES=("${AVAILABLE_MODULES[@]}")
fi

PROCESSED=()

for module in "${TARGET_MODULES[@]}"; do
  ci_file="${WORKFLOWS_DIR}/ci-${module}.yml"
  release_file="${WORKFLOWS_DIR}/release-${module}.yml"

  write_or_print "$ci_file" "$(generate_ci "$module")"
  write_or_print "$release_file" "$(generate_release "$module" "$BUMP_TYPE")"
  PROCESSED+=("$module")
done

echo ""
echo "Done. Processed modules: $(IFS=', '; echo "${PROCESSED[*]}")"
