package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dexisback/gheppo/internal/stats"
)

// StatsRow represents a single line in the stats panel with its visible character width.
type StatsRow struct {
	Content    string
	VisibleLen int
}

type statsRow = StatsRow

// RenderStatsPanel formats the activity and profile information into rows for the right panel.
func RenderStatsPanel(s *stats.Summary, theme Theme) []StatsRow {
	if s == nil {
		return nil
	}

	rows := []StatsRow{}

	// === ACTIVITY SECTION ===
	rows = append(rows, StatsRow{
		Content:    theme.Muted + "ACTIVITY" + theme.Reset,
		VisibleLen: 8,
	})

	// Total Contributions in Activity
	totalStr := formatNumber(s.Total)
	totalLabel := "CONTRIBUTIONS"
	if s.Total == 1 {
		totalLabel = "CONTRIBUTION"
	}
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + totalStr + theme.Reset + " " + theme.Secondary + totalLabel + theme.Reset,
		VisibleLen: len(totalStr) + 1 + len(totalLabel),
	})

	// Longest streak
	streakStr := formatNumber(s.LongestStreak)
	streakLabel := "DAYS"
	if s.LongestStreak == 1 {
		streakLabel = "DAY"
	}
	rows = append(rows, StatsRow{
		Content:    theme.ActivityHigh + theme.Bold + streakStr + theme.Reset + " " + theme.Secondary + streakLabel + theme.Reset,
		VisibleLen: len(streakStr) + 1 + len(streakLabel),
	})
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "LONGEST STREAK" + theme.Reset,
		VisibleLen: 14,
	})

	// Empty line separator between metric pairs
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Best day
	if !s.BestDay.IsZero() && s.BestDayCount > 0 {
		bestDayStr := strings.ToUpper(s.BestDay.Format("Jan 02"))
		bestCountStr := formatNumber(s.BestDayCount)
		bestCombined := bestDayStr + " · " + bestCountStr
		rows = append(rows, StatsRow{
			Content:    theme.Primary + theme.Bold + bestCombined + theme.Reset,
			VisibleLen: len(bestCombined),
		})
		rows = append(rows, StatsRow{
			Content:    theme.Secondary + "BEST DAY" + theme.Reset,
			VisibleLen: 8,
		})
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	}

	// Daily average
	avgStr := fmt.Sprintf("%s / DAY", formatFloat(s.DailyAverage))
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + avgStr + theme.Reset,
		VisibleLen: len(avgStr),
	})
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "DAILY AVERAGE" + theme.Reset,
		VisibleLen: 13,
	})

	// Empty line before profile
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// === PROFILE SECTION ===
	rows = append(rows, StatsRow{
		Content:    theme.Muted + "PROFILE" + theme.Reset,
		VisibleLen: 7,
	})

	// Username as identity anchor
	username := "@" + s.Login
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + username + theme.Reset,
		VisibleLen: len(username),
	})

	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Followers
	followersStr := fmt.Sprintf("%s FOLLOWERS", formatNumber(s.Followers))
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + followersStr + theme.Reset,
		VisibleLen: utf8.RuneCountInString(followersStr),
	})

	// Following
	followingStr := fmt.Sprintf("%s FOLLOWING", formatNumber(s.Following))
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + followingStr + theme.Reset,
		VisibleLen: utf8.RuneCountInString(followingStr),
	})

	// Repositories
	reposStr := fmt.Sprintf("%s REPOSITORIES", formatNumber(s.Repos))
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + reposStr + theme.Reset,
		VisibleLen: utf8.RuneCountInString(reposStr),
	})

	// Total stars
	starsStr := fmt.Sprintf("★ %s TOTAL STARS", formatNumber(s.TotalStars))
	rows = append(rows, StatsRow{
		Content:    theme.ActivityHigh + starsStr + theme.Reset,
		VisibleLen: utf8.RuneCountInString(starsStr),
	})

	return rows
}

// Backward compatibility helper
func buildStatsRows(s *stats.Summary, tokens PaletteTokens, mode ColorMode) []StatsRow {
	return RenderStatsPanel(s, tokens)
}
