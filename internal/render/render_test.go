package render

import (
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

// TestDetectColorMode verifies that terminal environment variables
// are mapped to the correct rendering mode.
func TestDetectColorMode(t *testing.T) {
	tests := []struct {
		name      string
		noColor   string
		colorTerm string
		term      string
		want      ColorMode
	}{
		{
			name:      "NO_COLOR",
			noColor:   "1",
			colorTerm: "truecolor",
			term:      "xterm-256color",
			want:      ColorASCII,
		},
		{
			name:      "truecolor",
			colorTerm: "truecolor",
			term:      "xterm-256color",
			want:      ColorTrueColor,
		},
		{
			name:      "24bit",
			colorTerm: "24bit",
			term:      "xterm-256color",
			want:      ColorTrueColor,
		},
		{
			name: "256 color",
			term: "xterm-256color",
			want: Color256,
		},
		{
			name: "unknown terminal",
			term: "dumb",
			want: ColorASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("COLORTERM", tt.colorTerm)
			t.Setenv("TERM", tt.term)

			got := DetectColorMode()

			if got != tt.want {
				t.Errorf(
					"DetectColorMode() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

// TestGridNilSummary verifies that rendering a nil summary
// is handled safely.
func TestGridNilSummary(t *testing.T) {
	output := Grid(nil)

	if output != "" {
		t.Errorf(
			"Grid(nil) = %q, want empty output",
			output,
		)
	}
}

// TestGridEmptySummary verifies that rendering an empty summary
// does not panic and still produces the summary line.
func TestGridEmptySummary(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	summary := &stats.Summary{}

	output := Grid(summary)

	if output == "" {
		t.Fatal("Grid(empty summary) returned empty output")
	}

	want := "0 contributions · 0 day streak · 0 longest streak"

	if !strings.Contains(output, want) {
		t.Errorf(
			"Grid(empty summary) = %q, want summary %q",
			output,
			want,
		)
	}
}

// TestGridContainsContributionData verifies that real cells and
// summary information appear in the rendered output.
func TestGridContainsContributionData(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "dumb")

	summary := &stats.Summary{
		Login:         "test-user",
		Total:         42,
		CurrentStreak: 5,
		LongestStreak: 10,
		Grid: [][]stats.Cell{
			{
				{
					Date:   time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
					Count:  0,
					Bucket: 0,
				},
				{
					Date:   time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
					Count:  5,
					Bucket: 2,
				},
			},
		},
	}

	output := Grid(summary)

	if output == "" {
		t.Fatal("Grid() returned empty output")
	}

	if !strings.Contains(output, "##") {
		t.Errorf(
			"rendered grid does not contain expected ASCII cells: %q",
			output,
		)
	}

	if !strings.Contains(output, "42 contributions") {
		t.Errorf(
			"rendered output does not contain total contribution count: %q",
			output,
		)
	}

	if !strings.Contains(output, "5 day streak") {
		t.Errorf(
			"rendered output does not contain current streak: %q",
			output,
		)
	}

	if !strings.Contains(output, "10 longest streak") {
		t.Errorf(
			"rendered output does not contain longest streak: %q",
			output,
		)
	}
}

// TestGridEmptyCells verifies that cells marked as Empty are rendered
// as spaces rather than contribution blocks.
func TestGridEmptyCells(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "dumb")

	summary := &stats.Summary{
		Total: 1,
		Grid: [][]stats.Cell{
			{
				{
					Empty: true,
				},
				{
					Date:   time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
					Count:  1,
					Bucket: 1,
				},
			},
		},
	}

	output := Grid(summary)

	if output == "" {
		t.Fatal("Grid() returned empty output")
	}

	// The first character should be the space representing
	// the empty cell.
	if output[0] != ' ' {
		t.Errorf(
			"first rendered character = %q, want space",
			output[0],
		)
	}
}

// TestCellColorASCII verifies the ASCII fallback representation.
func TestCellColorASCII(t *testing.T) {
	for bucket := 0; bucket <= 4; bucket++ {
		got := cellColor(ColorASCII, bucket)

		if got != "#" {
			t.Errorf(
				"cellColor(ColorASCII, %d) = %q, want %q",
				bucket,
				got,
				"#",
			)
		}
	}
}

// TestTrueColor verifies that each supported intensity bucket
// produces a truecolor ANSI escape sequence.
func TestTrueColor(t *testing.T) {
	for bucket := 0; bucket <= 4; bucket++ {
		got := trueColor(bucket)

		if got == "" {
			t.Errorf(
				"trueColor(%d) returned empty string",
				bucket,
			)
		}

		if !strings.Contains(got, "\033[38;2;") {
			t.Errorf(
				"trueColor(%d) = %q, want truecolor ANSI sequence",
				bucket,
				got,
			)
		}
	}
}

// TestTrueColorInvalidBucket verifies that unsupported buckets
// fall back to the zero-contribution color.
func TestTrueColorInvalidBucket(t *testing.T) {
	got := trueColor(99)

	want := trueColor(0)

	if got != want {
		t.Errorf(
			"trueColor(99) = %q, want fallback %q",
			got,
			want,
		)
	}
}

// TestColor256 verifies that each supported intensity bucket
// produces a 256-color ANSI escape sequence.
func TestColor256(t *testing.T) {
	for bucket := 0; bucket <= 4; bucket++ {
		got := color256(bucket)

		if got == "" {
			t.Errorf(
				"color256(%d) returned empty string",
				bucket,
			)
		}

		if !strings.Contains(got, "\033[38;5;") {
			t.Errorf(
				"color256(%d) = %q, want 256-color ANSI sequence",
				bucket,
				got,
			)
		}
	}
}

// TestColor256InvalidBucket verifies that unsupported buckets
// fall back to the zero-contribution color.
func TestColor256InvalidBucket(t *testing.T) {
	got := color256(99)

	want := color256(0)

	if got != want {
		t.Errorf(
			"color256(99) = %q, want fallback %q",
			got,
			want,
		)
	}
}

// TestCellColor256 verifies that Color256 actually uses the
// 256-color palette instead of the truecolor palette.
func TestCellColor256(t *testing.T) {
	got := cellColor(Color256, 3)

	if !strings.Contains(got, "\033[38;5;") {
		t.Errorf(
			"cellColor(Color256, 3) = %q, want 256-color ANSI sequence",
			got,
		)
	}
}

// TestCellColorTrueColor verifies that ColorTrueColor uses
// the truecolor palette.
func TestCellColorTrueColor(t *testing.T) {
	got := cellColor(ColorTrueColor, 3)

	if !strings.Contains(got, "\033[38;2;") {
		t.Errorf(
			"cellColor(ColorTrueColor, 3) = %q, want truecolor ANSI sequence",
			got,
		)
	}
}
