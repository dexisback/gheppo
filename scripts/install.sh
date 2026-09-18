#!/bin/sh

set -eu

REPO="dexisback/gheppo"
DEFAULT_VERSION="0.1.0"
VERSION="${GHEPPO_VERSION:-$DEFAULT_VERSION}"
VERSION="${VERSION#v}"

# Validate version format (e.g. 0.1.0, 1.2.3-beta.1)
case "$VERSION" in
    *[!0-9a-zA-Z._-]*|"")
        echo "Error: Invalid version format '$VERSION'." >&2
        exit 1
        ;;
esac

INSTALL_DIR="${GHEPPO_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_PATH="${INSTALL_DIR}/gheppo"

# 1. Detect Operating System
OS_TYPE="$(uname -s)"
case "$OS_TYPE" in
    Linux*)
        OS="linux"
        EXT="tar.gz"
        ;;
    Darwin*)
        OS="darwin"
        EXT="tar.gz"
        ;;
    CYGWIN*|MINGW*|MSYS*|Windows_NT*)
        OS="windows"
        EXT="zip"
        BINARY_PATH="${INSTALL_DIR}/gheppo.exe"
        ;;
    *)
        echo "Error: Unsupported operating system '$OS_TYPE'." >&2
        echo "Gheppo pre-built releases currently support Linux, Darwin (macOS), and Windows." >&2
        exit 1
        ;;
esac

# 2. Detect CPU Architecture
ARCH_TYPE="$(uname -m)"
case "$ARCH_TYPE" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported CPU architecture '$ARCH_TYPE'." >&2
        echo "Gheppo pre-built releases currently support amd64 (x86_64) and arm64 (aarch64)." >&2
        exit 1
        ;;
esac

# 3. Formulate Asset & Release URLs
TAG="v${VERSION}"
ASSET_NAME="gheppo_${VERSION}_${OS}_${ARCH}.${EXT}"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"
ARCHIVE_URL="${BASE_URL}/${ASSET_NAME}"
CHECKSUMS_URL="${BASE_URL}/checksums.txt"

# 4. Set up temporary working directory with auto-cleanup
TMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'gheppo')"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM HUP

# 5. Download helper (curl / wget)
download_file() {
    _url="$1"
    _dest="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$_url" -o "$_dest"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$_dest" "$_url"
    else
        echo "Error: Neither 'curl' nor 'wget' was found on your system." >&2
        echo "Please install curl or wget to continue." >&2
        exit 1
    fi
}

echo "Downloading Gheppo ${TAG} (${OS}/${ARCH})..."
ARCHIVE_FILE="${TMP_DIR}/${ASSET_NAME}"
CHECKSUMS_FILE="${TMP_DIR}/checksums.txt"

if ! download_file "$ARCHIVE_URL" "$ARCHIVE_FILE"; then
    echo "Error: Failed to download release archive from:" >&2
    echo "  $ARCHIVE_URL" >&2
    exit 1
fi

if ! download_file "$CHECKSUMS_URL" "$CHECKSUMS_FILE"; then
    echo "Error: Failed to download checksums from:" >&2
    echo "  $CHECKSUMS_URL" >&2
    exit 1
fi

# 6. Verify Checksum (Fail Closed)
echo "Verifying SHA256 checksum..."
EXPECTED_SUM="$(awk -v name="$ASSET_NAME" '($2 == name || $2 == "*" name) { print $1; exit }' "$CHECKSUMS_FILE" 2>/dev/null || true)"
if [ -z "$EXPECTED_SUM" ]; then
    echo "Error: Checksum entry for '${ASSET_NAME}' not found in checksums.txt." >&2
    exit 1
fi

case "$EXPECTED_SUM" in
    *[!0-9a-fA-F]*|???????????????????????????????????????????????????????????????)
        # Ensure it is exactly 64 hex characters (not 63 or fewer, and only hex digits)
        if [ "${#EXPECTED_SUM}" -ne 64 ]; then
            echo "Error: Invalid checksum format in checksums.txt: '$EXPECTED_SUM'." >&2
            exit 1
        fi
        ;;
    *)
        if [ "${#EXPECTED_SUM}" -ne 64 ]; then
            echo "Error: Invalid checksum format in checksums.txt: '$EXPECTED_SUM'." >&2
            exit 1
        fi
        ;;
esac

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SUM="$(sha256sum "$ARCHIVE_FILE" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SUM="$(shasum -a 256 "$ARCHIVE_FILE" | awk '{print $1}')"
else
    echo "Error: Checksum verification tool not found (requires 'sha256sum' or 'shasum -a 256')." >&2
    echo "Cannot verify release integrity. Failing closed." >&2
    exit 1
fi

