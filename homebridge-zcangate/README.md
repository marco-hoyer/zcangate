# homebridge-zcangate

Homebridge accessory plugin that exposes a Zehnder ComfoAir ventilation unit
(via [zcangate](../README.md)) as a HomeKit fan, so it can be controlled from
the Home app and Siri.

## Installation

From this directory:

```sh
npm install
npm run build
```

Then either `npm link` it into your Homebridge installation, or copy the
`homebridge-zcangate` directory (with its built `dist/`) into Homebridge's
`node_modules`.

## Configuration

Add an entry to Homebridge's `config.json` `accessories` array (or use
Homebridge UI X, which reads `config.schema.json` to render a form):

```json
{
  "accessory": "ZcangateVentilation",
  "name": "Ventilation",
  "apiBaseUrl": "http://raspberrypi:8080",
  "authToken": "",
  "pollInterval": 30
}
```

| Key | Type | Required | Default | Description |
|---|---|---|---|---|
| `name` | string | no | `Ventilation` | Accessory display name in the Home app. |
| `apiBaseUrl` | string | yes | — | Base URL of the zcangate HTTP server, e.g. `http://raspberrypi:8080`. |
| `authToken` | string | no | *(unset)* | Sent as `Authorization: Bearer <token>` on command requests only. Matches zcangate's `COMMAND_AUTH_TOKEN`. Leave unset if that's not configured on the server. |
| `pollInterval` | number (seconds) | no | `30` | How often to poll `/measurements` to refresh HomeKit state. Minimum `5`. |

## HomeKit mapping

| HomeKit characteristic | Behavior |
|---|---|
| `Active` (on/off) | Off sends `ventilation_level_0`. On sends `ventilation_level_N` for the last remembered non-zero speed (default level 1). |
| `RotationSpeed` (0–100%) | Quantized to the nearest of 4 buckets — 0/33/66/100 — mapped to `ventilation_level_0`.._3`. Changing speed while in Auto mode first switches to Manual. |
| `TargetFanState` (Auto/Manual) | Sends the `auto_mode` / `manual_mode` commands. Reconciled from the polled `ventilation_control_mode` measurement (`0`=auto, `1`=manual — the inverse of HAP's `TargetFanState`). |
| `CurrentFanState` (read-only) | Derived from the polled fan speed. |

Because zcangate passively listens to the whole CAN bus, `ventilation_control_mode` also reflects mode changes made from the physical remote or the official app, not just changes made through this plugin.
