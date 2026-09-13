package refresh

import (
	"github.com/dexisback/gheppo/internal/cache"
)

// MaybeRefresh checks whether the cached data is stale and,
// if so, attempts to start a detached background refresh.
//
// This function is deliberately non-blocking because it runs
// on Gheppo's normal shell startup path.
func MaybeRefresh() error {
	summary, ok := cache.Load()

	if !ok || summary == nil {
		return nil
	}

	if !cache.IsStale(summary) {
		return nil
	}

	// Try to acquire the cross-process refresh lock.
	// Only one Gheppo process should start a refresh at a time.
	release, acquired, lockToken := cache.TryAcquireRefreshLockWithToken()

	if !acquired {
		return nil
	}

	// Start the detached background sync process and pass it
	// the token belonging to the lock we just acquired.
	if err := spawnDetachedRefresh(lockToken); err != nil {
		// If the child process could not be started, we still own
		// the lock, so release it immediately.
		release()
		return err
	}

	// Ownership of the lock is intentionally transferred to
	// the detached refresh process.
	return nil
}