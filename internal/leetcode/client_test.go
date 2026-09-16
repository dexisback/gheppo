package leetcode

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchUserSummarySuccess(t *testing.T) {
	ts := time.Now().UTC().Unix()
	calendarJSON := fmt.Sprintf(`{"%d": 5}`, ts)

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		resp := fmt.Sprintf(`{
			"data": {
				"matchedUser": {
					"username": "tourist",
					"profile": {
						"ranking": 42,
						"reputation": 1337
					},
					"submitStatsGlobal": {
						"acSubmissionNum": [
							{"difficulty": "All", "count": 1500, "submissions": 3000},
							{"difficulty": "Easy", "count": 500, "submissions": 800},
							{"difficulty": "Medium", "count": 700, "submissions": 1400},
							{"difficulty": "Hard", "count": 300, "submissions": 800}
						]
					},
					"userCalendar": {
						"activeYears": [2025, 2026],
						"streak": 12,
						"totalActiveDays": 90,
						"submissionCalendar": %q
					}
				},
				"userContestRanking": {
					"attendedContestsCount": 35,
					"rating": 2450.5,
					"globalRanking": 120,
					"totalParticipants": 500000,
					"topPercentage": 0.02
				}
			}
		}`, calendarJSON)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	restore := SetEndpointForTesting(server.URL)
	defer restore()

	client := NewClient()
	summary, err := client.FetchUserSummary("tourist")
	if err != nil {
		t.Fatalf("FetchUserSummary() failed: %v", err)
	}

	if summary.Login != "tourist" {
		t.Errorf("summary.Login = %q, want tourist", summary.Login)
	}
	if summary.Source != "leetcode" {
		t.Errorf("summary.Source = %q, want leetcode", summary.Source)
	}
	if summary.ProblemsSolved != 1500 {
		t.Errorf("summary.ProblemsSolved = %d, want 1500", summary.ProblemsSolved)
	}
	if summary.EasySolved != 500 {
		t.Errorf("summary.EasySolved = %d, want 500", summary.EasySolved)
	}
	if summary.MediumSolved != 700 {
		t.Errorf("summary.MediumSolved = %d, want 700", summary.MediumSolved)
	}
	if summary.HardSolved != 300 {
		t.Errorf("summary.HardSolved = %d, want 300", summary.HardSolved)
	}
	if summary.ContestRating != 2450.5 {
		t.Errorf("summary.ContestRating = %f, want 2450.5", summary.ContestRating)
	}
	if summary.GlobalRanking != 120 {
		t.Errorf("summary.GlobalRanking = %d, want 120", summary.GlobalRanking)
	}
	if summary.Reputation != 1337 {
		t.Errorf("summary.Reputation = %d, want 1337", summary.Reputation)
	}
	if len(summary.Grid) != 52 {
		t.Errorf("summary.Grid weeks = %d, want 52", len(summary.Grid))
	}
}

func TestFetchUserSummaryNotFound(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{"data": {"matchedUser": null}}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(resp))
	})

	server := httptest.NewServer(mockHandler)
	defer server.Close()

	restore := SetEndpointForTesting(server.URL)
	defer restore()

	client := NewClient()
	_, err := client.FetchUserSummary("nonexistent_lc_user")
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetchUserSummaryEmptyUsername(t *testing.T) {
	client := NewClient()
	_, err := client.FetchUserSummary("   ")
	if err == nil {
		t.Fatal("expected error for empty username, got nil")
	}
}
