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

func TestGheppoFirstRun(t *testing.T) {
	binary := buildBinary(t)
	cacheDir := t.TempDir()

	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cacheDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gheppo returned error: %v\n%s", err, output)
	}
	got := string(output)

	if !strings.Contains(got, "gheppo isnt setup yet") {
		t.Fatalf("output does not contain setup instructions : %q", got)
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

	if !strings.Contains(output, "##") {
		t.Fatalf(
			"cached contribution grid was not rendered: %q",
			output,
		)
	}

	if !strings.Contains(output, "123 contributions") {
		t.Fatalf(
			"cached total was not rendered: %q",
			output,
		)
	}

	if !strings.Contains(output, "7 day streak") {
		t.Fatalf(
			"cached current streak was not rendered: %q",
			output,
		)
	}

	if !strings.Contains(output, "21 longest streak") {
		t.Fatalf(
			"cached longest streak was not rendered: %q",
			output,
		)
	}
}

//we're testing go build > actual gheppo binary > actual root command > actual cache.Load() > no cache > actual first run message
