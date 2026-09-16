package github

import (
	"encoding/json"
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
				"followers": { "totalCount": 150 },
				"following": { "totalCount": 45 },
				"repositoriesTotal": { "totalCount": 12 },
				"ownedRepositories": {
					"pageInfo": { "hasNextPage": false, "endCursor": "" },
					"nodes": [
						{ "stargazerCount": 50 },
						{ "stargazerCount": 38 }
					]
				},
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

	if result.Followers != 150 {
		t.Errorf("Followers = %d, want 150", result.Followers)
	}

	if result.Following != 45 {
		t.Errorf("Following = %d, want 45", result.Following)
	}

	if result.Repos != 12 {
		t.Errorf("Repos = %d, want 12", result.Repos)
	}

	if result.TotalStars != 88 {
		t.Errorf("TotalStars = %d, want 88", result.TotalStars)
	}

	if len(result.Weeks) != 1 {
		t.Fatalf("len(Weeks) = %d, want 1", len(result.Weeks))
	}
}

func TestFetchContributionCalendarPagination(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			// Page 1 has hasNextPage: true
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"login": "paginated-user",
						"followers": map[string]interface{}{"totalCount": 10},
						"following": map[string]interface{}{"totalCount": 5},
						"repositoriesTotal": map[string]interface{}{"totalCount": 150},
						"ownedRepositories": map[string]interface{}{
							"pageInfo": map[string]interface{}{
								"hasNextPage": true,
								"endCursor": "cursor-page-1",
							},
							"nodes": []map[string]interface{}{
								{"stargazerCount": 10},
							},
						},
						"contributionsCollection": map[string]interface{}{
							"contributionCalendar": map[string]interface{}{
								"totalContributions": 100,
								"weeks": []interface{}{},
							},
						},
					},
				},
			})
		} else {
			// Page 2
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"ownedRepositories": map[string]interface{}{
							"pageInfo": map[string]interface{}{
								"hasNextPage": false,
								"endCursor": "",
							},
							"nodes": []map[string]interface{}{
								{"stargazerCount": 25},
							},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	useTestURL(t, server.URL)

	client := NewClient("test-token")
	result, err := client.FetchContributionCalendar("paginated-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalStars != 35 {
		t.Errorf("TotalStars with pagination = %d, want 35", result.TotalStars)
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
