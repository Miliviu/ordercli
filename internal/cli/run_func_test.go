package cli

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRun_Help(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := Run(context.Background(), []string{"--config", cfgPath, "--help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
