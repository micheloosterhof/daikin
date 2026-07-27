// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Command-line tool to discover, register and control Daikin
// ABOUTME: air conditioners with BRP072C wifi adapters on the local network.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/micheloosterhof/daikin"
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: daikin <command> [arguments]

commands:
  discover                 scan the local network for Daikin units
  register <ip> <key>      register with a unit using its 13-digit key
  status [name]            show status of one unit, or all registered units
  set <name> [flags]       change settings on a unit

set flags:
  -power on|off
  -mode auto|dry|cool|heat|fan
  -temp <celsius>
  -fan auto|quiet|1|2|3|4|5
  -swing off|vertical|horizontal|both
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "discover":
		err = cmdDiscover()
	case "register":
		err = cmdRegister(os.Args[2:])
	case "status":
		err = cmdStatus(os.Args[2:])
	case "set":
		err = cmdSet(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "daikin: %v\n", err)
		os.Exit(1)
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

func cmdDiscover() error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	devices, err := daikin.Discover(3 * time.Second)
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		fmt.Println("no Daikin units found")
		return nil
	}
	for _, d := range devices {
		registered := ""
		if _, ok := cfg.FindDevice(d.Name); ok {
			registered = " (registered)"
		}
		fmt.Printf("%-20s %-15s %s%s\n", d.Name, d.IP, d.MAC, registered)
	}
	return nil
}

func cmdRegister(args []string) error {
	if len(args) != 2 {
		usage()
	}
	ip, key := args[0], args[1]
	cfg, path, err := loadConfig()
	if err != nil {
		return err
	}
	cfg.EnsureUUID()
	client := daikin.NewClient(ip, cfg.UUID)
	if err := client.Register(key); err != nil {
		return fmt.Errorf("registering with %s: %w", ip, err)
	}
	dev, err := client.BasicInfo()
	if err != nil {
		return err
	}
	cfg.UpsertDevice(daikin.ConfigDevice{Name: dev.Name, IP: ip, Key: key})
	if err := cfg.Save(path); err != nil {
		return err
	}
	fmt.Printf("registered %q at %s\n", dev.Name, ip)
	return nil
}

func printStatus(cfg *daikin.Config, dev daikin.ConfigDevice) error {
	client := daikin.NewClient(dev.IP, cfg.UUID)
	ci, err := client.ControlInfo()
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	si, err := client.SensorInfo()
	if err != nil {
		return fmt.Errorf("%s: %w", dev.Name, err)
	}
	power := "off"
	if ci.Power {
		power = "on"
	}
	fmt.Printf("%-20s %-4s %-5s set %s°C  room %s°C %s%%  outside %s°C  fan %s  swing %s\n",
		dev.Name, power, daikin.ModeName(ci.Mode), ci.SetTemp,
		si.HomeTemp, si.HomeHumidity, si.OutsideTemp,
		daikin.FanRateName(ci.FanRate), daikin.SwingName(ci.FanDir))
	return nil
}

func cmdStatus(args []string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(args) > 1 {
		usage()
	}
	if len(args) == 1 {
		dev, ok := cfg.FindDevice(args[0])
		if !ok {
			return fmt.Errorf("no registered device %q", args[0])
		}
		return printStatus(cfg, dev)
	}
	if len(cfg.Devices) == 0 {
		return fmt.Errorf("no registered devices; run: daikin register <ip> <key>")
	}
	failed := 0
	for _, dev := range cfg.Devices {
		if err := printStatus(cfg, dev); err != nil {
			fmt.Fprintf(os.Stderr, "daikin: %v\n", err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d devices did not respond", failed, len(cfg.Devices))
	}
	return nil
}

func cmdSet(args []string) error {
	fs := flag.NewFlagSet("set", flag.ExitOnError)
	power := fs.String("power", "", "on or off")
	mode := fs.String("mode", "", "auto, dry, cool, heat or fan")
	temp := fs.String("temp", "", "target temperature in celsius")
	fan := fs.String("fan", "", "auto, quiet or 1-5")
	swing := fs.String("swing", "", "off, vertical, horizontal or both")
	if len(args) < 1 {
		usage()
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
	ci, err := client.ControlInfo()
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

	if err := client.SetControl(ci); err != nil {
		return err
	}
	return printStatus(cfg, dev)
}
