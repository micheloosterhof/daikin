// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Tests for the HTTPS client that talks to a Daikin adapter using
// ABOUTME: the X-Daikin-uuid header, exercised against a local TLS test server.
package daikin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const testUUID = "7b9c9a4d3c8e4f0aa1b2c3d4e5f60718"

// newTestClient returns a Client aimed at a TLS test server and a pointer to
// the most recent request the server received.
func newTestClient(t *testing.T, response string) (*Client, *http.Request) {
	t.Helper()
	lastReq := &http.Request{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*lastReq = *r
		w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(srv.Listener.Addr().String(), testUUID)
	c.http = srv.Client()
	return c, lastReq
}

func TestRegister(t *testing.T) {
	c, req := newTestClient(t, "ret=OK")
	if err := c.Register("0512852162213"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if req.URL.Path != "/common/register_terminal" {
		t.Errorf("path = %q, want /common/register_terminal", req.URL.Path)
	}
	if got := req.URL.Query().Get("key"); got != "0512852162213" {
		t.Errorf("key = %q, want 0512852162213", got)
	}
	if got := req.Header.Get("X-Daikin-uuid"); got != testUUID {
		t.Errorf("X-Daikin-uuid = %q, want %q", got, testUUID)
	}
}

func TestControlInfo(t *testing.T) {
	c, req := newTestClient(t, "ret=OK,pow=1,mode=3,adv=,stemp=23.0,shum=0,f_rate=4,f_dir=2")
	ci, err := c.ControlInfo()
	if err != nil {
		t.Fatalf("ControlInfo returned error: %v", err)
	}
	if req.URL.Path != "/aircon/get_control_info" {
		t.Errorf("path = %q, want /aircon/get_control_info", req.URL.Path)
	}
	want := ControlInfo{Power: true, Mode: "3", SetTemp: "23.0", SetHumidity: "0", FanRate: "4", FanDir: "2"}
	if ci != want {
		t.Errorf("ControlInfo = %+v, want %+v", ci, want)
	}
}

func TestSensorInfo(t *testing.T) {
	c, _ := newTestClient(t, "ret=OK,htemp=24.0,hhum=65,otemp=30.0,err=0,cmpfreq=24")
	si, err := c.SensorInfo()
	if err != nil {
		t.Fatalf("SensorInfo returned error: %v", err)
	}
	want := SensorInfo{HomeTemp: "24.0", HomeHumidity: "65", OutsideTemp: "30.0", CompressorFreq: "24"}
	if si != want {
		t.Errorf("SensorInfo = %+v, want %+v", si, want)
	}
}

func TestSetControl(t *testing.T) {
	c, req := newTestClient(t, "ret=OK,adv=")
	ci := ControlInfo{Power: true, Mode: "3", SetTemp: "24.0", SetHumidity: "0", FanRate: "A", FanDir: "0"}
	if err := c.SetControl(ci); err != nil {
		t.Fatalf("SetControl returned error: %v", err)
	}
	if req.URL.Path != "/aircon/set_control_info" {
		t.Errorf("path = %q, want /aircon/set_control_info", req.URL.Path)
	}
	want := url.Values{"pow": {"1"}, "mode": {"3"}, "stemp": {"24.0"}, "shum": {"0"}, "f_rate": {"A"}, "f_dir": {"0"}}
	if got := req.URL.Query(); !urlValuesEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestBasicInfo(t *testing.T) {
	c, _ := newTestClient(t, officeBasicInfo)
	dev, err := c.BasicInfo()
	if err != nil {
		t.Fatalf("BasicInfo returned error: %v", err)
	}
	if dev.Name != "Office" || dev.MAC != "A841F4D64496" {
		t.Errorf("BasicInfo = %+v", dev)
	}
}

func TestRegisterAdapterError(t *testing.T) {
	c, _ := newTestClient(t, "ret=PARAM NG")
	if err := c.Register("badkey"); err == nil {
		t.Fatal("expected error for ret=PARAM NG")
	}
}

// The adapters drop connections under load, so a request that fails at the
// transport level is retried once.
func TestGetRetriesTransientError(t *testing.T) {
	calls := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Fatalf("hijack failed: %v", err)
			}
			conn.Close()
			return
		}
		w.Write([]byte("ret=OK,htemp=24.0,hhum=65,otemp=30.0,err=0,cmpfreq=24"))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(srv.Listener.Addr().String(), testUUID)
	c.http = srv.Client()
	if _, err := c.SensorInfo(); err != nil {
		t.Fatalf("SensorInfo should succeed after one retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("server saw %d calls, want 2", calls)
	}
}

// The adapter matches the UUID header name case-sensitively and rejects Go's
// canonical X-Daikin-Uuid form, so the header key must stay exactly as-is on
// the wire. Direct map access is intentional: Header.Get would canonicalize.
func TestUUIDHeaderCasing(t *testing.T) {
	c := NewClient("192.0.2.1", testUUID)
	req, err := c.newRequest("/common/basic_info", nil)
	if err != nil {
		t.Fatalf("newRequest returned error: %v", err)
	}
	if got := req.Header["X-Daikin-uuid"]; len(got) != 1 || got[0] != testUUID {
		t.Errorf("Header[X-Daikin-uuid] = %v, want [%s]", got, testUUID)
	}
	if _, ok := req.Header["X-Daikin-Uuid"]; ok {
		t.Error("header must not be stored under the canonical X-Daikin-Uuid key")
	}
}

func urlValuesEqual(a, b url.Values) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a.Get(k) != b.Get(k) {
			return false
		}
	}
	return true
}
