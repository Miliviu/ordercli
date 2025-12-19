package config

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Version          int       `json:"version"`
	BaseURL          string    `json:"base_url"`
	GlobalEntityID   string    `json:"global_entity_id,omitempty"`
	TargetCountryISO string    `json:"target_country_iso,omitempty"`
	DeviceID         string    `json:"device_id"`
	AccessToken      string    `json:"access_token,omitempty"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	ClientSecret     string    `json:"client_secret,omitempty"`
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "foodoracli", "config.json"), nil
}

func Load(path string) (Config, error) {
	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.DeviceID == "" {
		cfg.DeviceID = newDeviceID()
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.DeviceID == "" {
		cfg.DeviceID = newDeviceID()
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func New() Config {
	return Config{
		Version:  1,
		DeviceID: newDeviceID(),
	}
}

func (c Config) HasSession() bool {
	return c.AccessToken != "" && c.RefreshToken != ""
}

func (c Config) TokenLikelyExpired(now time.Time) bool {
	if c.AccessToken == "" {
		return true
	}
	if c.ExpiresAt.IsZero() {
		return false
	}
	return !c.ExpiresAt.After(now.Add(30 * time.Second))
}

func newDeviceID() string {
	// UUIDv4-ish; good enough for header X-Device.
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("dev-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}
