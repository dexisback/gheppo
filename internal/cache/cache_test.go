package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

// setupTestCache redirects the cache package to a temporary directory.
//
// This keeps tests completely isolated from the user's real Gheppo cache.
func setupTestCache(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	originalCacheDirectory := cacheDirectory

	cacheDirectory = func() (string, error) {
		return dir, nil
	}

	t.Cleanup(func() {
		cacheDirectory = originalCacheDirectory
	})

	return dir
}

// testSummary creates a small predictable summary that can be
// used by the cache tests.
func testSummary() *stats.Summary {
	return &stats.Summary{
		Login:         "test-user",
		Total:         42,
		CurrentStreak: 5,
		LongestStreak: 10,
		Grid: [][]stats.Cell{
			{
				{
					Date:  time.Date(2026, 9, 13, 0, 0, 0, 0, time.Local),
					Count: 3,
					Bucket: 2,
				},
			},
		},
	}
}

// TestSaveAndLoad verifies that a summary written to the cache
// can be loaded back with the same important data.
func TestSaveAndLoad(t *testing.T) {
	setupTestCache(t)

	original := testSummary()

	if err := Save(original); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	loaded, ok := Load()

	if !ok {
		t.Fatal("Load() returned false after Save()")
	}

	if loaded == nil {
		t.Fatal("Load() returned nil summary")
	}

	if loaded.Login != original.Login {
		t.Errorf(
			"Login = %q, want %q",
			loaded.Login,
			original.Login,
		)
	}

	if loaded.Total != original.Total {
		t.Errorf(
			"Total = %d, want %d",
			loaded.Total,
			original.Total,
		)
	}

	if loaded.CurrentStreak != original.CurrentStreak {
		t.Errorf(
			"CurrentStreak = %d, want %d",
			loaded.CurrentStreak,
			original.CurrentStreak,
		)
	}

	if loaded.LongestStreak != original.LongestStreak {
		t.Errorf(
			"LongestStreak = %d, want %d",
			loaded.LongestStreak,
			original.LongestStreak,
		)
	}

	if loaded.FetchedAt.IsZero() {
		t.Fatal("Loaded summary has zero FetchedAt")
	}
}

// TestLoadMissing verifies that Load() safely reports that no cache
// exists when the cache file has not been created yet.
func TestLoadMissing(t *testing.T) {
	setupTestCache(t)

	loaded, ok := Load()

	if ok {
		t.Fatal("Load() returned true when cache does not exist")
	}

	if loaded != nil {
		t.Fatal("Load() returned a summary when cache does not exist")
	}
}

// TestLoadCorrupt verifies that invalid JSON does not cause the
// cache package to panic or return a valid summary.
func TestLoadCorrupt(t *testing.T) {
	dir := setupTestCache(t)

	cachePath := filepath.Join(dir, cacheFileName)

	if err := os.WriteFile(
		cachePath,
		[]byte("this is not valid JSON"),
		0600,
	); err != nil {
		t.Fatalf("failed to create corrupt cache: %v", err)
	}

	loaded, ok := Load()

	if ok {
		t.Fatal("Load() returned true for corrupt cache")
	}

	if loaded != nil {
		t.Fatal("Load() returned a summary for corrupt cache")
	}
}

// TestSaveNil verifies that Save() rejects a nil summary.
func TestSaveNil(t *testing.T) {
	setupTestCache(t)

	if err := Save(nil); err == nil {
		t.Fatal("Save(nil) returned nil error")
	}
}

// TestIsStale verifies the refresh threshold.
//
// Data newer than six hours should be considered fresh.
// Data older than six hours should be considered stale.
func TestIsStale(t *testing.T) {
	fresh := testSummary()
	fresh.FetchedAt = time.Now().Add(-1 * time.Hour)

	if IsStale(fresh) {
		t.Fatal("IsStale() returned true for fresh cache")
	}

	stale := testSummary()
	stale.FetchedAt = time.Now().Add(-7 * time.Hour)

	if !IsStale(stale) {
		t.Fatal("IsStale() returned false for stale cache")
	}
}

// TestIsStaleNil verifies that a nil summary is always considered stale.
func TestIsStaleNil(t *testing.T) {
	if !IsStale(nil) {
		t.Fatal("IsStale(nil) returned false")
	}
}

