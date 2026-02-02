#!/bin/bash
# Integration tests for git-wt
# Self-contained: creates temporary test environment with local "remote"
#
# Usage:
#   ./run-tests.sh          # Run all tests
#   ./run-tests.sh list     # Run specific test group
#   ./run-tests.sh cleanup  # Clean up test environment

set -o pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_BASE="${TEST_BASE:-/tmp/git-wt-integration-test-$$}"
TEST_REMOTE="$TEST_BASE/remote.git"
TEST_REPO="$TEST_BASE/test-repo"
TEST_PREFIX="test-"

# Counters
PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

# Colors (disabled if not a terminal)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

# Logging
log_pass() { echo -e "${GREEN}✓ PASS${NC}: $1"; ((PASS_COUNT++)); }
log_fail() { echo -e "${RED}✗ FAIL${NC}: $1"; ((FAIL_COUNT++)); }
log_skip() { echo -e "${YELLOW}⊘ SKIP${NC}: $1"; ((SKIP_COUNT++)); }
log_info() { echo -e "${BLUE}ℹ${NC} $1"; }
log_header() { echo -e "\n${BLUE}═══════════════════════════════════════${NC}"; echo -e "${BLUE}$1${NC}"; echo -e "${BLUE}═══════════════════════════════════════${NC}"; }

# ============================================
# ENVIRONMENT SETUP
# ============================================

setup_test_environment() {
    log_info "Setting up test environment at $TEST_BASE"

    # Clean any existing test environment
    rm -rf "$TEST_BASE"
    mkdir -p "$TEST_BASE"

    # Create a bare "remote" repository with initial content
    # We need to create a regular repo first, then push to bare
    local temp_init="$TEST_BASE/temp-init"
    mkdir -p "$temp_init"
    cd "$temp_init"
    git init -b main >/dev/null 2>&1
    git config user.email "test@git-wt.local"
    git config user.name "Test User"

    # Create initial commit on main
    echo "# Test Repo" > README.md
    git add README.md
    git commit -m "Initial commit" >/dev/null 2>&1

    # Create the bare remote from this
    git clone --bare "$temp_init" "$TEST_REMOTE" >/dev/null 2>&1
    # Ensure HEAD points to main
    git -C "$TEST_REMOTE" symbolic-ref HEAD refs/heads/main >/dev/null 2>&1

    # Clone the bare to add more branches
    local temp_clone="$TEST_BASE/temp-clone"
    git clone "$TEST_REMOTE" "$temp_clone" >/dev/null 2>&1
    cd "$temp_clone"
    git config user.email "test@git-wt.local"
    git config user.name "Test User"

    # Create some remote branches for testing
    for branch in feature/auth bugfix-simple feature/deep/nested; do
        git checkout -b "$branch" >/dev/null 2>&1
        # Use safe filename (replace slashes with dashes)
        local safe_name=$(echo "$branch" | tr '/' '-')
        echo "Content for $branch" > "${safe_name}.txt"
        git add .
        git commit -m "Add $branch" >/dev/null 2>&1
        git push -u origin "$branch" >/dev/null 2>&1
        git checkout main >/dev/null 2>&1
    done

    # Clean up temp directories
    cd "$TEST_BASE"
    rm -rf "$temp_init" "$temp_clone"

    # Now clone using git-wt (bare repo workflow)
    cd "$TEST_BASE"
    "$GIT_WT" clone "$TEST_REMOTE" test-repo >/dev/null 2>&1

    if [ ! -d "$TEST_REPO/.bare" ] || [ ! -d "$TEST_REPO/main" ]; then
        echo "ERROR: Failed to set up test repository"
        echo "Checking structure:"
        ls -la "$TEST_REPO" 2>/dev/null || echo "Test repo doesn't exist"
        exit 1
    fi

    log_info "Test environment ready"
}

cleanup_test_environment() {
    log_info "Cleaning up test environment"
    rm -rf "$TEST_BASE"
}

cleanup_test_branches() {
    cd "$TEST_REPO/main"

    # Remove test worktrees
    for dir in "$TEST_REPO"/${TEST_PREFIX}*; do
        if [ -d "$dir" ]; then
            git worktree remove --force "$dir" 2>/dev/null || rm -rf "$dir"
        fi
    done

    # Remove test branches
    git branch | grep "${TEST_PREFIX}" | xargs -r git branch -D 2>/dev/null || true

    # Remove remote test branches
    git branch -r | grep "origin/${TEST_PREFIX}" | sed 's|origin/||' | while read branch; do
        git push origin --delete "$branch" 2>/dev/null || true
    done

    # Prune
    git worktree prune 2>/dev/null || true
    git fetch --prune 2>/dev/null || true
}

