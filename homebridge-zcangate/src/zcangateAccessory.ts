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
      if (typeof level === 'number') {
        this.applyLevel(level);
      }
      const controlMode = measurements.ventilation_control_mode;
      if (typeof controlMode === 'number') {
        this.applyControlMode(controlMode);
      }
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
    this.fanService.updateCharacteristic(this.api.hap.Characteristic.TargetFanState, this.cachedTargetState);
  }

  // ventilation_control_mode is 0=auto/1=manual, the inverse of HAP's TargetFanState (MANUAL=0/AUTO=1).
  private applyControlMode(deviceMode: number): void {
    this.cachedTargetState =
      deviceMode === 0
        ? this.api.hap.Characteristic.TargetFanState.AUTO
        : this.api.hap.Characteristic.TargetFanState.MANUAL;
    this.fanService.updateCharacteristic(this.api.hap.Characteristic.TargetFanState, this.cachedTargetState);
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
