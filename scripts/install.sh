#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_PATH="${INSTALL_DIR}/gheppo"

mkdir -p "$INSTALL_DIR"

echo "Building Gheppo..."

cd "$PROJECT_ROOT"
go build -o "$BINARY_PATH" .

chmod +x "$BINARY_PATH"

echo "Installed Gheppo to $BINARY_PATH"

case "${SHELL##*/}" in
    zsh)
        SHELL_RC="${HOME}/.zshrc"
        SHELL_INTEGRATION="$PROJECT_ROOT/scripts/shell/gheppo.zsh"
        ;;
    bash)
        SHELL_RC="${HOME}/.bashrc"
        SHELL_INTEGRATION="$PROJECT_ROOT/scripts/shell/gheppo.bash"
        ;;
    *)
        echo
        echo "Could not detect a supported shell."
        echo "Gheppo was installed, but shell integration was not configured."
        exit 0
        ;;
esac

GHEPPO_START="# >>> gheppo >>>"
GHEPPO_END="# <<< gheppo <<<"

if ! grep -Fqx "$GHEPPO_START" "$SHELL_RC" 2>/dev/null; then
    {
        printf '\n%s\n' "$GHEPPO_START"
        cat "$SHELL_INTEGRATION"
        printf '%s\n' "$GHEPPO_END"
    } >> "$SHELL_RC"

    echo "Added Gheppo shell integration to $SHELL_RC"
else
    echo "Gheppo shell integration is already configured in $SHELL_RC"
fi

echo
echo "Gheppo installation complete."
echo
echo "Run:"
echo "  gheppo auth login"
echo "  gheppo sync"
