# Homebridge plugin for zcangate: design

## Purpose

Expose the Zehnder ComfoAir ventilation unit — currently controllable only via
zcangate's HTTP API (`/measurements`, `/commands`) — as a HomeKit accessory, so
it can be controlled from the Home app and Siri via Homebridge.

## Scope (MVP)

- Single accessory: the ventilation fan (speed + auto/manual mode).
- Read-only reconciliation of fan speed/on-off from polled measurements.
- Out of scope for this iteration: bypass control, temperature/humidity
  sensors, multi-accessory platform support. These can be added later as
  separate specs if desired.

## Architecture

A new `homebridge-zcangate/` directory in this repo: a self-contained
Node/TypeScript package with its own `package.json`, compiled via `tsc` to
`dist/`. It is a Homebridge **accessory plugin** (not a platform plugin) that
registers exactly one accessory, using HomeKit's **Fanv2** service. It talks
to the existing Go server purely over HTTP; no changes to the Go code are
required.

### Components

- `src/index.ts` — registers the accessory class with the Homebridge API
  (`homebridge.registerAccessory(...)`).
- `src/zcangateClient.ts` — thin HTTP client:
  - `getMeasurements(): Promise<Record<string, number>>` — GET `/measurements`.
  - `runCommand(name: string): Promise<void>` — POST `/commands/{name}`, with
    `Authorization: Bearer <token>` header when a token is configured.
  - Wraps `fetch` with a request timeout; throws a typed error on
    network failure, timeout, or non-2xx response.
- `src/levelMapping.ts` — pure functions converting HomeKit `RotationSpeed`
  (0–100) to/from `ventilation_level_0`..`_3` commands. Kept separate from the
  accessory so it's trivially unit-testable without mocking HTTP or HAP.
- `src/zcangateAccessory.ts` — the Fanv2 accessory: wires up `Active`,
  `RotationSpeed`, `TargetFanState`, `CurrentFanState` characteristics, holds
  the cached/optimistic state described below, and runs the poll loop.
- `config.schema.json` — Homebridge UI X form (see Configuration below).

## HomeKit mapping

| HomeKit characteristic | Direction | zcangate mapping |
|---|---|---|
| `Active` (0/1) | write | `0` → `ventilation_level_0`. `1` → `ventilation_level_N` for the last remembered non-zero level (default `1` if none yet); also updates `RotationSpeed` to match. |
| `Active` (0/1) | read | Derived each poll from `ventilation_level`: `0` → `0`, `>0` → `1`. |
| `RotationSpeed` (0–100) | write | Quantized to the nearest of 4 buckets — 0/33/66/100 — mapped to `ventilation_level_0..3`. If the cached `TargetFanState` is currently `AUTO`, first sends `manual_mode` and flips the cached target state to `MANUAL` before sending the level command (adjusting speed implies taking manual control). |
| `RotationSpeed` (0–100) | read | Derived each poll from `ventilation_level` via the same bucket table (0/33/66/100). |
| `TargetFanState` (`AUTO`/`MANUAL`) | write | `AUTO` → `auto_mode` command. `MANUAL` → `manual_mode` command. Cached locally. |
| `TargetFanState` (`AUTO`/`MANUAL`) | read | **Not reconciled from polling.** Returns the last locally cached value (see Known limitation below). |
| `CurrentFanState` (`INACTIVE`/`IDLE`/`BLOWING_AIR`) | read | Derived each poll from `ventilation_level`: `0` → `INACTIVE`, `>0` → `BLOWING_AIR`. |

### Bucket table (`levelMapping.ts`)

| `ventilation_level` | `RotationSpeed` (read) | `RotationSpeed` range accepted on write |
|---|---|---|
| 0 | 0 | 0–16 |
| 1 | 33 | 17–49 |
| 2 | 66 | 50–83 |
| 3 | 100 | 84–100 |

### Known limitation: Auto/Manual is not read back from the device

`can/mapping.go` has no confirmed, labeled measurement for "current
auto/manual mode." The closest candidate, PDU 225 (`comfocontrol_mode`), is
unverified and sits among several `z_Unknown_*` fields, so it is not treated
as authoritative. `TargetFanState` is therefore **optimistically cached**:
it reflects the last mode this plugin itself set (via HomeKit or via the
auto-to-manual switch triggered by a speed change), not necessarily the
device's true current mode if it was changed by another controller (e.g. the
physical remote). This will be documented prominently in the plugin's
README.

## Error handling

- On HTTP failure (network error, timeout, non-2xx) during a characteristic
  `set` or on-demand `get`, the handler throws a HAP `HapStatusError` with
  `SERVICE_COMMUNICATION_FAILURE`, so the Home app shows "No Response"
  instead of silently accepting a wrong value.
- Poll-loop failures are logged via the Homebridge logger and skipped; the
  last known cached state is retained until the next successful poll (no
  characteristic is force-updated to an error state on a single missed
  poll).

## Configuration

`config.schema.json` (Homebridge UI X form):

