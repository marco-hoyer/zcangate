import { levelToCommand } from './levelMapping';

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
