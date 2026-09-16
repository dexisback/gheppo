package render

import (
	"fmt"
	"strings"

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
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Longest streak
	streakStr := formatNumber(s.LongestStreak)
	streakLabel := "LONGEST STREAK"
	if s.LongestStreak == 1 {
		streakLabel = "DAY"
	} else {
		streakLabel = "DAYS"
	}
	rows = append(rows, StatsRow{
		Content:    theme.ActivityHigh + theme.Bold + streakStr + theme.Reset + " " + theme.Secondary + streakLabel + theme.Reset,
		VisibleLen: len(streakStr) + 1 + len(streakLabel),
	})
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "LONGEST STREAK" + theme.Reset,
		VisibleLen: 14,
	})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Best day
	if !s.BestDay.IsZero() && s.BestDayCount > 0 {
		bestDayStr := s.BestDay.Format("Jan 02")
		bestDayStr = strings.ToUpper(bestDayStr)
		bestStr := fmt.Sprintf("%s · %s", bestDayStr, formatNumber(s.BestDayCount))
		rows = append(rows, StatsRow{
			Content:    theme.Primary + theme.Bold + bestStr + theme.Reset,
			VisibleLen: len(bestStr),
		})
		rows = append(rows, StatsRow{
			Content:    theme.Secondary + "BEST DAY" + theme.Reset,
			VisibleLen: 8,
		})
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	}

	// Daily average
	if s.DailyAverage > 0 {
		avgStr := fmt.Sprintf("%s / DAY", formatFloat(s.DailyAverage))
		rows = append(rows, StatsRow{
			Content:    theme.Primary + avgStr + theme.Reset,
			VisibleLen: len(avgStr),
		})
		rows = append(rows, StatsRow{
			Content:    theme.Secondary + "DAILY AVERAGE" + theme.Reset,
			VisibleLen: 13,
		})
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	}

	// === PROFILE SECTION ===
	rows = append(rows, StatsRow{
		Content:    theme.Muted + "PROFILE" + theme.Reset,
		VisibleLen: 7,
	})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Username
	username := "@" + s.Login
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + username + theme.Reset,
		VisibleLen: len(username),
	})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// Followers
	if s.Followers > 0 {
		followersStr := fmt.Sprintf("%s FOLLOWERS", formatNumber(s.Followers))
		rows = append(rows, StatsRow{
			Content:    theme.Primary + followersStr + theme.Reset,
			VisibleLen: len(followersStr),
		})
	}

	// Following
	if s.Following > 0 {
		followingStr := fmt.Sprintf("%s FOLLOWING", formatNumber(s.Following))
		rows = append(rows, StatsRow{
			Content:    theme.Secondary + followingStr + theme.Reset,
			VisibleLen: len(followingStr),
		})
	}

	// Repositories
	if s.Repos > 0 {
		reposStr := fmt.Sprintf("%s REPOSITORIES", formatNumber(s.Repos))
		rows = append(rows, StatsRow{
			Content:    theme.Secondary + reposStr + theme.Reset,
			VisibleLen: len(reposStr),
		})
	}

	// Total stars
	if s.TotalStars > 0 {
		starsStr := fmt.Sprintf("★ %s TOTAL STARS", formatNumber(s.TotalStars))
		rows = append(rows, StatsRow{
			Content:    theme.ActivityHigh + starsStr + theme.Reset,
			VisibleLen: len(starsStr),
		})
	}

	return rows
}

// Backward compatibility helper
func buildStatsRows(s *stats.Summary, tokens PaletteTokens, mode ColorMode) []StatsRow {
	return RenderStatsPanel(s, tokens)
}
