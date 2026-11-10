#!/usr/bin/env bash
# Picks the version bump for a release run.
# A manual dispatch passes INPUT_BUMP; a push to main leaves it empty, so the
# labels of the PR that produced the commit decide: release:major > release:minor > patch.
set -euo pipefail

BUMP="${INPUT_BUMP:-}"

if [ -z "$BUMP" ]; then
  LABELS=$(gh api "repos/${GITHUB_REPOSITORY}/commits/${GITHUB_SHA}/pulls" \
    --jq '.[0].labels[]?.name' 2>/dev/null || true)

  if grep -qx 'release:major' <<<"$LABELS"; then
    BUMP=major
  elif grep -qx 'release:minor' <<<"$LABELS"; then
    BUMP=minor
  else
    BUMP=patch
  fi
  echo "PR labels: ${LABELS:-<none>}"
fi

case "$BUMP" in
  patch|minor|major) ;;
  *) echo "ERROR: invalid bump type '$BUMP'"; exit 1 ;;
esac

echo "Bump type: $BUMP"
echo "bump_type=$BUMP" >> "$GITHUB_OUTPUT"
