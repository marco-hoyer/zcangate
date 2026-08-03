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