if [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
    echo "Error: Checksum verification failed for $ASSET_NAME!" >&2
    echo "  Expected: $EXPECTED_SUM" >&2
    echo "  Actual:   $ACTUAL_SUM" >&2
    exit 1
fi
echo "Checksum verified successfully."

# 7. Extract archive
echo "Extracting binary..."
EXTRACT_DIR="${TMP_DIR}/extracted"
mkdir -p "$EXTRACT_DIR"

case "$EXT" in
    tar.gz)
        if ! tar -xzf "$ARCHIVE_FILE" -C "$EXTRACT_DIR"; then
            echo "Error: Failed to extract tar archive." >&2
            exit 1
        fi
        ;;
    zip)
        if command -v unzip >/dev/null 2>&1; then
            if ! unzip -q -o "$ARCHIVE_FILE" -d "$EXTRACT_DIR"; then
                echo "Error: Failed to extract zip archive." >&2
                exit 1
            fi
        else
            echo "Error: 'unzip' command is required to extract the Windows archive." >&2
            exit 1
        fi
        ;;
esac

BIN_NAME="gheppo"
if [ "$OS" = "windows" ]; then
    BIN_NAME="gheppo.exe"
fi

if [ ! -f "${EXTRACT_DIR}/${BIN_NAME}" ]; then
    echo "Error: Binary '${BIN_NAME}' not found inside extracted archive." >&2
    exit 1
fi

# 8. Install binary
mkdir -p "$INSTALL_DIR"
cp -f "${EXTRACT_DIR}/${BIN_NAME}" "$BINARY_PATH"
chmod 755 "$BINARY_PATH"

echo "Installed Gheppo to $BINARY_PATH"

# 9. Configure Shell Integration
DETECTED_SHELL="${SHELL##*/}"
case "$DETECTED_SHELL" in
    zsh)
        SHELL_NAME="zsh"
        SHELL_RC="${HOME}/.zshrc"
        ;;
    bash)
        SHELL_NAME="bash"
        SHELL_RC="${HOME}/.bashrc"
        ;;
    *)
        SHELL_NAME=""
        SHELL_RC=""
        ;;
esac

GHEPPO_START="# >>> gheppo >>>"
GHEPPO_END="# <<< gheppo <<<"

get_shell_integration() {
    _sh="$1"
    # If local scripts exist (e.g. running from source checkout), use them
    _script_dir="$(cd "$(dirname "$0")/shell" 2>/dev/null && pwd || true)"
    if [ -n "$_script_dir" ] && [ -f "$_script_dir/gheppo.${_sh}" ]; then
        cat "$_script_dir/gheppo.${_sh}"
    elif [ "$_sh" = "zsh" ]; then
        cat << 'EOF'
[[ -o interactive ]] || return 0

if [[ -o zle ]]; then
    autoload -Uz add-zle-hook-widget

    _gheppo_once() {
        add-zle-hook-widget -d line-init _gheppo_once
        zle && zle -I
        command gheppo
    }

    add-zle-hook-widget line-init _gheppo_once
else
    command gheppo
fi
EOF
    elif [ "$_sh" = "bash" ]; then
        cat << 'EOF'
_gheppo_once() {
    if declare -p PROMPT_COMMAND 2>/dev/null | grep -q 'declare -a'; then
        PROMPT_COMMAND=("${_GHEPPO_ORIGINAL_PROMPT_COMMAND[@]}")
    else
        PROMPT_COMMAND="${_GHEPPO_ORIGINAL_PROMPT_COMMAND-}"
    fi

    unset _GHEPPO_ORIGINAL_PROMPT_COMMAND
    unset -f _gheppo_once

    command gheppo
}

if declare -p PROMPT_COMMAND 2>/dev/null | grep -q 'declare -a'; then
    _GHEPPO_ORIGINAL_PROMPT_COMMAND=("${PROMPT_COMMAND[@]}")
    PROMPT_COMMAND=("_gheppo_once" "${PROMPT_COMMAND[@]}")
else
    _GHEPPO_ORIGINAL_PROMPT_COMMAND="${PROMPT_COMMAND-}"
    PROMPT_COMMAND="_gheppo_once${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
EOF
    fi
}

if [ -n "$SHELL_RC" ]; then
    touch "$SHELL_RC"
    if ! grep -Fqx "$GHEPPO_START" "$SHELL_RC" 2>/dev/null; then
        {
            printf '\n%s\n' "$GHEPPO_START"
            get_shell_integration "$SHELL_NAME"
            printf '%s\n' "$GHEPPO_END"
        } >> "$SHELL_RC"
        echo "Added Gheppo shell integration to $SHELL_RC"
    else
        echo "Gheppo shell integration is already configured in $SHELL_RC"
    fi
else
    echo
    echo "Could not detect a supported shell (bash or zsh)."
    echo "Gheppo was installed, but shell integration was not configured automatically."
fi

# 10. Check PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*)
        ;;
    *)
        echo
        echo "Note: '$INSTALL_DIR' is not in your current PATH."
        echo "To run gheppo directly, add it to your PATH in $SHELL_RC:"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
        ;;
esac

echo
echo "Gheppo installation complete."
echo
echo "Run:"
echo "  gheppo auth login"
echo "  gheppo sync"