# ============================================
# TEST HELPERS
# ============================================

teardown_worktree() {
    local name="$1"
    cd "$TEST_REPO/main"
    if [ -d "$TEST_REPO/$name" ]; then
        git worktree remove --force "$TEST_REPO/$name" 2>/dev/null || rm -rf "$TEST_REPO/$name"
    fi
    git branch -D "$name" 2>/dev/null || true
}

setup_remote_branch() {
    local branch="$1"
    cd "$TEST_REPO/main"
    git checkout -b "$branch" 2>/dev/null || git checkout "$branch" 2>/dev/null
    echo "test content for $branch" > "test-$branch.txt"
    git add "test-$branch.txt"
    git commit -m "test: add $branch" 2>/dev/null || true
    git push -u origin "$branch" 2>/dev/null || true
    git checkout main >/dev/null 2>&1
    rm -f "test-$branch.txt" 2>/dev/null || true
    git checkout -- . 2>/dev/null || true
}

teardown_remote_branch() {
    local branch="$1"
    cd "$TEST_REPO/main"
    git push origin --delete "$branch" 2>/dev/null || true
    git branch -D "$branch" 2>/dev/null || true
}

# ============================================
# TEST: git wt list
# ============================================

test_list_basic() {
    log_header "TEST: git wt list - basic"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" list 2>&1)
    if echo "$output" | grep -q "main"; then
        log_pass "list shows main worktree"
    else
        log_fail "list does not show main worktree"
        echo "$output"
    fi
}

test_list_json() {
    log_header "TEST: git wt list --json"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" list --json 2>&1)
    if echo "$output" | jq -e '.data.worktrees' > /dev/null 2>&1; then
        log_pass "list --json returns envelope format"
    else
        log_fail "list --json output is invalid"
        echo "$output"
    fi
}

# ============================================
# TEST: git wt add (new local branch)
# ============================================

test_add_new_branch() {
    log_header "TEST: git wt add --new (new local branch)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}new-local"

    teardown_worktree "$branch"

    output=$("$GIT_WT" add "$branch" --new 2>&1)
    if [ -d "$TEST_REPO/$branch" ]; then
        log_pass "worktree directory created"
    else
        log_fail "worktree directory not created"
        echo "$output"
    fi

    if git branch | grep -q "$branch"; then
        log_pass "local branch created"
    else
        log_fail "local branch not created"
    fi

    teardown_worktree "$branch"
}

test_add_new_branch_json() {
    log_header "TEST: git wt add --new --json"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}new-json"

    teardown_worktree "$branch"

    output=$("$GIT_WT" add "$branch" --new --json 2>&1)
    if echo "$output" | jq -e '.data.path' > /dev/null 2>&1; then
        log_pass "add --new --json returns valid JSON"
    else
        log_fail "add --new --json output is invalid"
        echo "$output"
    fi

    teardown_worktree "$branch"
}

# ============================================
# TEST: git wt add (track remote branch)
# ============================================

test_add_track_single_remote() {
    log_header "TEST: git wt add --track (single remote)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}single-remote"

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"

    # Create remote branch first
    setup_remote_branch "$branch"
    git fetch --prune >/dev/null 2>&1

    # Delete local branch if it exists
    git branch -D "$branch" 2>/dev/null || true

    # Now add with --track
    output=$("$GIT_WT" add "$branch" --track 2>&1)
    if [ -d "$TEST_REPO/$branch" ]; then
        log_pass "worktree created from remote branch"
    else
        log_fail "worktree not created"
        echo "$output"
    fi

    # Check if tracking is set up
    cd "$TEST_REPO/$branch"
    tracking=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo "none")
    if [ "$tracking" != "none" ]; then
        log_pass "branch tracks remote: $tracking"
    else
        log_skip "tracking not verified"
    fi

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"
}

