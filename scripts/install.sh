#!/bin/sh

set -eu

REPO="dexisback/gheppo"
DEFAULT_VERSION="0.1.1"
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
# EXIT does the cleanup. The signal traps must *exit*: a trap that only runs
# cleanup lets a POSIX shell resume the script after Ctrl-C, which makes an
# interrupted install look stuck instead of terminating.
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

# 5. Download helper (curl / wget) with timeouts and retries
#
# A bare `curl -fsSL` has no connect/stall timeout and no retries: a stalled
# connection to the GitHub release CDN hangs forever, and one transient TLS
# or network blip (which macOS' LibreSSL-based system curl hits occasionally
# against release-assets.githubusercontent.com) fails the whole install.
DOWNLOAD_RETRIES="${GHEPPO_DOWNLOAD_RETRIES:-3}"
case "$DOWNLOAD_RETRIES" in
    ''|*[!0-9]*) DOWNLOAD_RETRIES=3 ;;
esac

if command -v curl >/dev/null 2>&1; then
    DL_TOOL="curl"
elif command -v wget >/dev/null 2>&1; then
    DL_TOOL="wget"
else
    DL_TOOL=""
fi

download_once() {
    _url="$1"
    _dest="$2"
    case "$DL_TOOL" in
        curl)
            # --connect-timeout bounds TCP/TLS setup; --speed-limit/--speed-time
            # abort a transfer that stalls below 1 KiB/s for 30s instead of
            # hanging indefinitely. </dev/null keeps the tool from eating the
            # rest of the script when installed via `curl ... | sh`.
            curl -fsSL \
                --connect-timeout 15 \
                --speed-limit 1024 --speed-time 30 \
                "$_url" -o "$_dest" </dev/null
            ;;
        wget)
            wget -q --timeout=20 --tries=1 -O "$_dest" "$_url" </dev/null
            ;;
        *)
            echo "Error: Neither 'curl' nor 'wget' was found on your system." >&2
            echo "Please install curl or wget to continue." >&2
            exit 1
            ;;
    esac
}

download_file() {
    _url="$1"
    _dest="$2"
    _attempt=1
    _rc=1
    while [ "$_attempt" -le "$DOWNLOAD_RETRIES" ]; do
        if [ "$_attempt" -gt 1 ]; then
            echo "  Retrying download (attempt $_attempt of $DOWNLOAD_RETRIES)..." >&2
            sleep 2
        fi
        _rc=0
        download_once "$_url" "$_dest" || _rc=$?
        if [ "$_rc" -eq 0 ] && [ -s "$_dest" ]; then
            return 0
        fi
        rm -f "$_dest" >/dev/null 2>&1 || true
        echo "  Download failed (exit code $_rc): $_url" >&2
        _attempt=$((_attempt + 1))
    done
    return 1
}

download_failure_hint() {
    echo >&2
    echo "This is usually a transient network problem or a restrictive network" >&2
    echo "(DNS, proxy, or firewall blocking the GitHub release CDN)." >&2
    echo "Try again, or download the file manually from:" >&2
    echo "  https://github.com/${REPO}/releases/tag/${TAG}" >&2
}

echo "Downloading Gheppo ${TAG} (${OS}/${ARCH})..."
ARCHIVE_FILE="${TMP_DIR}/${ASSET_NAME}"
CHECKSUMS_FILE="${TMP_DIR}/checksums.txt"

if ! download_file "$ARCHIVE_URL" "$ARCHIVE_FILE"; then
    echo "Error: Failed to download release archive after $DOWNLOAD_RETRIES attempts:" >&2
    echo "  $ARCHIVE_URL" >&2
    download_failure_hint
    exit 1
fi

if ! download_file "$CHECKSUMS_URL" "$CHECKSUMS_FILE"; then
    echo "Error: Failed to download checksums after $DOWNLOAD_RETRIES attempts:" >&2
    echo "  $CHECKSUMS_URL" >&2
    download_failure_hint
    exit 1
fi

# 6. Verify Checksum (Fail Closed)
echo "Verifying SHA256 checksum..."
EXPECTED_SUM="$(awk -v name="$ASSET_NAME" '($2 == name || $2 == "*" name) { print $1; exit }' "$CHECKSUMS_FILE" 2>/dev/null || true)"
if [ -z "$EXPECTED_SUM" ]; then
    echo "Error: Checksum entry for '${ASSET_NAME}' not found in checksums.txt." >&2
    exit 1
fi

# Ensure the expected sum is exactly 64 hex characters
case "$EXPECTED_SUM" in
    *[!0-9a-fA-F]*) EXPECTED_SUM_VALID=0 ;;
    *) EXPECTED_SUM_VALID=1 ;;
