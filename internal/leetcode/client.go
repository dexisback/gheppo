// Package leetcode communicates with LeetCode's public GraphQL API to retrieve
// user submission statistics and activity calendar.
package leetcode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dexisback/gheppo/internal/github"
	"github.com/dexisback/gheppo/internal/stats"
)

var defaultGraphQLEndpoint = "https://leetcode.com/graphql"
var graphqlEndpoint = defaultGraphQLEndpoint

// Client interacts with LeetCode's public GraphQL endpoint.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new LeetCode client.
func NewClient() *Client {
	return &Client{
		endpoint: graphqlEndpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetEndpointForTesting overrides the GraphQL endpoint in tests.
func SetEndpointForTesting(endpoint string) func() {
	prev := graphqlEndpoint
	graphqlEndpoint = endpoint
	return func() {
		graphqlEndpoint = prev
	}
}

type submissionNum struct {
	Difficulty  string `json:"difficulty"`
	Count       int    `json:"count"`
	Submissions int    `json:"submissions"`
}

type leetcodeUserResponse struct {
	Data struct {
		MatchedUser *struct {
			Username string `json:"username"`
			Profile  struct {
				Ranking    int `json:"ranking"`
				Reputation int `json:"reputation"`
			} `json:"profile"`
			SubmitStats struct {
				AcSubmissionNum []submissionNum `json:"acSubmissionNum"`
			} `json:"submitStats"`
			SubmitStatsGlobal struct {
				AcSubmissionNum []submissionNum `json:"acSubmissionNum"`
			} `json:"submitStatsGlobal"`
			UserCalendar struct {
				ActiveYears        []int  `json:"activeYears"`
				Streak             int    `json:"streak"`
				TotalActiveDays    int    `json:"totalActiveDays"`
				SubmissionCalendar string `json:"submissionCalendar"`
			} `json:"userCalendar"`
		} `json:"matchedUser"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const leetcodeQuery = `
query getUserProfile($username: String!) {
  matchedUser(username: $username) {
    username
    profile {
      ranking
      reputation
    }
    submitStats: submitStatsGlobal {
      acSubmissionNum {
        difficulty
        count
        submissions
      }
    }
    userCalendar {
      activeYears
      streak
      totalActiveDays
      submissionCalendar
    }
  }
}
`

// FetchUserSummary queries LeetCode public profile and transforms it into a standard Summary.
func (c *Client) FetchUserSummary(username string) (*stats.Summary, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("LeetCode username cannot be empty")
	}

	reqBody := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query: leetcodeQuery,
		Variables: map[string]interface{}{
			"username": username,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := c.endpoint
	if endpoint == "" {
		endpoint = graphqlEndpoint
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gheppo-terminal-client")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API error: HTTP %d", resp.StatusCode)
	}

	var userResp leetcodeUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(userResp.Errors) > 0 {
		return nil, fmt.Errorf("leetcode API error: %s", userResp.Errors[0].Message)
	}

	user := userResp.Data.MatchedUser
	if user == nil || user.Username == "" {
		return nil, fmt.Errorf("user %q not found on LeetCode", username)
	}

	// Parse submission calendar JSON
	countsByDate := make(map[string]int)
	if user.UserCalendar.SubmissionCalendar != "" {
		var rawMap map[string]int
		if err := json.Unmarshal([]byte(user.UserCalendar.SubmissionCalendar), &rawMap); err == nil {
			for tsStr, count := range rawMap {
				if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
					dateStr := time.Unix(ts, 0).UTC().Format("2006-01-02")
					countsByDate[dateStr] = count
				}
			}
		}
	}

	// Build 52 weeks of day entries aligned from Sunday to Saturday ending today
	now := time.Now().UTC()
	daysFromSunday := int(now.Weekday())
	endOfWeek := now.AddDate(0, 0, 6-daysFromSunday)
	startOfWeek := endOfWeek.AddDate(0, 0, -(52*7 - 1))

	var weeks []github.Week
	for w := 0; w < 52; w++ {
		var days []github.Day
		for d := 0; d < 7; d++ {
			currDate := startOfWeek.AddDate(0, 0, w*7+d)
			dateStr := currDate.Format("2006-01-02")
			days = append(days, github.Day{
				Date:  dateStr,
				Count: countsByDate[dateStr],
			})
		}
		weeks = append(weeks, github.Week{Days: days})
	}

	totalSolved := 0
	totalSubmissions := 0
	subList := user.SubmitStats.AcSubmissionNum
	if len(subList) == 0 {
		subList = user.SubmitStatsGlobal.AcSubmissionNum
	}
	for _, num := range subList {
		if strings.EqualFold(num.Difficulty, "All") {
			totalSolved = num.Count
			totalSubmissions = num.Submissions
			break
		}
	}

	cal := &github.ContributionCalendar{
		Login:      user.Username,
		Total:      totalSubmissions,
		Weeks:      weeks,
		Followers:  user.Profile.Ranking,
		Following:  user.UserCalendar.TotalActiveDays,
		Repos:      totalSolved,
		TotalStars: user.Profile.Reputation,
	}

	summary := stats.Summarize(cal)
	summary.Login = user.Username
	return summary, nil
}
