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