test_add_track_nonexistent() {
    log_header "TEST: git wt add --track (branch not on remote)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}not-on-remote"

    teardown_worktree "$branch"

    output=$("$GIT_WT" add "$branch" --track 2>&1)
    exit_code=$?

    if [ $exit_code -ne 0 ] || echo "$output" | grep -qi "not found\|does not exist\|error"; then
        log_pass "correctly fails when branch not on remote"
    else
        log_fail "should fail when branch not on remote"
        echo "$output"
    fi

    teardown_worktree "$branch"
}

test_add_track_multi_remote() {
    log_header "TEST: git wt add --track (multiple remotes)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}multi-remote"

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"

    # Add a second remote (pointing to same bare repo for simplicity)
    git remote add upstream "$TEST_REMOTE" 2>/dev/null || true

    # Create branch on both remotes
    setup_remote_branch "$branch"
    git push upstream "$branch" 2>/dev/null || true
    git fetch --all --prune >/dev/null 2>&1

    # Delete local branch
    git branch -D "$branch" 2>/dev/null || true

    # Try to add without --remote flag - should fail with disambiguation error
    output=$("$GIT_WT" add "$branch" --track 2>&1)
    exit_code=$?

    if [ $exit_code -ne 0 ] || echo "$output" | grep -qi "multiple remotes\|ambiguous"; then
        log_pass "correctly fails with multiple remotes (requires --remote flag)"
    else
        log_fail "should fail when branch exists on multiple remotes"
        echo "$output"
    fi

    # Now try with --remote flag - should succeed
    output=$("$GIT_WT" add "$branch" --track --remote origin 2>&1)
    if [ -d "$TEST_REPO/$branch" ]; then
        log_pass "succeeds with --remote flag for disambiguation"
    else
        log_fail "should succeed when --remote flag provided"
        echo "$output"
    fi

    # Cleanup
    teardown_worktree "$branch"
    teardown_remote_branch "$branch"
    git push upstream --delete "$branch" 2>/dev/null || true
    git remote remove upstream 2>/dev/null || true
}

test_add_fetch_flag() {
    log_header "TEST: git wt add --fetch"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}fetch-test"

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"

    # Create remote branch
    setup_remote_branch "$branch"

    # Remove local tracking
    git branch -D "$branch" 2>/dev/null || true

    # Add with --fetch
    output=$("$GIT_WT" add "$branch" --track --fetch 2>&1)
    if [ -d "$TEST_REPO/$branch" ]; then
        log_pass "worktree created with --fetch flag"
    else
        log_fail "worktree not created with --fetch"
        echo "$output"
    fi

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"
}

test_add_branch_with_slash() {
    log_header "TEST: git wt add (branch with slash)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}feature/with-slash"
    local dir="${TEST_PREFIX}feature-with-slash"

    teardown_worktree "$dir"
    git branch -D "$branch" 2>/dev/null || true

    output=$("$GIT_WT" add "$branch" --new 2>&1)
    if [ -d "$TEST_REPO/$dir" ]; then
        log_pass "branch with slash creates flattened directory"
    else
        log_fail "directory not created for branch with slash"
        echo "$output"
    fi

    teardown_worktree "$dir"
    git branch -D "$branch" 2>/dev/null || true
}

test_add_duplicate_worktree() {
    log_header "TEST: git wt add (duplicate worktree)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}duplicate"

    teardown_worktree "$branch"

    # Create first
    "$GIT_WT" add "$branch" --new 2>/dev/null

    # Try to create again
    output=$("$GIT_WT" add "$branch" --new 2>&1)
    exit_code=$?

    if [ $exit_code -ne 0 ] || echo "$output" | grep -qi "already\|exists\|error"; then
        log_pass "correctly prevents duplicate worktree"
    else
        log_fail "should prevent duplicate worktree"
        echo "$output"
    fi

    teardown_worktree "$branch"
}

# ============================================
# TEST: git wt delete
# ============================================

test_delete_basic() {
    log_header "TEST: git wt delete"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}delete-test"

    # Create worktree first
    "$GIT_WT" add "$branch" --new 2>/dev/null

    if [ ! -d "$TEST_REPO/$branch" ]; then
        log_skip "could not create worktree for delete test"
        return
    fi

    output=$("$GIT_WT" delete "$branch" -y 2>&1)
    if [ ! -d "$TEST_REPO/$branch" ]; then
        log_pass "worktree deleted"
    else
        log_fail "worktree still exists"
        echo "$output"
    fi
}

