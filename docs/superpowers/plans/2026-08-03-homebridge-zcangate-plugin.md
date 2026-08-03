# Homebridge zcangate Plugin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `homebridge-zcangate`, a Homebridge accessory plugin that exposes the Zehnder ComfoAir ventilation unit as a HomeKit Fanv2 accessory, controlling it via zcangate's existing HTTP API.

**Architecture:** A self-contained TypeScript package in `homebridge-zcangate/` with no changes to the Go server. Pure mapping logic (`levelMapping.ts`), an HTTP client (`zcangateClient.ts`), the HAP-facing accessory (`zcangateAccessory.ts`), and a thin registration entry point (`index.ts` + `settings.ts`), each independently unit-tested with Jest/ts-jest.

**Tech Stack:** TypeScript 5, Node.js >=18 (global `fetch`), Homebridge API >=1.6.0 (accessory plugin, not platform), Jest + ts-jest for tests.

## Global Constraints

- MVP scope only: single Fanv2 accessory (speed + auto/manual). No bypass switch, no temperature/humidity sensors, no platform-plugin support — these are explicitly out of scope per the spec.
- `TargetFanState` is optimistically cached locally and is **not** reconciled from polling (no confirmed device measurement exists for it).
- `ventilation_level` is the only measurement field consumed from `/measurements`; it is the authoritative source for `Active`, `RotationSpeed`, and `CurrentFanState`.
- Speed bucket table: level 0 ↔ 0%, level 1 ↔ 33%, level 2 ↔ 66%, level 3 ↔ 100%. Write-side boundaries: 0–16→0, 17–49→1, 50–83→2, 84–100→3.
- `pollInterval` config default is 30 seconds, minimum 5.
- `authToken` header (`Authorization: Bearer <token>`) is sent only on command requests, and only when configured — never on `/measurements` GET, matching zcangate's `COMMAND_AUTH_TOKEN` behavior which only guards `POST /commands/{name}`.
- No integration/e2e tests against real hardware; all tests are unit tests with mocked HTTP/HAP.
- Plugin package lives at `homebridge-zcangate/` inside this repo; it is its own npm package with its own `package.json`, `tsconfig.json`, `.gitignore`.

---

### Task 1: Package scaffold + level/percentage mapping

**Files:**
- Create: `homebridge-zcangate/package.json`
- Create: `homebridge-zcangate/tsconfig.json`
- Create: `homebridge-zcangate/jest.config.js`
- Create: `homebridge-zcangate/.gitignore`
- Create: `homebridge-zcangate/src/levelMapping.ts`
- Test: `homebridge-zcangate/src/levelMapping.test.ts`

**Interfaces:**
- Produces (used by Task 3 `zcangateAccessory.ts`):
  - `levelToPercent(level: number): number`
  - `percentToLevel(percent: number): number`
  - `levelToCommand(level: number): string`

- [ ] **Step 1: Create the package scaffold files**

`homebridge-zcangate/package.json`:

```json
{
  "name": "homebridge-zcangate",
  "version": "0.1.0",
  "description": "Homebridge plugin to control a Zehnder ComfoAir ventilation unit via the zcangate HTTP API",
  "main": "dist/index.js",
  "engines": {
    "node": ">=18.0.0",
    "homebridge": ">=1.6.0"
  },
  "keywords": [
    "homebridge-plugin"
  ],
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "prepublishOnly": "npm run build"
  },
  "devDependencies": {
    "@types/jest": "^29.5.0",
    "@types/node": "^20.0.0",
    "homebridge": "^1.8.0",
    "jest": "^29.7.0",
    "ts-jest": "^29.1.0",
    "typescript": "^5.4.0"
  },
  "peerDependencies": {
    "homebridge": ">=1.6.0"
  }
}
```

`homebridge-zcangate/tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020", "DOM"],
    "declaration": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "node",
    "resolveJsonModule": true
  },
  "include": ["src/**/*.ts"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

`homebridge-zcangate/jest.config.js`:

```js
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testMatch: ['**/src/**/*.test.ts'],
};
```

`homebridge-zcangate/.gitignore`:

```
node_modules/
dist/
*.log
```

- [ ] **Step 2: Install dependencies**

Run: `cd homebridge-zcangate && npm install`
Expected: completes without error, creates `node_modules/` and `package-lock.json`.

- [ ] **Step 3: Write the failing test for level/percentage mapping**

`homebridge-zcangate/src/levelMapping.test.ts`:

```ts
import { levelToCommand, levelToPercent, percentToLevel } from './levelMapping';

