//this sits in b/w main (which will call into this instead of spawning a raw goroutine) and cache (which alr has the lock)
//this file fixes the bug -- goroutine might die when main() returns 
package refresh

//refresh.go answers "should we refresh rn", process.go answers "how do we actually launch something 
// workflow -> loads the cache, checks if there's stale cache (cache.IsStale()) -> if stale, try cache.TryAcquireRefreshLock() -> if lock acquird, hand off to process.go to spawn a detached child process (NOT A GOROUTINE) -> return immediately either way so as to NOT to block the "fast path"