| Key | Type | Required | Default | Notes |
|---|---|---|---|---|
| `name` | string | no | `"Ventilation"` | Accessory display name. |
| `apiBaseUrl` | string | yes | — | e.g. `http://raspberrypi:8080`. |
| `authToken` | string | no | *(unset)* | Sent as `Authorization: Bearer <token>` on command requests only when set; omitted entirely otherwise. Matches zcangate's `COMMAND_AUTH_TOKEN`. |
| `pollInterval` | number (seconds) | no | `30` | Minimum `5`. |

## Testing

Unit tests with Jest + ts-jest, no real HTTP/hardware involved:

- `levelMapping.ts` — percentage↔level bucketing, including boundary values
  (16/17/49/50/83/84/100).
- `zcangateAccessory.ts` — characteristic get/set handlers with
  `zcangateClient` mocked. Covers: quantization → correct command sent,
  auto→manual switch on speed change, `Active` on/off remembering last
  non-zero level, HTTP failure → `HapStatusError` thrown, poll loop updating
  cached state.
- `zcangateClient.ts` — request construction (URL, headers, bearer token
  present/absent), timeout behavior, error mapping — using a mocked `fetch`.

No integration/e2e tests against real hardware are included; the plugin
author will manually verify against a live unit. This limitation is noted in
the plugin's README.

## Out of scope / future work

- Bypass control (Switch accessory).
- Temperature/humidity sensor accessories.
- Converting to a platform plugin if multiple accessories are added later.
- Native HomeKit mid-drag slider snapping for `RotationSpeed` (see the final
  addendum below — two attempts both regressed, considered not reliably
  achievable in the current Home app).

## Addendum (2026-08-03): TargetFanState reconciliation resolved

The "Known limitation" above has been resolved. A `DEBUG_CAN` capture from
real hardware showed that CAN PDU 72 — previously undecoded in
`can/mapping.go` (the mapping jumped from PDU 71 to PDU 81) — flips cleanly
between `00` and `01` in lockstep with `auto_mode`/`manual_mode` commands,
confirmed across two full toggle cycles in both directions.

- `can/mapping.go` now decodes PDU 72 as `ventilation_control_mode`
  (`0`=auto, `1`=manual), so it's included in `/measurements`.
- `zcangateAccessory.poll()` reads this field and reconciles
  `TargetFanState`, inverting polarity since HAP's `TargetFanState` is
  `MANUAL=0`/`AUTO=1` — the opposite of the device's encoding.
- Because zcangate passively listens to the whole CAN bus, this also
  surfaces mode changes made via the physical remote or official app, not
  just changes made through this plugin — a stronger result than the
  original spec anticipated.

`TargetFanState` is therefore no longer purely a local optimistic cache;
it converges to the device's real state on the next successful poll.

## Addendum (2026-08-03): RotationSpeed changed from percentage to raw level

A later revision briefly had `RotationSpeed` report the bucket-table
percentages above (0/33/66/100) with `setProps({ minValue: 0, maxValue: 100,
minStep: 100/3 })`. This caused a persistent loading spinner on the fan tile
in the Home app: the declared step grid (`0, 33.33.., 66.67.., 100`) never
exactly matched the rounded integer percentages the plugin actually pushed
(`0, 33, 66, 100`), so HomeKit could never confirm the value was on-grid.

The fix, following the pattern used by other multi-speed Homebridge fan
plugins (e.g. `homebridge-tuya-lan`): `RotationSpeed` now reports the raw
speed level directly — `setProps({ minValue: 0, maxValue: 3, minStep: 1 })` —
and `getRotationSpeed`/`setRotationSpeed` operate on that 0–3 scale with no
percentage conversion. HomeKit itself computes the displayed percentage from
the characteristic's min/max range, so the Home app still shows 0%/33%/67%/
100%, but every value the plugin pushes is a plain integer that trivially
satisfies the integer step — eliminating the grid mismatch.

`levelToPercent`/`percentToLevel` in `levelMapping.ts` were removed as a
result (only `levelToCommand` remains); the "Bucket table" section above and
the `RotationSpeed` rows in the HomeKit mapping table are superseded by this
addendum.

## Addendum (2026-08-03): raw-level RotationSpeed reverted — HomeKit range wasn't rescaled

The previous addendum's fix did not work in practice. Live testing showed the
Home app slider still displayed a 0–100% range and snapped to 0%/1% —
confirming that the Home app does **not** rescale `RotationSpeed`'s displayed
percentage based on a custom `maxValue`; it always treats the raw value as a
literal percentage. The `homebridge-tuya-lan` reference this fix was based on
came from an AI-summarized description of that plugin's PR, not a verified
read of its actual behavior, and turned out not to generalize here. Restarting
the Home app and fully removing/re-adding the accessory (ruling out client-side
metadata caching) did not change the result either.

`RotationSpeed` has been reverted to the original percentage-based design
(0–100%, `levelToPercent`/`percentToLevel` restored, no `setProps` override
at all). Between this and the prior addendum's fractional-`minStep` spinner
bug, two independent attempts at a native HomeKit "hard snap" for this
characteristic both produced worse regressions than the plain percentage
approach. Native mid-drag snapping via declared characteristic properties is
therefore considered **not reliably achievable** for this characteristic in
the current Home app, and is out of scope going forward (see "Out of scope /
future work"). The slider drags smoothly; quantization still happens
server-side in `setRotationSpeed`, so the value settles on the nearest valid
percentage immediately after release.