describe('levelToPercent', () => {
  it.each([
    [0, 0],
    [1, 33],
    [2, 66],
    [3, 100],
  ])('maps level %i to %i%%', (level, percent) => {
    expect(levelToPercent(level)).toBe(percent);
  });

  it('defaults unknown levels to 0', () => {
    expect(levelToPercent(99)).toBe(0);
  });
});

describe('percentToLevel', () => {
  it.each([
    [0, 0],
    [16, 0],
    [17, 1],
    [49, 1],
    [50, 2],
    [83, 2],
    [84, 3],
    [100, 3],
  ])('maps %i%% to level %i', (percent, level) => {
    expect(percentToLevel(percent)).toBe(level);
  });
});

describe('levelToCommand', () => {
  it.each([
    [0, 'ventilation_level_0'],
    [1, 'ventilation_level_1'],
    [2, 'ventilation_level_2'],
    [3, 'ventilation_level_3'],
  ])('maps level %i to command %s', (level, command) => {
    expect(levelToCommand(level)).toBe(command);
  });
});
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `cd homebridge-zcangate && npx jest levelMapping`
Expected: FAIL — `Cannot find module './levelMapping'`.

- [ ] **Step 5: Implement the mapping functions**

`homebridge-zcangate/src/levelMapping.ts`:

```ts
const LEVEL_TO_PERCENT: Record<number, number> = {
  0: 0,
  1: 33,
  2: 66,
  3: 100,
};

export function levelToPercent(level: number): number {
  return LEVEL_TO_PERCENT[level] ?? 0;
}

export function percentToLevel(percent: number): number {
  if (percent <= 16) {
    return 0;
  }
  if (percent <= 49) {
    return 1;
  }
  if (percent <= 83) {
    return 2;
  }
  return 3;
}

export function levelToCommand(level: number): string {
  return `ventilation_level_${level}`;
}
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `cd homebridge-zcangate && npx jest levelMapping`
Expected: PASS, all `it.each` cases green.

- [ ] **Step 7: Commit**

```bash
git add homebridge-zcangate/package.json homebridge-zcangate/tsconfig.json \
  homebridge-zcangate/jest.config.js homebridge-zcangate/.gitignore \
  homebridge-zcangate/src/levelMapping.ts homebridge-zcangate/src/levelMapping.test.ts \
  homebridge-zcangate/package-lock.json
git commit -m "Scaffold homebridge-zcangate package with level/percentage mapping"
```

---

### Task 2: HTTP client (`zcangateClient`)

**Files:**
- Create: `homebridge-zcangate/src/zcangateClient.ts`
- Test: `homebridge-zcangate/src/zcangateClient.test.ts`

**Interfaces:**
- Consumes: none (uses global `fetch`/`AbortController`).
- Produces (used by Task 3 `zcangateAccessory.ts`):
  - `class ZcangateApiError extends Error { constructor(message: string, cause?: unknown) }`
  - `interface ZcangateClientOptions { baseUrl: string; authToken?: string; timeoutMs?: number }`
  - `class ZcangateClient { constructor(options: ZcangateClientOptions); getMeasurements(): Promise<Record<string, number>>; runCommand(name: string): Promise<void>; }`

- [ ] **Step 1: Write the failing tests**

`homebridge-zcangate/src/zcangateClient.test.ts`:

```ts
import { ZcangateApiError, ZcangateClient } from './zcangateClient';

function fakeResponse(ok: boolean, status: number, body: unknown): Response {
  return {
    ok,
    status,
    json: async () => body,
  } as Response;
}