esac
if [ "$EXPECTED_SUM_VALID" -eq 0 ] || [ "${#EXPECTED_SUM}" -ne 64 ]; then
    echo "Error: Invalid checksum format in checksums.txt: '$EXPECTED_SUM'." >&2
    exit 1
fi

find_powershell() {
    command -v powershell.exe 2>/dev/null || command -v pwsh 2>/dev/null || true
}

PS_EXE="$(find_powershell)"

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SUM="$(sha256sum "$ARCHIVE_FILE" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SUM="$(shasum -a 256 "$ARCHIVE_FILE" | awk '{print $1}')"
elif [ -n "$PS_EXE" ]; then
    # Minimal Git Bash environments without sha256sum/shasum: use PowerShell.
    _win_archive="$(cygpath -w "$ARCHIVE_FILE" 2>/dev/null || echo "$ARCHIVE_FILE")"
    ACTUAL_SUM="$("$PS_EXE" -NoProfile -Command "(Get-FileHash -LiteralPath '$_win_archive' -Algorithm SHA256).Hash.ToLower()" </dev/null | tr -d '\r\n')"
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
        # Git for Windows ships no `unzip`, so fall back through the options
        # that exist on a stock Windows 10+ machine before giving up.
        _unzipped=0
        if command -v unzip >/dev/null 2>&1; then
            if unzip -q -o "$ARCHIVE_FILE" -d "$EXTRACT_DIR"; then
                _unzipped=1
            fi
        fi
        # Windows 10+ includes bsdtar (System32\tar.exe), which extracts zips.
        _win_tar="${SYSTEMROOT:-/c/Windows}/System32/tar.exe"
        if [ "$_unzipped" -eq 0 ] && [ -x "$_win_tar" ]; then
            _win_archive="$(cygpath -w "$ARCHIVE_FILE" 2>/dev/null || echo "$ARCHIVE_FILE")"
            _win_extract="$(cygpath -w "$EXTRACT_DIR" 2>/dev/null || echo "$EXTRACT_DIR")"
            if "$_win_tar" -xf "$_win_archive" -C "$_win_extract" </dev/null; then
                _unzipped=1
            fi
        fi
        # Last resort: PowerShell's Expand-Archive (any modern Windows).
        if [ "$_unzipped" -eq 0 ] && [ -n "$PS_EXE" ]; then
            _win_archive="$(cygpath -w "$ARCHIVE_FILE" 2>/dev/null || echo "$ARCHIVE_FILE")"
            _win_extract="$(cygpath -w "$EXTRACT_DIR" 2>/dev/null || echo "$EXTRACT_DIR")"
            if "$PS_EXE" -NoProfile -Command "Expand-Archive -LiteralPath '$_win_archive' -DestinationPath '$_win_extract' -Force" </dev/null; then
                _unzipped=1
            fi
        fi
        if [ "$_unzipped" -ne 1 ]; then
            echo "Error: Failed to extract zip archive (tried unzip, System32 tar.exe, and PowerShell Expand-Archive)." >&2
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

# 10. Persistent PATH configuration (automatic)
#
# The shell integration above runs `command gheppo`, which only resolves when
# the install directory is on the PATH in every shell that sources the rc
# file. A one-off `export PATH=...` survives only the current session, so the
# installer writes a managed PATH entry automatically. The current session's
# PATH is irrelevant here; only the rc file decides what future shells see.
#
#   GHEPPO_ADD_PATH=0  -> opt out and print the manual steps instead

GHEPPO_PATH_START="# >>> gheppo PATH >>>"
GHEPPO_PATH_END="# <<< gheppo PATH <<<"

install_dir_on_path() {
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) return 0 ;;
        *) return 1 ;;
    esac
}

path_entry_in_rc() {
    _path_rc="$1"
    [ -n "$_path_rc" ] || return 1
    [ -f "$_path_rc" ] || return 1
    # Match both an absolute reference to the install directory and, for the
    # default location, the portable "$HOME/.local/bin" / "~/.local/bin" forms.
    if grep -Fq -- "$INSTALL_DIR" "$_path_rc" 2>/dev/null; then
        return 0
    fi
    if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
        grep -Fq -- ".local/bin" "$_path_rc" 2>/dev/null
    else
        return 1
    fi
}

