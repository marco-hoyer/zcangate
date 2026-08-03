import type { API, AccessoryPluginConstructor } from 'homebridge';
import { ACCESSORY_NAME, PLUGIN_NAME } from './settings';
import { ZcangateAccessory } from './zcangateAccessory';

export = (api: API): void => {
  api.registerAccessory(PLUGIN_NAME, ACCESSORY_NAME, ZcangateAccessory as unknown as AccessoryPluginConstructor);
};
