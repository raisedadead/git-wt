#!/bin/bash
# @name: github-pr
# @description: Checkout PR branch for review (uses actual PR head branch)
# @events: pre_create
# @requires: gh

source "$WT_LIB/helpers.sh"
wt_requires gh
wt_requires jq

# Get the correct repo for PR operations
# Handles fork pattern (upstream) vs direct origin pattern
get_pr_repo() {
    local upstream_url origin_url repo

    # Check for upstream remote (fork pattern)
    upstream_url=$(git remote get-url upstream 2>/dev/null)
    if [ -n "$upstream_url" ]; then
        # Extract owner/repo from upstream URL
        repo=$(echo "$upstream_url" | sed -E 's|.*github\.com[:/]([^/]+/[^/.]+)(\.git)?$|\1|')
        echo "$repo"
        return
    fi

    # Fall back to origin
    origin_url=$(git remote get-url origin 2>/dev/null)
    if [ -n "$origin_url" ]; then
        repo=$(echo "$origin_url" | sed -E 's|.*github\.com[:/]([^/]+/[^/.]+)(\.git)?$|\1|')
        echo "$repo"
        return
    fi

    # No suitable remote found
    echo ""
}

# Select a PR from the list of open PRs
# Returns the selected PR number
select_pr_from_list() {
    local repo="$1"
    local prs pr_count

    wt_info "Fetching open PRs from $repo..."

    # Fetch open PRs with relevant fields
    prs=$(gh pr list -R "$repo" --state open --json number,title,author --jq '.[] | "#\(.number) \(.title) (\(.author.login))"' 2>/dev/null)

    if [ -z "$prs" ]; then
        echo ""
        return
    fi

    pr_count=$(echo "$prs" | wc -l | tr -d ' ')

    wt_info "Found $pr_count open PR(s)"
    echo ""

    # Add manual entry option
    local options
    options=$(printf "%s\nEnter PR number manually" "$prs")

    # Use bash select for interactive menu
    echo "Select a PR to review:"
    PS3="Enter selection: "

    local selection
    select selection in $options; do
        if [ -n "$selection" ]; then
            if [ "$selection" = "Enter PR number manually" ]; then
                echo "manual"
                return
            fi
            # Extract PR number from selection (format: "#123 title (author)")
            echo "$selection" | sed -E 's/^#([0-9]+).*/\1/'
            return
        else
            echo "Invalid selection. Please try again."
        fi
    done
}

# Get PR number from environment or prompt
pr_num="${WT_PR:-}"

if [ -z "$pr_num" ]; then
    # Only prompt if stdin is a TTY (interactive mode)
    if [ -t 0 ]; then
        # Try to show PR list for selection
        repo=$(get_pr_repo)
        if [ -n "$repo" ]; then
            pr_num=$(select_pr_from_list "$repo")

            if [ "$pr_num" = "manual" ] || [ -z "$pr_num" ]; then
                # Fallback: No open PRs found or user chose manual entry
                if [ -z "$pr_num" ]; then
                    wt_info "No open PRs found. Enter PR number manually."
                fi
                read -p "PR number: " pr_num
            fi
        else
            # Couldn't detect repo, fall back to manual entry
            wt_info "Could not detect repository. Enter PR number manually."
            read -p "PR number: " pr_num
        fi
    else
        wt_error "PR number is required. Use --pr <number> or run interactively."
    fi
fi

if [ -z "$pr_num" ]; then
    wt_error "PR number is required"
fi

# Fetch PR from GitHub
wt_info "Fetching PR #$pr_num..."

# If we detected a repo, use it for the PR view
repo=$(get_pr_repo)
if [ -n "$repo" ]; then
    pr=$(gh pr view "$pr_num" -R "$repo" --json number,title,author,headRefName,url,state 2>&1)
else
    pr=$(gh pr view "$pr_num" --json number,title,author,headRefName,url,state 2>&1)
fi

if [ $? -ne 0 ]; then
    wt_error "Failed to fetch PR #$pr_num: $pr"
fi

number=$(echo "$pr" | jq -r '.number')
title=$(echo "$pr" | jq -r '.title')
author=$(echo "$pr" | jq -r '.author.login')
branch=$(echo "$pr" | jq -r '.headRefName')
url=$(echo "$pr" | jq -r '.url')
state=$(echo "$pr" | jq -r '.state')

# Use the actual PR head branch so gh pr browse works
wt_set_branch "$branch"
wt_set_meta "pr_number" "$number"
wt_set_meta "pr_title" "$title"
wt_set_meta "pr_author" "$author"
wt_set_meta "pr_url" "$url"
wt_set_meta "pr_state" "$state"
wt_set_meta "track_remote" "true"

wt_info "PR #$number by @$author: $title"
wt_info "Branch: $branch (state: $state)"
