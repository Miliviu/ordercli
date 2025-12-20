package cli

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestParseCookieHeader(t *testing.T) {
	m := parseCookieHeader("a=1; b=2 ;  c=3")
	if m["a"] != "1" || m["b"] != "2" || m["c"] != "3" {
		t.Fatalf("unexpected: %#v", m)
	}
}

func TestJWTExpiryAndClientID(t *testing.T) {
	payload := map[string]any{
		"exp":       float64(123),
		"client_id": "android",
	}
	b, _ := json.Marshal(payload)
	tok := "x." + base64.RawURLEncoding.EncodeToString(b) + ".y"

	if exp, ok := jwtExpiry(tok); !ok || exp != 123 {
		t.Fatalf("exp=%d ok=%v", exp, ok)
	}
	if cid, ok := jwtClientID(tok); !ok || cid != "android" {
		t.Fatalf("cid=%q ok=%v", cid, ok)
	}
}
