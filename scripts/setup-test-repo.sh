#!/bin/bash
# Setup script for experiments-by-mrugesh/test-repo
# Run once to configure the test repository for integration tests
# Usage: just setup-test-repo

set -euo pipefail

REPO="experiments-by-mrugesh/test-repo"
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Setting up test repository: $REPO"
echo "Using temp directory: $TEMP_DIR"

# Clone the repo
cd "$TEMP_DIR"
gh repo clone "$REPO" repo
cd repo

# Configure git
git config user.email "test@git-wt.local"
git config user.name "Test Setup"

# === MAIN BRANCH FILES ===
echo "Setting up main branch files..."

cat > README.md << 'EOF'
# Test Repository

This repository is used for git-wt integration testing.
Do not delete branches without updating the tests.

## Branches

- `main` - default branch
- `develop` - development branch
- `feature/auth` - feature branch with slash
- `feature/nested/deep` - deeply nested feature branch
- `bugfix-123` - simple bugfix branch
EOF

cat > .envrc << 'EOF'
# Test envrc for direnv hook testing
export TEST_VAR="integration-test"
EOF

cat > .git-wt.toml << 'EOF'
default_base_branch = "main"
default_remote = "origin"

[hooks]
post_add = ["echo 'test hook ran'"]
EOF

mkdir -p src
cat > src/app.js << 'EOF'
// Placeholder file for realistic repo structure
console.log("Hello from test repo");
EOF

git add -A
git commit -m "chore: set up test repo structure" || echo "No changes to commit on main"
git push origin main

# === CREATE BRANCHES ===
echo "Creating feature branches..."

# develop branch
git checkout -b develop 2>/dev/null || git checkout develop
cat > develop.txt << 'EOF'
develop branch content
EOF
git add develop.txt
git commit -m "chore: add develop branch marker" || echo "No changes"
git push -u origin develop --force

# feature/auth branch
git checkout main
git checkout -b feature/auth 2>/dev/null || git checkout feature/auth
cat > auth.txt << 'EOF'
auth feature content
EOF
git add auth.txt
git commit -m "feat: add auth feature marker" || echo "No changes"
git push -u origin feature/auth --force

# feature/nested/deep branch
git checkout main
git checkout -b feature/nested/deep 2>/dev/null || git checkout feature/nested/deep
cat > nested.txt << 'EOF'
nested feature content
EOF
git add nested.txt
git commit -m "feat: add nested feature marker" || echo "No changes"
git push -u origin feature/nested/deep --force

# bugfix-123 branch
git checkout main
git checkout -b bugfix-123 2>/dev/null || git checkout bugfix-123
cat > bugfix.txt << 'EOF'
bugfix content
EOF
git add bugfix.txt
git commit -m "fix: add bugfix marker" || echo "No changes"
git push -u origin bugfix-123 --force

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Verifying branches..."
gh api repos/$REPO/branches --jq '.[].name' | sort

echo ""
echo "Test repo is ready: https://github.com/$REPO"
