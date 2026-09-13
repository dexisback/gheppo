//imports stats.Summary and does nothing but walk Summary.Grid and paints colored chars.

package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/dexisback/gheppo/internal/stats"
)

type ColorMode int

const (
	ColorTrueColor ColorMode = iota
	Color256
	ColorASCII
)

// detect-color-mode determines what level of terminal color support is avaiable rn
func DetectColorMode() ColorMode {
	if os.Getenv("NO_COLOR") != "" {
		return ColorASCII
	}

	colorterm := strings.ToLower(os.Getenv("COLORTERM"))

	if colorterm == "truecolor" || colorterm == "24bit" {
		return ColorTrueColor
	}

	term := strings.ToLower(os.Getenv("TERM"))

	if strings.Contains(term, "256color") {
		return Color256
	}

	return ColorASCII
}

//grid renders a processed contribution summary
//all date calculations, weekday alignment, streak calculations, and intentsity bucketing have alr been handled by the stats file

func Grid(s *stats.Summary) string {
	if s == nil {
		return ""
	}

	//else:
	mode := DetectColorMode()

	var out strings.Builder

	for _, week := range s.Grid {
		for _, cell := range week {
			if cell.Empty {
				out.WriteByte(' ')
				continue
			}
			out.WriteString(cellColor(mode, cell.Bucket))
		}

		out.WriteByte('\n')
	}

	out.WriteString(fmt.Sprintf("%d contributions · %d day streak · %d longest streak", s.Total, s.CurrentStreak, s.LongestStreak))
	return out.String()
}

// func cellColor(mode ColorMode, bucket int ) string {
// 	if mode == ColorASCII{
// 		return "#"
// 	}
// 	//else:
// 	if bucket <= 0{
// 		return "\033[38;5;238m■\033[0m"

// 	}

// 	return trueColor(bucket)
// }

func cellColor(mode ColorMode, bucket int) string {
	if mode == ColorASCII {
		return "#"
	}

	if mode == Color256 {
		return color256(bucket)
	}

	return trueColor(bucket)
}

func trueColor(bucket int) string {
	colors := map[int]string{
		0: "\033[38;2;235;237;240m■\033[0m",
		1: "\033[38;2;155;233;168m■\033[0m",
		2: "\033[38;2;64;196;99m■\033[0m",
		3: "\033[38;2;38;166;65m■\033[0m",
		4: "\033[38;2;22;101;52m■\033[0m",
	}

	if color, ok := colors[bucket]; ok {
		return color
	}
	return colors[0]
}

func color256(bucket int) string {
	colors := map[int]string{
		0: "\033[38;5;238m■\033[0m",
		1: "\033[38;5;151m■\033[0m",
		2: "\033[38;5;77m■\033[0m",
		3: "\033[38;5;71m■\033[0m",
		4: "\033[38;5;29m■\033[0m",
	}

	if color, ok := colors[bucket]; ok {
		return color
	}

	return colors[0]
}
