# homebridge-zcangate

Homebridge accessory plugin that exposes a Zehnder ComfoAir ventilation unit
(via [zcangate](../README.md)) as a HomeKit fan, so it can be controlled from
the Home app and Siri.

## Installation

Once published to npm (see [Publishing](#publishing) below), install it like
any other Homebridge plugin — either via Homebridge UI X's plugin search, or:

```sh
npm install -g homebridge-zcangate
```

### Installing from source (unpublished changes)

This plugin lives in a subdirectory of the `zcangate` repo, so a plain
`npm install github:marco-hoyer/zcangate` won't find its `package.json`.
To run it from a local checkout instead:

```sh
cd homebridge-zcangate
npm install
npm run build
npm link
```

Then in your Homebridge installation's directory: `npm link homebridge-zcangate`.

## Publishing

For maintainers cutting a new release:

```sh
cd homebridge-zcangate
npm version patch   # or minor / major — also runs `npm run build` via prepublishOnly
npm publish
git push --follow-tags
```

`npm version` bumps the version in `package.json`, and `prepublishOnly`
ensures `dist/` is rebuilt from the latest source before `npm publish` packs
the tarball. The `files` field in `package.json` keeps the published package
limited to `dist/`, `config.schema.json`, `README.md`, and `LICENSE` — no
source, tests, or dev config ship to consumers.

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
