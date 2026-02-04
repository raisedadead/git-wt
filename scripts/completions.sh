#!/bin/bash
# Generate shell completions for wt
# Called by goreleaser before build and by `just install-completions`

set -e

rm -rf completions
mkdir -p completions

# Use provided binary, or fall back to go run for goreleaser
if [[ -n "$1" && -x "$1" ]]; then
    CMD="$1"
    echo "Generating completions using binary: $CMD"
else
    CMD="go run ./cmd/wt"
    echo "Generating completions using: $CMD"
fi

$CMD completion bash > completions/wt.bash
$CMD completion zsh > completions/_wt
$CMD completion fish > completions/wt.fish

echo "Generated:"
ls -la completions/
