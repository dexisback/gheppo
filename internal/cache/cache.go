package cache

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

const (
	staleAfter = 6 * time.Hour
	lockAfter  = 2 * time.Minute

	cacheDirName  = "gheppo"
	cacheFileName = "data.json"
	lockFileName  = "refresh.lock"
)

// cachedData contains the cached summary and the time it was fetched at.
// We keep these as explicit fields instead of embedding Summary,
// so the JSON structure is clear and unambiguous.
type cachedData struct {
	Summary   *stats.Summary `json:"summary"`
	FetchedAt time.Time      `json:"fetchedAt"`
}

// refreshLock represents ownership of the refresh lock.
// Each process that successfully acquires the lock gets a unique token.
// This prevents an old process from accidentally deleting a newer process's lock.
type refreshLock struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
}

// cacheDir returns Gheppo's cache directory, creating it if necessary.
func cacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir = filepath.Join(dir, cacheDirName)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return dir, nil
}

// Load reads the cached summary. ok is false when no valid cache exists yet.
func Load() (*stats.Summary, bool) {
	dir, err := cacheDir()
	if err != nil {
		return nil, false
	}

	path := filepath.Join(dir, cacheFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}

		return nil, false
	}

	var cached cachedData

	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	if cached.Summary == nil {
		return nil, false
	}

	// Restore the cache's fetched timestamp into the summary.
	cached.Summary.FetchedAt = cached.FetchedAt

	return cached.Summary, true
}

// Save automatically writes the summary to the cache.
// This data is first written to a temporary file and then renamed
// over the existing cache file. Both files live in the same dir,
// so the rename is atomic.
func Save(s *stats.Summary) error {
	if s == nil {
		return os.ErrInvalid
	}

	dir, err := cacheDir()
	if err != nil {
		return err
	}

	fetchedAt := time.Now()

	// Keep the in-memory summary and the cache timestamp in sync.
	s.FetchedAt = fetchedAt

	cached := cachedData{
		Summary:   s,
		FetchedAt: fetchedAt,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	tmpPath := filepath.Join(dir, cacheFileName+".tmp")
	finalPath := filepath.Join(dir, cacheFileName)

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}

// IsStale reports whether the cached data is older than the refresh threshold.
func IsStale(s *stats.Summary) bool {
	if s == nil {
		return true
	}

	return time.Since(s.FetchedAt) > staleAfter
}

// TryAcquireRefreshLock attempts to acquire the cross-process refresh lock.
//
// release must be called by whoever successfully acquired the lock.
//
// acquired is false when another process successfully already owns the lock.
//
// Each successful lock acquisition gets a unique token so that only
// the process that owns the lock can release it.
func TryAcquireRefreshLock() (release func(), acquired bool) {
	dir, err := cacheDir()
	if err != nil {
		return func() {}, false
	}

	lockPath := filepath.Join(dir, lockFileName)

	// First check whether an existing lock is stale.
	if info, err := os.Stat(lockPath); err == nil {
		if time.Since(info.ModTime()) > lockAfter {
			_ = os.Remove(lockPath)
		}
	}

	// Generate a unique token that identifies this lock owner.
	token, err := generateLockToken()
	if err != nil {
		return func() {}, false
	}

	lock := refreshLock{
		Token:     token,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(lock)
	if err != nil {
		return func() {}, false
	}

	// O_EXCL makes creation atomic across processes.
	// Only one process can successfully create the lock file.
	file, err := os.OpenFile(
		lockPath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0600,
	)

	if err != nil {
		return func() {}, false
	}

	// Write this process's ownership information into the lock.
	if _, err := file.Write(data); err != nil {
		file.Close()
		_ = os.Remove(lockPath)
		return func() {}, false
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(lockPath)
		return func() {}, false
	}

	// release only removes the lock if this process still owns it.
	release = func() {
		data, err := os.ReadFile(lockPath)
		if err != nil {
			return
		}

		var current refreshLock

		if err := json.Unmarshal(data, &current); err != nil {
			return
		}

		// If the token doesn't match, another process owns this lock now.
		// Do not remove it.
		if current.Token != token {
			return
		}

		_ = os.Remove(lockPath)
	}

	return release, true
}

// ReleaseRefreshLock removes the refresh lock.
//
// This is used by the detached background sync process.
func ReleaseRefreshLock() {
	dir, err := cacheDir()
	if err != nil {
		return
	}

	lockPath := filepath.Join(dir, lockFileName)

	_ = os.Remove(lockPath)
}

// generateLockToken creates a random token used to identify
// the process that successfully acquired the refresh lock.
func generateLockToken() (string, error) {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}