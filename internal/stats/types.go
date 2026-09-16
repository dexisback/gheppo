//takes raw calendar and turns it into buckets/streak data that the render would actually want ki types

package stats

import "time"

// Summary contains all processed contribution data that the renderer needs
type Summary struct {
	Source        string    `json:"source,omitempty"`
	Login         string    `json:"login"`
	Total         int       `json:"total"`
	CurrentStreak int       `json:"currentStreak"`
	LongestStreak int       `json:"longestStreak"`
	FetchedAt     time.Time `json:"fetchedAt"`

	Grid [][]Cell `json:"grid"` // grid is indexed as [week][weekday]

	// Profile data (GitHub & generic)
	Followers  int `json:"followers"`
	Following  int `json:"following"`
	Repos      int `json:"repos"`
	TotalStars int `json:"totalStars"`

	// LeetCode specific metrics
	ProblemsSolved int     `json:"problemsSolved,omitempty"`
	EasySolved     int     `json:"easySolved,omitempty"`
	MediumSolved   int     `json:"mediumSolved,omitempty"`
	HardSolved     int     `json:"hardSolved,omitempty"`
	ContestRating  float64 `json:"contestRating,omitempty"`
	GlobalRanking  int     `json:"globalRanking,omitempty"`
	Reputation     int     `json:"reputation,omitempty"`

	// Computed stats
	BestDay      time.Time `json:"bestDay"`
	BestDayCount int       `json:"bestDayCount"`
	DailyAverage float64   `json:"dailyAverage"`
}

type Cell struct {
	Date   time.Time
	Count  int
	Bucket int

	Empty bool //for empty boxes
}