// TestRefreshLock verifies the basic cross-process lock behavior.
//
// The first acquisition should succeed.
// A second acquisition should fail while the lock exists.
// After releasing the first lock, another acquisition should succeed.
func TestRefreshLock(t *testing.T) {
	setupTestCache(t)

	releaseFirst, acquiredFirst := TryAcquireRefreshLock()

	if !acquiredFirst {
		t.Fatal("first lock acquisition failed")
	}

	defer releaseFirst()

	_, acquiredSecond := TryAcquireRefreshLock()

	if acquiredSecond {
		t.Fatal("second lock acquisition succeeded while lock was already held")
	}

	releaseFirst()

	releaseThird, acquiredThird := TryAcquireRefreshLock()

	if !acquiredThird {
		t.Fatal("lock acquisition failed after previous lock was released")
	}

	releaseThird()
}

// TestRefreshLockOwnership verifies that a process cannot remove
// a lock belonging to another process.
//
// This is important when stale-lock recovery causes a new process
// to acquire a lock while an older process is still finishing.
func TestRefreshLockOwnership(t *testing.T) {
	setupTestCache(t)

	_, acquired, firstToken := TryAcquireRefreshLockWithToken()

	if !acquired {
		t.Fatal("first lock acquisition failed")
	}

	// Attempt to release the lock with the wrong token.
	ReleaseRefreshLockWithToken("wrong-token")

	// The lock should still exist because the token was incorrect.
	_, acquiredSecond := TryAcquireRefreshLock()

	if acquiredSecond {
		t.Fatal("lock was incorrectly released by the wrong token")
	}

	// Release using the correct token.
	ReleaseRefreshLockWithToken(firstToken)

	// The lock should now be available.
	releaseThird, acquiredThird := TryAcquireRefreshLock()

	if !acquiredThird {
		t.Fatal("lock was not released by the correct token")
	}

	releaseThird()
}

// TestRefreshLockToken verifies that every successful lock acquisition
// receives a non-empty ownership token.
func TestRefreshLockToken(t *testing.T) {
	setupTestCache(t)

	release, acquired, token := TryAcquireRefreshLockWithToken()

	if !acquired {
		t.Fatal("lock acquisition failed")
	}

	defer release()

	if token == "" {
		t.Fatal("successful lock acquisition returned an empty token")
	}
}




func TestIsStaleFuture(t *testing.T) {
	future := testSummary()
	future.FetchedAt = time.Now().Add(1 * time.Hour)

	if IsStale(future) {
		t.Fatal("IsStale() returned true for future cache")
	}
}


func TestSaveReplacesExistingCache(t *testing.T) {
	setupTestCache(t)

	first := testSummary()
	first.Total = 42

	if err := Save(first); err != nil {
		t.Fatalf("first Save() returned error: %v", err)
	}

	second := testSummary()
	second.Total = 99

	if err := Save(second); err != nil {
		t.Fatalf("second Save() returned error: %v", err)
	}

	loaded, ok := Load()
	if !ok {
		t.Fatal("Load() returned false after second Save()")
	}

	if loaded.Total != 99 {
		t.Fatalf("Total = %d, want 99", loaded.Total)
	}
}


func TestLoadNilSummary(t *testing.T) {
	dir := setupTestCache(t)

	cachePath := filepath.Join(dir, cacheFileName)

	data := []byte(`{
		"summary": null,
		"fetchedAt": "2026-09-13T00:00:00Z"
	}`)

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	loaded, ok := Load()

	if ok {
		t.Fatal("Load() returned true for nil summary")
	}

	if loaded != nil {
		t.Fatal("Load() returned a summary for nil summary")
	}
}


//stale lock recovery:
func TestRefreshLockRecoversStaleLock(t *testing.T) {
	dir := setupTestCache(t)

	lockPath := filepath.Join(dir, lockFileName)

	if err := os.WriteFile(
		lockPath,
		[]byte(`{"token":"old-token","createdAt":"2026-09-12T00:00:00Z"}`),
		0600,
	); err != nil {
		t.Fatalf("failed to create stale lock: %v", err)
	}

	staleTime := time.Now().Add(-3 * time.Minute)

	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatalf("failed to make lock stale: %v", err)
	}

	release, acquired, token := TryAcquireRefreshLockWithToken()

	if !acquired {
		t.Fatal("failed to acquire lock after stale lock recovery")
	}

	if token == "" {
		t.Fatal("stale lock recovery returned empty token")
	}

	release()
}
