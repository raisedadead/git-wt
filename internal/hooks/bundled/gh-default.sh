#!/bin/bash
# @name: gh-default
# @description: Auto-configure GitHub CLI default repository
# @events: post_clone
# @requires: gh

cd "$GIT_WT_PATH" || exit 0

# Skip if gh not installed
command -v gh &>/dev/null || exit 0

# Skip if already configured
if gh repo set-default --view &>/dev/null; then
    exit 0
fi

# Get remotes
origin=$(git remote get-url origin 2>/dev/null)
upstream=$(git remote get-url upstream 2>/dev/null)

# Detection logic:
# 1. origin + upstream exist → use upstream (fork pattern)
# 2. only origin exists → use origin
# 3. other patterns → skip silently

if [[ -n "$upstream" && -n "$origin" ]]; then
    gh repo set-default "$upstream" 2>/dev/null
elif [[ -n "$origin" && -z "$upstream" ]]; then
    gh repo set-default "$origin" 2>/dev/null
fi
