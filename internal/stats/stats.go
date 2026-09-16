package stats

import (
	"sort"
	"time"

	"github.com/dexisback/gheppo/internal/github"
)

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// Summarize converts GitHub's contribution calendar into a renderer-ready
// summary containing aligned grid data, intensity buckets, and streaks.
func Summarize(cal *github.ContributionCalendar) *Summary {
	if cal == nil {
		return &Summary{}
	}

	summary := &Summary{
		Login:      cal.Login,
		Total:      cal.Total,
		Followers:  cal.Followers,
		Following:  cal.Following,
		Repos:      cal.Repos,
		TotalStars: cal.TotalStars,
	}

	// Flatten all days while preserving chronological order.
	days := make([]github.Day, 0)

	for _, week := range cal.Weeks {
		days = append(days, week.Days...)
	}

	if len(days) == 0 {
		return summary
	}

	// GitHub normally returns chronological data, but sorting here makes
	// the streak calculation independent of the API's ordering.
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	parsedDays := make([]parsedDay, 0, len(days))

	for _, day := range days {
		date, err := time.Parse("2006-01-02", day.Date)
		if err != nil {
			continue
		}

		parsedDays = append(parsedDays, parsedDay{
			Date:  date,
			Count: day.Count,
		})
	}

	if len(parsedDays) == 0 {
		return summary
	}

	maxCount := 0

	for _, day := range parsedDays {
		if day.Count > maxCount {
			maxCount = day.Count
		}
	}

	summary.Grid = buildGrid(parsedDays, maxCount)

	// Filter days to the current calendar year (YTD)
	currentYear := time.Now().Year()
	if len(parsedDays) > 0 {
		currentYear = parsedDays[len(parsedDays)-1].Date.Year()
	}

	var currentYearDays []parsedDay
	ytdTotal := 0
	for _, day := range parsedDays {
		if day.Date.Year() == currentYear {
			currentYearDays = append(currentYearDays, day)
			ytdTotal += day.Count
		}
	}

	if len(currentYearDays) > 0 {
		summary.Total = ytdTotal
		summary.BestDay, summary.BestDayCount = findBestDay(currentYearDays)
		summary.DailyAverage = float64(ytdTotal) / float64(len(currentYearDays))
		summary.CurrentStreak, summary.LongestStreak = calculateStreaks(parsedDays)
	} else {
		summary.Total = cal.Total
		summary.BestDay, summary.BestDayCount = findBestDay(parsedDays)
		if len(parsedDays) > 0 {
			summary.DailyAverage = float64(summary.Total) / float64(len(parsedDays))
		}
		summary.CurrentStreak, summary.LongestStreak = calculateStreaks(parsedDays)
	}

	return summary
}

type parsedDay struct {
	Date  time.Time
	Count int
}

// buildGrid converts chronological days into a [week][weekday] grid.
func buildGrid(days []parsedDay, maxCount int) [][]Cell {
	if len(days) == 0 {
		return nil
	}

	nonZeroCounts := make([]int, 0, len(days))
	for _, day := range days {
		if day.Count > 0 {
			nonZeroCounts = append(nonZeroCounts, day.Count)
		}
	}
	sort.Ints(nonZeroCounts)

	firstDay := days[0]
	firstWeekday := int(firstDay.Date.Weekday())

	grid := make([][]Cell, 0)

	// The first week may begin partway through the Sunday-Saturday row.
	// Insert empty cells before the first real day so it lands on its
	// actual weekday.
	firstWeek := make([]Cell, firstWeekday)

	for i := 0; i < firstWeekday; i++ {
		firstWeek[i] = Cell{
			Empty: true,
		}
	}

	for i := 0; i < 7-firstWeekday && i < len(days); i++ {
		day := days[i]

		firstWeek = append(firstWeek, Cell{
			Date:   day.Date,
			Count:  day.Count,
			Bucket: calculateBucket(day.Count, maxCount, nonZeroCounts),
		})
	}

	grid = append(grid, firstWeek)

	consumed := 7 - firstWeekday

	// Every subsequent week gets seven cells.
	for consumed < len(days) {
		week := make([]Cell, 0, 7)

		for weekday := 0; weekday < 7 && consumed < len(days); weekday++ {
			day := days[consumed]

			week = append(week, Cell{
				Date:   day.Date,
				Count:  day.Count,
				Bucket: calculateBucket(day.Count, maxCount, nonZeroCounts),
			})

			consumed++
		}

		grid = append(grid, week)
	}

	return grid
}

