#!/usr/bin/env bash
#
# sync.sh — Declarative GitHub org repo config sync.
#
# Reads repos.json and ensures repo settings and branch protection
# match the desired state across all managed repos.
#
# Usage:
#   ./github/sync.sh --check              # Report drift (exit 1 if any)
#   ./github/sync.sh --apply              # Apply desired config to all repos
#   ./github/sync.sh --apply --repo NAME  # Apply to a single repo
#
# Requirements: gh (GitHub CLI), jq

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG="${SCRIPT_DIR}/repos.json"

# --- Helpers ----------------------------------------------------------------

die() { echo "error: $*" >&2; exit 2; }

usage() {
  cat <<EOF
Usage: $(basename "$0") [--check | --apply] [--repo NAME]

  --check   Compare actual vs desired, report diffs (exit 1 if drifted)
  --apply   Apply desired config to all repos (or one with --repo)
  --repo    Target a single repo
EOF
  exit 0
}

# --- Parse args -------------------------------------------------------------

MODE=""
TARGET_REPO=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check) MODE="check"; shift ;;
    --apply) MODE="apply"; shift ;;
    --repo)  TARGET_REPO="$2"; shift 2 ;;
    --help|-h) usage ;;
    *) die "unknown option: $1" ;;
  esac
done

[[ -z "$MODE" ]] && die "specify --check or --apply"
[[ -f "$CONFIG" ]] || die "config not found: $CONFIG"
command -v gh >/dev/null || die "gh CLI not found"
command -v jq >/dev/null || die "jq not found"

# --- Load config ------------------------------------------------------------

ORG=$(jq -r '.org' "$CONFIG")
REPOS=$(jq -r '.repos[]' "$CONFIG")
DRIFT=0

# If --repo was given, filter the list.
if [[ -n "$TARGET_REPO" ]]; then
  if ! jq -e --arg r "$TARGET_REPO" '.repos[] | select(. == $r)' "$CONFIG" >/dev/null 2>&1; then
    die "repo '$TARGET_REPO' not in config"
  fi
  REPOS="$TARGET_REPO"
fi

# --- Repo settings ----------------------------------------------------------

