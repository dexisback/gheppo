//this page's only job is to talking to github api  -- sending the GraphQL request -> getting raw JSON back -> decoding it into go structs 
//its job is NOT to know about contribution buckets, calendards with meaning or rewndering. (that layer is internal/stats)

package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const graphqlURL= "https://api.github.com/graphql"


type Client struct {
	token string 
	httpClient   *http.Client 
}


func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}


// ContributionCalendar is Gheppo's internal representation of github contri calendar



type ContributionCalendar struct {
	Login string
	Total int
	Weeks []Week
}


type Week struct {
	Days []Day
}

type Day struct {
	Date  string
	Count int
}

//----

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

	result := &ContributionCalendar{
		Login: response.Data.User.Login,
		Total: calendar.TotalContributions,
		Weeks: make([]Week, 0, len(calendar.Weeks)),
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


// These structs mirror GitHub's GraphQL response.
// They stay private to this package so GitHub's API shape doesnt leak into the rest of gheppo

type graphQLResponse struct {
	Data graphQLData `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

type graphQLError struct {
	Message   string `json:"message"`

}


type graphQLData struct {
	User   *githubUser    `json:"user"`
}

type githubUser  struct {
	Login   string  `json:"login"`
	ContributionsCollection      contributionsCollection    `json:"contributionCollection"`
}

type contributionsCollection struct {
	ContributionCalendar contributionCalendar `json:"contributionCalendar"`
}

type contributionCalendar struct {
	TotalContributions int            `json:"totalContributions"`
	Weeks              []githubWeek   `json:"weeks"`
}

type githubWeek struct {
	ContributionDays []githubDay `json:"contributionDays"`
}


type githubDay struct {
	Date             string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}


//we no longer do the assumption that Days[0] is Sunday
//the architecture is : github graphql -> internal/github -> ContributionsCalendar -> internal/stats -> bucketed contribution data -> internal/render -> terminal


