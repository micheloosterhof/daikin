// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Tests for fetching the adapter's power consumption history:
// ABOUTME: daily figures for the last week and monthly figures per year.
package daikin

import (
	"slices"
	"testing"
)

func TestWeekPower(t *testing.T) {
	c, req := newTestClient(t, "ret=OK,today_runtime=133,datas=7000/6200/8300/7900/5200/3300/3500")
	wp, err := c.WeekPower(t.Context())
	if err != nil {
		t.Fatalf("WeekPower returned error: %v", err)
	}
	if req.URL.Path != "/aircon/get_week_power" {
		t.Errorf("path = %q, want /aircon/get_week_power", req.URL.Path)
	}
	if wp.TodayRuntime != 133 {
		t.Errorf("TodayRuntime = %d, want 133", wp.TodayRuntime)
	}
	wantDays := []int{7000, 6200, 8300, 7900, 5200, 3300, 3500}
	if !slices.Equal(wp.Days, wantDays) {
		t.Errorf("Days = %v, want %v", wp.Days, wantDays)
	}
}

func TestWeekPowerMalformed(t *testing.T) {
	c, _ := newTestClient(t, "ret=OK,today_runtime=133,datas=7000/x/8300")
	if _, err := c.WeekPower(t.Context()); err == nil {
		t.Fatal("expected error for non-numeric datas")
	}
}

func TestYearPower(t *testing.T) {
	c, req := newTestClient(t, "ret=OK,previous_year=87/114/148/115/242/246/187/145/154/187/190/105,this_year=106/113/185/256/206/84/158")
	yp, err := c.YearPower(t.Context())
	if err != nil {
		t.Fatalf("YearPower returned error: %v", err)
	}
	if req.URL.Path != "/aircon/get_year_power" {
		t.Errorf("path = %q, want /aircon/get_year_power", req.URL.Path)
	}
	wantPrev := []int{87, 114, 148, 115, 242, 246, 187, 145, 154, 187, 190, 105}
	if !slices.Equal(yp.PreviousYear, wantPrev) {
		t.Errorf("PreviousYear = %v, want %v", yp.PreviousYear, wantPrev)
	}
	wantThis := []int{106, 113, 185, 256, 206, 84, 158}
	if !slices.Equal(yp.ThisYear, wantThis) {
		t.Errorf("ThisYear = %v, want %v", yp.ThisYear, wantThis)
	}
}
