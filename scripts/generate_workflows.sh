#!/usr/bin/env bash
# generate_workflows.sh
# Auto-generates ci-MODULE.yml and release-MODULE.yml for each Go module in the repo.
# Usage: ./scripts/generate_workflows.sh [--bump-type patch|minor|major] [--dry-run]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORKFLOWS_DIR="${REPO_ROOT}/.github/workflows"

BUMP_TYPE="patch"
DRY_RUN=false

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
    *)
      echo "Unknown argument: $1" >&2
      echo "Usage: $0 [--bump-type patch|minor|major] [--dry-run]" >&2
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
  else
    echo "$content" > "$file"
    echo "Written: ${file}"
  fi
}

mkdir -p "${WORKFLOWS_DIR}"

PROCESSED=()

for module_dir in "${MODULE_DIRS[@]}"; do
  module="$(basename "${module_dir}")"

  # Skip ignored modules
  skip=false
  for ignored in "${IGNORED_MODULES[@]}"; do
    [[ "$module" == "$ignored" ]] && skip=true && break
  done
  if [[ "$skip" == "true" ]]; then
    echo "Skipping: ${module}"
    continue
  fi

  ci_file="${WORKFLOWS_DIR}/ci-${module}.yml"
  release_file="${WORKFLOWS_DIR}/release-${module}.yml"

  write_or_print "$ci_file" "$(generate_ci "$module")"
  write_or_print "$release_file" "$(generate_release "$module" "$BUMP_TYPE")"
  PROCESSED+=("$module")
done

echo ""
echo "Done. Processed modules: $(IFS=', '; echo "${PROCESSED[*]}")"