add_path_block() {
    _path_rc="$1"
    if [ -z "$_path_rc" ]; then
        return 1
    fi
    if [ ! -f "$_path_rc" ] || ! grep -Fqx "$GHEPPO_PATH_START" "$_path_rc" 2>/dev/null; then
        # Requirement 8: if the directory is already present anywhere in the rc
        # file (user's own PATH line, pipx-style config, ...), don't add a
        # duplicate PATH entry.
        if path_entry_in_rc "$_path_rc"; then
            echo "'$INSTALL_DIR' is already present in $_path_rc; not adding a duplicate PATH entry."
            return 0
        fi
        {
            printf '\n%s\n' "$GHEPPO_PATH_START"
            printf 'export PATH="%s:$PATH"\n' "$INSTALL_DIR"
            printf '%s\n' "$GHEPPO_PATH_END"
        } >> "$_path_rc" || return 1
    fi
    echo "Added persistent PATH entry to $_path_rc"
    return 0
}

add_windows_user_path() {
    # Git Bash sessions source ~/.bashrc (the POSIX block above covers them);
    # this additionally makes gheppo.exe resolvable from PowerShell and cmd.
    _win_dir="$(cygpath -w "$INSTALL_DIR" 2>/dev/null || true)"
    if [ -z "$_win_dir" ]; then
        echo "Skipping Windows user PATH update: could not resolve the Windows form of '$INSTALL_DIR'."
        return 1
    fi
    _ps="$(find_powershell)"
    if [ -z "$_ps" ]; then
        echo "Skipping Windows user PATH update: PowerShell not found."
        return 1
    fi
    if "$_ps" -NoProfile -Command "\$k = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', \$true); if (-not \$k) { exit 1 }; \$o = [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames; try { \$kind = \$k.GetValueKind('Path') } catch { \$kind = [Microsoft.Win32.RegistryValueKind]::String }; \$raw = \$k.GetValue('Path', '', \$o); \$exp = \$k.GetValue('Path', ''); if ((\$raw -and \$raw.Split(';') -contains '$_win_dir') -or (\$exp -and \$exp.Split(';') -contains '$_win_dir')) { exit 0 }; if (\$raw) { \$new = \$raw.TrimEnd(';') + ';$_win_dir' } else { \$new = '$_win_dir' }; \$k.SetValue('Path', \$new, \$kind); \$k.Close()"; then
        echo "Added '$_win_dir' to the Windows user PATH (new PowerShell/cmd windows)."
        return 0
    fi
    echo "Could not update the Windows user PATH automatically."
    return 1
}

print_manual_path_steps() {
    echo
    echo "'$INSTALL_DIR' is not on your PATH."
    echo "Add it to your shell rc file to render the card in every new terminal:"
    echo
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc   # zsh"
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc   # bash"
    if [ "$OS" = "windows" ]; then
        echo
        echo "For PowerShell/cmd, add the Windows form of this directory to your"
        echo "user PATH via System Properties > Environment Variables."
    fi
    echo
    echo "You can re-run the installer to configure the PATH automatically"
    echo "(GHEPPO_ADD_PATH=0 skips PATH setup)."
}

PATH_RC="$SHELL_RC"
if [ "$OS" = "windows" ] && [ -z "$PATH_RC" ]; then
    PATH_RC="${HOME}/.bashrc"
fi

PATH_DECISION=""
case "${GHEPPO_ADD_PATH-}" in
    1|true|True|TRUE|yes|Yes|YES) PATH_DECISION="add" ;;
    0|false|False|FALSE|no|No|NO) PATH_DECISION="declined" ;;
esac

if [ -z "$PATH_DECISION" ] && [ -z "$PATH_RC" ] && [ "$OS" != "windows" ]; then
    PATH_DECISION="declined"
fi

# The rc file, not the current session's PATH, decides what future shells see.
if [ -z "$PATH_DECISION" ] && [ -n "$PATH_RC" ] && path_entry_in_rc "$PATH_RC"; then
    PATH_DECISION="already"
fi

if [ -z "$PATH_DECISION" ]; then
    PATH_DECISION="add"
fi

case "$PATH_DECISION" in
    add)
        path_configured=""
        if [ -n "$PATH_RC" ]; then
            if add_path_block "$PATH_RC"; then
                path_configured=1
            fi
        fi
        if [ "$OS" = "windows" ]; then
            if add_windows_user_path; then
                path_configured=1
            fi
        fi
        if [ -z "$path_configured" ]; then
            print_manual_path_steps
        elif [ -n "$PATH_RC" ]; then
            echo
            echo "PATH configured: '$INSTALL_DIR' will be on the PATH in new terminals."
            echo "Open a new terminal, or run 'source $PATH_RC' to update the current one."
        fi
        if ! install_dir_on_path; then
            echo
            echo "For the current terminal session, run:"
            echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        fi
        ;;
    declined)
        print_manual_path_steps
        ;;
    already)
        ;;
esac

echo
echo "Gheppo installation complete."
echo
echo "Run:"
echo "  gheppo auth login"
echo "  gheppo sync"
echo
echo "Open a new terminal window to see your contribution card."
