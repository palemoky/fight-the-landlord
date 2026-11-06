import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';
import { describe, expect, it } from 'vitest';
import { Card } from '../src/shared/cards/Card';
import { normalizeCard, summarizeHand } from '../src/shared/cards/cardModel';

describe('card face model', () => {
  it('normalizes standard suit cards to rank and corner symbols', () => {
    expect(normalizeCard({ suit: 0, rank: 14, color: 0 })).toMatchObject({
      rankLabel: 'A',
      suitSymbol: '♠',
      color: 'black',
      label: 'A♠',
      isJoker: false
    });
    expect(normalizeCard({ suit: 3, rank: 10, color: 1 })).toMatchObject({
      rankLabel: '10',
      suitSymbol: '♦',
      color: 'red',
      label: '10♦'
    });
  });

  it('recognizes jokers by rank even when suit data is inconsistent', () => {
    expect(normalizeCard({ suit: 0, rank: 17, color: 1 })).toMatchObject({
      label: '大王',
      rankLabel: 'JOKER',
      color: 'red',
      isJoker: true
    });
    expect(normalizeCard({ suit: 4, rank: 16, color: 0 })).toMatchObject({
      label: '小王',
      rankLabel: 'JOKER',
      color: 'black',
      isJoker: true
    });
  });

  it('renders jokers without B/R corner markers', () => {
    render(<Card card={{ suit: 4, rank: 17, color: 1 }} />);
    const joker = screen.getByRole('button', { name: '大王' });
    expect(joker).toHaveTextContent('JOKER');
    expect(joker.querySelector('.card__joker-index')).toBeNull();
  });

  it('summarizes common action selections', () => {
    expect(summarizeHand([{ suit: 4, rank: 16, color: 0 }, { suit: 4, rank: 17, color: 1 }])).toBe('王炸');
    expect(summarizeHand([{ suit: 1, rank: 8, color: 1 }, { suit: 3, rank: 8, color: 1 }])).toBe('对子');
  });
});
