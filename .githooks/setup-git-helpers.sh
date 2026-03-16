#!/bin/bash
# Setup git aliases and configuration for quick feature branch creation

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Setting up git helpers...${NC}"

# Set up git hooks directory
git config core.hooksPath .githooks
echo -e "${GREEN}✓${NC} Git hooks path set to .githooks"

# Create git aliases for quick branch creation
git config alias.feature '!bash .githooks/create-feature-branch.sh'
echo -e "${GREEN}✓${NC} Git alias created: ${YELLOW}git feature <name>${NC}"

git config alias.feature-main '!bash .githooks/create-feature-branch.sh $1 main'
echo -e "${GREEN}✓${NC} Git alias created: ${YELLOW}git feature-main <name>${NC}"

git config alias.feature-develop '!bash .githooks/create-feature-branch.sh $1 develop'
echo -e "${GREEN}✓${NC} Git alias created: ${YELLOW}git feature-develop <name>${NC}"

echo ""
echo -e "${BLUE}Setup complete!${NC}"
echo ""
echo -e "${YELLOW}Usage examples:${NC}"
echo -e "  ${GREEN}git feature my-feature${NC}           # Create from current branch"
echo -e "  ${GREEN}git feature-main my-feature${NC}      # Create from main"
echo -e "  ${GREEN}git feature-develop my-feature${NC}   # Create from develop"
echo ""
echo -e "${YELLOW}These will create branches like:${NC} ${GREEN}username/my-feature${NC}"
