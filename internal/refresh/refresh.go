//this sits in b/w main (which will call into this instead of spawning a raw goroutine) and cache (which alr has the lock)
//this file fixes the bug -- goroutine might die when main() returns 
package refresh

//refresh.go answers "should we refresh rn", process.go answers "how do we actually launch something 
// workflow -> loads the cache, checks if there's stale cache (cache.IsStale()) -> if stale, try cache.TryAcquireRefreshLock() -> if lock acquird, hand off to process.go to spawn a detached child process (NOT A GOROUTINE) -> return immediately either way so as to NOT to block the "fast path"








import (
	"github.com/dexisback/gheppo/internal/cache"
)


//mayberefresh checks whether the cache needs refreshing and if it needs -> start a detached background refreshing process
func MaybeRefresh() error {
	summary, ok := cache.Load()

	if !ok || summary == nil {
		return nil
	}

	if !cache.IsStale(summary) {
		return nil
	}

	release, acquired := cache.TryAcquireRefreshLock()

	if !acquired {
		return nil
	}

	if err := spawnDetachedRefresh(); err != nil {
		release()
		return err
	}

	// Ownership of the lock is intentionally transferred to the
	// detached refresh process.
	return nil
}

//note: Do not call release() after successful spawning.
// Otherwise the whole point of the cross-process lock disappears: another terminal could immediately start another refresh while the detached child is still fetching.


