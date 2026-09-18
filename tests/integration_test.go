// we build the actual gheppo binary, give it an isolated cache directory, and run the actual CLI. We won't hit GitHub.
package tests

import (
	"encoding/json"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testCell struct {
	Date   time.Time `json:"Date"`
	Count  int       `json:"Count"`
	Bucket int       `json:"Bucket"`
	Empty  bool      `json:"Empty"`
}

type testSummary struct {
	Login         string       `json:"Login"`
	Total         int          `json:"Total"`
	CurrentStreak int          `json:"CurrentStreak"`
	LongestStreak int          `json:"LongestStreak"`
	FetchedAt     time.Time    `json:"FetchedAt"`
	Grid          [][]testCell `json:"Grid"`
}

type cachedData struct {
	Summary   *testSummary `json:"summary"`
	FetchedAt time.Time    `json:"fetchedAt"`
}

func buildBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "gheppo")

	cmd := exec.Command("go", "build", "-o", binary, "..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build gheppo : %v\n%s", err, output)

	}
	return binary
}

func writeCache(t *testing.T, cacheHome string, summary *testSummary) {
	t.Helper()
	cacheDir := filepath.Join(cacheHome, "gheppo")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("failed to create cache directory : %v", err)
	}

	data := cachedData{
		Summary:   summary,
		FetchedAt: summary.FetchedAt,
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to encode cache : %v", err)
	}

	cachedPath := filepath.Join(cacheDir, "data.json")
	if err := os.WriteFile(cachedPath, encoded, 0600); err != nil {
		t.Fatalf("failed to write cache : %v", err)

	}
}

func runGheppo(t *testing.T, binary, cacheHome string) string {
	t.Helper()

	cmd := exec.Command(binary)

	cmd.Env = append(
		os.Environ(),
		"XDG_CACHE_HOME="+cacheHome,
		"NO_COLOR=1",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf(
			"gheppo returned error: %v\n%s",
			err,
			output,
		)
	}

	return string(output)
}

// func TestGheppoFirstRun(t *testing.T) {
// 	binary := buildBinary(t)
// 	cacheDir := t.TempDir()

// 	cmd := exec.Command(binary)
// 	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheDir)

// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		t.Fatalf("gheppo returned error: %v\n%s", err, output)
// 	}
// 	got := string(output)

// 	if !strings.Contains(got, "gheppo isnt setup yet") {
// 		t.Fatalf("output does not contain setup instructions : %q", got)
// 	}

// }

func TestGheppoFirstRun(t *testing.T) {
	binary := buildBinary(t)
	cacheDir := t.TempDir()

	configDir := t.TempDir()
	cmd := exec.Command(binary)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"XDG_CACHE_HOME=" + cacheDir,
		"XDG_CONFIG_HOME=" + configDir,
		"DBUS_SESSION_BUS_ADDRESS=/dev/null",
		"XDG_RUNTIME_DIR=" + t.TempDir(),
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf(
			"gheppo returned error: %v\n%s",
			err,
			output,
		)
	}

	got := string(output)

	want := "Gheppo isn't set up yet.\n\nRun:\n  gheppo auth login\n  gheppo sync\n"

	if got != want {
		t.Fatalf(
			"unexpected output:\n%q\nwant:\n%q",
			got,
			want,
		)
	}
}

func TestGheppoLoadsCachedSummary(t *testing.T) {
	binary := buildBinary(t)
	cacheHome := t.TempDir()

	summary := &testSummary{
		Login:         "integration-test",
		Total:         123,
		CurrentStreak: 7,
		LongestStreak: 21,
		FetchedAt:     time.Now(),
		Grid: [][]testCell{
			{
				{
					Date:   time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
					Count:  5,
					Bucket: 4,
				},
				{
					Date:   time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
					Count:  2,
					Bucket: 2,
				},
			},
		},
	}

	writeCache(t, cacheHome, summary)

	output := runGheppo(t, binary, cacheHome)

	// if !strings.Contains(output, "##") {
	// 	t.Fatalf(
	// 		"cached contribution grid was not rendered: %q",
	// 		output,
	// 	)
	// }
	if !strings.Contains(output, "#") && !strings.Contains(output, "·") && !strings.Contains(output, "■") {
		t.Fatal("cached contribution grid was not rendered")
	}

	if !strings.Contains(strings.ToUpper(output), "123 CONTRIBUTIONS") {
		t.Fatalf(
			"cached total was not rendered: %q",
			output,
		)
	}

	// Current streak is no longer shown by default per design spec
	// Check for longest streak instead
	if !strings.Contains(strings.ToUpper(output), "21 DAYS") {
		t.Fatalf(
			"cached longest streak days was not rendered: %q",
			output,
		)
	}

	if !strings.Contains(strings.ToUpper(output), "LONGEST STREAK") {
		t.Fatalf(
			"cached longest streak label was not rendered: %q",
			output,
		)
	}
}

