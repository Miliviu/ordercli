package cli

import (
	"testing"

	"github.com/steipete/ordercli/internal/config"
)

func TestCookieHost(t *testing.T) {
	if got := cookieHost("https://mj.fd-api.com/api/v5/"); got != "mj.fd-api.com" {
		t.Fatalf("got %q", got)
	}
	if got := cookieHost("not a url"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestAppHeaders_AT(t *testing.T) {
	st := &state{}
	st.cfg = config.New()
	cfg := st.foodora()
	cfg.TargetCountryISO = "AT"
	p := st.appHeaders()
	if p.AppName != "at.mjam" || p.FPAPIKey == "" || p.UserAgent == "" {
		t.Fatalf("unexpected: %#v", p)
	}
}
