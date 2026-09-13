package stats

import (
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/github"
)

// TestSummarizeNil verifies that a nil contribution calendar
// safely produces an empty summary.
func TestSummarizeNil(t *testing.T) {
	summary := Summarize(nil)

	if summary == nil {
		t.Fatal("Summarize(nil) returned nil")
	}

	if summary.Login != "" {
		t.Errorf("Login = %q, want empty string", summary.Login)
	}

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}

	if summary.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %d, want 0", summary.CurrentStreak)
	}

	if summary.LongestStreak != 0 {
		t.Errorf("LongestStreak = %d, want 0", summary.LongestStreak)
	}
}

// TestSummarizeEmptyCalendar verifies that an empty GitHub calendar
// produces a summary without grid or streak data.
func TestSummarizeEmptyCalendar(t *testing.T) {
	cal := &github.ContributionCalendar{
		Login: "test-user",
		Total: 0,
	}

	summary := Summarize(cal)

	if summary.Login != "test-user" {
		t.Errorf(
			"Login = %q, want %q",
			summary.Login,
			"test-user",
		)
	}

	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}

	if summary.Grid != nil {
		t.Fatal("Grid should be nil for an empty calendar")
	}

	if summary.CurrentStreak != 0 {
		t.Errorf("CurrentStreak = %d, want 0", summary.CurrentStreak)
	}

	if summary.LongestStreak != 0 {
		t.Errorf("LongestStreak = %d, want 0", summary.LongestStreak)
	}
}

// TestSummarizeSortsDays verifies that Summarize does not depend
// on GitHub's days already being in chronological order.
func TestSummarizeSortsDays(t *testing.T) {
	cal := &github.ContributionCalendar{
		Login: "test-user",
		Total: 3,
		Weeks: []github.Week{
			{
				Days: []github.Day{
					{
						Date:  "2026-09-03",
						Count: 1,
					},
					{
						Date:  "2026-09-01",
						Count: 1,
					},
					{
						Date:  "2026-09-02",
						Count: 1,
					},
				},
			},
		},
	}

	summary := Summarize(cal)

	if len(summary.Grid) == 0 {
		t.Fatal("Grid is empty")
	}

	var dates []time.Time

	for _, week := range summary.Grid {
		for _, cell := range week {
			if !cell.Empty {
				dates = append(dates, cell.Date)
			}
		}
	}

	if len(dates) != 3 {
		t.Fatalf("got %d real cells, want 3", len(dates))
	}

	if !dates[0].Before(dates[1]) || !dates[1].Before(dates[2]) {
		t.Fatalf("dates are not chronological: %v", dates)
	}
}

