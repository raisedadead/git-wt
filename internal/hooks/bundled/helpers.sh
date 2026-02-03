#!/bin/bash
# Helper functions for git-wt hooks
# Source this file: source "$GIT_WT_LIB/helpers.sh"

# Set the suggested branch name
wt_set_branch() {
    echo "GIT_WT_BRANCH=$1" >> "$GIT_WT_OUTPUT"
}

# Store metadata (available to subsequent hooks and command)
wt_set_meta() {
    echo "GIT_WT_META_$1=$2" >> "$GIT_WT_OUTPUT"
}

# Request user input with a default value
wt_prompt() {
    echo "GIT_WT_PROMPT_$1=$2" >> "$GIT_WT_OUTPUT"
}

# Slugify text for branch names
# Converts to lowercase, replaces spaces with hyphens, removes special chars
wt_slugify() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | \
    tr -cd '[:alnum:]-' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//' | cut -c1-50
}

# Check if a command is available, error if not
wt_requires() {
    if ! command -v "$1" &>/dev/null; then
        wt_error "$1 is required but not installed"
    fi
}

# Signal an error and abort the hook
wt_error() {
    echo "GIT_WT_ERROR=$1" >> "$GIT_WT_OUTPUT"
    exit 1
}

# Log an info message (shown to user)
wt_info() {
    echo "$1"
}

# Log a warning message
wt_warn() {
    echo "GIT_WT_WARNING=$1" >> "$GIT_WT_OUTPUT"
}
