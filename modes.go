// SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
// SPDX-License-Identifier: BSD-3-Clause

// ABOUTME: Translates between human-friendly names for mode, fan rate and
// ABOUTME: swing and the numeric codes the adapter protocol uses.
package daikin

import "fmt"

var (
	modeCodes = map[string]string{"auto": "0", "dry": "2", "cool": "3", "heat": "4", "fan": "6"}
	modeNames = map[string]string{"0": "auto", "1": "auto", "7": "auto", "2": "dry", "3": "cool", "4": "heat", "6": "fan"}

	fanRateCodes = map[string]string{"auto": "A", "quiet": "B", "1": "3", "2": "4", "3": "5", "4": "6", "5": "7"}
	fanRateNames = map[string]string{"A": "auto", "B": "quiet", "3": "1", "4": "2", "5": "3", "6": "4", "7": "5"}

	swingCodes = map[string]string{"off": "0", "vertical": "1", "horizontal": "2", "both": "3"}
	swingNames = map[string]string{"0": "off", "1": "vertical", "2": "horizontal", "3": "both"}
)

func lookup(kind string, m map[string]string, key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("unknown %s %q", kind, key)
	}
	return v, nil
}

// ModeCode converts a mode name (auto, dry, cool, heat, fan) to its code.
func ModeCode(name string) (string, error) { return lookup("mode", modeCodes, name) }

// ModeName converts a mode code to its name, or returns the code verbatim.
func ModeName(code string) string {
	if n, ok := modeNames[code]; ok {
		return n
	}
	return code
}

// FanRateCode converts a fan rate name (auto, quiet, 1-5) to its code.
func FanRateCode(name string) (string, error) { return lookup("fan rate", fanRateCodes, name) }

// FanRateName converts a fan rate code to its name, or returns the code verbatim.
func FanRateName(code string) string {
	if n, ok := fanRateNames[code]; ok {
		return n
	}
	return code
}

// SwingCode converts a swing name (off, vertical, horizontal, both) to its code.
func SwingCode(name string) (string, error) { return lookup("swing", swingCodes, name) }

// SwingName converts a swing code to its name, or returns the code verbatim.
func SwingName(code string) string {
	if n, ok := swingNames[code]; ok {
		return n
	}
	return code
}
