#!/usr/bin/env bash
# check_duplicates.sh - Detect duplicates in BDD suite
# Detects duplicate scenario names, step regexes, and requirement tags

set -e

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BDD_DIR="$(dirname "$SCRIPT_DIR")"
FEATURES_DIR="$BDD_DIR/features"
STEPS_DIR="$BDD_DIR/step_definitions"

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "======================================"
echo "BDD Suite Duplicate Detector"
echo "======================================"
echo

# Track issues
DUPLICATE_SCENARIOS=0
DUPLICATE_STEPS=0
DUPLICATE_TAGS=0

# ============================================================================
# 1. Check for duplicate scenario names
# ============================================================================
echo "1. Checking for duplicate scenario names..."
echo

# Find all scenario names, extract them, sort, and count duplicates
SCENARIO_DUPLICATES=$(find "$FEATURES_DIR" -name "*.feature" -exec cat {} \; 2>/dev/null | \
    grep -E '^Scenario:' | \
    sed 's/^Scenario:[[:space:]]*//' | \
    sort | \
    uniq -d)

if [[ -n "$SCENARIO_DUPLICATES" ]]; then
    echo -e "${RED}✗ Found duplicate scenario name(s):${NC}"
    echo "$SCENARIO_DUPLICATES" | while IFS= read -r scenario; do
        scenario=$(echo "$scenario" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        count=$(find "$FEATURES_DIR" -name "*.feature" -exec cat {} \; 2>/dev/null | \
            grep -F "^Scenario: $scenario" | wc -l | tr -d ' ')
        echo -e "  ${YELLOW}• \"$scenario\" (appears $count times)${NC}"
    done
    DUPLICATE_SCENARIOS=$(echo "$SCENARIO_DUPLICATES" | wc -l | tr -d ' ')
else
    echo -e "${GREEN}✓ No duplicate scenario names found${NC}"
fi

echo

# ============================================================================
# 2. Check for duplicate step regexes
# ============================================================================
echo "2. Checking for duplicate step regexes..."
echo

# Find all step registrations, extract regexes, sort, and count duplicates
STEP_DUPLICATES=$(grep -h 'suite\.\\(Given\\|When\\|Then\\)(' "$STEPS_DIR"/*.go 2>/dev/null | \
    grep -oE '\`\^.+\`' | \
    sort | \
    uniq -d)

if [[ -n "$STEP_DUPLICATES" ]]; then
    echo -e "${RED}✗ Found duplicate step regex(es):${NC}"
    echo "$STEP_DUPLICATES" | while IFS= read -r regex; do
        echo -e "  ${YELLOW}• $regex${NC}"
        # Show files where this regex appears
        grep -n "suite\.\\(Given\\|When\\|Then\\)($regex" "$STEPS_DIR"/*.go 2>/dev/null | \
            sed 's|'"$STEPS_DIR"'||' | \
            while IFS= read -r line; do
                echo -e "    $line"
            done
    done
    DUPLICATE_STEPS=$(echo "$STEP_DUPLICATES" | wc -l | tr -d ' ')
else
    echo -e "${GREEN}✓ No duplicate step regexes found${NC}"
fi

echo

# ============================================================================
# 3. Check for duplicate requirement tags
# ============================================================================
echo "3. Checking for duplicate requirement tags..."
echo

# Find all @requirement: tags, extract them, sort, and count duplicates
TAG_DUPLICATES=$(grep -h '@requirement:' "$FEATURES_DIR"/*.feature 2>/dev/null | \
    grep -oE '@requirement:[A-Z0-9_-]+' | \
    sort | \
    uniq -d)

if [[ -n "$TAG_DUPLICATES" ]]; then
    echo -e "${RED}✗ Found duplicate requirement tag(s):${NC}"
    echo "$TAG_DUPLICATES" | while IFS= read -r tag; do
        count=$(grep -h "$tag" "$FEATURES_DIR"/*.feature 2>/dev/null | wc -l | tr -d ' ')
        echo -e "  ${YELLOW}• $tag (appears $count times)${NC}"
    done
    DUPLICATE_TAGS=$(echo "$TAG_DUPLICATES" | wc -l | tr -d ' ')
else
    echo -e "${GREEN}✓ No duplicate requirement tags found${NC}"
fi

echo

# ============================================================================
# Summary
# ============================================================================
echo "======================================"
echo "Summary"
echo "======================================"
echo

TOTAL_ISSUES=$((DUPLICATE_SCENARIOS + DUPLICATE_STEPS + DUPLICATE_TAGS))

if [[ $TOTAL_ISSUES -eq 0 ]]; then
    echo -e "${GREEN}✓ No duplicates found! BDD suite is clean.${NC}"
    exit 0
else
    echo -e "${RED}✗ Found $TOTAL_ISSUES duplicate issue(s):${NC}"
    echo "  • Duplicate scenario names: $DUPLICATE_SCENARIOS"
    echo "  • Duplicate step regexes: $DUPLICATE_STEPS"
    echo "  • Duplicate requirement tags: $DUPLICATE_TAGS"
    echo
    echo -e "${YELLOW}Please consolidate duplicates to improve maintainability.${NC}"
    exit 1
fi