describe('ZcangateClient', () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it('returns parsed JSON from getMeasurements on success', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, { ventilation_level: 2 }));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    const result = await client.getMeasurements();

    expect(result).toEqual({ ventilation_level: 2 });
    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/measurements',
      expect.objectContaining({ signal: expect.anything() }),
    );
  });

  it('throws ZcangateApiError when getMeasurements gets a non-2xx status', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(false, 500, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.getMeasurements()).rejects.toThrow(ZcangateApiError);
  });

  it('throws ZcangateApiError when the network request fails', async () => {
    (global.fetch as jest.Mock).mockRejectedValue(new Error('network down'));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.getMeasurements()).rejects.toThrow(ZcangateApiError);
  });

  it('sends a bearer token header when configured', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test', authToken: 'secret' });

    await client.runCommand('auto_mode');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/commands/auto_mode',
      expect.objectContaining({
        method: 'POST',
        headers: { Authorization: 'Bearer secret' },
      }),
    );
  });

  it('omits the Authorization header when no token is configured', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await client.runCommand('auto_mode');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/commands/auto_mode',
      expect.objectContaining({ headers: {} }),
    );
  });

  it('throws ZcangateApiError when runCommand gets a non-2xx status', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(false, 404, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.runCommand('unknown_command')).rejects.toThrow(ZcangateApiError);
  });

  it('aborts the request after the configured timeout', async () => {
    jest.useFakeTimers();
    (global.fetch as jest.Mock).mockImplementation((_url: string, init: RequestInit) => {
      return new Promise((_resolve, reject) => {
        init.signal?.addEventListener('abort', () => {
          reject(Object.assign(new Error('Aborted'), { name: 'AbortError' }));
        });
      });
    });
    const client = new ZcangateClient({ baseUrl: 'http://unit-test', timeoutMs: 1000 });

    const promise = client.getMeasurements();
    const assertion = expect(promise).rejects.toThrow(ZcangateApiError);
    jest.advanceTimersByTime(1000);
    await assertion;

    jest.useRealTimers();
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd homebridge-zcangate && npx jest zcangateClient`
Expected: FAIL — `Cannot find module './zcangateClient'`.

- [ ] **Step 3: Implement the client**

`homebridge-zcangate/src/zcangateClient.ts`:

```ts
export class ZcangateApiError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = 'ZcangateApiError';
  }
}

export interface ZcangateClientOptions {
  baseUrl: string;
  authToken?: string;
  timeoutMs?: number;
}

export class ZcangateClient {
  constructor(private readonly options: ZcangateClientOptions) {}

  async getMeasurements(): Promise<Record<string, number>> {
    const res = await this.fetchWithTimeout(`${this.options.baseUrl}/measurements`, {});
    if (!res.ok) {
      throw new ZcangateApiError(`GET /measurements failed with status ${res.status}`);
    }
    return (await res.json()) as Record<string, number>;
  }

  async runCommand(name: string): Promise<void> {
    const headers: Record<string, string> = {};
    if (this.options.authToken) {
      headers.Authorization = `Bearer ${this.options.authToken}`;
    }
    const res = await this.fetchWithTimeout(`${this.options.baseUrl}/commands/${name}`, {
      method: 'POST',
      headers,
    });
    if (!res.ok) {
      throw new ZcangateApiError(`POST /commands/${name} failed with status ${res.status}`);
    }
  }

