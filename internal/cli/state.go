package cli

import (
	"errors"
	"os"

	"github.com/steipete/foodoracli/internal/config"
)

type state struct {
	configPath string
	cfg        config.Config
	dirty      bool
}

func (s *state) load() error {
	if s.configPath == "" {
		p, err := config.DefaultPath()
		if err != nil {
			return err
		}
		s.configPath = p
	}
	cfg, err := config.Load(s.configPath)
	if err != nil {
		return err
	}
	s.cfg = cfg
	return nil
}

func (s *state) save() error {
	if !s.dirty {
		return nil
	}
	if s.configPath == "" {
		return errors.New("internal: configPath unset")
	}
	if err := config.Save(s.configPath, s.cfg); err != nil {
		return err
	}
	s.dirty = false
	return nil
}

func (s *state) markDirty() { s.dirty = true }

func (s *state) clientSecretOrEnv() string {
	if s.cfg.ClientSecret != "" {
		return s.cfg.ClientSecret
	}
	return os.Getenv("FOODORA_CLIENT_SECRET")
}