func calculateBucket(count, maxCount int, nonZeroCounts []int) int {
	if count <= 0 || maxCount <= 0 {
		return 0
	}
	if len(nonZeroCounts) < 4 {
		return bucket(count, maxCount)
	}

	q1 := nonZeroCounts[len(nonZeroCounts)/4]
	q2 := nonZeroCounts[len(nonZeroCounts)/2]
	q3 := nonZeroCounts[(3*len(nonZeroCounts))/4]

	if q3 <= q1 {
		return bucket(count, maxCount)
	}

	if count <= q1 {
		return 1
	}
	if count <= q2 {
		return 2
	}
	if count <= q3 {
		return 3
	}
	return 4
}

// bucket converts a contribution count into an intensity level:
//
// 0 = no contributions
// 1 = low
// 2 = medium-low
// 3 = medium-high
// 4 = high
func bucket(count, max int) int {
	if count <= 0 || max <= 0 {
		return 0
	}

	// Divide the positive range into four intensity levels.
	level := (count * 4) / max

	if level < 1 {
		level = 1
	}

	if level > 4 {
		level = 4
	}

	return level
}

// func calculateStreaks(days []parsedDay) (current int, longest int) {
// 	if len(days) == 0 {
// 		return 0, 0
// 	}

// 	// Longest streak: straightforward linear scan.
// 	run := 0

// 	for _, day := range days {
// 		if day.Count > 0 {
// 			run++

// 			if run > longest {
// 				longest = run
// 			}
// 		} else {
// 			run = 0
// 		}
// 	}

// 	// Current streak is different from longest streak.
// 	//
// 	// If today has zero contributions, today's day is still in progress,
// 	// so start looking from yesterday rather than immediately breaking
// 	// the streak.
// 	// today := time.Now().Truncate(24 * time.Hour)
// 	today := dateOnly(time.Now())

// 	last := len(days) - 1

// 	if sameDay(days[last].Date, today) && days[last].Count == 0 {
// 		last--
// 	}

// 	if last < 0 {
// 		return 0, longest
// 	}

// 	// Walk backwards from the most recent completed/contributing day.
// 	for i := last; i >= 0; i-- {
// 		if days[i].Count == 0 {
// 			break
// 		}

// 		current++
// 	}

// 	return current, longest
// }

func calculateStreaks(days []parsedDay) (current, longest int) {
	if len(days) == 0 {
		return 0, 0
	}

	run := 0
	for i, day := range days {
		if day.Count == 0 {
			run = 0
			continue
		}

		if i > 0 && !sameDay(days[i-1].Date.AddDate(0, 0, 1), day.Date) {
			run = 0
		}

		run++

		if run > longest {
			longest = run
		}
	}

	today := dateOnly(time.Now())
	last := len(days) - 1

	if sameDay(days[last].Date, today) && days[last].Count == 0 {
		last--
	}

	if last < 0 {
		return 0, longest
	}

	for i := last; i >= 0; i-- {
		if days[i].Count == 0 {
			break
		}

		if i < last && !sameDay(days[i].Date.AddDate(0, 0, 1), days[i+1].Date) {
			break
		}

		current++
	}

	return current, longest
}
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()

	return ay == by && am == bm && ad == bd
}

// findBestDay returns the date and count of the day with the highest contribution count.
// If multiple days tie, the most recent is returned. If all days are zero, zero values are returned.
func findBestDay(days []parsedDay) (time.Time, int) {
	if len(days) == 0 {
		return time.Time{}, 0
	}

	var bestDay time.Time
	bestCount := 0

	for _, day := range days {
		if day.Count >= bestCount && day.Count > 0 {
			bestDay = day.Date
			bestCount = day.Count
		}
	}

	return bestDay, bestCount
}

// RecomputeDerived recomputes BestDay, BestDayCount, and DailyAverage from the Grid
// if they are missing or uninitialized (e.g. from an older cache file).
func RecomputeDerived(s *Summary) {
	if s == nil || len(s.Grid) == 0 {
		return
	}

	var bestDay time.Time
	bestCount := 0
	totalDays := 0
	totalCount := 0

	for _, week := range s.Grid {
		for _, cell := range week {
			if cell.Empty || cell.Date.IsZero() {
				continue
			}
			totalDays++
			totalCount += cell.Count
			if cell.Count >= bestCount && cell.Count > 0 {
				bestDay = cell.Date
				bestCount = cell.Count
			}
		}
	}

	if (s.BestDay.IsZero() || s.BestDayCount == 0) && bestCount > 0 {
		s.BestDay = bestDay
		s.BestDayCount = bestCount
	}

	if s.DailyAverage == 0 && totalDays > 0 {
		if s.Total > 0 {
			s.DailyAverage = float64(s.Total) / float64(totalDays)
		} else {
			s.DailyAverage = float64(totalCount) / float64(totalDays)
		}
	}
}

