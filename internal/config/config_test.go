package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := New()
	cfg.BaseURL = "https://hu.fd-api.com/api/v5/"
	cfg.AccessToken = "a"
	cfg.RefreshToken = "r"
	cfg.ExpiresAt = time.Unix(123, 0).UTC()

	if err := Save(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	if st, err := os.Stat(path); err != nil {
		t.Fatalf("stat: %v", err)
	} else if st.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 perms, got %o", st.Mode().Perm())
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.BaseURL != cfg.BaseURL || got.AccessToken != "a" || got.RefreshToken != "r" {
		t.Fatalf("unexpected cfg: %#v", got)
	}
	if got.DeviceID == "" {
		t.Fatalf("expected device id")
	}
}
