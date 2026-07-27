// ABOUTME: Tests for the JSON config store holding the client UUID and the
// ABOUTME: registered devices with their keys.
package daikin

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.json")
	c := &Config{
		UUID: testUUID,
		Devices: []ConfigDevice{
			{Name: "Office", IP: "192.168.179.237", Key: "0512852162213"},
		},
	}
	if err := c.Save(path); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if got.UUID != c.UUID || len(got.Devices) != 1 || got.Devices[0] != c.Devices[0] {
		t.Errorf("LoadConfig = %+v, want %+v", got, c)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	c, err := LoadConfig(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("LoadConfig on missing file returned error: %v", err)
	}
	if len(c.Devices) != 0 {
		t.Errorf("expected empty config, got %+v", c)
	}
}

func TestEnsureUUID(t *testing.T) {
	c := &Config{}
	c.EnsureUUID()
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(c.UUID) {
		t.Errorf("EnsureUUID produced %q, want 32 lowercase hex chars", c.UUID)
	}
	first := c.UUID
	c.EnsureUUID()
	if c.UUID != first {
		t.Error("EnsureUUID must not replace an existing UUID")
	}
}

func TestFindDevice(t *testing.T) {
	c := &Config{Devices: []ConfigDevice{
		{Name: "Office", IP: "192.168.179.237", Key: "k1"},
		{Name: "Kids Bedroom", IP: "192.168.179.150", Key: "k2"},
	}}
	for _, q := range []string{"office", "Office", "192.168.179.237"} {
		d, ok := c.FindDevice(q)
		if !ok || d.Name != "Office" {
			t.Errorf("FindDevice(%q) = %+v, %v; want Office", q, d, ok)
		}
	}
	if _, ok := c.FindDevice("garage"); ok {
		t.Error("FindDevice(garage) should not match")
	}
}

func TestUpsertDevice(t *testing.T) {
	c := &Config{}
	c.UpsertDevice(ConfigDevice{Name: "Office", IP: "192.168.179.237", Key: "k1"})
	c.UpsertDevice(ConfigDevice{Name: "Office", IP: "192.168.179.99", Key: "k1"})
	if len(c.Devices) != 1 {
		t.Fatalf("expected 1 device after upsert, got %d", len(c.Devices))
	}
	if c.Devices[0].IP != "192.168.179.99" {
		t.Errorf("upsert did not update IP: %+v", c.Devices[0])
	}
}
