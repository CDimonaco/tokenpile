package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigDir       string
	DataDir         string
	DBPath          string
	PricingOverride string
	IdentityKeyPath string
	IdentityPubPath string
	CredentialsPath string
}

func Resolve() Paths {
	cfg := configDir()
	data := dataDir()

	return Paths{
		ConfigDir:       cfg,
		DataDir:         data,
		DBPath:          filepath.Join(data, "tokenpile.db"),
		PricingOverride: filepath.Join(cfg, "pricing.yaml"),
		IdentityKeyPath: filepath.Join(cfg, "identity.key"),
		IdentityPubPath: filepath.Join(cfg, "identity.pub"),
		CredentialsPath: filepath.Join(cfg, "credentials"),
	}
}

func configDir() string {
	if dir := os.Getenv("TOKENPILE_CONFIG_DIR"); dir != "" {
		return dir
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tokenpile")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "tokenpile")
	}

	return filepath.Join(".", ".config", "tokenpile")
}

func dataDir() string {
	if dir := os.Getenv("TOKENPILE_DATA_DIR"); dir != "" {
		return dir
	}

	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "tokenpile")
	}

	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "tokenpile")
	}

	return filepath.Join(".", ".local", "share", "tokenpile")
}

func EnsureDirs(p Paths) error {
	for _, dir := range []string{p.ConfigDir, p.DataDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	return nil
}
