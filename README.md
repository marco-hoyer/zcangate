# zcangate

CAN bus gateway for Zehnder ComfoAir ventilation units. Connects to an SLCAN serial adapter, decodes CN1F measurement frames, stores them in InfluxDB, and exposes an HTTP API for reading current values and sending control commands.

## How it works

The ComfoAir unit speaks a proprietary CAN protocol (CN1F) on top of a standard CAN bus. This gateway:

1. Opens the SLCAN serial adapter and puts it into active mode.
2. Reads CAN frames and decodes them into named measurements (temperatures, fan speeds, etc.).
3. Publishes measurements to InfluxDB for time-series storage and dashboarding.
4. Serves an HTTP API so other systems can poll current values or send commands.

Device node IDs are discovered automatically from the heartbeat frames the device emits — no static configuration needed.

## Hardware requirements

To connect a host computer to the ComfoAir CAN bus you need an MCP2515-based CAN module, such as [this one](https://amzn.to/4uq0rov). The MCP2515 is a standalone CAN controller with SPI interface, typically paired with a TJA1050 CAN transceiver on the same board.

**Wiring:**
- Connect the module to the host via SPI (CS, MOSI, MISO, SCLK) and wire INT to a free GPIO pin.
- Connect CANH and CANL to the ComfoAir CAN bus terminals.
- Power the module from 5 V.

**SLCAN setup on Linux (e.g. Raspberry Pi):**

The software speaks the SLCAN serial protocol, so the MCP2515 must be exposed as a serial CAN interface using `slcand`:

```sh
# Load the kernel driver (add to /etc/modules for persistence)
sudo modprobe mcp251x

# Create the SLCAN interface
sudo slcand -o -s6 -t hw -S 115200 /dev/spidev0.0 can0
sudo ip link set can0 up

# Create a virtual serial port that zcangate can open
sudo slcan_attach -f -s6 -o /dev/ttyACM0 can0
```

Set `SERIAL_PORT` to the resulting device path (default `/dev/ttyACM0`).

## Building

```
go build -o zcangate .
```

## Configuration

All settings are read from environment variables at startup.

| Variable             | Default          | Description                                               |
|----------------------|------------------|-----------------------------------------------------------|
| `SERIAL_PORT`        | `/dev/ttyACM0`   | Serial port of the SLCAN adapter                          |
| `SERIAL_BAUD`        | `115200`         | Baud rate for the serial connection                       |
| `HTTP_ADDR`          | `:8080`          | TCP address the HTTP server listens on                    |
| `HEARTBEAT_TIMEOUT`  | `60s`            | Time without a measurement before the process exits (Go duration string, e.g. `90s`, `2m`) |
| `INFLUXDB_URL`       | *(required)*     | InfluxDB server URL, e.g. `http://influxdb:8086`          |
| `INFLUXDB_TOKEN`     | *(required)*     | InfluxDB authentication token                             |
| `INFLUXDB_ORG`       | `home`           | InfluxDB organisation                                     |
| `INFLUXDB_BUCKET`    | `home-metrics`   | InfluxDB bucket                                           |
| `INFLUXDB_LOCATION`  | `home`           | Value written to the `location` tag on every point        |
| `COMMAND_AUTH_TOKEN` | *(unset)*        | Bearer token required to execute commands. Auth is disabled when unset. |
| `DEBUG_CAN`          | *(unset)*        | Set to any non-empty value to enable verbose CAN frame logging |

## HTTP API

| Method | Path                          | Description                          |
|--------|-------------------------------|--------------------------------------|
| GET    | `/measurements`               | All current measurements as JSON     |
| GET    | `/measurements/{name}`        | Single measurement by name           |
| GET    | `/commands`                   | List of available command names      |
| POST   | `/commands/{name}`            | Execute a command by name — requires `Authorization: Bearer <token>` header when `COMMAND_AUTH_TOKEN` is set |

### Available commands

| Command                                 | Description                              |
|-----------------------------------------|------------------------------------------|
| `auto_mode`                             | Switch to automatic ventilation mode     |
| `manual_mode`                           | Switch to manual ventilation mode        |
| `ventilation_level_0`–`_3`             | Set fan speed (0 = off, 3 = high)        |
| `ventilation_mode_balanced`             | Balanced supply and exhaust              |
| `ventilation_mode_supply_only_1h`       | Supply-only mode for 1 hour              |
| `ventilation_mode_outlet_only_1h`       | Exhaust-only mode for 1 hour             |
| `temperature_profile_normal`            | Normal temperature profile               |
| `temperature_profile_cool`              | Cool temperature profile                 |
| `temperature_profile_warm`              | Warm temperature profile                 |
| `bypass_auto`                           | Automatic bypass control                 |
| `bypass_activated_1h`                   | Open bypass for 1 hour                   |
| `bypass_deactivated_1h`                 | Close bypass for 1 hour                  |
| `passive_temperature_control_off`       | Disable passive temperature control      |
| `passive_temperature_control_auto_only` | Passive temperature control: auto only   |
| `passive_temperature_control_on`        | Enable passive temperature control       |
| `passive_humidity_control_off`          | Disable passive humidity control         |
| `passive_humidity_control_auto_only`    | Passive humidity control: auto only      |
| `passive_humidity_control_on`           | Enable passive humidity control          |
| `humidity_protection_off`               | Disable humidity protection              |
| `humidity_protection_auto_only`         | Humidity protection: auto only           |
| `humidity_protection_on`                | Enable humidity protection               |
