#!/bin/bash
# Helper script to create a feature branch with your GitHub username
# Usage: .githooks/create-feature-branch.sh <branch-name> [base-branch]
# Example: .githooks/create-feature-branch.sh add-auth-middleware main

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get inputs
FEATURE_NAME="${1:-}"
BASE_BRANCH="${2:-main}"

if [ -z "$FEATURE_NAME" ]; then
  echo -e "${YELLOW}Usage: .githooks/create-feature-branch.sh <branch-name> [base-branch]${NC}"
  echo "Example: .githooks/create-feature-branch.sh add-auth-middleware main"
  exit 1
fi

# Get GitHub username from git config or git remote
GITHUB_USER=$(git config user.name 2>/dev/null || echo "")

if [ -z "$GITHUB_USER" ]; then
  # Try to extract from remote URL
  GITHUB_USER=$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\).*/\1/')
fi

if [ -z "$GITHUB_USER" ]; then
  echo -e "${YELLOW}⚠️  Could not determine GitHub username${NC}"
  read -p "Enter your GitHub username: " GITHUB_USER
fi

BRANCH_NAME="$GITHUB_USER/$FEATURE_NAME"

echo -e "${BLUE}Creating branch: ${GREEN}$BRANCH_NAME${BLUE} from ${GREEN}$BASE_BRANCH${NC}"

# Fetch latest to ensure base branch is up to date
git fetch origin "$BASE_BRANCH" --quiet 2>/dev/null || true

# Create and checkout the branch
git checkout -b "$BRANCH_NAME" "origin/$BASE_BRANCH" 2>/dev/null || git checkout -b "$BRANCH_NAME" "$BASE_BRANCH"

echo -e "${GREEN}✅ Branch created and checked out: $BRANCH_NAME${NC}"
echo -e "${BLUE}Current branch:${NC} $(git branch --show-current)"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Make your changes"
echo "  2. Commit: git commit -m 'Your message'"
echo "  3. Push: git push -u origin $BRANCH_NAME"
echo "  4. Create PR on GitHub"
