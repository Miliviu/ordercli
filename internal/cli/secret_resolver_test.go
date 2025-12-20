package cli

import (
	"context"
	"os"
	"testing"

	"github.com/steipete/ordercli/internal/config"
)

func TestResolveClientSecret_FromConfig(t *testing.T) {
	st := &state{cfg: config.New()}
	cfg := st.foodora()
	cfg.ClientSecret = "s"
	cfg.OAuthClientID = "android"

	sec, err := st.resolveClientSecret(context.Background(), "android")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sec.Secret != "s" || !sec.FromConfig {
		t.Fatalf("unexpected: %#v", sec)
	}
}

func TestResolveClientSecret_FromEnv(t *testing.T) {
	st := &state{cfg: config.New()}
	cfg := st.foodora()
	cfg.ClientSecret = ""

	old, had := os.LookupEnv("FOODORA_CLIENT_SECRET")
	_ = os.Setenv("FOODORA_CLIENT_SECRET", "envs")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("FOODORA_CLIENT_SECRET", old)
		} else {
			_ = os.Unsetenv("FOODORA_CLIENT_SECRET")
		}
	})

	sec, err := st.resolveClientSecret(context.Background(), "android")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if sec.Secret != "envs" || !sec.FromEnv {
		t.Fatalf("unexpected: %#v", sec)
	}
}

func TestRemoteConfigKeyCandidates(t *testing.T) {
	st := &state{cfg: config.New()}
	cfg := st.foodora()
	cfg.BaseURL = "https://mj.fd-api.com/api/v5/"
	cfg.TargetCountryISO = "AT"

	keys := st.remoteConfigKeyCandidates()
	if len(keys) < 2 {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}
