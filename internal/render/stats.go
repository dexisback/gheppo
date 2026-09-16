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
	barWidth := 16
	if maxWidth >= 30 {
		barWidth = 18
	}
	if maxWidth >= 40 {
		barWidth = 22
	}
	progressStr, progressVisLen := RenderYearProgress(year, yearProgress, barWidth, theme)
	rows = append(rows, StatsRow{
		Content:    progressStr,
		VisibleLen: progressVisLen,
	})

	// === ROWS 1-3: EMPTY GAPS ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// === ROW 4: LONGEST STREAK (Inline value + label) ===
	streakStr := formatNumber(s.LongestStreak)
	streakUnit := "DAYS"
	if s.LongestStreak == 1 {
		streakUnit = "DAY"
	}
	if theme.Mode == ColorASCII {
		streakLine := fmt.Sprintf("%s %s LONGEST STREAK", streakStr, streakUnit)
		rows = append(rows, StatsRow{
			Content:    streakLine,
			VisibleLen: len(streakLine),
		})
	} else {
		streakLine := fmt.Sprintf("%s%s%s%s %s%s LONGEST STREAK%s",
			theme.Bold, theme.Primary, streakStr, theme.Reset,
			theme.Secondary, streakUnit, theme.Reset,
		)
		visLen := len(streakStr) + 1 + len(streakUnit) + 1 + len("LONGEST STREAK")
		rows = append(rows, StatsRow{
			Content:    streakLine,
			VisibleLen: visLen,
		})
	}

	// === ROW 5: DAILY AVERAGE (Inline value + label, cohesive activity section) ===
	avgNum := formatFloat(s.DailyAverage)
	if theme.Mode == ColorASCII {
		avgLine := fmt.Sprintf("%s / DAY DAILY AVERAGE", avgNum)
		rows = append(rows, StatsRow{
			Content:    avgLine,
			VisibleLen: len(avgLine),
		})
	} else {
		avgLine := fmt.Sprintf("%s%s%s%s %s/ DAY DAILY AVERAGE%s",
			theme.Bold, theme.Primary, avgNum, theme.Reset,
			theme.Secondary, theme.Reset,
		)
		visLen := len(avgNum) + 1 + len("/ DAY DAILY AVERAGE")
		rows = append(rows, StatsRow{
			Content:    avgLine,
			VisibleLen: visLen,
		})
	}

	// === ROW 6: BEST DAY (Inline date + count + label) ===
	var bestDayLine string
	var bestDayVisLen int
	if !s.BestDay.IsZero() && s.BestDayCount > 0 {
		bestDayStr := strings.ToUpper(s.BestDay.Format("Jan 02"))
		bestCountStr := formatNumber(s.BestDayCount)
		if theme.Mode == ColorASCII {
			bestDayLine = fmt.Sprintf("%s . %s BEST DAY", bestDayStr, bestCountStr)
			bestDayVisLen = len(bestDayStr) + 3 + len(bestCountStr) + 9
		} else {
			bestDayLine = fmt.Sprintf("%s%s%s%s %s·%s %s%s%s%s %sBEST DAY%s",
				theme.Primary, theme.Bold, bestDayStr, theme.Reset,
				theme.Secondary, theme.Reset,
				theme.ActivityHigh, theme.Bold, bestCountStr, theme.Reset,
				theme.Secondary, theme.Reset,
			)
			bestDayVisLen = len(bestDayStr) + 3 + len(bestCountStr) + 1 + len("BEST DAY")
		}
	} else {
		if theme.Mode == ColorASCII {
			bestDayLine = "-- . 0 BEST DAY"
			bestDayVisLen = 15
		} else {
			bestDayLine = fmt.Sprintf("%s%s--%s %s·%s %s%s0%s %sBEST DAY%s",
				theme.Primary, theme.Bold, theme.Reset,
				theme.Secondary, theme.Reset,
				theme.ActivityHigh, theme.Bold, theme.Reset,
				theme.Secondary, theme.Reset,
			)
			bestDayVisLen = 2 + 3 + 1 + 1 + len("BEST DAY")
		}
	}
	rows = append(rows, StatsRow{
		Content:    bestDayLine,
		VisibleLen: bestDayVisLen,
	})

	// === ROWS 7-8: EMPTY GAPS ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	subDivider := theme.Border + "│ " + theme.Reset
	if theme.Mode == ColorASCII {
		subDivider = "| "
	}

	// === ROWS 9-13: PROFILE & METRICS (Source specific) ===
	if s.Source == "leetcode" {
		col1Width := 15

		var label1, label2 string
		var val1, val2 string

		if s.ContestRating > 0 {
			label1 = "CONTEST RATING"
			label2 = "GLOBAL RANK"
			val1 = formatNumber(int(s.ContestRating))
			if s.GlobalRanking > 0 {
				val2 = "#" + formatNumber(s.GlobalRanking)
			} else {
				val2 = "--"
			}
		} else {
			label1 = "GLOBAL RANK"
			label2 = "REPUTATION"
			if s.GlobalRanking > 0 {
				val1 = "#" + formatNumber(s.GlobalRanking)
			} else {
				val1 = "UNRANKED"
			}
			val2 = formatNumber(s.Reputation)
		}

		// Row 9: Contest Rating & Rank labels
		pad9 := col1Width - len(label1)
		if pad9 < 1 {
			pad9 = 1
		}
		row9 := fmt.Sprintf("%s%s%s%s%s%s%s%s",
			theme.Secondary, label1, theme.Reset,
			strings.Repeat(" ", pad9),
			subDivider,
			theme.Secondary, label2, theme.Reset,
		)
		rows = append(rows, StatsRow{
			Content:    row9,
			VisibleLen: len(label1) + pad9 + 2 + len(label2),
		})

		// Row 10: Contest Rating & Rank values
		pad10 := col1Width - len(val1)
		if pad10 < 1 {
			pad10 = 1
		}
		if theme.Mode == ColorASCII {
			row10 := fmt.Sprintf("%s%s%s%s",
				val1,
				strings.Repeat(" ", pad10),
				subDivider,
				val2,
			)
			rows = append(rows, StatsRow{
				Content:    row10,
				VisibleLen: len(val1) + pad10 + 2 + len(val2),
			})
		} else {
			row10 := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s",
				theme.Primary, theme.Bold, val1, theme.Reset,
				strings.Repeat(" ", pad10),
				subDivider,
				theme.Primary, theme.Bold, val2, theme.Reset,
			)
			rows = append(rows, StatsRow{
				Content:    row10,
				VisibleLen: len(val1) + pad10 + 2 + len(val2),
			})
		}

		// Row 11: Empty GAP
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

		// Rows 12-13: EASY, MEDIUM, HARD breakdown
		easyStr := formatNumber(s.EasySolved)
		mediumStr := formatNumber(s.MediumSolved)
		hardStr := formatNumber(s.HardSolved)

		colW := 9
		padEasy := colW - len("EASY")
		if padEasy < 1 {
			padEasy = 1
		}
		padMed := colW - len("MEDIUM")
		if padMed < 1 {
			padMed = 1
		}

		row12 := fmt.Sprintf("%sEASY%s%s%s%sMEDIUM%s%s%s%sHARD%s",
			theme.Secondary, theme.Reset,
			strings.Repeat(" ", padEasy),
			subDivider,
			theme.Secondary, theme.Reset,
			strings.Repeat(" ", padMed),
			subDivider,
			theme.Secondary, theme.Reset,
		)
		visLen12 := len("EASY") + padEasy + 2 + len("MEDIUM") + padMed + 2 + len("HARD")
		rows = append(rows, StatsRow{
			Content:    row12,
			VisibleLen: visLen12,
		})

		padValEasy := colW - len(easyStr)
		if padValEasy < 1 {
			padValEasy = 1
		}
		padValMed := colW - len(mediumStr)
		if padValMed < 1 {
			padValMed = 1
		}

		if theme.Mode == ColorASCII {
			row13 := fmt.Sprintf("%s%s%s%s%s%s%s",
				easyStr,
				strings.Repeat(" ", padValEasy),
				subDivider,
				mediumStr,
				strings.Repeat(" ", padValMed),
				subDivider,
				hardStr,
			)
			rows = append(rows, StatsRow{
				Content:    row13,
				VisibleLen: len(easyStr) + padValEasy + 2 + len(mediumStr) + padValMed + 2 + len(hardStr),
			})
		} else {
			row13 := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
				theme.Primary, theme.Bold, easyStr, theme.Reset,
				strings.Repeat(" ", padValEasy),
				subDivider,
				theme.Primary, theme.Bold, mediumStr, theme.Reset,
				strings.Repeat(" ", padValMed),
				subDivider,
				theme.Primary, theme.Bold, hardStr, theme.Reset,
			)
			rows = append(rows, StatsRow{
				Content:    row13,
				VisibleLen: len(easyStr) + padValEasy + 2 + len(mediumStr) + padValMed + 2 + len(hardStr),
			})
		}

		// Rows 14-15: Safety buffer
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
		rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

		return rows
	}

	// === GITHUB PROFILE & REPOSITORIES (Default) ===
	followersNum := formatNumber(s.Followers)
	followingNum := formatNumber(s.Following)
	reposNum := formatNumber(s.Repos)
	starsNum := formatNumber(s.TotalStars)

	col1Width := 13

	// Row 9: Followers & Following labels
	pad9 := col1Width - len("FOLLOWERS")
	if pad9 < 1 {
		pad9 = 1
	}
	row9 := fmt.Sprintf("%sFOLLOWERS%s%s%s%sFOLLOWING%s",
		theme.Secondary, theme.Reset,
		strings.Repeat(" ", pad9),
		subDivider,
		theme.Secondary, theme.Reset,
	)
	rows = append(rows, StatsRow{
		Content:    row9,
		VisibleLen: 9 + pad9 + 2 + 9,
	})

	// Row 10: Followers & Following numbers
	pad10 := col1Width - len(followersNum)
	if pad10 < 1 {
		pad10 = 1
	}
	if theme.Mode == ColorASCII {
		row10 := fmt.Sprintf("%s%s%s%s",
			followersNum,
			strings.Repeat(" ", pad10),
			subDivider,
			followingNum,
		)
		rows = append(rows, StatsRow{
			Content:    row10,
			VisibleLen: len(followersNum) + pad10 + 2 + len(followingNum),
		})
	} else {
		row10 := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s",
			theme.Primary, theme.Bold, followersNum, theme.Reset,
			strings.Repeat(" ", pad10),
			subDivider,
			theme.Primary, theme.Bold, followingNum, theme.Reset,
		)
		rows = append(rows, StatsRow{
			Content:    row10,
			VisibleLen: len(followersNum) + pad10 + 2 + len(followingNum),
		})
	}

	// === ROW 11: EMPTY GAP ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	// === ROWS 12-13: REPOSITORIES & TOTAL STARS ===
	pad12 := col1Width - len("REPOSITORIES")
	if pad12 < 1 {
		pad12 = 1
	}
	row12 := fmt.Sprintf("%sREPOSITORIES%s%s%s%sTOTAL STARS%s",
		theme.Secondary, theme.Reset,
		strings.Repeat(" ", pad12),
		subDivider,
		theme.Secondary, theme.Reset,
	)
	rows = append(rows, StatsRow{
		Content:    row12,
		VisibleLen: 12 + pad12 + 2 + 11,
	})

	pad13 := col1Width - len(reposNum)
	if pad13 < 1 {
		pad13 = 1
	}
	if theme.Mode == ColorASCII {
		starPart := fmt.Sprintf("* %s", starsNum)
		row13 := fmt.Sprintf("%s%s%s%s",
			reposNum,
			strings.Repeat(" ", pad13),
			subDivider,
			starPart,
		)
		rows = append(rows, StatsRow{
			Content:    row13,
			VisibleLen: len(reposNum) + pad13 + 2 + len(starPart),
		})
	} else {
		starPart := fmt.Sprintf("%s★%s %s%s%s%s",
			theme.ActivityHigh, theme.Reset,
			theme.Primary, theme.Bold, starsNum, theme.Reset,
		)
		row13 := fmt.Sprintf("%s%s%s%s%s%s%s",
			theme.Primary, theme.Bold, reposNum, theme.Reset,
			strings.Repeat(" ", pad13),
			subDivider,
			starPart,
		)
		rows = append(rows, StatsRow{
			Content:    row13,
			VisibleLen: len(reposNum) + pad13 + 2 + 2 + len(starsNum),
		})
	}

	// === ROWS 14-15: EMPTY GAPS (Safety buffer) ===
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})
	rows = append(rows, StatsRow{Content: "", VisibleLen: 0})

	return rows
}

// Backward compatibility helper
func buildStatsRows(s *stats.Summary, tokens PaletteTokens, mode ColorMode) []StatsRow {
	return RenderStatsPanel(s, tokens)
}
