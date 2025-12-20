package config

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfig_Deliveroo(t *testing.T) {
	cfg := New()
	d := cfg.Deliveroo()
	if d == nil {
		t.Fatalf("expected deliveroo config")
	}
	d.Market = "uk"
	if cfg.Providers.Deliveroo == nil || cfg.Providers.Deliveroo.Market != "uk" {
		t.Fatalf("unexpected: %#v", cfg)
	}
}

func TestLegacyPaths(t *testing.T) {
	p1, err := LegacyPathFoodcli()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasSuffix(p1, string(filepath.Separator)+"foodcli"+string(filepath.Separator)+"config.json") {
		t.Fatalf("unexpected: %q", p1)
	}

	p2, err := LegacyPathFoodoracli()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.HasSuffix(p2, string(filepath.Separator)+"foodoracli"+string(filepath.Separator)+"config.json") {
		t.Fatalf("unexpected: %q", p2)
	}
}

func TestFoodoraConfig_AccessTokenExpiresAt(t *testing.T) {
	payload := map[string]any{"exp": float64(123)}
	b, _ := json.Marshal(payload)
	tok := "x." + base64.RawURLEncoding.EncodeToString(b) + ".y"

	c := FoodoraConfig{AccessToken: tok}
	exp, ok := c.AccessTokenExpiresAt()
	if !ok || exp.Unix() != 123 {
		t.Fatalf("exp=%v ok=%v", exp, ok)
	}
}

func TestJWTExpiry_BadInputs(t *testing.T) {
	if _, ok := jwtExpiry(""); ok {
		t.Fatalf("expected false")
	}
	if _, ok := jwtExpiry("nope"); ok {
		t.Fatalf("expected false")
	}
	if _, ok := jwtExpiry("a.b.c"); ok {
		t.Fatalf("expected false")
	}
}

func TestFoodoraConfig_TokenLikelyExpired_UsesJWT(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	payload := map[string]any{"exp": float64(now.Add(10 * time.Minute).Unix())}
	b, _ := json.Marshal(payload)
	tok := "x." + base64.RawURLEncoding.EncodeToString(b) + ".y"

	c := FoodoraConfig{AccessToken: tok}
	if c.TokenLikelyExpired(now) {
		t.Fatalf("expected not expired")
	}
}
