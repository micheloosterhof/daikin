// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Tests for the daikin CLI entry point, exercising run() against a
// ABOUTME: temp config dir and a local TLS server standing in for an adapter.
package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		{"daikin"},
		{"daikin", "bogus"},
		{"daikin", "register", "onlyip"},
		{"daikin", "status", "one", "two"},
		{"daikin", "set"},
		{"daikin", "power", "one", "two"},
	} {
		var out, errb bytes.Buffer
		if err := run(t.Context(), args, &out, &errb); !errors.Is(err, errUsage) {
			t.Errorf("run(%v) = %v, want errUsage", args, err)
		}
	}
}

func TestRunStatusNoDevices(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out, errb bytes.Buffer
	err := run(t.Context(), []string{"daikin", "status"}, &out, &errb)
	if err == nil || !strings.Contains(err.Error(), "no registered devices") {
		t.Fatalf("run status = %v, want no registered devices error", err)
	}
}

// newAdapterServer stands in for a BRP072C adapter and records the query of
// the most recent set_control_info request.
func newAdapterServer(t *testing.T) (*httptest.Server, *url.Values) {
	t.Helper()
	lastSet := &url.Values{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body string
		switch r.URL.Path {
		case "/common/register_terminal":
			body = "ret=OK"
		case "/common/basic_info":
			body = "ret=OK,pow=1,name=%4f%66%66%69%63%65,mac=A841F4D64496"
		case "/aircon/get_control_info":
			body = "ret=OK,pow=1,mode=3,stemp=23.0,shum=0,f_rate=A,f_dir=0"
		case "/aircon/get_sensor_info":
			body = "ret=OK,htemp=24.0,hhum=65,otemp=30.0,cmpfreq=24"
		case "/aircon/set_control_info":
			*lastSet = r.URL.Query()
			body = "ret=OK"
		case "/aircon/get_week_power":
			body = "ret=OK,today_runtime=133,datas=7000/6200/8300/7900/5200/3300/3500"
		case "/aircon/get_year_power":
			body = "ret=OK,previous_year=87/114/148/115/242/246/187/145/154/187/190/105,this_year=106/113/185/256/206/84/158"
		default:
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, lastSet
}

func TestRunPower(t *testing.T) {
	confHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", confHome)
	srv, _ := newAdapterServer(t)
	addr := srv.Listener.Addr().String()

	var out, errb bytes.Buffer
	if err := run(t.Context(), []string{"daikin", "register", addr, "0512852162213"}, &out, &errb); err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	out.Reset()
	if err := run(t.Context(), []string{"daikin", "power", "Office"}, &out, &errb); err != nil {
		t.Fatalf("power returned error: %v", err)
	}
	// today 133 min and 3500 Wh, 41400 Wh over 7 days, 1108 kWh this year,
	// 1920 kWh previous year
	for _, want := range []string{"Office", "2h13m", "3500", "41400", "1108", "1920"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("power output = %q, missing %q", out.String(), want)
		}
	}
}

func TestRunRegisterStatusSet(t *testing.T) {
	confHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", confHome)
	srv, lastSet := newAdapterServer(t)
	addr := srv.Listener.Addr().String()

	var out, errb bytes.Buffer
	if err := run(t.Context(), []string{"daikin", "register", addr, "0512852162213"}, &out, &errb); err != nil {
		t.Fatalf("register returned error: %v", err)
	}
	if !strings.Contains(out.String(), `registered "Office"`) {
		t.Errorf("register output = %q, want registered \"Office\"", out.String())
	}
	if _, err := os.Stat(filepath.Join(confHome, "daikin", "config.json")); err != nil {
		t.Errorf("config file not written: %v", err)
	}

	out.Reset()
	if err := run(t.Context(), []string{"daikin", "status", "Office"}, &out, &errb); err != nil {
		t.Fatalf("status returned error: %v", err)
	}
	for _, want := range []string{"Office", "cool", "24.0"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("status output = %q, missing %q", out.String(), want)
		}
	}

	out.Reset()
	if err := run(t.Context(), []string{"daikin", "set", "Office", "-power", "off"}, &out, &errb); err != nil {
		t.Fatalf("set returned error: %v", err)
	}
	if got := lastSet.Get("pow"); got != "0" {
		t.Errorf("set_control_info pow = %q, want 0", got)
	}
	// Unchanged settings are merged in from the current state.
	if got := lastSet.Get("mode"); got != "3" {
		t.Errorf("set_control_info mode = %q, want 3", got)
	}
}