// TestBucket verifies the four contribution intensity levels.
func TestBucket(t *testing.T) {
	tests := []struct {
		name  string
		count int
		max   int
		want  int
	}{
		{
			name:  "zero contributions",
			count: 0,
			max:   10,
			want:  0,
		},
		{
			name:  "negative contributions",
			count: -1,
			max:   10,
			want:  0,
		},
		{
			name:  "zero maximum",
			count: 5,
			max:   0,
			want:  0,
		},
		{
			name:  "low",
			count: 1,
			max:   10,
			want:  1,
		},
		{
			name:  "medium-low",
			count: 5,
			max:   10,
			want:  2,
		},
		{
			name:  "medium-high",
			count: 8,
			max:   10,
			want:  3,
		},
		{
			name:  "high",
			count: 10,
			max:   10,
			want:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bucket(tt.count, tt.max)

			if got != tt.want {
				t.Errorf(
					"bucket(%d, %d) = %d, want %d",
					tt.count,
					tt.max,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestBuildGridWeekdayAlignment verifies that the first contribution
// is placed under its actual weekday.
//
// 2026-09-13 is Sunday, so the first week should begin at index 0.
func TestBuildGridWeekdayAlignment(t *testing.T) {
	days := []parsedDay{
		{
			Date:  time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
			Count: 5,
		},
		{
			Date:  time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
			Count: 2,
		},
	}

	grid := buildGrid(days, 5)

	if len(grid) != 1 {
		t.Fatalf("got %d weeks, want 1", len(grid))
	}

	if len(grid[0]) != 2 {
		t.Fatalf("got %d cells, want 2", len(grid[0]))
	}

	if grid[0][0].Empty {
		t.Fatal("Sunday cell should not be empty")
	}

	if !sameDay(
		grid[0][0].Date,
		time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	) {
		t.Errorf("first cell has wrong date: %v", grid[0][0].Date)
	}

	if grid[0][0].Bucket != 4 {
		t.Errorf(
			"first cell bucket = %d, want 4",
			grid[0][0].Bucket,
		)
	}
}

// TestBuildGridLeadingEmptyCells verifies that empty cells are inserted
// before the first real contribution when the first day begins mid-week.
//
// 2026-09-08 is Tuesday, so Sunday and Monday should be empty.
func TestBuildGridLeadingEmptyCells(t *testing.T) {
	days := []parsedDay{
		{
			Date:  time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC),
			Count: 3,
		},
		{
			Date:  time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC),
			Count: 2,
		},
	}

	grid := buildGrid(days, 3)

	if len(grid) != 1 {
		t.Fatalf("got %d weeks, want 1", len(grid))
	}

	// Two leading empty cells + two real contribution cells.
	if len(grid[0]) != 4 {
		t.Fatalf(
			"got %d cells in first week, want 4",
			len(grid[0]),
		)
	}

	if !grid[0][0].Empty {
		t.Fatal("Sunday cell should be empty")
	}

	if !grid[0][1].Empty {
		t.Fatal("Monday cell should be empty")
	}

	if grid[0][2].Empty {
		t.Fatal("Tuesday cell should contain the first day")
	}

	if grid[0][2].Count != 3 {
		t.Errorf(
			"Tuesday count = %d, want 3",
			grid[0][2].Count,
		)
	}

	if grid[0][3].Empty {
		t.Fatal("Wednesday cell should contain the second day")
	}

	if grid[0][3].Count != 2 {
		t.Errorf(
			"Wednesday count = %d, want 2",
			grid[0][3].Count,
		)
	}
}

// TestCalculateStreaks verifies both current and longest streaks.
func TestCalculateStreaks(t *testing.T) {
	days := []parsedDay{
		{
			Date:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Count: 3,
		},
		{
			Date:  time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
			Count: 2,
		},
		{
			Date:  time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
			Count: 0,
		},
		{
			Date:  time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC),
			Count: 5,
		},
		{
			Date:  time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
			Count: 1,
		},
	}

	current, longest := calculateStreaks(days)

	if longest != 2 {
		t.Errorf(
			"LongestStreak = %d, want 2",
			longest,
		)
	}

	// The last day is contributing, so the current streak should
	// contain the final two contributing days.
	if current != 2 {
		t.Errorf(
			"CurrentStreak = %d, want 2",
			current,
		)
	}
}

// TestCalculateStreaksEmpty verifies that an empty day list produces
// zero for both streak values.
func TestCalculateStreaksEmpty(t *testing.T) {
	current, longest := calculateStreaks(nil)

	if current != 0 {
		t.Errorf("CurrentStreak = %d, want 0", current)
	}

	if longest != 0 {
		t.Errorf("LongestStreak = %d, want 0", longest)
	}
}

// TestSummarizeInvalidDates verifies that invalid GitHub dates are
// ignored rather than causing Summarize to fail.
func TestSummarizeInvalidDates(t *testing.T) {
	cal := &github.ContributionCalendar{
		Login: "test-user",
		Total: 5,
		Weeks: []github.Week{
			{
				Days: []github.Day{
					{
						Date:  "not-a-date",
						Count: 100,
					},
					{
						Date:  "2026-09-01",
						Count: 5,
					},
				},
			},
		},
	}

	summary := Summarize(cal)

	if summary == nil {
		t.Fatal("Summarize() returned nil")
	}

	if len(summary.Grid) == 0 {
		t.Fatal("Grid should contain the valid date")
	}

	for _, week := range summary.Grid {
		for _, cell := range week {
			if !cell.Empty && cell.Count == 100 {
				t.Fatal("invalid date should not appear in the grid")
			}
		}
	}
}

// TestSameDay verifies that sameDay compares calendar dates rather
// than exact timestamps.
func TestSameDay(t *testing.T) {
	a := time.Date(
		2026,
		9,
		13,
		8,
		30,
		0,
		0,
		time.UTC,
	)

	b := time.Date(
		2026,
		9,
		13,
		22,
		45,
		0,
		0,
		time.UTC,
	)

	if !sameDay(a, b) {
		t.Fatal("sameDay() returned false for the same calendar date")
	}

	c := time.Date(
		2026,
		9,
		14,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	if sameDay(a, c) {
		t.Fatal("sameDay() returned true for different calendar dates")
	}
}