  private async fetchWithTimeout(url: string, init: RequestInit): Promise<Response> {
    const controller = new AbortController();
    const timeoutMs = this.options.timeoutMs ?? 5000;
    const timeoutHandle = setTimeout(() => controller.abort(), timeoutMs);
    try {
      return await fetch(url, { ...init, signal: controller.signal });
    } catch (err) {
      throw new ZcangateApiError(`Request to ${url} failed`, err);
    } finally {
      clearTimeout(timeoutHandle);
    }
  }
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd homebridge-zcangate && npx jest zcangateClient`
Expected: PASS, all 7 tests green.

- [ ] **Step 5: Commit**

```bash
git add homebridge-zcangate/src/zcangateClient.ts homebridge-zcangate/src/zcangateClient.test.ts
git commit -m "Add zcangateClient HTTP client for measurements and commands"
```

---

### Task 3: HAP-facing accessory (`zcangateAccessory`)

**Files:**
- Create: `homebridge-zcangate/src/zcangateAccessory.ts`
- Test: `homebridge-zcangate/src/zcangateAccessory.test.ts`

**Interfaces:**
- Consumes:
  - `levelToPercent`, `percentToLevel`, `levelToCommand` from `./levelMapping` (Task 1).
  - `ZcangateClient`, `ZcangateClientOptions` from `./zcangateClient` (Task 2).
- Produces (used by Task 4 `index.ts`):
  - `interface ZcangateAccessoryConfig extends AccessoryConfig { apiBaseUrl: string; authToken?: string; pollInterval?: number }`
  - `class ZcangateAccessory implements AccessoryPlugin { constructor(log: Logging, config: ZcangateAccessoryConfig, api: API, client?: ZcangateClient); getServices(): Service[]; poll(): Promise<void>; }`

- [ ] **Step 1: Write the failing tests**

`homebridge-zcangate/src/zcangateAccessory.test.ts`:

```ts
import type { API, Logging } from 'homebridge';
import { ZcangateAccessory, ZcangateAccessoryConfig } from './zcangateAccessory';
import { ZcangateClient } from './zcangateClient';

class FakeCharacteristic {
  value: unknown;
  private getHandler?: () => Promise<unknown>;
  private setHandler?: (value: unknown) => Promise<void>;

  onGet(handler: () => Promise<unknown>): this {
    this.getHandler = handler;
    return this;
  }

  onSet(handler: (value: unknown) => Promise<void>): this {
    this.setHandler = handler;
    return this;
  }

  async triggerGet(): Promise<unknown> {
    if (!this.getHandler) {
      throw new Error('no get handler registered');
    }
    return this.getHandler();
  }

  async triggerSet(value: unknown): Promise<void> {
    if (!this.setHandler) {
      throw new Error('no set handler registered');
    }
    return this.setHandler(value);
  }

  updateValue(value: unknown): this {
    this.value = value;
    return this;
  }
}

class FakeService {
  private characteristics = new Map<unknown, FakeCharacteristic>();

  getCharacteristic(ctor: unknown): FakeCharacteristic {
    let ch = this.characteristics.get(ctor);
    if (!ch) {
      ch = new FakeCharacteristic();
      this.characteristics.set(ctor, ch);
    }
    return ch;
  }

  updateCharacteristic(ctor: unknown, value: unknown): this {
    this.getCharacteristic(ctor).updateValue(value);
    return this;
  }

