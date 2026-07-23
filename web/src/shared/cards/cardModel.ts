import type { CardInfo } from '../../protocol/types';

export type CardSuit = 'spade' | 'heart' | 'club' | 'diamond' | 'joker';
export type CardColor = 'red' | 'black';

export interface CardFace {
  rank: number;
  suit: CardSuit;
  color: CardColor;
  rankLabel: string;
  suitSymbol: string;
  label: string;
  isJoker: boolean;
  jokerColor?: CardColor;
}

const rankNames: Record<number, string> = {
  3: '3',
  4: '4',
  5: '5',
  6: '6',
  7: '7',
  8: '8',
  9: '9',
  10: '10',
  11: 'J',
  12: 'Q',
  13: 'K',
  14: 'A',
  15: '2',
  16: 'JOKER',
  17: 'JOKER'
};

const suitMeta: Record<number, { suit: CardSuit; symbol: string; color: CardColor; name: string }> = {
  0: { suit: 'spade', symbol: '♠', color: 'black', name: '黑桃' },
  1: { suit: 'heart', symbol: '♥', color: 'red', name: '红桃' },
  2: { suit: 'club', symbol: '♣', color: 'black', name: '梅花' },
  3: { suit: 'diamond', symbol: '♦', color: 'red', name: '方块' },
  4: { suit: 'joker', symbol: '', color: 'black', name: '王' }
};

export function cardKey(card: CardInfo): string {
  return `${card.suit}_${card.rank}`;
}

export function normalizeCard(card: CardInfo): CardFace {
  const isJoker = card.rank >= 16 || card.suit === 4;
  if (isJoker) {
    const jokerColor: CardColor = card.rank === 17 || card.color === 1 ? 'red' : 'black';
    return {
      rank: card.rank,
      suit: 'joker',
      color: jokerColor,
      rankLabel: 'JOKER',
      suitSymbol: jokerColor === 'red' ? '★' : '☆',
      label: jokerColor === 'red' ? '大王' : '小王',
      isJoker: true,
      jokerColor
    };
  }

  const meta = suitMeta[card.suit] ?? suitMeta[0];
  const rankLabel = rankNames[card.rank] ?? String(card.rank);
  return {
    rank: card.rank,
    suit: meta.suit,
    color: meta.color,
    rankLabel,
    suitSymbol: meta.symbol,
    label: `${rankLabel}${meta.symbol}`,
    isJoker: false
  };
}

export function cardLabel(card: CardInfo): string {
  return normalizeCard(card).label;
}

export function cardColor(card: CardInfo): CardColor {
  return normalizeCard(card).color;
}

export function sortCards(cards: CardInfo[]): CardInfo[] {
  return [...cards].sort((a, b) => b.rank - a.rank || a.suit - b.suit);
}

export function summarizeHand(cards: CardInfo[]): string {
  if (cards.length === 0) return '';
  const counts = new Map<number, number>();
  for (const card of cards) counts.set(card.rank, (counts.get(card.rank) ?? 0) + 1);
  const groups = [...counts.values()].sort((a, b) => b - a);
  if (cards.length === 1) return '单张';
  if (cards.length === 2 && counts.has(16) && counts.has(17)) return '王炸';
  if (cards.length === 2 && groups[0] === 2) return '对子';
  if (cards.length === 3 && groups[0] === 3) return '三张';
  if (cards.length === 4 && groups[0] === 4) return '炸弹';
  if (cards.length === 4 && groups[0] === 3) return '三带一';
  if (cards.length === 5 && groups[0] === 3 && groups[1] === 2) return '三带二';
  if (isStraight([...counts.keys()], cards.length, 1)) return '顺子';
  if (cards.length >= 6 && cards.length % 2 === 0 && isStraight([...counts.keys()], cards.length / 2, 2)) return '连对';
  return `${cards.length} 张`;
}

export function generateSimpleSuggestions(hand: CardInfo[], lastPlayed: CardInfo[], mustPlay: boolean): CardInfo[][] {
  const sorted = sortCards(hand);
  if (mustPlay || lastPlayed.length === 0) return sorted.length ? [[sorted[sorted.length - 1]]] : [];

  const targetLength = lastPlayed.length;
  const targetRank = Math.max(...lastPlayed.map((card) => card.rank));
  const byRank = new Map<number, CardInfo[]>();
  for (const card of sorted) {
    const group = byRank.get(card.rank) ?? [];
    group.push(card);
    byRank.set(card.rank, group);
  }

  return [...byRank.entries()]
    .filter(([rank, cards]) => rank > targetRank && cards.length >= targetLength)
    .sort(([a], [b]) => a - b)
    .slice(0, 8)
    .map(([, cards]) => cards.slice(0, targetLength));
}

export function initialCounter(myCards: CardInfo[]): Record<number, number> {
  const counts: Record<number, number> = {};
  for (let rank = 3; rank <= 15; rank += 1) counts[rank] = 4;
  counts[16] = 1;
  counts[17] = 1;
  for (const card of myCards) counts[card.rank] = Math.max(0, (counts[card.rank] ?? 0) - 1);
  return counts;
}

export function deductCounter(counter: Record<number, number>, cards: CardInfo[]): Record<number, number> {
  const next = { ...counter };
  for (const card of cards) next[card.rank] = Math.max(0, (next[card.rank] ?? 0) - 1);
  return next;
}

function isStraight(ranks: number[], expectedLength: number, groupSize: number): boolean {
  if (ranks.length !== expectedLength) return false;
  if (ranks.some((rank) => rank >= 15)) return false;
  const sorted = [...ranks].sort((a, b) => a - b);
  for (let index = 1; index < sorted.length; index += 1) {
    if (sorted[index] !== sorted[index - 1] + 1) return false;
  }
  return groupSize > 0;
}
