import { useAppStore, useChatStore } from './appStore';
import type { CardInfo, PlayerInfo } from '../protocol/types';
import { initialCounter, sortCards } from '../shared/cards/cardModel';

const players: PlayerInfo[] = [
  { id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 16, online: true },
  { id: 'p2', name: '山月', seat: 1, ready: true, is_landlord: true, cards_count: 13, online: true },
  { id: 'p3', name: '松风', seat: 2, ready: true, is_landlord: false, cards_count: 15, online: true }
];

const hand: CardInfo[] = [
  c(4, 17), c(4, 16), c(1, 15), c(0, 14), c(2, 14), c(1, 13), c(3, 13), c(0, 12),
  c(1, 11), c(2, 10), c(3, 10), c(0, 9), c(2, 9), c(1, 8), c(0, 7), c(3, 6), c(2, 3)
];

export function seedDemoState(mode: string): void {
  const sortedHand = sortCards(hand);
  const isBidding = mode === 'bidding';
  const demoPlayers = isBidding ? players.map((player) => ({ ...player, is_landlord: false })) : players;
  useChatStore.setState({
    messages: [
      { sender_name: '山月', content: '这局节奏很快。', scope: 'room', time: Date.now() / 1000 },
      { sender_name: '青竹', content: '我先看一手。', scope: 'room', time: Date.now() / 1000 }
    ]
  });
  useAppStore.setState({
    connected: true,
    phase: mode === 'lobby' ? 'lobby' : mode === 'result' ? 'game_over' : isBidding ? 'bidding' : 'playing',
    playerId: 'p1',
    playerName: '青竹',
    roomCode: '836219',
    players: demoPlayers,
    onlineCount: 128,
    hand: sortedHand,
    bottomCards: [c(1, 5), c(2, 11), c(3, 15)],
    bottomCardsRevealed: !isBidding,
    currentTurn: isBidding ? 'p1' : 'p1',
    lastPlayed: isBidding ? [] : [c(2, 8), c(3, 8)],
    lastPlayedBy: isBidding ? '' : 'p3',
    lastPlayedName: isBidding ? '' : '松风',
    lastHandType: isBidding ? '' : '对子',
    mustPlay: false,
    canBeat: true,
    multiplier: 3,
    timeout: 25,
    timerStart: Date.now(),
    cardCounter: initialCounter(sortedHand),
    seatActions: isBidding ? {
      p2: { type: 'bid', player_id: 'p2', player_name: '灞辨湀', label: '不叫' }
    } : {
      p2: { type: 'pass', player_id: 'p2', player_name: '灞辨湀', label: '不出' },
      p3: { type: 'play', player_id: 'p3', player_name: '鏉鹃', cards: [c(2, 8), c(3, 8)], hand_type: '对子' }
    },
    isGrabTurn: false,
    recentActions: isBidding ? [
      { type: 'bid', player_id: 'p2', player_name: '山月', label: '不叫' }
    ] : [
      { type: 'play', player_id: 'p3', player_name: '松风', cards: [c(2, 8), c(3, 8)], hand_type: '对子' },
      { type: 'pass', player_id: 'p2', player_name: '山月' }
    ],
    scores: [
      { player_id: 'p1', player_name: '青竹', is_landlord: false, score: 6 },
      { player_id: 'p2', player_name: '山月', is_landlord: true, score: -12 },
      { player_id: 'p3', player_name: '松风', is_landlord: false, score: 6 }
    ],
    playerHands: [
      { player_id: 'p2', player_name: '山月', cards: [c(1, 4), c(2, 4), c(3, 7)] },
      { player_id: 'p3', player_name: '松风', cards: [c(0, 5), c(2, 6)] }
    ],
    winnerId: 'p1',
    winnerName: '青竹',
    winnerIsLandlord: false,
    finalMultiplier: 3
  });
}

function c(suit: number, rank: number): CardInfo {
  return { suit, rank, color: suit === 1 || suit === 3 || rank === 17 ? 1 : 0 };
}
