// Package github communicates with GitHub's GraphQL API to retrieve
// contribution calendar and user profile metadata.
package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var graphqlURL = "https://api.github.com/graphql"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ContributionCalendar represents GitHub contribution calendar and profile metadata.
type ContributionCalendar struct {
	Login string
	Total int
	Weeks []Week

	// Profile metadata
	Followers  int
	Following  int
	Repos      int
	TotalStars int
}

type Week struct {
	Days []Day
}

type Day struct {
	Date  string
	Count int
}

// FetchContributionCalendar fetches the contribution calendar and profile metadata for a user.
func (c *Client) FetchContributionCalendar(login string) (*ContributionCalendar, error) {
	if c.token == "" {
		return nil, errors.New("GitHub token is empty")
	}

	if login == "" {
		return nil, errors.New("GitHub login is empty")
	}

	query := `
		query($login: String!) {
			user(login: $login) {
				login
				followers {
					totalCount
				}
				following {
					totalCount
				}
				repositoriesTotal: repositories {
					totalCount
				}
				ownedRepositories: repositories(first: 100, ownerAffiliations: OWNER) {
					pageInfo {
						hasNextPage
						endCursor
					}
					nodes {
						stargazerCount
					}
				}
				contributionsCollection {
					contributionCalendar {
						totalContributions
						weeks {
							contributionDays {
								date
								contributionCount
							}
						}
					}
				}
			}
		}
	`

	requestBody := struct {
		Query     string `json:"query"`
		Variables struct {
			Login string `json:"login"`
		} `json:"variables"`
	}{
		Query: query,
	}
	requestBody.Variables.Login = login

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("encode GraphQL request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		graphqlURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create GitHub request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send GitHub request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %s", resp.Status)
	}

	var response graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode GitHub response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GitHub GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data.User == nil {
		return nil, fmt.Errorf("GitHub user %q not found", login)
	}

	calendar := response.Data.User.ContributionsCollection.ContributionCalendar

	// Calculate total stars across first page of owned repositories
	totalStars := 0
	if response.Data.User.OwnedRepositories.Nodes != nil {
		for _, repo := range response.Data.User.OwnedRepositories.Nodes {
			totalStars += repo.StargazerCount
		}
	}

	// Handle pagination if user has > 100 repositories
	cursor := response.Data.User.OwnedRepositories.PageInfo.EndCursor
	hasNext := response.Data.User.OwnedRepositories.PageInfo.HasNextPage
	for hasNext && cursor != "" {
		nextPage, err := c.fetchNextRepoPage(login, cursor)
		if err != nil {
			break // gracefully keep accumulated stars
		}
		for _, repo := range nextPage.Nodes {
			totalStars += repo.StargazerCount
		}
		cursor = nextPage.PageInfo.EndCursor
		hasNext = nextPage.PageInfo.HasNextPage
	}

	// Determine repository count from repositoriesTotal or fallback to length
	repoCount := response.Data.User.RepositoriesTotal.TotalCount
	if repoCount == 0 && response.Data.User.RepositoriesFallback.TotalCount > 0 {
		repoCount = response.Data.User.RepositoriesFallback.TotalCount
	}

	result := &ContributionCalendar{
		Login:      response.Data.User.Login,
		Total:      calendar.TotalContributions,
		Weeks:      make([]Week, 0, len(calendar.Weeks)),
		Followers:  response.Data.User.Followers.TotalCount,
		Following:  response.Data.User.Following.TotalCount,
		Repos:      repoCount,
		TotalStars: totalStars,
	}

	for _, githubWeek := range calendar.Weeks {
		week := Week{
			Days: make([]Day, 0, len(githubWeek.ContributionDays)),
		}

		for _, githubDay := range githubWeek.ContributionDays {
			week.Days = append(week.Days, Day{
				Date:  githubDay.Date,
				Count: githubDay.ContributionCount,
			})
		}

		result.Weeks = append(result.Weeks, week)
	}

	return result, nil
}

func (c *Client) fetchNextRepoPage(login, cursor string) (*repositoryConnection, error) {
	query := `
		query($login: String!, $after: String!) {
			user(login: $login) {
				ownedRepositories: repositories(first: 100, ownerAffiliations: OWNER, after: $after) {
					pageInfo {
						hasNextPage
						endCursor
					}
					nodes {
						stargazerCount
					}
				}
			}
		}
	`

	requestBody := struct {
		Query     string `json:"query"`
		Variables struct {
			Login string `json:"login"`
			After string `json:"after"`
		} `json:"variables"`
	}{
		Query: query,
	}
	requestBody.Variables.Login = login
	requestBody.Variables.After = cursor

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, graphqlURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}

	var response graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if response.Data.User == nil {
		return nil, errors.New("user nil in next page")
	}

	return &response.Data.User.OwnedRepositories, nil
}

// GraphQL response structs
type graphQLResponse struct {
	Data   graphQLData    `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLData struct {
	User *githubUser `json:"user"`
}

type githubUser struct {
	Login                   string                  `json:"login"`
	Followers               totalCount              `json:"followers"`
	Following               totalCount              `json:"following"`
	RepositoriesTotal       totalCount              `json:"repositoriesTotal"`
	RepositoriesFallback    totalCount              `json:"repositories"`
	OwnedRepositories       repositoryConnection    `json:"ownedRepositories"`
	ContributionsCollection contributionsCollection `json:"contributionsCollection"`
}

type totalCount struct {
	TotalCount int `json:"totalCount"`
}

type repositoryConnection struct {
	PageInfo pageInfo     `json:"pageInfo"`
	Nodes    []repository `json:"nodes"`
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type repository struct {
	StargazerCount int `json:"stargazerCount"`
}

type contributionsCollection struct {
	ContributionCalendar contributionCalendar `json:"contributionCalendar"`
}

type contributionCalendar struct {
	TotalContributions int          `json:"totalContributions"`
	Weeks              []githubWeek `json:"weeks"`
}

type githubWeek struct {
	ContributionDays []githubDay `json:"contributionDays"`
}

type githubDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}
