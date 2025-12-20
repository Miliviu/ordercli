package cli

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steipete/ordercli/internal/browserauth"
	"github.com/steipete/ordercli/internal/foodora"
)

func TestFoodoraLogin_BrowserFlow_StoresCookiesAndUA(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	setEnv(t, "FOODORA_CLIENT_SECRET", "secret")

	_, _, err := runCLI(cfgPath, []string{"foodora", "config", "set", "--country", "AT"}, "")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	orig := browserOAuthTokenPassword
	defer func() { browserOAuthTokenPassword = orig }()

	browserOAuthTokenPassword = func(ctx context.Context, req foodora.OAuthPasswordRequest, opts browserauth.PasswordOptions) (foodora.AuthToken, *foodora.MfaChallenge, browserauth.Session, error) {
		return foodora.AuthToken{AccessToken: "a", RefreshToken: "r", ExpiresIn: 3600}, nil, browserauth.Session{
			Host:         "mj.fd-api.com",
			CookieHeader: "cf=1",
			UserAgent:    "Android-app-25.3.0(250300134)",
		}, nil
	}

	out, _, err := runCLI(cfgPath, []string{"foodora", "login", "--email", "a@example.com", "--password", "pw", "--browser", "--wait-for-otp=false"}, "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected out=%q", out)
	}

	out, _, err = runCLI(cfgPath, []string{"foodora", "config", "show"}, "")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(out, "cookies_by_host=*** (1)") || !strings.Contains(out, "http_user_agent=") {
		t.Fatalf("unexpected out: %s", out)
	}
}
