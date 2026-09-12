package refresh
//process.go is the mechanism layer: `refresh.go` decides whether to refresh; this file only knows HOW to launch the refresh safely as a detached child process

import(
	"os"
	"os/exec"
)
const backgroundRefreshEnv = "GHEPPO_BACKGROUND_REFRESH"


func spawnDetachedRefresh() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, "sync")

	// The refresh must be completely silent.
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	cmd.Env = append(os.Environ(),  backgroundRefreshEnv+"=1")   //this means the child gets : GHEPPO_BACKGROUND_REFRESH=1   , while a normal `gheppo sync` doesnt

	cmd.SysProcAttr = detachedProcessAttributes()

	return cmd.Start()
}

//SysProcAttr is platform specific, so we're not putting both implementations in this file, rather splitting into process_unix and process_windows (fuck windows)


