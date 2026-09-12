package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNoToken
		}
		return nil, fmt.Errorf("get credentials from keychain: %w", err)
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

// Login interactively asks for a GitHub PAT, validates it,
// resolves the authenticated GitHub username, and stores both.
func Login() error {
	fmt.Print("GitHub Personal Access Token: ")

	tokenBytes, err := term.ReadPassword(0)
	fmt.Println()

	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}

	token := strings.TrimSpace(string(tokenBytes))

	if token == "" {
		return errors.New("token cannot be empty")
	}

	login, err := resolveLogin(token)
	if err != nil {
		return fmt.Errorf("validate GitHub token: %w", err)
	}

	credentials := &Credentials{
		Token: token,
		Login: login,
	}

	if err := SaveCredentials(credentials); err != nil {
		return err
	}

	fmt.Printf("Authenticated as @%s.\n", login)

	return nil
}

// resolveLogin validates the PAT against GitHub and retrieves
// the username associated with the authenticated account.
func resolveLogin(token string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.github.com/user",
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



