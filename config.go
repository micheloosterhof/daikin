// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: JSON config store for the CLI: the client UUID shared across all
// ABOUTME: adapters and the list of registered devices with their keys.
package daikin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ConfigDevice is one registered unit: its display name, current IP and the
// 13-digit key printed on the adapter.
type ConfigDevice struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Key  string `json:"key"`
}

// Config holds everything the CLI persists between runs.
type Config struct {
	UUID    string         `json:"uuid"`
	Devices []ConfigDevice `json:"devices"`
}

// DefaultConfigPath returns $XDG_CONFIG_HOME/daikin/config.json, falling back
// to ~/.config/daikin/config.json.
func DefaultConfigPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "daikin", "config.json"), nil
}

// LoadConfig reads the config at path. A missing file yields an empty config.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // reading the caller-chosen config path is this function's purpose
	if errors.Is(err, fs.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save writes the config to path, creating parent directories as needed.
func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// EnsureUUID generates a client UUID if the config does not have one yet.
func (c *Config) EnsureUUID() {
	if c.UUID != "" {
		return
	}
	b := make([]byte, 16)
	rand.Read(b)
	c.UUID = hex.EncodeToString(b)
}

// FindDevice looks a device up by case-insensitive name or by IP.
func (c *Config) FindDevice(query string) (ConfigDevice, bool) {
	for _, d := range c.Devices {
		if strings.EqualFold(d.Name, query) || d.IP == query {
			return d, true
		}
	}
	return ConfigDevice{}, false
}

// UpsertDevice adds the device or, when a device with the same name exists,
// replaces it.
func (c *Config) UpsertDevice(dev ConfigDevice) {
	for i, d := range c.Devices {
		if strings.EqualFold(d.Name, dev.Name) {
			c.Devices[i] = dev
			return
		}
	}
	c.Devices = append(c.Devices, dev)
}