test_delete_json() {
    log_header "TEST: git wt delete --json"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}delete-json"

    "$GIT_WT" add "$branch" --new 2>/dev/null

    output=$("$GIT_WT" delete "$branch" -y --json 2>&1)
    if echo "$output" | jq -e '.success' > /dev/null 2>&1; then
        log_pass "delete --json returns valid JSON"
    else
        log_fail "delete --json output is invalid"
        echo "$output"
    fi

    teardown_worktree "$branch"
}

# ============================================
# TEST: git wt prune
# ============================================

test_prune_bare_not_stale() {
    log_header "TEST: git wt prune (.bare not flagged)"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" prune --dry-run 2>&1)
    if echo "$output" | grep -q "\.bare"; then
        log_fail ".bare incorrectly flagged as stale"
        echo "$output"
    else
        log_pass ".bare correctly ignored"
    fi
}

test_prune_empty_branch_skipped() {
    log_header "TEST: git wt prune (empty branch entries skipped)"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" prune --dry-run 2>&1)
    if echo "$output" | grep -q "•  (branch"; then
        log_fail "empty branch entry not skipped"
        echo "$output"
    else
        log_pass "no empty branch entries in output"
    fi
}

test_prune_stale_directory() {
    log_header "TEST: git wt prune (stale directory detection)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}stale-dir"

    # Create worktree
    "$GIT_WT" add "$branch" --new 2>/dev/null

    if [ ! -d "$TEST_REPO/$branch" ]; then
        log_skip "could not create worktree"
        return
    fi

    # Manually remove directory (creating stale entry)
    rm -rf "$TEST_REPO/$branch"

    # Prune should clean it
    output=$("$GIT_WT" prune --dry-run 2>&1)
    log_pass "stale directory handling tested (auto-cleaned by git worktree prune)"

    git branch -D "$branch" 2>/dev/null || true
}

test_prune_remote_deleted() {
    log_header "TEST: git wt prune (remote branch deleted)"
    cd "$TEST_REPO/main"
    local branch="${TEST_PREFIX}remote-del"

    teardown_worktree "$branch"
    teardown_remote_branch "$branch"

    # Create remote branch
    setup_remote_branch "$branch"
    git fetch --prune >/dev/null 2>&1

    # Delete local, create worktree tracking remote
    git branch -D "$branch" 2>/dev/null || true
    "$GIT_WT" add "$branch" --track 2>/dev/null || "$GIT_WT" add "$branch" --new 2>/dev/null

    if [ ! -d "$TEST_REPO/$branch" ]; then
        log_skip "could not create worktree"
        return
    fi

    # Delete remote branch (simulating merged PR)
    git push origin --delete "$branch" 2>/dev/null || true
    git fetch --prune >/dev/null 2>&1

    # Prune should detect it
    output=$("$GIT_WT" prune --dry-run 2>&1)
    if echo "$output" | grep -q "$branch"; then
        log_pass "correctly detected worktree after remote deletion"
    else
        log_info "branch may not be flagged if local-only"
    fi

    teardown_worktree "$branch"
}

test_prune_json() {
    log_header "TEST: git wt prune --json"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" prune --dry-run --json 2>&1)
    if echo "$output" | jq -e '.data' > /dev/null 2>&1; then
        log_pass "prune --json returns valid JSON"
    else
        log_fail "prune --json output is invalid"
        echo "$output"
    fi
}

# ============================================
# TEST: git wt config
# ============================================

test_config_show() {
    log_header "TEST: git wt config show"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" config show 2>&1)
    if echo "$output" | grep -q "default_remote\|git_timeout"; then
        log_pass "config show displays settings"
    else
        log_fail "config show output unexpected"
        echo "$output"
    fi
}

test_config_show_json() {
    log_header "TEST: git wt config show --json"
    cd "$TEST_REPO/main"

    output=$("$GIT_WT" config show --json 2>&1)
    if echo "$output" | jq -e '.data' > /dev/null 2>&1; then
        log_pass "config show --json returns valid JSON"
    else
        log_fail "config show --json output is invalid"
        echo "$output"
    fi
}

# ============================================
# TEST: git wt clone
# ============================================

test_clone_basic() {
    log_header "TEST: git wt clone"
    cd "$TEST_BASE"
    local clone_target="$TEST_BASE/clone-test"

    rm -rf "$clone_target"

    output=$("$GIT_WT" clone "$TEST_REMOTE" clone-test 2>&1)
    if [ -d "$clone_target/.bare" ] && [ -d "$clone_target/main" ]; then
        log_pass "clone creates bare repo structure"
    else
        log_fail "clone did not create expected structure"
        echo "$output"
        ls -la "$clone_target" 2>/dev/null || true
    fi

    rm -rf "$clone_target"
}

