package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// func setupTestKeyring(t *testing.T) {
// 	t.Helper()

// 	keyring.MockInit()

// 	t.Cleanup(func() {
// 		keyring.MockReset()
// 	})
// }

func setupTestKeyring(t *testing.T) {
	t.Helper()
	keyring.MockInit()
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_MCP_TOKEN", "")
}

func TestGetCredentialsNoToken(t *testing.T) {
	setupTestKeyring(t)

	credentials, err := GetCredentials()

	if !errors.Is(err, ErrNoToken) {
		t.Fatalf("GetCredentials() error = %v, want ErrNoToken", err)
	}

	if credentials != nil {
		t.Fatal("GetCredentials() returned credentials when none were stored")
	}
}

func TestSaveAndGetCredentials(t *testing.T) {
	setupTestKeyring(t)

	original := &Credentials{
		Token: "test-token",
		Login: "test-user",
	}

	if err := SaveCredentials(original); err != nil {
		t.Fatalf("SaveCredentials() returned error: %v", err)
	}

	loaded, err := GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() returned error: %v", err)
	}

	if loaded.Token != original.Token {
		t.Fatalf("Token = %q, want %q", loaded.Token, original.Token)
	}

	if loaded.Login != original.Login {
		t.Fatalf("Login = %q, want %q", loaded.Login, original.Login)
	}
}

func TestSaveCredentialsTrimsValues(t *testing.T) {
	setupTestKeyring(t)

	credentials := &Credentials{
		Token: "  test-token  ",
		Login: "  test-user  ",
	}

	if err := SaveCredentials(credentials); err != nil {
		t.Fatalf("SaveCredentials() returned error: %v", err)
	}

	if credentials.Token != "test-token" {
		t.Fatalf("Token = %q, want %q", credentials.Token, "test-token")
	}

	if credentials.Login != "test-user" {
		t.Fatalf("Login = %q, want %q", credentials.Login, "test-user")
	}

	loaded, err := GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials() returned error: %v", err)
	}

	if loaded.Token != "test-token" {
		t.Fatalf("loaded Token = %q, want %q", loaded.Token, "test-token")
	}

	if loaded.Login != "test-user" {
		t.Fatalf("loaded Login = %q, want %q", loaded.Login, "test-user")
	}
}

func TestSaveCredentialsValidation(t *testing.T) {
	setupTestKeyring(t)

	tests := []struct {
		name        string
		credentials *Credentials
	}{
		{
			name:        "nil credentials",
			credentials: nil,
		},
		{
			name: "empty token",
			credentials: &Credentials{
				Token: "",
				Login: "test-user",
			},
		},
		{
			name: "empty login",
			credentials: &Credentials{
				Token: "test-token",
				Login: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SaveCredentials(tt.credentials); err == nil {
				t.Fatal("SaveCredentials() returned nil error")
			}
		})
	}
}

func TestGetCredentialsMalformedData(t *testing.T) {
	setupTestKeyring(t)

	if err := keyring.Set(serviceName, username, "not valid JSON"); err != nil {
		t.Fatalf("keyring.Set() returned error: %v", err)
	}

	credentials, err := GetCredentials()

	if err == nil {
		t.Fatal("GetCredentials() returned nil error for malformed data")
	}

	if credentials != nil {
		t.Fatal("GetCredentials() returned credentials for malformed data")
	}
}

func TestGetCredentialsEmptyToken(t *testing.T) {
	setupTestKeyring(t)

	if err := keyring.Set(
		serviceName,
		username,
		`{"token":"","login":"test-user"}`,
	); err != nil {
		t.Fatalf("keyring.Set() returned error: %v", err)
	}

	credentials, err := GetCredentials()

	if !errors.Is(err, ErrNoToken) {
		t.Fatalf("GetCredentials() error = %v, want ErrNoToken", err)
	}

	if credentials != nil {
		t.Fatal("GetCredentials() returned credentials with empty token")
	}
}

