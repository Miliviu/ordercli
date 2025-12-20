package foodora

import (
	"testing"
	"time"
)

func TestFormatReorderTime(t *testing.T) {
	tt := time.Date(2025, 12, 20, 1, 2, 3, 0, time.FixedZone("X", 3600))
	if got, want := FormatReorderTime(tt), "2025-12-20T01:02:03+0100"; got != want {
		t.Fatalf("FormatReorderTime() = %q, want %q", got, want)
	}
}
