package refresh

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/stats"
)

func setupTestCache(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	return filepath.Join(dir, "gheppo")
}

func writeTestCache(t *testing.T, summary *stats.Summary) {
	t.Helper()

	cacheDir := setupTestCache(t)

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}

	data := struct {
		Summary   *stats.Summary `json:"summary"`
		FetchedAt time.Time      `json:"fetchedAt"`
	}{
		Summary:   summary,
		FetchedAt: summary.FetchedAt,
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to encode test cache: %v", err)
	}

	cachePath := filepath.Join(cacheDir, "data.json")

	if err := os.WriteFile(cachePath, encoded, 0600); err != nil {
		t.Fatalf("failed to write test cache: %v", err)
	}
}

func testSummary(fetchedAt time.Time) *stats.Summary {
	return &stats.Summary{
		Login:         "test-user",
		Total:         42,
		CurrentStreak: 5,
		LongestStreak: 10,
		FetchedAt:     fetchedAt,
		Grid: [][]stats.Cell{
			{
				{
					Date:   time.Now(),
					Count:  3,
					Bucket: 2,
				},
			},
		},
	}
}

func TestMaybeRefreshFreshCache(t *testing.T) {
	setupTestCache(t)

	summary := testSummary(time.Now().Add(-1 * time.Hour))

	if err := cache.Save(summary); err != nil {
		t.Fatalf("cache.Save() returned error: %v", err)
	}

	originalSpawn := spawnRefresh

	spawnCalled := false

	spawnRefresh = func(lockToken string) error {
		spawnCalled = true
		return nil
	}

	t.Cleanup(func() {
		spawnRefresh = originalSpawn
	})

	if err := MaybeRefresh(); err != nil {
		t.Fatalf("MaybeRefresh() returned error: %v", err)
	}

	if spawnCalled {
		t.Fatal("background refresh was started for fresh cache")
	}
}

func TestMaybeRefreshStaleCache(t *testing.T) {
	setupTestCache(t)

	summary := testSummary(time.Now().Add(-7 * time.Hour))

	writeTestCache(t, summary)

	originalSpawn := spawnRefresh

	var receivedToken string

	spawnRefresh = func(lockToken string) error {
		receivedToken = lockToken
		return nil
	}

	t.Cleanup(func() {
		spawnRefresh = originalSpawn

		if receivedToken != "" {
			cache.ReleaseRefreshLockWithToken(receivedToken)
		}
	})

	if err := MaybeRefresh(); err != nil {
		t.Fatalf("MaybeRefresh() returned error: %v", err)
	}

	if receivedToken == "" {
		t.Fatal("background refresh was not started for stale cache")
	}
}

func TestMaybeRefreshExistingLock(t *testing.T) {
	setupTestCache(t)

	summary := testSummary(time.Now().Add(-7 * time.Hour))

	writeTestCache(t, summary)

	release, acquired, token := cache.TryAcquireRefreshLockWithToken()

	if !acquired {
		t.Fatal("failed to acquire test refresh lock")
	}

	defer release()

	if token == "" {
		t.Fatal("test lock returned an empty token")
	}

	originalSpawn := spawnRefresh

	spawnCalled := false

	spawnRefresh = func(lockToken string) error {
		spawnCalled = true
		return nil
	}

	t.Cleanup(func() {
		spawnRefresh = originalSpawn
	})

	if err := MaybeRefresh(); err != nil {
		t.Fatalf("MaybeRefresh() returned error: %v", err)
	}

	if spawnCalled {
		t.Fatal("background refresh started while another process owned the lock")
	}
}

func TestMaybeRefreshSpawnFailure(t *testing.T) {
	setupTestCache(t)

	summary := testSummary(time.Now().Add(-7 * time.Hour))

	writeTestCache(t, summary)

	originalSpawn := spawnRefresh

	expectedErr := errors.New("failed to start background process")

	var receivedToken string

	spawnRefresh = func(lockToken string) error {
		receivedToken = lockToken
		return expectedErr
	}

	t.Cleanup(func() {
		spawnRefresh = originalSpawn
	})

	err := MaybeRefresh()

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"MaybeRefresh() error = %v, want %v",
			err,
			expectedErr,
		)
	}

	if receivedToken == "" {
		t.Fatal("background refresh was not attempted")
	}

	_, acquired := cache.TryAcquireRefreshLock()

	if !acquired {
		t.Fatal("refresh lock was not released after spawn failure")
	}
}

func TestMaybeRefreshTransfersLockOwnership(t *testing.T) {
	setupTestCache(t)

	summary := testSummary(time.Now().Add(-7 * time.Hour))

	writeTestCache(t, summary)

	originalSpawn := spawnRefresh

	var receivedToken string

	spawnRefresh = func(lockToken string) error {
		receivedToken = lockToken
		return nil
	}

	t.Cleanup(func() {
		spawnRefresh = originalSpawn

		if receivedToken != "" {
			cache.ReleaseRefreshLockWithToken(receivedToken)
		}
	})

	if err := MaybeRefresh(); err != nil {
		t.Fatalf("MaybeRefresh() returned error: %v", err)
	}

	if receivedToken == "" {
		t.Fatal("background refresh did not receive lock token")
	}

	_, acquired := cache.TryAcquireRefreshLock()

	if acquired {
		t.Fatal("lock was released even though refresh process started successfully")
	}

	cache.ReleaseRefreshLockWithToken(receivedToken)

	release, acquired := cache.TryAcquireRefreshLock()

	if !acquired {
		t.Fatal("lock was not available after background refresh released it")
	}

	release()
}