func TestGetCredentialsEmptyLogin(t *testing.T) {
	setupTestKeyring(t)

	if err := keyring.Set(
		serviceName,
		username,
		`{"token":"test-token","login":""}`,
	); err != nil {
		t.Fatalf("keyring.Set() returned error: %v", err)
	}

	credentials, err := GetCredentials()

	if err == nil {
		t.Fatal("GetCredentials() returned nil error for empty login")
	}

	if credentials != nil {
		t.Fatal("GetCredentials() returned credentials with empty login")
	}
}

func TestClearToken(t *testing.T) {
	setupTestKeyring(t)

	if err := SaveCredentials(&Credentials{
		Token: "test-token",
		Login: "test-user",
	}); err != nil {
		t.Fatalf("SaveCredentials() returned error: %v", err)
	}

	if err := ClearToken(); err != nil {
		t.Fatalf("ClearToken() returned error: %v", err)
	}

	credentials, err := GetCredentials()

	if !errors.Is(err, ErrNoToken) {
		t.Fatalf("GetCredentials() error = %v, want ErrNoToken", err)
	}

	if credentials != nil {
		t.Fatal("credentials still exist after ClearToken()")
	}
}

func TestClearTokenWhenAlreadyLoggedOut(t *testing.T) {
	setupTestKeyring(t)

	if err := ClearToken(); err != nil {
		t.Fatalf("ClearToken() returned error: %v", err)
	}
}

func TestResolveLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"test-user"}`))
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() {
		http.DefaultClient = originalClient
	})

	originalURL := githubUserURL
	githubUserURL = server.URL
	t.Cleanup(func() {
		githubUserURL = originalURL
	})

	login, err := resolveLogin("test-token")
	if err != nil {
		t.Fatalf("resolveLogin() returned error: %v", err)
	}

	if login != "test-user" {
		t.Fatalf("login = %q, want %q", login, "test-user")
	}
}

func TestPromptAndLogin(t *testing.T) {
	setupTestKeyring(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"gh-tester"}`))
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() {
		http.DefaultClient = originalClient
	})

	restoreURL := SetGitHubUserURLForTesting(server.URL)
	defer restoreURL()

	in := strings.NewReader("valid-token\n")
	out := new(strings.Builder)

	creds, err := PromptAndLogin(in, out)
	if err != nil {
		t.Fatalf("PromptAndLogin failed: %v", err)
	}

	if creds.Login != "gh-tester" || creds.Token != "valid-token" {
		t.Fatalf("unexpected creds: %+v", creds)
	}

	stored, err := GetCredentials()
	if err != nil {
		t.Fatalf("GetCredentials failed: %v", err)
	}
	if stored.Login != "gh-tester" || stored.Token != "valid-token" {
		t.Fatalf("unexpected stored credentials: %+v", stored)
	}
}

func TestPromptAndLoginEmptyToken(t *testing.T) {
	setupTestKeyring(t)

	in := strings.NewReader("\n")
	out := new(strings.Builder)

	_, err := PromptAndLogin(in, out)
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean username",
			input: "dexisback",
			want:  "dexisback",
		},
		{
			name:  "username with ansi color escape",
			input: "\x1b[31mdexisback\x1b[0m",
			want:  "dexisback",
		},
		{
			name:  "username with control characters",
			input: "user\r\n\x07\x08name",
			want:  "username",
		},
		{
			name:  "username with osc escape",
			input: "\x1b]0;fake-title\x07attacker",
			want:  "attacker",
		},
		{
			name:  "unicode and allowed characters",
			input: "user_name-123.test",
			want:  "user_name-123.test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeUsername(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeUsername(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveLoginSanitizesLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"login":"\u001b[31mmalicious\u001b[0m"}`))
	}))
	defer server.Close()

	restoreURL := SetGitHubUserURLForTesting(server.URL)
	defer restoreURL()

	login, err := resolveLogin("dummy-token")
	if err != nil {
		t.Fatalf("resolveLogin returned error: %v", err)
	}

	if login != "malicious" {
		t.Errorf("login was not sanitized: got %q, want %q", login, "malicious")
	}
}
