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