func TestZshIntegrationScript(t *testing.T) {
	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available on this system")
	}

	scriptPath, err := filepath.Abs("../scripts/shell/gheppo.zsh")
	if err != nil {
		t.Fatalf("locating script: %v", err)
	}

	// 1. Verify syntax
	syntaxCmd := exec.Command(zshPath, "-n", scriptPath)
	if out, err := syntaxCmd.CombinedOutput(); err != nil {
		t.Fatalf("syntax check failed for gheppo.zsh: %v\n%s", err, out)
	}

	// 2. Verify non-interactive execution succeeds cleanly (code 0)
	execCmd := exec.Command(zshPath, "-c", "source "+scriptPath)
	if out, err := execCmd.CombinedOutput(); err != nil {
		t.Fatalf("non-interactive execution failed for gheppo.zsh: %v\n%s", err, out)
	}
}

func TestBashIntegrationScript(t *testing.T) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available on this system")
	}

	scriptPath, err := filepath.Abs("../scripts/shell/gheppo.bash")
	if err != nil {
		t.Fatalf("locating script: %v", err)
	}

	// 1. Verify syntax
	syntaxCmd := exec.Command(bashPath, "-n", scriptPath)
	if out, err := syntaxCmd.CombinedOutput(); err != nil {
		t.Fatalf("syntax check failed for gheppo.bash: %v\n%s", err, out)
	}

	// 2. Verify non-interactive execution succeeds cleanly (code 0)
	execCmd := exec.Command(bashPath, "-c", "source "+scriptPath)
	if out, err := execCmd.CombinedOutput(); err != nil {
		t.Fatalf("non-interactive execution failed for gheppo.bash: %v\n%s", err, out)
	}
}

func TestGheppoLeetCodeIntegration(t *testing.T) {
	binary := buildBinary(t)
	cacheHome := t.TempDir()
	configHome := t.TempDir()

	// 1. Write LeetCode configuration
	cfgDir := filepath.Join(configHome, "gheppo")
	_ = os.MkdirAll(cfgDir, 0700)
	cfgJSON := `{"theme":"gruvbox","source":"leetcode","leetcodeUsername":"lc_integration_user"}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(cfgJSON), 0600); err != nil {
		t.Fatalf("writing config.json: %v", err)
	}

	// 2. Write LeetCode cache
	cacheDir := filepath.Join(cacheHome, "gheppo")
	_ = os.MkdirAll(cacheDir, 0700)
	lcSummary := &testSummary{
		Login:         "lc_integration_user",
		Total:         450,
		CurrentStreak: 10,
		LongestStreak: 30,
		FetchedAt:     time.Now(),
		Grid: [][]testCell{
			{
				{
					Date:   time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
					Count:  8,
					Bucket: 4,
				},
			},
		},
	}
	lcData := cachedData{
		Summary:   lcSummary,
		FetchedAt: lcSummary.FetchedAt,
	}
	encoded, _ := json.Marshal(lcData)
	if err := os.WriteFile(filepath.Join(cacheDir, "data_leetcode.json"), encoded, 0600); err != nil {
		t.Fatalf("writing data_leetcode.json: %v", err)
	}

	// 3. Execute gheppo
	cmd := exec.Command(binary)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"XDG_CACHE_HOME=" + cacheHome,
		"XDG_CONFIG_HOME=" + configHome,
		"NO_COLOR=1",
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gheppo returned error: %v\n%s", err, output)
	}

	outStr := string(output)
	if !strings.Contains(outStr, "@lc_integration_user") {
		t.Errorf("output missing LeetCode username: %q", outStr)
	}
	if !strings.Contains(strings.ToUpper(outStr), "450 CONTRIBUTIONS") && !strings.Contains(strings.ToUpper(outStr), "450") {
		t.Errorf("output missing LeetCode total: %q", outStr)
	}
}
