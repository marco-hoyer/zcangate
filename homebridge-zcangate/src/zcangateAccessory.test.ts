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

  it('reconciles TargetFanState to MANUAL from a poll when ventilation_control_mode is 1', async () => {
    client.getMeasurements.mockResolvedValue({ ventilation_level: 2, ventilation_control_mode: 1 });

    await accessory.poll();

    expect(fanService.getCharacteristic(TargetFanState).value).toBe(TargetFanState.MANUAL);
  });

  it('reconciles TargetFanState to AUTO from a poll when ventilation_control_mode is 0', async () => {
    client.getMeasurements.mockResolvedValue({ ventilation_level: 2, ventilation_control_mode: 0 });

    await accessory.poll();

    expect(fanService.getCharacteristic(TargetFanState).value).toBe(TargetFanState.AUTO);
  });

  it('leaves TargetFanState untouched when a poll omits ventilation_control_mode', async () => {
    await fanService.getCharacteristic(TargetFanState).triggerSet(TargetFanState.AUTO);
    client.getMeasurements.mockResolvedValue({ ventilation_level: 2 });

    await accessory.poll();

    expect(fanService.getCharacteristic(TargetFanState).value).toBe(TargetFanState.AUTO);
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
