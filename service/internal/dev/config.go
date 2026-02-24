package dev

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultIDPKID       = "dev-idp-rs256"
	defaultKASRSAKID    = "kas-rsa-1"
	defaultKASECKID     = "kas-ec-1"
	defaultIssuerHost   = "http://127.0.0.1"
	defaultListenHost   = "127.0.0.1"
	defaultTokenTTL     = time.Hour
	defaultClientID     = "opentdf-dev"
	defaultAdminRole    = "opentdf-admin"
	defaultStandardRole = "opentdf-standard"
)

type Options struct {
	IDPPort      int
	PlatformPort int
	Regenerate   bool
}

type IDPConfig struct {
	Issuer   string        `yaml:"issuer"`
	Listen   string        `yaml:"listen"`
	Audience string        `yaml:"audience"`
	TokenTTL time.Duration `yaml:"token_ttl"`
	Key      IDPKeyConfig  `yaml:"key"`
	Clients  []IDPClient   `yaml:"clients"`
}

type IDPKeyConfig struct {
	KID        string `yaml:"kid"`
	PrivateKey string `yaml:"private_key"`
}

type IDPClient struct {
	ID     string   `yaml:"id"`
	Secret string   `yaml:"secret"`
	Roles  []string `yaml:"roles"`
	Scopes []string `yaml:"scopes"`
}

func EnsureConfigs(layout Layout, opts Options) (IDPConfig, error) {
	rootKey, err := ensureRootKey(layout.RootKeyPath, opts.Regenerate)
	if err != nil {
		return IDPConfig{}, err
	}

	idpConfig, err := ensureIDPConfig(layout, opts)
	if err != nil {
		return IDPConfig{}, err
	}

	if err := ensurePlatformConfig(layout, opts, rootKey, idpConfig); err != nil {
		return IDPConfig{}, err
	}

	return idpConfig, nil
}

func LoadIDPConfig(path string) (IDPConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return IDPConfig{}, err
	}

	var cfg IDPConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return IDPConfig{}, err
	}

	return cfg, nil
}

func ensureIDPConfig(layout Layout, opts Options) (IDPConfig, error) {
	if !opts.Regenerate {
		if _, err := os.Stat(layout.IDPConfigPath); err == nil {
			return LoadIDPConfig(layout.IDPConfigPath)
		}
	}

	if err := ensureIDPKey(layout.IDPPrivateKeyPath, opts.Regenerate); err != nil {
		return IDPConfig{}, err
	}

	secret, err := randomSecret(defaultSecretBytes)
	if err != nil {
		return IDPConfig{}, err
	}

	idpConfig := IDPConfig{
		Issuer:   fmt.Sprintf("%s:%d", defaultIssuerHost, opts.IDPPort),
		Listen:   fmt.Sprintf("%s:%d", defaultListenHost, opts.IDPPort),
		Audience: fmt.Sprintf("%s:%d", defaultIssuerHost, opts.PlatformPort),
		TokenTTL: defaultTokenTTL,
		Key: IDPKeyConfig{
			KID:        defaultIDPKID,
			PrivateKey: layout.IDPPrivateKeyPath,
		},
		Clients: []IDPClient{
			{
				ID:     defaultClientID,
				Secret: secret,
				Roles:  []string{defaultAdminRole, defaultStandardRole},
			},
		},
	}

	if err := writeYAML(layout.IDPConfigPath, idpConfig, defaultFilePerm); err != nil {
		return IDPConfig{}, err
	}

	return idpConfig, nil
}

func ensurePlatformConfig(layout Layout, opts Options, rootKey string, idpConfig IDPConfig) error {
	if !opts.Regenerate {
		if _, err := os.Stat(layout.PlatformConfigPath); err == nil {
			return nil
		}
	}

	cfg := map[string]any{
		"logger": map[string]any{
			"level":  "info",
			"type":   "text",
			"output": "stderr",
		},
		"services": map[string]any{
			"kas": map[string]any{
				"registered_kas_uri": fmt.Sprintf("%s:%d", defaultIssuerHost, opts.PlatformPort),
				"root_key":           rootKey,
				"preview": map[string]any{
					"key_management": true,
				},
			},
			"entityresolution": map[string]any{
				"mode": "claims",
			},
		},
		"server": map[string]any{
			"public_hostname": "localhost",
			"port":            opts.PlatformPort,
			"tls": map[string]any{
				"enabled": false,
			},
			"auth": map[string]any{
				"enabled":     true,
				"enforceDPoP": false,
				"audience":    idpConfig.Audience,
				"issuer":      idpConfig.Issuer,
				"policy": map[string]any{
					"groups_claim":    "roles",
					"username_claim":  "sub",
					"client_id_claim": "azp",
					"extension": strings.Join([]string{
						"g, opentdf-admin, role:admin",
						"g, opentdf-standard, role:standard",
					}, "\n"),
				},
			},
		},
	}

	return writeYAML(layout.PlatformConfigPath, cfg, defaultFilePerm)
}

func ensureRootKey(path string, regenerate bool) (string, error) {
	if !regenerate {
		if raw, err := os.ReadFile(path); err == nil {
			return strings.TrimSpace(string(raw)), nil
		}
	}

	keyBytes := make([]byte, defaultRootKeyBytes)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("generate root key: %w", err)
	}
	rootKey := hex.EncodeToString(keyBytes)
	if err := os.WriteFile(path, []byte(rootKey+"\n"), defaultFilePerm); err != nil {
		return "", fmt.Errorf("write root key: %w", err)
	}
	return rootKey, nil
}

func LoadRootKey(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func ensureIDPKey(path string, regenerate bool) error {
	return EnsureRSAKeyPair(path, "", regenerate)
}

func writeYAML(path string, v any, perm os.FileMode) error {
	raw, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), defaultDirPerm); err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, perm); err != nil {
		return err
	}
	return nil
}

func randomSecret(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("secret size must be positive")
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