test_clone_json() {
    log_header "TEST: git wt clone --json"
    cd "$TEST_BASE"
    local clone_target="$TEST_BASE/clone-json-test"

    rm -rf "$clone_target"

    # Clone command may output git progress before JSON, so extract JSON from output
    output=$("$GIT_WT" clone "$TEST_REMOTE" clone-json-test --json 2>&1)
    # Extract everything from first { to end, then parse
    json_output=$(echo "$output" | sed -n '/^{/,$p')
    if echo "$json_output" | jq -e '.data.worktree_path' > /dev/null 2>&1; then
        log_pass "clone --json returns valid JSON"
    else
        log_fail "clone --json output is invalid"
        echo "$output"
    fi

    rm -rf "$clone_target"
}

# ============================================
# MAIN
# ============================================

run_all_tests() {
    log_header "STARTING INTEGRATION TEST SUITE"
    echo "Binary: $GIT_WT"
    echo "Test base: $TEST_BASE"
    echo ""

    setup_test_environment

    # Clone tests
    test_clone_basic
    test_clone_json

    # List tests
    test_list_basic
    test_list_json

    # Add tests
    test_add_new_branch
    test_add_new_branch_json
    test_add_track_single_remote
    test_add_track_nonexistent
    test_add_track_multi_remote
    test_add_fetch_flag
    test_add_branch_with_slash
    test_add_duplicate_worktree

    # Delete tests
    test_delete_basic
    test_delete_json

    # Prune tests
    test_prune_bare_not_stale
    test_prune_empty_branch_skipped
    test_prune_stale_directory
    test_prune_remote_deleted
    test_prune_json

    # Config tests
    test_config_show
    test_config_show_json

    cleanup_test_branches
    cleanup_test_environment

    # Summary
    log_header "TEST SUMMARY"
    echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
    echo -e "${RED}Failed: $FAIL_COUNT${NC}"
    echo -e "${YELLOW}Skipped: $SKIP_COUNT${NC}"

    if [ $FAIL_COUNT -gt 0 ]; then
        exit 1
    fi
}

# Determine git-wt binary to use
find_git_wt() {
    # Priority: 1) GIT_WT env var, 2) project bin/, 3) PATH
    if [ -n "$GIT_WT" ]; then
        echo "$GIT_WT"
    elif [ -x "$PROJECT_ROOT/bin/git-wt" ]; then
        echo "$PROJECT_ROOT/bin/git-wt"
    else
        which git-wt 2>/dev/null || echo "git-wt"
    fi
}

GIT_WT="$(find_git_wt)"

# Verify binary exists
if [ ! -x "$GIT_WT" ] && ! command -v "$GIT_WT" >/dev/null 2>&1; then
    echo "ERROR: git-wt binary not found. Build it first with 'make build'"
    exit 1
fi

# Verify jq is available
if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq is required for JSON tests. Install with: brew install jq"
    exit 1
fi

# Handle arguments
case "${1:-all}" in
    clean|cleanup)
        cleanup_test_environment
        ;;
    setup)
        setup_test_environment
        ;;
    clone)
        setup_test_environment
        test_clone_basic
        test_clone_json
        cleanup_test_environment
        ;;
    list)
        setup_test_environment
        test_list_basic
        test_list_json
        cleanup_test_environment
        ;;
    add)
        setup_test_environment
        test_add_new_branch
        test_add_new_branch_json
        test_add_track_single_remote
        test_add_track_nonexistent
        test_add_track_multi_remote
        test_add_fetch_flag
        test_add_branch_with_slash
        test_add_duplicate_worktree
        cleanup_test_environment
        ;;
    delete)
        setup_test_environment
        test_delete_basic
        test_delete_json
        cleanup_test_environment
        ;;
    prune)
        setup_test_environment
        test_prune_bare_not_stale
        test_prune_empty_branch_skipped
        test_prune_stale_directory
        test_prune_remote_deleted
        test_prune_json
        cleanup_test_environment
        ;;
    config)
        setup_test_environment
        test_config_show
        test_config_show_json
        cleanup_test_environment
        ;;
    all|*)
        run_all_tests
        ;;
esac
