<!--
SPDX-FileCopyrightText: 2026 Michel Oosterhof <michel@oosterhof.net>
SPDX-License-Identifier: CC0-1.0
-->

# daikin

Command-line tool and Go library to discover and control Daikin air
conditioners with BRP072C wifi adapters on the local network.

## Install

```
go install github.com/micheloosterhof/daikin/cmd/daikin@latest
```

## Usage

```
daikin discover                 # scan the local network for units
daikin register <ip> <key>      # register with a unit (13-digit key on the adapter sticker)
daikin status [name]            # show one unit, or all registered units
daikin set <name> [flags]       # change settings
```

`set` flags: `-power on|off`, `-mode auto|dry|cool|heat|fan`, `-temp <celsius>`,
`-fan auto|quiet|1..5`, `-swing off|vertical|horizontal|both`.

Registration generates a client UUID, registers it with the adapter over HTTPS
using the unit's key, and stores both in `~/.config/daikin/config.json`. It
only needs to be done once per unit.

## Protocol notes

- Units are discovered by broadcasting `DAIKIN_UDP/common/basic_info` on UDP
  port 30050.
- These adapters serve HTTPS only (self-signed certificate) and require an
  `X-Daikin-uuid` header on every request — with that exact casing; the
  firmware rejects the canonical `X-Daikin-Uuid` form with HTTP 403.
- `set_control_info` requires the full setting set (power, mode, temperature,
  humidity, fan rate, fan direction), so the CLI reads the current state and
  merges changes into it.
