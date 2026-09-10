package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const (
	serviceName = "gheppo"
	username    = "github"
)

var ErrNoToken  = errors.New("No token found") //// ErrNoToken means Gheppo has not been authenticated yet.

// getToken returns the stored github token
// breaking news -- this function is intentionally non interactive
func GetToken() (string, error) {
	token, err := keyring.Get(serviceName, username)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNoToken
		}

		return "", fmt.Errorf("get token from keychain: %w", err)
	}

	token = strings.TrimSpace(token)

	if token == "" {
		return "", ErrNoToken
	}

	return token, nil
}

// SaveToken stores the GitHub token securely in the OS keychain.
func SaveToken(token string) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return errors.New("token cannot be empty")
	}

	if err := keyring.Set(serviceName, username, token); err != nil {
		return fmt.Errorf("save token to keychain: %w", err)
	}

	return nil
}


// ClearToken removes the stored GitHub token.
func ClearToken() error {
	err := keyring.Delete(serviceName, username)

	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("clear token from keychain: %w", err)
	}

	return nil
}


// Login interactively asks for a GitHub PAT and stores it securely.
// This is only called by `gheppo auth login`.
// It must never be called from the normal shell startup path.

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

	if err := SaveToken(token); err != nil {
		return err
	}

	fmt.Println("GitHub token saved securely.")

	return nil
}