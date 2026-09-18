package auth

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const (
	serviceName = "gheppo"
	username    = "github"
)

var ErrNoToken = errors.New("no token found")

type Credentials struct {
	Token string
	Login string
}

type storedCredentials struct {
	Token string `json:"token"`
	Login string `json:"login"`
}

// GetCredentials returns the stored GitHub credentials.
// This function is intentionally non-interactive.
func GetCredentials() (*Credentials, error) {
	data, err := keyring.Get(serviceName, username)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			return nil, fmt.Errorf("get credentials from keychain: %w", err)
		}

		// Fallback to environment variables if keyring has no stored token
		envKeys := []string{"GITHUB_TOKEN", "GH_TOKEN", "GITHUB_MCP_TOKEN"}
		for _, key := range envKeys {
			token := strings.TrimSpace(os.Getenv(key))
			if token != "" {
				login, err := resolveLogin(token)
				if err == nil && login != "" {
					return &Credentials{
						Token: token,
						Login: login,
					}, nil
				}
			}
		}

		return nil, ErrNoToken
	}

	var stored storedCredentials
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fmt.Errorf("decode stored credentials: %w", err)
	}

	stored.Token = strings.TrimSpace(stored.Token)
	stored.Login = strings.TrimSpace(stored.Login)

	if stored.Token == "" {
		return nil, ErrNoToken
	}

	if stored.Login == "" {
		return nil, errors.New("stored credentials are missing GitHub login")
	}

	return &Credentials{
		Token: stored.Token,
		Login: stored.Login,
	}, nil
}

// SaveCredentials stores the GitHub token and login securely
// in the OS keychain.
func SaveCredentials(credentials *Credentials) error {
	if credentials == nil {
		return errors.New("credentials cannot be nil")
	}

	credentials.Token = strings.TrimSpace(credentials.Token)
	credentials.Login = strings.TrimSpace(credentials.Login)

	if credentials.Token == "" {
		return errors.New("token cannot be empty")
	}

	if credentials.Login == "" {
		return errors.New("GitHub login cannot be empty")
	}

	data, err := json.Marshal(storedCredentials{
		Token: credentials.Token,
		Login: credentials.Login,
	})
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}

	if err := keyring.Set(serviceName, username, string(data)); err != nil {
		return fmt.Errorf("save credentials to keychain: %w", err)
	}

	return nil
}

// ClearToken removes the stored GitHub credentials.
func ClearToken() error {
	err := keyring.Delete(serviceName, username)

	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("clear credentials from keychain: %w", err)
	}

	return nil
}

// PromptAndLogin interactively asks for a GitHub PAT using the provided reader and writer,
// validates it, resolves the authenticated GitHub username, and stores both in keyring.
func PromptAndLogin(in io.Reader, out io.Writer) (*Credentials, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprint(out, "GitHub Personal Access Token: ")

	var token string
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		tokenBytes, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(out)
		if err != nil {
			return nil, fmt.Errorf("read token: %w", err)
		}
		token = strings.TrimSpace(string(tokenBytes))
	} else {
		scanner := bufio.NewScanner(in)
		if scanner.Scan() {
			token = strings.TrimSpace(scanner.Text())
		}
		fmt.Fprintln(out)
	}

	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	fmt.Fprintln(out, "Validating GitHub token...")
	login, err := resolveLogin(token)
	if err != nil {
		return nil, fmt.Errorf("validate GitHub token: %w", err)
	}

	credentials := &Credentials{
		Token: token,
		Login: login,
	}

	if err := SaveCredentials(credentials); err != nil {
		return nil, err
	}

	fmt.Fprintf(out, "Authenticated as @%s.\n", login)
	return credentials, nil
}

// Login interactively asks for a GitHub PAT, validates it,
// resolves the authenticated GitHub username, and stores both.
func Login() error {
	_, err := PromptAndLogin(os.Stdin, os.Stdout)
	return err
}
var githubUserURL = "https://api.github.com/user" // for testing

// SetGitHubUserURLForTesting overrides the GitHub user API endpoint for unit tests.
func SetGitHubUserURLForTesting(url string) func() {
	prev := githubUserURL
	githubUserURL = url
	return func() {
		githubUserURL = prev
	}
}

// resolveLogin validates the PAT against GitHub and retrieves
// the username associated with the authenticated account.
func resolveLogin(token string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		githubUserURL,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("create GitHub request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request authenticated user: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub returned status %s", response.Status)
	}

	var user struct {
		Login string `json:"login"`
	}

	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("decode GitHub user: %w", err)
	}

	login := strings.TrimSpace(user.Login)

	if login == "" {
		return "", errors.New("GitHub response did not contain a login")
	}

	return login, nil
}



