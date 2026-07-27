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
daikin power [name]             # show energy consumption
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

## API endpoints

The BRP072C42 (firmware 1.16.0, `adp_kind=3`, `en_secure=1`) keeps most of the
HTTP API known from the earlier BRP069/BRP072A adapters, moved to HTTPS with
the registered-UUID header. Endpoints verified present:

| Endpoint | Returns |
|----------|---------|
| `/common/basic_info` | identity, name, MAC, power, firmware version |
| `/common/get_remote_method` | polling configuration |
| `/common/get_notify` | auto-off settings |
| `/common/get_datetime` | adapter clock and region |
| `/common/get_wifi_setting` | home-network SSID and key |
| `/common/get_holiday` | holiday mode state |
| `/aircon/get_model_info` | model capabilities (`elec=0`: no power metering) |
| `/aircon/get_control_info` | settable state |
| `/aircon/get_sensor_info` | room/outside temperature, humidity, compressor frequency |
| `/aircon/get_price` | configured electricity price |
| `/aircon/get_target` | consumption target |
| `/aircon/get_week_power` | today's runtime (min) and 7 daily figures (Wh, today last) |
| `/aircon/get_year_power` | monthly figures (kWh, January first) for this and last year |
| `/aircon/get_day_power_ex`, `get_week_power_ex`, `get_year_power_ex` | heat/cool split consumption |
| `/aircon/get_monitordata` | raw internal state |

Removed relative to the older firmware: `/aircon/get_timer`, `/aircon/get_program`
and `/aircon/get_scdltimer` all return 404 — on-device scheduling moved to the
Daikin cloud service.

On units without electricity metering (`elec=0`) the consumption figures are
firmware estimates, not measurements.

## Security notes

- The adapter's own access point (`DaikinAP…` SSID, `adp_mode=ap_run`) cannot
  be disabled on the BRP072C42: no API endpoint controls it, the buttons only
  switch connection modes, and Daikin documents the AP as permanently on for
  this model. The AP is WPA2-protected with the per-unit key on the sticker.
- `/common/get_wifi_setting` returns the home network's wifi password in
  cleartext to any registered UUID. Anyone who can read the 13-digit key off
  an adapter's sticker can register and retrieve it.
