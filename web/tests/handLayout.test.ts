import { describe, expect, it } from 'vitest';
import type { CardInfo } from '../src/protocol/types';
import { buildHandLayout } from '../src/shared/cards/Hand';

function c(suit: number, rank: number): CardInfo {
  return { suit, rank, color: suit === 1 || suit === 3 ? 1 : 0 };
}

describe('hand layout model', () => {
  it('keeps logical order while grouping equal ranks', () => {
    const layout = buildHandLayout([c(4, 17), c(4, 16), c(0, 14), c(1, 14), c(2, 10)]);

    expect(layout.map((item) => item.index)).toEqual([0, 1, 2, 3, 4]);
    expect(layout[2].groupIndex).toBe(layout[3].groupIndex);
    expect(layout[2].groupSize).toBe(2);
  });

  it('provides independent row coordinates for portrait two-row hands', () => {
    const layout = buildHandLayout([
      c(4, 17), c(4, 16), c(0, 15), c(1, 14), c(2, 13), c(3, 12), c(0, 11), c(1, 10)
    ]);

    expect(new Set(layout.map((item) => item.row))).toEqual(new Set([0, 1]));
    expect(layout.every((item) => Number.isFinite(item.singleX) && Number.isFinite(item.compactX) && Number.isFinite(item.rowX))).toBe(true);
  });
});