check_settings() {
  local repo="$1"
  local actual
  actual=$(gh api "repos/${ORG}/${repo}" 2>/dev/null) || { echo "  [$repo] settings: ERROR (could not fetch)"; DRIFT=1; return; }

  local drifts=()
  while IFS= read -r key; do
    local desired actual_val
    desired=$(jq -r --arg k "$key" '.settings[$k]' "$CONFIG")
    actual_val=$(echo "$actual" | jq -r --arg k "$key" '.[$k]')
    if [[ "$actual_val" != "$desired" ]]; then
      drifts+=("  - ${key}: actual=${actual_val}, desired=${desired}")
    fi
  done < <(jq -r '.settings | keys[]' "$CONFIG")

  if [[ ${#drifts[@]} -eq 0 ]]; then
    echo "  [$repo] settings: OK"
  else
    echo "  [$repo] settings: DRIFT"
    printf '%s\n' "${drifts[@]}"
    DRIFT=1
  fi
}

apply_settings() {
  local repo="$1"
  local payload
  payload=$(jq -c '.settings' "$CONFIG")
  gh api -X PATCH "repos/${ORG}/${repo}" --input - <<< "$payload" >/dev/null
  echo "  [$repo] settings: APPLIED"
}

# --- Branch protection ------------------------------------------------------

check_branch_protection() {
  local repo="$1"
  local branch
  branch=$(jq -r '.branch_protection.branch' "$CONFIG")

  local actual
  if ! actual=$(gh api "repos/${ORG}/${repo}/branches/${branch}/protection" 2>&1); then
    echo "  [$repo] branch_protection: MISSING"
    DRIFT=1
    return
  fi

  # Verify we got valid protection data (not a 404 message).
  if echo "$actual" | jq -e '.message' >/dev/null 2>&1; then
    echo "  [$repo] branch_protection: MISSING"
    DRIFT=1
    return
  fi

  local drifts=()

  # Helper: extract .foo.enabled, defaulting null to false.
  _bp_enabled() { echo "$actual" | jq -r "($1) // false"; }

  # enforce_admins
  local desired_enforce actual_enforce
  desired_enforce=$(jq -r '.branch_protection.enforce_admins' "$CONFIG")
  actual_enforce=$(_bp_enabled '.enforce_admins.enabled')
  [[ "$actual_enforce" != "$desired_enforce" ]] && drifts+=("  - enforce_admins: actual=${actual_enforce}, desired=${desired_enforce}")

  # required_status_checks.strict
  local desired_strict actual_strict
  desired_strict=$(jq -r '.branch_protection.required_status_checks.strict' "$CONFIG")
  actual_strict=$(_bp_enabled '.required_status_checks.strict')
  [[ "$actual_strict" != "$desired_strict" ]] && drifts+=("  - required_status_checks.strict: actual=${actual_strict}, desired=${desired_strict}")

  # required_pull_request_reviews.required_approving_review_count
  local desired_reviews actual_reviews
  desired_reviews=$(jq -r '.branch_protection.required_pull_request_reviews.required_approving_review_count' "$CONFIG")
  actual_reviews=$(echo "$actual" | jq -r '.required_pull_request_reviews.required_approving_review_count // 0')
  [[ "$actual_reviews" != "$desired_reviews" ]] && drifts+=("  - required_approving_review_count: actual=${actual_reviews}, desired=${desired_reviews}")

  # required_conversation_resolution
  local desired_convo actual_convo
  desired_convo=$(jq -r '.branch_protection.required_conversation_resolution' "$CONFIG")
  actual_convo=$(_bp_enabled '.required_conversation_resolution.enabled')
  [[ "$actual_convo" != "$desired_convo" ]] && drifts+=("  - required_conversation_resolution: actual=${actual_convo}, desired=${desired_convo}")

  # allow_force_pushes
  local desired_force actual_force
  desired_force=$(jq -r '.branch_protection.allow_force_pushes' "$CONFIG")
  actual_force=$(_bp_enabled '.allow_force_pushes.enabled')
  [[ "$actual_force" != "$desired_force" ]] && drifts+=("  - allow_force_pushes: actual=${actual_force}, desired=${desired_force}")

  # allow_deletions
  local desired_del actual_del
  desired_del=$(jq -r '.branch_protection.allow_deletions' "$CONFIG")
  actual_del=$(_bp_enabled '.allow_deletions.enabled')
  [[ "$actual_del" != "$desired_del" ]] && drifts+=("  - allow_deletions: actual=${actual_del}, desired=${desired_del}")

  # block_creations
  local desired_block actual_block
  desired_block=$(jq -r '.branch_protection.block_creations' "$CONFIG")
  actual_block=$(_bp_enabled '.block_creations.enabled')
  [[ "$actual_block" != "$desired_block" ]] && drifts+=("  - block_creations: actual=${actual_block}, desired=${desired_block}")

  if [[ ${#drifts[@]} -eq 0 ]]; then
    echo "  [$repo] branch_protection: OK"
  else
    echo "  [$repo] branch_protection: DRIFT"
    printf '%s\n' "${drifts[@]}"
    DRIFT=1
  fi
}

apply_branch_protection() {
  local repo="$1"
  local branch
  branch=$(jq -r '.branch_protection.branch' "$CONFIG")

  local desired_enforce desired_strict desired_reviews desired_convo
  local desired_force desired_del desired_block
  desired_enforce=$(jq -r '.branch_protection.enforce_admins' "$CONFIG")
  desired_strict=$(jq -r '.branch_protection.required_status_checks.strict' "$CONFIG")
  desired_reviews=$(jq -r '.branch_protection.required_pull_request_reviews.required_approving_review_count' "$CONFIG")
  desired_convo=$(jq -r '.branch_protection.required_conversation_resolution' "$CONFIG")
  desired_force=$(jq -r '.branch_protection.allow_force_pushes' "$CONFIG")
  desired_del=$(jq -r '.branch_protection.allow_deletions' "$CONFIG")
  desired_block=$(jq -r '.branch_protection.block_creations' "$CONFIG")

  local restrictions
  restrictions=$(jq -c '.branch_protection.restrictions' "$CONFIG")

  local payload
  payload=$(jq -n \
    --argjson enforce "$desired_enforce" \
    --argjson strict "$desired_strict" \
    --argjson reviews "$desired_reviews" \
    --argjson convo "$desired_convo" \
    --argjson force "$desired_force" \
    --argjson del "$desired_del" \
    --argjson block "$desired_block" \
    --argjson restrictions "$restrictions" \
    '{
      enforce_admins: $enforce,
      required_status_checks: { strict: $strict, contexts: [] },
      required_pull_request_reviews: {
        required_approving_review_count: $reviews,
        dismiss_stale_reviews: false,
        require_code_owner_reviews: false
      },
      required_conversation_resolution: $convo,
      allow_force_pushes: $force,
      allow_deletions: $del,
      block_creations: $block,
      restrictions: $restrictions
    }')

  gh api -X PUT "repos/${ORG}/${repo}/branches/${branch}/protection" \
    --input - <<< "$payload" >/dev/null
  echo "  [$repo] branch_protection: APPLIED"
}

# --- Main loop --------------------------------------------------------------

echo ""
for repo in $REPOS; do
  echo "[$repo]"

  if [[ "$MODE" == "check" ]]; then
    check_settings "$repo"
    check_branch_protection "$repo"
  else
    apply_settings "$repo"
    apply_branch_protection "$repo"
  fi

  echo ""
done

if [[ "$MODE" == "check" ]]; then
  if [[ $DRIFT -eq 0 ]]; then
    echo "All repos match desired config."
    exit 0
  else
    echo "Drift detected — run with --apply to fix."
    exit 1
  fi
else
  echo "Done. Run --check to verify."
fi
