package render

import (
	"fmt"
	"strings"
	"time"

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
	return RenderStatsPanelWithWidth(s, theme, 30)
}

// RenderStatsPanelWithWidth formats stats panel rows adapted to maxWidth.
func RenderStatsPanelWithWidth(s *stats.Summary, theme Theme, maxWidth int) []StatsRow {
	if s == nil {
		return nil
	}

	rows := []StatsRow{}

	// === ROW 0: YEAR PROGRESS BAR ===
	year := time.Now().Year()
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	yearProgress := CalculateYearProgress()
	barWidth := 10
	if maxWidth >= 30 {
		barWidth = 12
	}
	progressStr, progressVisLen := RenderYearProgress(year, yearProgress, barWidth, theme)
	rows = append(rows, StatsRow{
		Content:    progressStr,
		VisibleLen: progressVisLen,
	})

	// === ROW 1: EMPTY ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// === ROW 2: ACTIVITY SECTION HEADING ===
	rows = append(rows, StatsRow{
		Content:    theme.Muted + "ACTIVITY" + theme.Reset,
		VisibleLen: 8,
	})

	// === ROW 3: TOTAL CONTRIBUTIONS ===
	totalStr := formatNumber(s.Total)
	totalLabel := "CONTRIBUTIONS"
	if s.Total == 1 {
		totalLabel = "CONTRIBUTION"
	}
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + totalStr + theme.Reset + " " + theme.Secondary + totalLabel + theme.Reset,
		VisibleLen: len(totalStr) + 1 + len(totalLabel),
	})

	// === ROW 4: LONGEST STREAK VALUE ===
	streakStr := formatNumber(s.LongestStreak)
	streakLabel := "DAYS"
	if s.LongestStreak == 1 {
		streakLabel = "DAY"
	}
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + streakStr + " " + streakLabel + theme.Reset,
		VisibleLen: len(streakStr) + 1 + len(streakLabel),
	})

	// === ROW 5: LONGEST STREAK LABEL ===
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "LONGEST STREAK" + theme.Reset,
		VisibleLen: 14,
	})

	// === ROW 6: BEST DAY VALUE ===
	var bestContent string
	var bestVisLen int
	if !s.BestDay.IsZero() && s.BestDayCount > 0 {
		bestDayStr := strings.ToUpper(s.BestDay.Format("Jan 02"))
		bestCountStr := formatNumber(s.BestDayCount)
		bestContent = fmt.Sprintf("%s%s%s%s %s·%s %s%s%s%s",
			theme.Primary, theme.Bold, bestDayStr, theme.Reset,
			theme.Secondary, theme.Reset,
			theme.ActivityHigh, theme.Bold, bestCountStr, theme.Reset,
		)
		bestVisLen = len(bestDayStr) + 3 + len(bestCountStr)
	} else {
		bestContent = fmt.Sprintf("%s%s--%s %s·%s %s%s0%s",
			theme.Primary, theme.Bold, theme.Reset,
			theme.Secondary, theme.Reset,
			theme.ActivityHigh, theme.Bold, theme.Reset,
		)
		bestVisLen = 6
	}
	rows = append(rows, StatsRow{
		Content:    bestContent,
		VisibleLen: bestVisLen,
	})

	// === ROW 7: BEST DAY LABEL ===
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "BEST DAY" + theme.Reset,
		VisibleLen: 8,
	})

	// === ROW 8: DAILY AVERAGE VALUE ===
	avgNum := formatFloat(s.DailyAverage)
	rows = append(rows, StatsRow{
		Content: fmt.Sprintf("%s%s%s%s %s/ DAY%s",
			theme.Primary, theme.Bold, avgNum, theme.Reset,
			theme.Secondary, theme.Reset,
		),
		VisibleLen: len(avgNum) + 6,
	})

	// === ROW 9: DAILY AVERAGE LABEL ===
	rows = append(rows, StatsRow{
		Content:    theme.Secondary + "DAILY AVERAGE" + theme.Reset,
		VisibleLen: 13,
	})

	// === ROW 10: EMPTY ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// === ROW 11: PROFILE SECTION HEADING ===
	rows = append(rows, StatsRow{
		Content:    theme.Muted + "PROFILE" + theme.Reset,
		VisibleLen: 7,
	})

	// === ROW 12: USERNAME ===
	username := "@" + s.Login
	rows = append(rows, StatsRow{
		Content:    theme.Primary + theme.Bold + username + theme.Reset,
		VisibleLen: len(username),
	})

	// === ROWS 13-16: PROFILE 2-COLUMN GRID WITH SUB-DIVIDER ===
	followersNum := formatNumber(s.Followers)
	followingNum := formatNumber(s.Following)
	reposNum := formatNumber(s.Repos)
	starsNum := formatNumber(s.TotalStars)

	subDivider := theme.Border + "│ " + theme.Reset
	if theme.Mode == ColorASCII {
		subDivider = "| "
	}
	col1Width := 13

	// Row 13: Followers & Following labels
	pad13 := col1Width - len("FOLLOWERS")
	if pad13 < 1 {
		pad13 = 1
	}
	row13 := fmt.Sprintf("%sFOLLOWERS%s%s%s%sFOLLOWING%s",
		theme.Secondary, theme.Reset,
		strings.Repeat(" ", pad13),
		subDivider,
		theme.Secondary, theme.Reset,
	)
	rows = append(rows, StatsRow{
		Content:    row13,
		VisibleLen: 9 + pad13 + 2 + 9,
	})

	// Row 14: Followers & Following numbers
	pad14 := col1Width - len(followersNum)
	if pad14 < 1 {
		pad14 = 1
	}
	row14 := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s",
		theme.Primary, theme.Bold, followersNum, theme.Reset,
		strings.Repeat(" ", pad14),
		subDivider,
		theme.Primary, theme.Bold, followingNum, theme.Reset,
	)
	rows = append(rows, StatsRow{
		Content:    row14,
		VisibleLen: len(followersNum) + pad14 + 2 + len(followingNum),
	})

	// Row 15: Repositories & Total Stars labels
	pad15 := col1Width - len("REPOSITORIES")
	if pad15 < 1 {
		pad15 = 1
	}
	row15 := fmt.Sprintf("%sREPOSITORIES%s%s%s%sTOTAL STARS%s",
		theme.Secondary, theme.Reset,
		strings.Repeat(" ", pad15),
		subDivider,
		theme.Secondary, theme.Reset,
	)
	rows = append(rows, StatsRow{
		Content:    row15,
		VisibleLen: 12 + pad15 + 2 + 11,
	})

	// Row 16: Repositories & Total Stars numbers
	pad16 := col1Width - len(reposNum)
	if pad16 < 1 {
		pad16 = 1
	}
	var starPart string
	if theme.Mode == ColorASCII {
		starPart = fmt.Sprintf("* %s", starsNum)
	} else {
		starPart = fmt.Sprintf("%s★%s %s%s%s%s",
			theme.ActivityHigh, theme.Reset,
			theme.Primary, theme.Bold, starsNum, theme.Reset,
		)
	}
	row16 := fmt.Sprintf("%s%s%s%s%s%s%s",
		theme.Primary, theme.Bold, reposNum, theme.Reset,
		strings.Repeat(" ", pad16),
		subDivider,
		starPart,
	)
	rows = append(rows, StatsRow{
		Content:    row16,
		VisibleLen: len(reposNum) + pad16 + 2 + 2 + len(starsNum),
	})

	return rows
}

// Backward compatibility helper
func buildStatsRows(s *stats.Summary, tokens PaletteTokens, mode ColorMode) []StatsRow {
	return RenderStatsPanel(s, tokens)
}
