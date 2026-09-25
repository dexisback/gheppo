# Gheppo Architecture & Technical Specification

This document details the internal architecture, concurrency guarantees, and engineering decisions of **Gheppo**.

---

## 1. Core Architectural Invariant

The primary engineering constraint of Gheppo is **terminal startup speed**. 

```
┌─────────────────────────────────────────────────────────┐
│                     Critical Path                       │
│  Shell Launch ──► gheppo ──► Read Cache ──► Render (<5ms) │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼ (If cache > 6h old)
┌─────────────────────────────────────────────────────────┐
│                  Asynchronous Path                      │
│  Acquire File Lock ──► Spawn Detached Sync ──► GitHub   │
└─────────────────────────────────────────────────────────┘
```

- **Zero Blocking I/O**: Network requests, DNS resolution, and remote API calls are strictly forbidden on the startup path.
- **Cache-First Execution**: The binary reads local JSON data and prints formatted ANSI output in <5ms.
- **Fail-Safe Startup**: Refresh failures, network outages, or rate limits never produce errors or block the interactive prompt.

---

## 2. Dependency Direction & Package Boundaries

Gheppo enforces strict unidirectional dependencies using Go `internal/` packages:

```
cmd/ (Cobra CLI routing)
 ├── internal/auth     (OS Keychain credential store)
 ├── internal/cache    (Local storage, atomic writes, file lock)
 ├── internal/refresh  (Detached process lifecycle management)
 ├── internal/render   (Terminal capability detection & ANSI output)
 ├── internal/stats    (Grid alignment, bucketing & streak calculation)
 └── internal/github   (GitHub GraphQL client & schema isolation)
```

Lower-level packages maintain strict isolation:
- `internal/stats` has no awareness of Cobra, rendering, or file systems.
- `internal/render` has no knowledge of GitHub or caching mechanics.
- `internal/github` isolates GraphQL response schemas from domain models.

---

## 3. Data Pipeline & Representation

```
GitHub GraphQL API
       │ (JSON via POST)
       ▼
internal/github (Client.FetchContributionCalendar)
       │ Output: *github.ContributionCalendar
       ▼
internal/stats (stats.Summarize)
       │ Output: *stats.Summary (Grid, Streaks, Buckets)
       ▼
internal/cache (cache.Save)
       │ Persistence: data.json (Atomic rename)
       ▼
internal/render (render.Grid)
       ▼
Terminal stdout (TrueColor / 256color / ASCII)
```

### Domain Data Structures

```go
type Summary struct {
    Login         string
    Total         int
    CurrentStreak int
    LongestStreak int
    FetchedAt     time.Time
    Grid          [][]Cell // [week_index][weekday_index 0..6]
}

type Cell struct {
    Date   time.Time
    Count  int
    Bucket int  // 0: none, 1: low, 2: med-low, 3: med-high, 4: high
    Empty  bool // Padding cell for calendar alignment
}
```

---

## 4. Authentication & Keychain Management (`internal/auth`)

Credentials are never stored in plaintext configuration files or shell dotfiles.

- **Storage**: Backed by `github.com/zalando/go-keyring`, using native platform security:
  - **macOS**: Apple Keychain Services.
  - **Linux**: Secret Service API / DBus (gnome-keyring, KWallet).
  - **Windows**: Windows Credential Manager.
- **Intake**: Personal Access Tokens (PAT) are collected using `golang.org/x/term` with terminal echo disabled.
- **Validation**: On login, the PAT is verified against `GET https://api.github.com/user` to resolve and store the authenticated `login` alongside the token.
- **Error Taxonomy**: Distinct `ErrNoToken` sentinel error allows callers to differentiate between an unauthenticated state and operating system keychain failures.

---

## 5. Persistence & Atomic I/O (`internal/cache`)

The cache directory resides in the platform-standard location:
- Linux: `$XDG_CACHE_HOME/gheppo` (defaults to `~/.cache/gheppo`)
- macOS: `~/Library/Caches/gheppo`
- Windows: `%LocalAppData%\gheppo`

Directory permissions are enforced at `0700` and file permissions at `0600`.

### Atomic Write Pattern

To prevent concurrent shell sessions from reading partially written JSON:
1. Serialize `cachedData` struct to memory.
2. Write payload to temporary file: `data.json.tmp`.
3. Perform atomic filesystem rename: `os.Rename("data.json.tmp", "data.json")`.

### Staleness & Clock Skew

- **Staleness Threshold**: `6 hours`.
- **Clock Skew Protection**: Cache timestamps in the future are treated as fresh to prevent infinite refresh loops caused by local system clock adjustments.

---

## 6. Concurrency & Background Refresh Engine (`internal/refresh`)

When multiple terminal tabs open simultaneously, a process stampede against the GitHub API is prevented using an atomic file lock.

```
Session A ──► IsStale? (Yes) ──► TryAcquireRefreshLock ──► Acquired (Token: 0x4f..) ──► Spawn Detached Sync
Session B ──► IsStale? (Yes) ──► TryAcquireRefreshLock ──► Busy ─────────────────────► Exit Startup Path
Session C ──► IsStale? (Yes) ──► TryAcquireRefreshLock ──► Busy ─────────────────────► Exit Startup Path
```

