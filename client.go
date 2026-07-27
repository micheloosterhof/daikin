// ABOUTME: HTTPS client for the Daikin BRP072C wifi adapter: registration,
// ABOUTME: control and sensor queries authenticated by an X-Daikin-uuid header.
package daikin

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client talks to one Daikin adapter over HTTPS. The adapter only accepts
// requests carrying a UUID previously registered with the unit's key.
type Client struct {
	host    string
	baseURL string
	uuid    string
	http    *http.Client
}

// ControlInfo holds the settable state of a unit. Temperature and humidity
// are kept as strings because the adapter reports non-numeric values such as
// "M" and "--" in dry and fan modes.
type ControlInfo struct {
	Power       bool
	Mode        string
	SetTemp     string
	SetHumidity string
	FanRate     string
	FanDir      string
}

// SensorInfo holds the read-only sensor state of a unit.
type SensorInfo struct {
	HomeTemp       string
	HomeHumidity   string
	OutsideTemp    string
	CompressorFreq string
}

// NewClient returns a client for the adapter at host. The adapter uses a
// self-signed certificate, so verification is disabled.
func NewClient(host, uuid string) *Client {
	return &Client{
		host:    host,
		baseURL: "https://" + host,
		uuid:    uuid,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *Client) get(path string, query url.Values) (map[string]string, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Daikin-uuid", c.uuid)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d (is the UUID registered?)", path, resp.StatusCode)
	}
	kv, err := ParseResponse(string(body))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return kv, nil
}

// Register associates the client's UUID with the adapter using the 13-digit
// key printed on the unit. This only needs to succeed once per adapter.
func (c *Client) Register(key string) error {
	_, err := c.get("/common/register_terminal", url.Values{"key": {key}})
	return err
}

// BasicInfo fetches the unit's identity as a Device.
func (c *Client) BasicInfo() (Device, error) {
	kv, err := c.get("/common/basic_info", nil)
	if err != nil {
		return Device{}, err
	}
	name, err := DecodeName(kv["name"])
	if err != nil {
		return Device{}, err
	}
	return Device{IP: c.host, Name: name, MAC: kv["mac"], PowerOn: kv["pow"] == "1"}, nil
}

// ControlInfo fetches the unit's current settable state.
func (c *Client) ControlInfo() (ControlInfo, error) {
	kv, err := c.get("/aircon/get_control_info", nil)
	if err != nil {
		return ControlInfo{}, err
	}
	return ControlInfo{
		Power:       kv["pow"] == "1",
		Mode:        kv["mode"],
		SetTemp:     kv["stemp"],
		SetHumidity: kv["shum"],
		FanRate:     kv["f_rate"],
		FanDir:      kv["f_dir"],
	}, nil
}

// SensorInfo fetches the unit's sensor readings.
func (c *Client) SensorInfo() (SensorInfo, error) {
	kv, err := c.get("/aircon/get_sensor_info", nil)
	if err != nil {
		return SensorInfo{}, err
	}
	return SensorInfo{
		HomeTemp:       kv["htemp"],
		HomeHumidity:   kv["hhum"],
		OutsideTemp:    kv["otemp"],
		CompressorFreq: kv["cmpfreq"],
	}, nil
}

// SetControl applies the given state to the unit. The adapter requires all
// fields on every set, so callers should start from the current ControlInfo.
func (c *Client) SetControl(ci ControlInfo) error {
	pow := "0"
	if ci.Power {
		pow = "1"
	}
	_, err := c.get("/aircon/set_control_info", url.Values{
		"pow":    {pow},
		"mode":   {ci.Mode},
		"stemp":  {ci.SetTemp},
		"shum":   {ci.SetHumidity},
		"f_rate": {ci.FanRate},
		"f_dir":  {ci.FanDir},
	})
	return err
}
