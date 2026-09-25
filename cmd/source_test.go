package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/leetcode"
	"github.com/dexisback/gheppo/internal/stats"
)

func TestSourceCommandList(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"source"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("source command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Current source: github") {
		t.Errorf("output missing current source: %q", out)
	}
	if !strings.Contains(out, "github") || !strings.Contains(out, "leetcode") {
		t.Errorf("output missing available sources: %q", out)
	}
}

func TestSourceCommandSetGitHub(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"source", "github"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("source command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Source set to 'github'.") {
		t.Errorf("unexpected output: %q", out)
	}

	if src := config.GetSource(); src != "github" {
		t.Errorf("config.GetSource() = %q, want github", src)
	}
}

func TestSourceCommandSetLeetCode(t *testing.T) {
	// Isolate the cache directory: this test syncs and saves a summary, and
	// without isolation it would overwrite the real user cache.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	tempDir := t.TempDir()
	restoreCfg := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restoreCfg()

	ts := time.Now().UTC().Unix()
	calendarJSON := fmt.Sprintf(`{"%d": 3}`, ts)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := fmt.Sprintf(`{
			"data": {
				"matchedUser": {
					"username": "leet_coder",
					"profile": {"ranking": 100, "reputation": 500},
					"submitStatsGlobal": {
						"acSubmissionNum": [{"difficulty": "All", "count": 250, "submissions": 600}]
					},
					"userCalendar": {
						"activeYears": [2026],
						"streak": 5,
						"totalActiveDays": 40,
						"submissionCalendar": %q
					}
				}
			}
		}`, calendarJSON)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	}))
	defer server.Close()

	restoreLC := leetcode.SetEndpointForTesting(server.URL)
	defer restoreLC()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"source", "leetcode", "leet_coder"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("source command failed: %v", err)
	}

	if src := config.GetSource(); src != "leetcode" {
		t.Errorf("config.GetSource() = %q, want leetcode", src)
	}
	if u := config.GetLeetCodeUsername(); u != "leet_coder" {
		t.Errorf("config.GetLeetCodeUsername() = %q, want leet_coder", u)
	}

	// Verify cached summary exists for leetcode
	loaded, ok := cache.LoadForSource(config.SourceLeetCode)
	if !ok || loaded.Login != "leet_coder" {
		t.Errorf("cache.LoadForSource(leetcode) = (%+v, %v), want leet_coder", loaded, ok)
	}
}

func TestSourceCommandSetInvalid(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"source", "invalid_src"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid source, got nil")
	}
	if !strings.Contains(err.Error(), "unknown source") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSourceSwitchingPreservesData(t *testing.T) {
	// Isolate the cache directory: this test saves fixture summaries for both
	// sources, and without isolation it would overwrite the real user cache.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	// 1. Save data for GitHub
	ghSummary := &stats.Summary{Login: "gh_user", Total: 500}
	if err := cache.SaveForSource(config.SourceGitHub, ghSummary); err != nil {
		t.Fatalf("failed to save github cache: %v", err)
	}
	_ = config.SetSource(config.SourceGitHub)

	// 2. Save data for LeetCode
	lcSummary := &stats.Summary{Login: "lc_user", Total: 300}
	if err := cache.SaveForSource(config.SourceLeetCode, lcSummary); err != nil {
		t.Fatalf("failed to save leetcode cache: %v", err)
	}
	_ = config.SetLeetCodeUsername("lc_user")
	_ = config.SetSource(config.SourceLeetCode)

	// 3. Switch back to GitHub
	_ = config.SetSource(config.SourceGitHub)
	loadedGH, okGH := cache.Load()
	if !okGH || loadedGH.Login != "gh_user" {
		t.Errorf("after switching to github, got (%+v, %v), want gh_user", loadedGH, okGH)
	}

	// 4. Switch back to LeetCode
	_ = config.SetSource(config.SourceLeetCode)
	loadedLC, okLC := cache.Load()
	if !okLC || loadedLC.Login != "lc_user" {
		t.Errorf("after switching to leetcode, got (%+v, %v), want lc_user", loadedLC, okLC)
	}
}