  setCharacteristic(ctor: unknown, value: unknown): this {
    return this.updateCharacteristic(ctor, value);
  }
}

const Active = { ACTIVE: 1, INACTIVE: 0 };
const RotationSpeed = {};
const TargetFanState = { MANUAL: 0, AUTO: 1 };
const CurrentFanState = { INACTIVE: 0, IDLE: 1, BLOWING_AIR: 2 };
const Manufacturer = {};
const Model = {};

class FakeHapStatusError extends Error {
  constructor(public status: number) {
    super(`HAP status ${status}`);
  }
}

function createFakeApi(): API {
  return {
    hap: {
      Service: {
        Fanv2: FakeService,
        AccessoryInformation: FakeService,
      },
      Characteristic: {
        Active,
        RotationSpeed,
        TargetFanState,
        CurrentFanState,
        Manufacturer,
        Model,
      },
      HapStatusError: FakeHapStatusError,
      HAPStatus: { SERVICE_COMMUNICATION_FAILURE: -70402 },
    },
  } as unknown as API;
}

function createFakeLogging(): Logging {
  return {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
    debug: jest.fn(),
    log: jest.fn(),
    success: jest.fn(),
  } as unknown as Logging;
}

function createConfig(overrides: Partial<ZcangateAccessoryConfig> = {}): ZcangateAccessoryConfig {
  return {
    name: 'Ventilation',
    apiBaseUrl: 'http://unit-test',
    pollInterval: 3600,
    ...overrides,
  } as ZcangateAccessoryConfig;
}

function createFakeClient(): jest.Mocked<ZcangateClient> {
  return {
    getMeasurements: jest.fn(),
    runCommand: jest.fn(),
  } as unknown as jest.Mocked<ZcangateClient>;
}

describe('ZcangateAccessory', () => {
  let api: API;
  let client: jest.Mocked<ZcangateClient>;
  let accessory: ZcangateAccessory;
  let fanService: FakeService;

  beforeEach(() => {
    jest.useFakeTimers();
    api = createFakeApi();
    client = createFakeClient();
    client.runCommand.mockResolvedValue(undefined);
    accessory = new ZcangateAccessory(createFakeLogging(), createConfig(), api, client);
    fanService = accessory.getServices()[1] as unknown as FakeService;
  });

  afterEach(() => {
    jest.clearAllTimers();
    jest.useRealTimers();
  });

  it('exposes an AccessoryInformation and a Fanv2 service', () => {
    expect(accessory.getServices()).toHaveLength(2);
  });

  it('turns the fan off by sending ventilation_level_0', async () => {
    await fanService.getCharacteristic(Active).triggerSet(0);

    expect(client.runCommand).toHaveBeenCalledWith('ventilation_level_0');
    expect(fanService.getCharacteristic(RotationSpeed).value).toBe(0);
  });

  it('turns the fan on at the last remembered non-zero level, defaulting to 1', async () => {
    await fanService.getCharacteristic(Active).triggerSet(1);

    expect(client.runCommand).toHaveBeenCalledWith('ventilation_level_1');
  });

  it('remembers the last non-zero level across an off/on cycle', async () => {
    await fanService.getCharacteristic(RotationSpeed).triggerSet(66);
    await fanService.getCharacteristic(Active).triggerSet(0);
    client.runCommand.mockClear();

    await fanService.getCharacteristic(Active).triggerSet(1);

    expect(client.runCommand).toHaveBeenCalledWith('ventilation_level_2');
  });

  it('quantizes RotationSpeed to the nearest level', async () => {
    await fanService.getCharacteristic(RotationSpeed).triggerSet(84);

    expect(client.runCommand).toHaveBeenCalledWith('ventilation_level_3');
  });

  it('switches to manual mode before applying a speed change while in auto mode', async () => {
    await fanService.getCharacteristic(TargetFanState).triggerSet(TargetFanState.AUTO);
    client.runCommand.mockClear();

    await fanService.getCharacteristic(RotationSpeed).triggerSet(66);

    expect(client.runCommand).toHaveBeenNthCalledWith(1, 'manual_mode');
    expect(client.runCommand).toHaveBeenNthCalledWith(2, 'ventilation_level_2');
    expect(fanService.getCharacteristic(TargetFanState).value).toBe(TargetFanState.MANUAL);
  });

  it('does not resend manual_mode when already in manual mode', async () => {
    await fanService.getCharacteristic(RotationSpeed).triggerSet(66);

    expect(client.runCommand).toHaveBeenCalledTimes(1);
    expect(client.runCommand).toHaveBeenCalledWith('ventilation_level_2');
  });

  it('sends auto_mode when TargetFanState is set to AUTO', async () => {
    await fanService.getCharacteristic(TargetFanState).triggerSet(TargetFanState.AUTO);

    expect(client.runCommand).toHaveBeenCalledWith('auto_mode');
  });

  it('sends manual_mode when TargetFanState is set to MANUAL', async () => {
    await fanService.getCharacteristic(TargetFanState).triggerSet(TargetFanState.MANUAL);

    expect(client.runCommand).toHaveBeenCalledWith('manual_mode');
  });

  it('caches TargetFanState locally without querying the device', async () => {
    await fanService.getCharacteristic(TargetFanState).triggerSet(TargetFanState.AUTO);

    const value = await fanService.getCharacteristic(TargetFanState).triggerGet();

    expect(value).toBe(TargetFanState.AUTO);
    expect(client.getMeasurements).not.toHaveBeenCalled();
  });

  it('updates cached state from a poll', async () => {
    client.getMeasurements.mockResolvedValue({ ventilation_level: 3 });

    await accessory.poll();

    expect(fanService.getCharacteristic(Active).value).toBe(1);
    expect(fanService.getCharacteristic(RotationSpeed).value).toBe(100);
    expect(fanService.getCharacteristic(CurrentFanState).value).toBe(CurrentFanState.BLOWING_AIR);
  });

  it('logs and keeps the last known state when a poll fails', async () => {
    client.getMeasurements.mockResolvedValue({ ventilation_level: 2 });
    await accessory.poll();
    client.getMeasurements.mockRejectedValue(new Error('device offline'));

    await accessory.poll();

    expect(fanService.getCharacteristic(RotationSpeed).value).toBe(66);
  });

  it('throws a HapStatusError when a command fails', async () => {
    client.runCommand.mockRejectedValue(new Error('network down'));

    await expect(fanService.getCharacteristic(Active).triggerSet(0)).rejects.toThrow(FakeHapStatusError);
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd homebridge-zcangate && npx jest zcangateAccessory`
Expected: FAIL — `Cannot find module './zcangateAccessory'`.

- [ ] **Step 3: Implement the accessory**

`homebridge-zcangate/src/zcangateAccessory.ts`:

```ts
import type { AccessoryConfig, AccessoryPlugin, API, CharacteristicValue, Logging, Service } from 'homebridge';
import { levelToCommand, levelToPercent, percentToLevel } from './levelMapping';
import { ZcangateClient } from './zcangateClient';

export interface ZcangateAccessoryConfig extends AccessoryConfig {
  apiBaseUrl: string;
  authToken?: string;
  pollInterval?: number;
}

export class ZcangateAccessory implements AccessoryPlugin {
  private readonly fanService: Service;
  private readonly informationService: Service;
  private readonly client: ZcangateClient;

  private cachedLevel = 0;
  private cachedTargetState: number;
  private lastNonZeroLevel = 1;

  constructor(
    private readonly log: Logging,
    private readonly config: ZcangateAccessoryConfig,
    private readonly api: API,
    client?: ZcangateClient,
  ) {
    this.client = client ?? new ZcangateClient({ baseUrl: config.apiBaseUrl, authToken: config.authToken });
    this.cachedTargetState = this.api.hap.Characteristic.TargetFanState.MANUAL;

    this.informationService = new this.api.hap.Service.AccessoryInformation()
      .setCharacteristic(this.api.hap.Characteristic.Manufacturer, 'Zehnder')
      .setCharacteristic(this.api.hap.Characteristic.Model, 'ComfoAir (via zcangate)');

    this.fanService = new this.api.hap.Service.Fanv2(config.name);

    this.fanService
      .getCharacteristic(this.api.hap.Characteristic.Active)
      .onGet(this.getActive.bind(this))
      .onSet(this.setActive.bind(this));

    this.fanService
      .getCharacteristic(this.api.hap.Characteristic.RotationSpeed)
      .onGet(this.getRotationSpeed.bind(this))
      .onSet(this.setRotationSpeed.bind(this));

    this.fanService
      .getCharacteristic(this.api.hap.Characteristic.TargetFanState)
      .onGet(this.getTargetFanState.bind(this))
      .onSet(this.setTargetFanState.bind(this));

    this.fanService.getCharacteristic(this.api.hap.Characteristic.CurrentFanState).onGet(this.getCurrentFanState.bind(this));

    const pollIntervalMs = (config.pollInterval ?? 30) * 1000;
    setInterval(() => {
      this.poll();
    }, pollIntervalMs);
  }

  getServices(): Service[] {
    return [this.informationService, this.fanService];
  }

  async poll(): Promise<void> {
    try {
      const measurements = await this.client.getMeasurements();
      const level = measurements.ventilation_level;
      if (typeof level !== 'number') {
        return;
      }
      this.applyLevel(level);
    } catch (err) {
      this.log.warn(`Failed to poll zcangate measurements: ${(err as Error).message}`);
    }
  }

  private applyLevel(level: number): void {
    this.cachedLevel = level;
    if (level > 0) {
      this.lastNonZeroLevel = level;
    }
    this.fanService.updateCharacteristic(this.api.hap.Characteristic.Active, level > 0 ? 1 : 0);
    this.fanService.updateCharacteristic(this.api.hap.Characteristic.RotationSpeed, levelToPercent(level));
    this.fanService.updateCharacteristic(
      this.api.hap.Characteristic.CurrentFanState,
      level > 0 ? this.api.hap.Characteristic.CurrentFanState.BLOWING_AIR : this.api.hap.Characteristic.CurrentFanState.INACTIVE,
    );
  }

  private async getActive(): Promise<CharacteristicValue> {
    return this.cachedLevel > 0 ? 1 : 0;
  }

  private async setActive(value: CharacteristicValue): Promise<void> {
    const level = value === 0 ? 0 : this.lastNonZeroLevel || 1;
    await this.sendLevelCommand(level);
  }

  private async getRotationSpeed(): Promise<CharacteristicValue> {
    return levelToPercent(this.cachedLevel);
  }

  private async setRotationSpeed(value: CharacteristicValue): Promise<void> {
    const level = percentToLevel(value as number);
    if (this.cachedTargetState === this.api.hap.Characteristic.TargetFanState.AUTO) {
      await this.runCommandSafely('manual_mode');
      this.cachedTargetState = this.api.hap.Characteristic.TargetFanState.MANUAL;
      this.fanService.updateCharacteristic(this.api.hap.Characteristic.TargetFanState, this.cachedTargetState);
    }
    await this.sendLevelCommand(level);
  }

  private async getTargetFanState(): Promise<CharacteristicValue> {
    return this.cachedTargetState;
  }

  private async setTargetFanState(value: CharacteristicValue): Promise<void> {
    const isAuto = value === this.api.hap.Characteristic.TargetFanState.AUTO;
    await this.runCommandSafely(isAuto ? 'auto_mode' : 'manual_mode');
    this.cachedTargetState = value as number;
  }

  private async getCurrentFanState(): Promise<CharacteristicValue> {
    return this.cachedLevel > 0
      ? this.api.hap.Characteristic.CurrentFanState.BLOWING_AIR
      : this.api.hap.Characteristic.CurrentFanState.INACTIVE;
  }

  private async sendLevelCommand(level: number): Promise<void> {
    await this.runCommandSafely(levelToCommand(level));
    this.applyLevel(level);
  }

  private async runCommandSafely(command: string): Promise<void> {
    try {
      await this.client.runCommand(command);
    } catch (err) {
      this.log.error(`zcangate command '${command}' failed: ${(err as Error).message}`);
      throw new this.api.hap.HapStatusError(this.api.hap.HAPStatus.SERVICE_COMMUNICATION_FAILURE);
    }
  }
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd homebridge-zcangate && npx jest zcangateAccessory`
Expected: PASS, all 13 tests green. If TypeScript errors surface about HAP types (e.g. `Characteristic` static members), adjust the fake's shape in the test file to match — do not weaken the accessory's real typings.

- [ ] **Step 5: Commit**

```bash
git add homebridge-zcangate/src/zcangateAccessory.ts homebridge-zcangate/src/zcangateAccessory.test.ts
git commit -m "Add ZcangateAccessory Fanv2 HomeKit accessory"
```

---

### Task 4: Plugin registration entry point

**Files:**
- Create: `homebridge-zcangate/src/settings.ts`
- Create: `homebridge-zcangate/src/index.ts`
- Test: `homebridge-zcangate/src/index.test.ts`

**Interfaces:**
- Consumes: `ZcangateAccessory` from `./zcangateAccessory` (Task 3).
- Produces: `PLUGIN_NAME = 'homebridge-zcangate'`, `ACCESSORY_NAME = 'ZcangateVentilation'` (must match `pluginAlias` in Task 5's `config.schema.json`).

- [ ] **Step 1: Write the failing test**

`homebridge-zcangate/src/index.test.ts`:

```ts
import type { API } from 'homebridge';
import plugin from './index';
import { ACCESSORY_NAME, PLUGIN_NAME } from './settings';
import { ZcangateAccessory } from './zcangateAccessory';

describe('plugin entry point', () => {
  it('registers ZcangateAccessory under the expected plugin and accessory name', () => {
    const registerAccessory = jest.fn();
    const fakeApi = { registerAccessory } as unknown as API;

    plugin(fakeApi);

    expect(registerAccessory).toHaveBeenCalledWith(PLUGIN_NAME, ACCESSORY_NAME, ZcangateAccessory);
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd homebridge-zcangate && npx jest index.test`
Expected: FAIL — `Cannot find module './index'` (and `./settings`).

- [ ] **Step 3: Implement settings and the entry point**

`homebridge-zcangate/src/settings.ts`:

```ts
export const PLUGIN_NAME = 'homebridge-zcangate';
export const ACCESSORY_NAME = 'ZcangateVentilation';
```

`homebridge-zcangate/src/index.ts`:

```ts
import type { API } from 'homebridge';
import { ACCESSORY_NAME, PLUGIN_NAME } from './settings';
import { ZcangateAccessory } from './zcangateAccessory';

export = (api: API): void => {
  api.registerAccessory(PLUGIN_NAME, ACCESSORY_NAME, ZcangateAccessory);
};
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `cd homebridge-zcangate && npx jest index.test`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add homebridge-zcangate/src/settings.ts homebridge-zcangate/src/index.ts homebridge-zcangate/src/index.test.ts
git commit -m "Wire ZcangateAccessory into the Homebridge plugin entry point"
```

---

### Task 5: Config UI schema and README

**Files:**
- Create: `homebridge-zcangate/config.schema.json`
- Create: `homebridge-zcangate/README.md`

**Interfaces:**
- Consumes: `ACCESSORY_NAME` value `'ZcangateVentilation'` from Task 4 (must equal `pluginAlias` below).
- Produces: none consumed by later tasks.

- [ ] **Step 1: Write the config schema**

`homebridge-zcangate/config.schema.json`:

```json
{
  "pluginAlias": "ZcangateVentilation",
  "pluginType": "accessory",
  "singular": true,
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "title": "Name",
        "type": "string",
        "default": "Ventilation",
        "required": true
      },
      "apiBaseUrl": {
        "title": "zcangate API base URL",
        "type": "string",
        "default": "http://raspberrypi:8080",
        "required": true,
        "description": "Base URL of the zcangate HTTP server, e.g. http://raspberrypi:8080"
      },
      "authToken": {
        "title": "Auth token",
        "type": "string",
        "required": false,
        "description": "Bearer token to send when executing commands. Leave empty if COMMAND_AUTH_TOKEN is not set on the zcangate server."
      },
      "pollInterval": {
        "title": "Poll interval (seconds)",
        "type": "integer",
        "default": 30,
        "minimum": 5,
        "required": false
      }
    }
  }
}
```

- [ ] **Step 2: Verify the schema is valid JSON**

Run: `cd homebridge-zcangate && node -e "JSON.parse(require('fs').readFileSync('config.schema.json', 'utf8')); console.log('valid json')"`
Expected: prints `valid json`.

- [ ] **Step 3: Write the README**

`homebridge-zcangate/README.md`:

```md
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
| `TargetFanState` (Auto/Manual) | Sends the `auto_mode` / `manual_mode` commands. |
| `CurrentFanState` (read-only) | Derived from the polled fan speed. |

## Known limitation: Auto/Manual isn't read back from the device

zcangate has no confirmed measurement field reporting whether the unit is
currently in automatic or manual mode. This plugin therefore **caches the
Auto/Manual state locally** — it reflects the last mode set through this
plugin, not necessarily the device's true current mode if changed by another
controller (e.g. the physical remote).
```

- [ ] **Step 4: Commit**

```bash
git add homebridge-zcangate/config.schema.json homebridge-zcangate/README.md
git commit -m "Add config schema and README for homebridge-zcangate"
```

---

### Task 6: Full build verification

**Files:** None (verification only — no new files created or modified).

**Interfaces:** None.

- [ ] **Step 1: Run the full test suite**

Run: `cd homebridge-zcangate && npm test`
Expected: PASS — all test suites from Tasks 1–4 green (levelMapping, zcangateClient, zcangateAccessory, index).

- [ ] **Step 2: Run the production build**

Run: `cd homebridge-zcangate && npm run build`
Expected: completes with no TypeScript errors; `dist/index.js`, `dist/settings.js`, `dist/zcangateClient.js`, `dist/zcangateAccessory.js` are created (test files are excluded from the build per `tsconfig.json`).

- [ ] **Step 3: Confirm no test files leaked into the build output**

Run: `cd homebridge-zcangate && find dist -name '*.test.js'`
Expected: no output (empty).

No commit needed for this task — it only verifies work already committed in Tasks 1–5.
