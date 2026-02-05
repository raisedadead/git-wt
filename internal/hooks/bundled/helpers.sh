#!/bin/sh
# Helper functions for wt hooks
# Source this file: . "$WT_LIB/helpers.sh"

# Set the suggested branch name
wt_set_branch() {
    [ -n "$WT_OUTPUT" ] || return 0
    echo "WT_BRANCH=$1" >> "$WT_OUTPUT"
}

# Store metadata (available to subsequent hooks and command)
wt_set_meta() {
    [ -n "$WT_OUTPUT" ] || return 0
    echo "WT_META_$1=$2" >> "$WT_OUTPUT"
}

# Request user input with a default value
wt_prompt() {
    [ -n "$WT_OUTPUT" ] || return 0
    echo "WT_PROMPT_$1=$2" >> "$WT_OUTPUT"
}

# Slugify text for branch names
# Converts to lowercase, replaces spaces with hyphens, removes special chars
wt_slugify() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | \
    tr -cd '[:alnum:]-' | sed 's/--*/-/g' | sed 's/^-//' | sed 's/-$//' | cut -c1-50
}

# Check if a command is available, error if not
wt_requires() {
    if ! command -v "$1" >/dev/null 2>&1; then
        wt_error "$1 is required but not installed"
    fi
}

# Signal an error and abort the hook
wt_error() {
    [ -n "$WT_OUTPUT" ] || { echo "error: $1" >&2; exit 1; }
    echo "WT_ERROR=$1" >> "$WT_OUTPUT"
    exit 1
}

# Log an info message (shown to user)
wt_info() {
    echo "$1"
}

# Log a warning message
wt_warn() {
    [ -n "$WT_OUTPUT" ] || { echo "warning: $1" >&2; return 0; }
    echo "WT_WARNING=$1" >> "$WT_OUTPUT"
}
