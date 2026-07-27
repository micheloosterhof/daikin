// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Discovers Daikin wifi adapters on the local network by broadcasting
// ABOUTME: the DAIKIN_UDP basic_info probe on UDP port 30050.
package daikin

import (
	"fmt"
	"net"
	"sort"
	"time"
)

const (
	discoveryProbe = "DAIKIN_UDP/common/basic_info"
	discoveryPort  = 30050
)

// Device describes one Daikin unit found on the network.
type Device struct {
	IP      string
	Name    string
	MAC     string
	PowerOn bool
}

// DeviceFromBasicInfo builds a Device from a basic_info response body.
func DeviceFromBasicInfo(ip, body string) (Device, error) {
	kv, err := ParseResponse(body)
	if err != nil {
		return Device{}, fmt.Errorf("parsing basic_info from %s: %w", ip, err)
	}
	name, err := DecodeName(kv["name"])
	if err != nil {
		return Device{}, fmt.Errorf("decoding name from %s: %w", ip, err)
	}
	return Device{IP: ip, Name: name, MAC: kv["mac"], PowerOn: kv["pow"] == "1"}, nil
}

// Discover broadcasts the discovery probe and collects responses until the
// timeout elapses. Units that answer with unparseable payloads are skipped.
func Discover(timeout time.Duration) ([]Device, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	dst := &net.UDPAddr{IP: net.IPv4bcast, Port: discoveryPort}
	if _, err := conn.WriteTo([]byte(discoveryProbe), dst); err != nil {
		return nil, err
	}

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	var devices []Device
	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}
		ip, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			continue
		}
		dev, err := DeviceFromBasicInfo(ip, string(buf[:n]))
		if err != nil {
			continue
		}
		devices = append(devices, dev)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].Name < devices[j].Name })
	return devices, nil
}
