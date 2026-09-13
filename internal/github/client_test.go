package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestServer(t *testing.T, response string, status int) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
}

func useTestURL(t *testing.T, url string) {
	t.Helper()

	originalURL := graphqlURL
	graphqlURL = url

	t.Cleanup(func() {
		graphqlURL = originalURL
	})
}

func TestFetchContributionCalendarRejectsEmptyToken(t *testing.T) {
	client := NewClient("")

	_, err := client.FetchContributionCalendar("test-user")

	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestFetchContributionCalendarRejectsEmptyLogin(t *testing.T) {
	client := NewClient("test-token")

	_, err := client.FetchContributionCalendar("")

	if err == nil {
		t.Fatal("expected error for empty login")
	}
}

func TestFetchContributionCalendar(t *testing.T) {
	server := setupTestServer(t, `{
		"data": {
			"user": {
				"login": "test-user",
				"contributionsCollection": {
					"contributionCalendar": {
						"totalContributions": 42,
						"weeks": [
							{
								"contributionDays": [
									{
										"date": "2026-09-12",
										"contributionCount": 3
									},
									{
										"date": "2026-09-13",
										"contributionCount": 5
									}
								]
							}
						]
					}
				}
			}
		}
	}`, http.StatusOK)
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")

	result, err := client.FetchContributionCalendar("test-user")
	if err != nil {
		t.Fatalf("FetchContributionCalendar() returned error: %v", err)
	}

	if result.Login != "test-user" {
		t.Fatalf("Login = %q, want %q", result.Login, "test-user")
	}

	if result.Total != 42 {
		t.Fatalf("Total = %d, want 42", result.Total)
	}

	if len(result.Weeks) != 1 {
		t.Fatalf("len(Weeks) = %d, want 1", len(result.Weeks))
	}

	if len(result.Weeks[0].Days) != 2 {
		t.Fatalf("len(Days) = %d, want 2", len(result.Weeks[0].Days))
	}

	if result.Weeks[0].Days[0].Date != "2026-09-12" {
		t.Fatalf(
			"first day date = %q, want 2026-09-12",
			result.Weeks[0].Days[0].Date,
		)
	}

	if result.Weeks[0].Days[0].Count != 3 {
		t.Fatalf(
			"first day count = %d, want 3",
			result.Weeks[0].Days[0].Count,
		)
	}

	if result.Weeks[0].Days[1].Count != 5 {
		t.Fatalf(
			"second day count = %d, want 5",
			result.Weeks[0].Days[1].Count,
		)
	}
}

func TestFetchContributionCalendarHTTPError(t *testing.T) {
	server := setupTestServer(
		t,
		`{"message":"Bad credentials"}`,
		http.StatusUnauthorized,
	)
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")

	_, err := client.FetchContributionCalendar("test-user")

	if err == nil {
		t.Fatal("expected error for HTTP failure")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("error = %q, want HTTP status", err.Error())
	}
}

func TestFetchContributionCalendarMalformedJSON(t *testing.T) {
	server := setupTestServer(
		t,
		`not valid json`,
		http.StatusOK,
	)
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")

	_, err := client.FetchContributionCalendar("test-user")

	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestFetchContributionCalendarGraphQLError(t *testing.T) {
	server := setupTestServer(t, `{
		"errors": [
			{
				"message": "Bad credentials"
			}
		]
	}`, http.StatusOK)
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")

	_, err := client.FetchContributionCalendar("test-user")

	if err == nil {
		t.Fatal("expected GraphQL error")
	}

	if !strings.Contains(err.Error(), "Bad credentials") {
		t.Fatalf(
			"error = %q, want GraphQL message",
			err.Error(),
		)
	}
}

func TestFetchContributionCalendarUserNotFound(t *testing.T) {
	server := setupTestServer(t, `{
		"data": {
			"user": null
		}
	}`, http.StatusOK)
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")

	_, err := client.FetchContributionCalendar("missing-user")

	if err == nil {
		t.Fatal("expected error for missing user")
	}

	if !strings.Contains(err.Error(), "missing-user") {
		t.Fatalf(
			"error = %q, want username in error",
			err.Error(),
		)
	}
}
