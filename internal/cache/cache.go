package cache 


import (
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


type cachedSummary struct {
	*stats.Summary    //contains the summary and the time it was fetched at 
	FetchedAt time.Time `json:"fetchedAt"`
}




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


//Load reads the cached summary. ok is false when no cache exists yet
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

	var cached cachedSummary

	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	if cached.Summary == nil {
		return nil, false
	}

	cached.Summary.FetchedAt = cached.FetchedAt

	return cached.Summary, true
}


//save automatically writes the summary to the cache . this data is first written to a temperory file and then renamed over the existing cache file. both files live in the same dir, so the rename is ez and atomic
func Save(s *stats.Summary) error {
	if s == nil {
		return os.ErrInvalid
	}

	dir, err := cacheDir()
	if err != nil {
		return err
	}

	cached := cachedSummary{
		Summary:   s,
		FetchedAt: time.Now(),
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


//isStale reports whether the cached data is older than the refresh threshold and returns time by how much
func IsStale(s *stats.Summary) bool {
	if s == nil {
		return true
	}

	return time.Since(s.FetchedAt) > staleAfter
}

//tryacquirerefreshlock attempts to acquire the cross-process refresh lock.
//release must be called by whoever successfully acquired the lock
//NOTE: acquired is false when another process succesfully alr owns the lock
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

	// O_EXCL makes creation atomic across processes.
	file, err := os.OpenFile(
		lockPath,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0600,
	)

	if err != nil {
		return func() {}, false
	}

	file.Close()

	release = func() {
		_ = os.Remove(lockPath)
	}

	return release, true
}


func ReleaseRefreshLock(){
	dir, err := cacheDir()
	if err != nil {
		return
	}
	lockPath := filepath.Join(dir, lockFileName)

	_ = os.Remove(lockPath)
}


