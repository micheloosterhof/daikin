// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Command-line tool to discover, register and control Daikin
// ABOUTME: air conditioners with BRP072C wifi adapters on the local network.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/micheloosterhof/daikin"
)

// errUsage signals that the arguments did not name a valid invocation;
// main responds by printing usage and exiting with status 2.
var errUsage = errors.New("usage")

func usage(w io.Writer) {
	fmt.Fprintf(w, `usage: daikin <command> [arguments]

commands:
  discover                 scan the local network for Daikin units
  register <ip> <key>      register with a unit using its 13-digit key
  status [name]            show status of one unit, or all registered units
  set <name> [flags]       change settings on a unit
  power [name]             show energy consumption of one unit, or all units

set flags:
  -power on|off
  -mode auto|dry|cool|heat|fan
  -temp <celsius>
  -fan auto|quiet|1|2|3|4|5
  -swing off|vertical|horizontal|both
`)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	stop()
	if errors.Is(err, errUsage) {
		usage(os.Stderr)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "daikin: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 {
		return errUsage
	}
	switch args[1] {
	case "discover":
		return cmdDiscover(ctx, stdout)
	case "register":
		return cmdRegister(ctx, args[2:], stdout)
	case "status":
		return cmdStatus(ctx, args[2:], stdout, stderr)
	case "set":
		return cmdSet(ctx, args[2:], stdout, stderr)
	case "power":
		return cmdPower(ctx, args[2:], stdout, stderr)
	default:
		return errUsage
	}
}

func loadConfig() (*daikin.Config, string, error) {
	path, err := daikin.DefaultConfigPath()
	if err != nil {
		return nil, "", err
	}
	cfg, err := daikin.LoadConfig(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func cmdDiscover(ctx context.Context, stdout io.Writer) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	devices, err := daikin.Discover(ctx, 3*time.Second)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		fmt.Fprintln(stdout, "no Daikin units found")
		return nil
	}
	for _, d := range devices {
		registered := ""
		if _, ok := cfg.FindDevice(d.Name); ok {
			registered = " (registered)"
		}
		fmt.Fprintf(stdout, "%-20s %-15s %s%s\n", d.Name, d.IP, d.MAC, registered)
	}
	return nil
}

func cmdRegister(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 2 {
		return errUsage
	}
	ip, key := args[0], args[1]
	cfg, path, err := loadConfig()
	if err != nil {
		return err
	}
	cfg.EnsureUUID()
	client := daikin.NewClient(ip, cfg.UUID)
	if err := client.Register(ctx, key); err != nil {
		return fmt.Errorf("registering with %s: %w", ip, err)
	}
	dev, err := client.BasicInfo(ctx)
	if err != nil {
		return err
	}
	cfg.UpsertDevice(daikin.ConfigDevice{Name: dev.Name, IP: ip, Key: key})
	if err := cfg.Save(path); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "registered %q at %s\n", dev.Name, ip)
	return nil
}

func printStatus(ctx context.Context, cfg *daikin.Config, dev daikin.ConfigDevice, stdout io.Writer) error {
	client := daikin.NewClient(dev.IP, cfg.UUID)
	ci, err := client.ControlInfo(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	si, err := client.SensorInfo(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	power := "off"
	if ci.Power {
		power = "on"
	}
	fmt.Fprintf(stdout, "%-20s %-4s %-5s set %s°C  room %s°C %s%%  outside %s°C  fan %s  swing %s\n",
		dev.Name, power, daikin.ModeName(ci.Mode), ci.SetTemp,
		si.HomeTemp, si.HomeHumidity, si.OutsideTemp,
		daikin.FanRateName(ci.FanRate), daikin.SwingName(ci.FanDir))
	return nil
}

// runOnDevices runs f against the device named in args, or against every
// registered device when args is empty, reporting per-device failures to
// stderr and failing if any device did not respond.
func runOnDevices(args []string, stderr io.Writer, f func(*daikin.Config, daikin.ConfigDevice) error) error {
	if len(args) > 1 {
		return errUsage
	}
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(args) == 1 {
		name := args[0]
		dev, ok := cfg.FindDevice(name)
		if !ok {
			return fmt.Errorf("no registered device %q", name)
		}
		return f(cfg, dev)
	}
	if len(cfg.Devices) == 0 {
		return fmt.Errorf("no registered devices; run: daikin register <ip> <key>")
	}
	failed := 0
	for _, dev := range cfg.Devices {
		if err := f(cfg, dev); err != nil {
			fmt.Fprintf(stderr, "daikin: %v\n", err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d devices did not respond", failed, len(cfg.Devices))
	}
	return nil
}

func cmdStatus(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return runOnDevices(args, stderr, func(cfg *daikin.Config, dev daikin.ConfigDevice) error {
		return printStatus(ctx, cfg, dev, stdout)
	})
}

// printPower shows consumption figures; on units without electricity
// metering these are firmware estimates.
func printPower(ctx context.Context, cfg *daikin.Config, dev daikin.ConfigDevice, stdout io.Writer) error {
	client := daikin.NewClient(dev.IP, cfg.UUID)
	wp, err := client.WeekPower(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	yp, err := client.YearPower(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	today, week := 0, 0
	for _, wh := range wp.Days {
		week += wh
	}
	if n := len(wp.Days); n > 0 {
		today = wp.Days[n-1]
	}
	year, prevYear := 0, 0
	for _, kwh := range yp.ThisYear {
		year += kwh
	}
	for _, kwh := range yp.PreviousYear {
		prevYear += kwh
	}
	fmt.Fprintf(stdout, "%-20s today %dh%02dm %d Wh  7-day %d Wh  year %d kWh  prev year %d kWh\n",
		dev.Name, wp.TodayRuntime/60, wp.TodayRuntime%60, today, week, year, prevYear)
	return nil
}

func cmdPower(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return runOnDevices(args, stderr, func(cfg *daikin.Config, dev daikin.ConfigDevice) error {
		return printPower(ctx, cfg, dev, stdout)
	})
}

func cmdSet(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(stderr)
	power := fs.String("power", "", "on or off")
	mode := fs.String("mode", "", "auto, dry, cool, heat or fan")
	temp := fs.String("temp", "", "target temperature in celsius")
	fan := fs.String("fan", "", "auto, quiet or 1-5")
	swing := fs.String("swing", "", "off, vertical, horizontal or both")
	if len(args) < 1 {
		return errUsage
	}
	name := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	dev, ok := cfg.FindDevice(name)
	if !ok {
		return fmt.Errorf("no registered device %q", name)
	}
	client := daikin.NewClient(dev.IP, cfg.UUID)
	ci, err := client.ControlInfo(ctx)
	if err != nil {
		return err
	}

	switch *power {
	case "":
	case "on":
		ci.Power = true
	case "off":
		ci.Power = false
	default:
		return fmt.Errorf("invalid -power %q, want on or off", *power)
	}
	if *mode != "" {
		if ci.Mode, err = daikin.ModeCode(*mode); err != nil {
			return err
		}
	}
	if *temp != "" {
		ci.SetTemp = *temp
	}
	if *fan != "" {
		if ci.FanRate, err = daikin.FanRateCode(*fan); err != nil {
			return err
		}
	}
	if *swing != "" {
		if ci.FanDir, err = daikin.SwingCode(*swing); err != nil {
			return err
		}
	}

	if err := client.SetControl(ctx, ci); err != nil {
		return err
	}
	return printStatus(ctx, cfg, dev, stdout)
}
