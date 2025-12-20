package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steipete/ordercli/internal/chromecookies"
)

func TestCookiesChromeCmd_StoresCookies(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")

	_, _, err := runCLI(cfgPath, []string{"foodora", "config", "set", "--country", "AT"}, "")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	orig := chromeLoadCookieHeader
	defer func() { chromeLoadCookieHeader = orig }()
	chromeLoadCookieHeader = func(ctx context.Context, opts chromecookies.Options) (chromecookies.Result, error) {
		return chromecookies.Result{CookieHeader: "cf=1", CookieCount: 1}, nil
	}

	out, _, err := runCLI(cfgPath, []string{"foodora", "cookies", "chrome", "--url", "https://www.foodora.at/"}, "")
	if err != nil {
		t.Fatalf("cookies chrome: %v", err)
	}
	if !strings.Contains(out, "ok host=mj.fd-api.com cookies=1") {
		t.Fatalf("unexpected out: %s", out)
	}

	out, _, err = runCLI(cfgPath, []string{"foodora", "config", "show"}, "")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(out, "cookies_by_host=*** (1)") {
		t.Fatalf("unexpected out: %s", out)
	}
}

func TestSessionChromeCmd_ImportsTokens(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")

	_, _, err := runCLI(cfgPath, []string{"foodora", "config", "set", "--country", "AT"}, "")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}

	jwt := func(payload map[string]any) string {
		b, _ := json.Marshal(payload)
		return "x." + base64.RawURLEncoding.EncodeToString(b) + ".y"
	}

	orig := chromeLoadCookieHeader
	defer func() { chromeLoadCookieHeader = orig }()
	chromeLoadCookieHeader = func(ctx context.Context, opts chromecookies.Options) (chromecookies.Result, error) {
		access := jwt(map[string]any{"client_id": "android", "exp": float64(2000)})
		return chromecookies.Result{
			CookieHeader: "token=" + access + "; refresh_token=ref; device_token=dev",
			CookieCount:  3,
		}, nil
	}

	out, _, err := runCLI(cfgPath, []string{"foodora", "session", "chrome", "--url", "https://www.foodora.at/"}, "")
	if err != nil {
		t.Fatalf("session chrome: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected out: %q", out)
	}

	out, _, err = runCLI(cfgPath, []string{"foodora", "config", "show"}, "")
	if err != nil {
		t.Fatalf("config show: %v", err)
	}
	if !strings.Contains(out, "access_token=***") || !strings.Contains(out, "refresh_token=***") || !strings.Contains(out, "oauth_client_id=android") {
		t.Fatalf("unexpected out: %s", out)
	}
}