### 1. Atomic Lock Acquisition
- Lock file: `refresh.lock`.
- Created with `os.O_CREATE | os.O_EXCL | os.O_WRONLY`. The operating system guarantees atomic creation across processes.

### 2. Lock Ownership & Token Validation
- Each lock acquisition generates a cryptographically random 16-byte hex token (`crypto/rand`).
- The token and timestamp are encoded into `refresh.lock`.
- Lock release requires providing the matching token (`ReleaseRefreshLockWithToken`). This prevents a slow or resurrected process from removing a newly acquired lock belonging to another session.

### 3. Stale Lock Recovery
- If `refresh.lock` modification time exceeds **2 minutes**, it is considered orphaned (e.g., parent terminated abruptly) and is reclaimed on the next attempt.

### 4. Process Detachment
- Background sync is executed as a detached child process: `gheppo sync`.
- Dispatched with `GHEPPO_BACKGROUND_REFRESH=1` and `GHEPPO_REFRESH_LOCK_TOKEN=<token>`.
- Standard I/O streams are detached (`Stdin = Stdout = Stderr = nil`).
- **Unix**: Detached session established via `syscall.SysProcAttr{Setsid: true}`.
- **Windows**: Detached via Windows process flags.
- When `gheppo sync` finishes (success or failure), a `defer` hook releases the lock token.

---

## 7. Statistics & Streak Calculation (`internal/stats`)

### Grid Construction & Weekday Alignment
- GitHub returns contribution weeks chronologically.
- `buildGrid` normalizes the first week by determining the weekday of `Days[0]`.
- Empty cells (`Empty: true`) pad the first column so days match their true Sunday–Saturday weekday row.
- Rendering keeps every column a complete 7-box rectangle, matching GitHub's own graph: padding cells and future days of the current week render as "no contribution" cells instead of blanks.

### Contribution Bucketing
Counts are dynamically bucketed into 5 intensity tiers based on the highest contribution count (`maxCount`) in the calendar window:

$$\text{level} = \min\left(4, \max\left(1, \frac{\text{count} \times 4}{\text{maxCount}}\right)\right)$$

### Streak Invariants

Streaks are calculated over chronologically sorted days:

1. **Date Continuity Check**: A streak increments only when $\text{Date}_{i} = \text{Date}_{i-1} + 1\text{ day}$. Calendar gaps reset the running count to zero.
2. **Zero-Count Reset**: Days with 0 contributions reset the active streak.
3. **In-Progress Day Handling**: When evaluating `CurrentStreak`, if the current calendar day has 0 contributions, evaluation starts from yesterday to avoid breaking an active streak before the day concludes.

---

## 8. Terminal Color Negotiation (`internal/render`)

The renderer converts `stats.Summary` into terminal escape sequences using a zero-allocation palette lookup.

### Detection Hierarchy

```
NO_COLOR != "" ───────────────► ASCII Mode ('#')
COLORTERM == "truecolor|24bit" ─► TrueColor (24-bit ANSI RGB)
TERM contains "256color" ───────► 256-Color ANSI
Fallback ───────────────────────► ASCII Mode
```

### Palettes

- **TrueColor**:
  - `0`: `#EBEDF0` / `\033[38;2;235;237;240m■\033[0m`
  - `1`: `#9BE9A8` / `\033[38;2;155;233;168m■\033[0m`
  - `2`: `#40C463` / `\033[38;2;64;196;99m■\033[0m`
  - `3`: `#26A641` / `\033[38;2;38;166;65m■\033[0m`
  - `4`: `#166534` / `\033[38;2;22;101;52m■\033[0m`
- **256-Color**: ANSI codes `238`, `151`, `77`, `71`, `29`.
- **ASCII**: `#` character without escape codes.

---

## 9. Shell Integration Mechanics

### Zsh (Powerlevel10k & Instant Prompt Safe)
Placing output commands directly in `.zshrc` or in `precmd` hooks triggers warnings in Powerlevel10k Instant Prompt because standard file descriptors remain redirected to a temporary buffer until prompt expansion completes. Gheppo hooks into the ZLE `line-init` lifecycle via a self-deregistering widget:

```zsh
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
```

- **`zle-line-init` execution**: Runs when the line editor begins reading user input, guarantees stdout/stderr descriptors have been restored by P10k, and avoids the `[WARNING]: Console output during zsh initialization detected` alert.
- **`zle -I`**: Invalidates ZLE's current display buffer so the prompt is redrawn cleanly underneath the printed heatmap.
- **Non-ZLE / Non-Interactive Fallbacks**: Safely returns in non-interactive subshells and falls back to direct execution if ZLE is disabled.

### Bash (Preserving `PROMPT_COMMAND`)
Bash users may have existing `PROMPT_COMMAND` definitions configured as strings or arrays (Bash 5.1+). Gheppo wraps and restores the exact original definition on first execution:

```bash
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
```

---

## 10. Verification & Quality Assurance

- **Race Detector**: All packages pass `go test -race ./...`.
- **Mock Backends**: Keychain integration tests use `keyring.MockInit()`.
- **Hermetic HTTP Testing**: `internal/github` tests use `net/http/httptest` to validate GraphQL network failures, malformed payloads, rate limits, and schema errors.
- **Integration Test Isolation**: Integration tests build the compiled binary and execute in isolated environments with custom `XDG_CACHE_HOME` temp directories.
