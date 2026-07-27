// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Fetches the adapter's power consumption history: today's runtime,
// ABOUTME: daily figures for the last week and monthly figures per year.
package daikin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// WeekPower holds the last week of consumption. On units without electricity
// metering (elec=0 in model info) the figures are firmware estimates.
type WeekPower struct {
	TodayRuntime int   // minutes the unit ran today
	Days         []int // daily consumption in Wh, oldest first, today last
}

// YearPower holds monthly consumption in kWh, January first. ThisYear runs
// up to the current month.
type YearPower struct {
	PreviousYear []int
	ThisYear     []int
}

func parseIntList(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, "/")
	vals := make([]int, len(parts))
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("malformed number list %q", s)
		}
		vals[i] = v
	}
	return vals, nil
}

// WeekPower fetches today's runtime and the daily consumption of the last week.
func (c *Client) WeekPower(ctx context.Context) (WeekPower, error) {
	kv, err := c.get(ctx, "/aircon/get_week_power", nil)
	if err != nil {
		return WeekPower{}, err
	}
	runtime, err := strconv.Atoi(kv["today_runtime"])
	if err != nil {
		return WeekPower{}, fmt.Errorf("malformed today_runtime %q", kv["today_runtime"])
	}
	days, err := parseIntList(kv["datas"])
	if err != nil {
		return WeekPower{}, err
	}
	return WeekPower{TodayRuntime: runtime, Days: days}, nil
}

// YearPower fetches monthly consumption for this year and the previous year.
func (c *Client) YearPower(ctx context.Context) (YearPower, error) {
	kv, err := c.get(ctx, "/aircon/get_year_power", nil)
	if err != nil {
		return YearPower{}, err
	}
	prev, err := parseIntList(kv["previous_year"])
	if err != nil {
		return YearPower{}, err
	}
	this, err := parseIntList(kv["this_year"])
	if err != nil {
		return YearPower{}, err
	}
	return YearPower{PreviousYear: prev, ThisYear: this}, nil
}
