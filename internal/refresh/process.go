package refresh

import (
	"os"
	"os/exec"
)

const (
	backgroundRefreshEnv = "GHEPPO_BACKGROUND_REFRESH"
	refreshLockTokenEnv  = "GHEPPO_REFRESH_LOCK_TOKEN"
)

// spawnDetachedRefresh launches a new Gheppo process running `sync`
// in the background.
//
// The child process receives the refresh lock token through an
// environment variable so that it can safely release only the
// lock that belongs to the refresh it started.
func spawnDetachedRefresh(lockToken string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Launch another copy of Gheppo and tell it to run `sync`.
	cmd := exec.Command(executable, "sync")

	// The background refresh must not interact with the terminal.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Preserve the current environment and add information
	// that tells the child process that it is a background refresh.
	cmd.Env = append(
		os.Environ(),
		backgroundRefreshEnv+"=1",
		refreshLockTokenEnv+"="+lockToken,
	)

	// The actual process-detachment settings are platform-specific.
	cmd.SysProcAttr = detachedProcessAttributes()

	// Start the detached process and return immediately.
	return cmd.Start()
}