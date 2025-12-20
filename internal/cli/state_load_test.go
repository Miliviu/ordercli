package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/steipete/ordercli/internal/config"
)

func TestStateLoad_MigratesLegacyConfig(t *testing.T) {
	tmpHome := t.TempDir()
	withEnvMap(t, map[string]string{
		"HOME":            tmpHome,
		"XDG_CONFIG_HOME": filepath.Join(tmpHome, ".config"),
	})

	legacyPath, err := config.LegacyPathFoodcli()
	if err != nil {
		t.Fatalf("legacy path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacyCfg := config.FoodoraConfig{
		BaseURL:      "https://hu.fd-api.com/api/v5/",
		AccessToken:  "a",
		RefreshToken: "r",
	}
	if err := config.Save(legacyPath, config.Config{Providers: config.Providers{Foodora: &legacyCfg}, Version: 1}); err != nil {
		t.Fatalf("save legacy: %v", err)
	}

	var st state
	if err := st.load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if !st.dirty {
		t.Fatalf("expected dirty migration")
	}
	if st.configPath == "" {
		t.Fatalf("expected configPath set")
	}

	if err := st.save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(st.configPath); err != nil {
		t.Fatalf("expected new config file: %v", err)
	}
}

func withEnvMap(t *testing.T, m map[string]string) {
	t.Helper()
	old := map[string]string{}
	had := map[string]bool{}
	for k, v := range m {
		if ov, ok := os.LookupEnv(k); ok {
			old[k] = ov
			had[k] = true
		}
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv %s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k := range m {
			if had[k] {
				_ = os.Setenv(k, old[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})
}
