package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/leetcode"
	"github.com/zalando/go-keyring"
)

func TestAuthLoginLeetCodeWithArgs(t *testing.T) {
	tempDir := t.TempDir()
	restoreCfg := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restoreCfg()

	ts := time.Now().UTC().Unix()
	calendarJSON := fmt.Sprintf(`{"%d": 4}`, ts)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := fmt.Sprintf(`{
			"data": {
				"matchedUser": {
					"username": "lc_auth_user",
					"profile": {"ranking": 42, "reputation": 999},
					"submitStatsGlobal": {
						"acSubmissionNum": [{"difficulty": "All", "count": 150, "submissions": 300}]
					},
					"userCalendar": {
						"activeYears": [2026],
						"streak": 7,
						"totalActiveDays": 50,
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
	rootCmd.SetArgs([]string{"auth", "login", "leetcode", "lc_auth_user"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth login leetcode failed: %v", err)
	}

	if src := config.GetSource(); src != "leetcode" {
		t.Errorf("config.GetSource() = %q, want leetcode", src)
	}
	if u := config.GetLeetCodeUsername(); u != "lc_auth_user" {
		t.Errorf("config.GetLeetCodeUsername() = %q, want lc_auth_user", u)
	}

	summary, ok := cache.LoadForSource(config.SourceLeetCode)
	if !ok || summary.Login != "lc_auth_user" {
		t.Errorf("cache.LoadForSource(leetcode) = (%+v, %v), want lc_auth_user", summary, ok)
	}
}

func TestAuthLoginLeetCodePrompt(t *testing.T) {
	tempDir := t.TempDir()
	restoreCfg := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restoreCfg()

	ts := time.Now().UTC().Unix()
	calendarJSON := fmt.Sprintf(`{"%d": 2}`, ts)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := fmt.Sprintf(`{
			"data": {
				"matchedUser": {
					"username": "lc_prompted_user",
					"profile": {"ranking": 10, "reputation": 100},
					"submitStatsGlobal": {
						"acSubmissionNum": [{"difficulty": "All", "count": 20, "submissions": 30}]
					},
					"userCalendar": {
						"activeYears": [2026],
						"streak": 2,
						"totalActiveDays": 10,
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

	inBuf := strings.NewReader("lc_prompted_user\n")
	outBuf := new(bytes.Buffer)

	rootCmd.SetIn(inBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(outBuf)
	rootCmd.SetArgs([]string{"auth", "login", "leetcode"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth login leetcode prompt failed: %v", err)
	}

	if u := config.GetLeetCodeUsername(); u != "lc_prompted_user" {
		t.Errorf("config.GetLeetCodeUsername() = %q, want lc_prompted_user", u)
	}
}

func TestAuthLoginInvalidSource(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"auth", "login", "unknown_provider"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown source, got nil")
	}
	if !strings.Contains(err.Error(), "unknown source") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuthStatusAndLogout(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_MCP_TOKEN", "")

	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	// 1. GitHub status when logged out
	_ = config.SetSource(config.SourceGitHub)
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"auth", "status"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth status failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Not logged in to github") {
		t.Errorf("unexpected status output: %q", buf.String())
	}

	// 2. GitHub status when logged in
	_ = auth.SaveCredentials(&auth.Credentials{Token: "test-token", Login: "gh_status_user"})
	buf.Reset()
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth status failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Logged in to github as @gh_status_user") {
		t.Errorf("unexpected status output: %q", buf.String())
	}

	// 3. GitHub logout
	buf.Reset()
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth logout failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Logged out of GitHub.") {
		t.Errorf("unexpected logout output: %q", buf.String())
	}

	// 4. LeetCode status & logout
	_ = config.SetSource(config.SourceLeetCode)
	_ = config.SetLeetCodeUsername("lc_status_user")

	buf.Reset()
	rootCmd.SetArgs([]string{"auth", "status"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth status for leetcode failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Active source: leetcode (@lc_status_user)") {
		t.Errorf("unexpected status output: %q", buf.String())
	}

	buf.Reset()
	rootCmd.SetArgs([]string{"auth", "logout"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("auth logout for leetcode failed: %v", err)
	}
	if !strings.Contains(buf.String(), "Logged out of LeetCode.") {
		t.Errorf("unexpected logout output: %q", buf.String())
	}
	if u := config.GetLeetCodeUsername(); u != "" {
		t.Errorf("config.GetLeetCodeUsername() = %q, want empty after logout", u)
	}
}
