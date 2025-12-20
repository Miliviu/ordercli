package cli

import (
	"testing"
)

func TestAsString(t *testing.T) {
	if asString(nil) != "" {
		t.Fatalf("expected empty")
	}
	if asString(true) != "true" {
		t.Fatalf("unexpected")
	}
	if asString(float64(3)) != "3" {
		t.Fatalf("unexpected")
	}
	if asString(float64(3.5)) != "3.5" {
		t.Fatalf("unexpected")
	}
}

func TestAsInt(t *testing.T) {
	if asInt(float64(3.9)) != 3 {
		t.Fatalf("unexpected")
	}
	if asInt(" 12 ") != 12 {
		t.Fatalf("unexpected")
	}
}

func TestNested(t *testing.T) {
	m := map[string]any{"a": map[string]any{"b": map[string]any{"c": "x"}}}
	if got := nested(m, "a", "b", "c"); got != "x" {
		t.Fatalf("got %v", got)
	}
	if nested(m, "a", "nope") != nil {
		t.Fatalf("expected nil")
	}
}

func TestFormatMoney(t *testing.T) {
	if formatMoney(float64(0)) != "" {
		t.Fatalf("expected empty")
	}
	if formatMoney(float64(1.2)) != "1.20" {
		t.Fatalf("unexpected")
	}
	if formatMoney("0.00") != "" {
		t.Fatalf("expected empty")
	}
	if formatMoney("2.50") != "2.50" {
		t.Fatalf("unexpected")
	}
}

func TestFormatOrderTime(t *testing.T) {
	if formatOrderTime(nil) != "" {
		t.Fatalf("expected empty")
	}
	// unix seconds
	if got := formatOrderTime(float64(1_700_000_000)); got == "" {
		t.Fatalf("expected non-empty")
	}
	// unix millis
	if got := formatOrderTime(float64(1_700_000_000_000)); got == "" {
		t.Fatalf("expected non-empty")
	}
}
