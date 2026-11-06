import { beforeEach, describe, expect, it } from 'vitest';
import { MsgType } from '../src/protocol/types';
import { useAppStore } from '../src/stores/appStore';

describe('app store message flow', () => {
  beforeEach(() => {
    localStorage.clear();
    useAppStore.setState({
      connected: false,
      playerId: '',
      playerName: '',
      reconnectToken: '',
      phase: 'connecting',
      players: [],
      roomCode: '',
      hand: [],
      bottomCards: [],
      bottomCardsRevealed: false,
      currentTurn: '',
      selectedCards: new Set(),
      recentActions: [],
      seatActions: {},
      cardCounter: {}
    });
  });

  it('enters lobby after connected', () => {
    useAppStore.getState().handleMessage({
      type: MsgType.Connected,
      payload: { player_id: 'p1', player_name: '青竹', reconnect_token: 'tok' }
    });
    expect(useAppStore.getState().phase).toBe('lobby');
    expect(useAppStore.getState().playerId).toBe('p1');
  });

  it('keeps bottom cards hidden after deal cards before landlord is confirmed', () => {
    useAppStore.getState().handleMessage({
      type: MsgType.DealCards,
      payload: {
        cards: [{ suit: 0, rank: 14, color: 0 }],
        bottom_cards: [
          { suit: 1, rank: 5, color: 1 },
          { suit: 2, rank: 11, color: 0 },
          { suit: 3, rank: 15, color: 1 }
        ]
      }
    });
    expect(useAppStore.getState().bottomCards).toHaveLength(3);
    expect(useAppStore.getState().bottomCardsRevealed).toBe(false);
  });

  it('normalizes missing game player counts and online flags from protobuf defaults', () => {
    useAppStore.setState({ playerId: 'p1' });
    useAppStore.getState().handleMessage({
      type: MsgType.GameStart,
      payload: {
        players: [
          { id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 0, online: false },
          { id: 'p2', name: '山月', seat: 1, ready: true, is_landlord: false, cards_count: 0, online: false },
          { id: 'p3', name: '松风', seat: 2, ready: true, is_landlord: false, cards_count: 0, online: false }
        ]
      }
    });
    useAppStore.getState().handleMessage({
      type: MsgType.DealCards,
      payload: {
        cards: [
          { suit: 0, rank: 14, color: 0 },
          { suit: 1, rank: 13, color: 1 }
        ],
        bottom_cards: []
      }
    });
    expect(useAppStore.getState().players.map((player) => player.cards_count)).toEqual([2, 17, 17]);
    expect(useAppStore.getState().players.every((player) => player.online)).toBe(true);
  });

  it('reveals bottom cards and appends them after landlord is confirmed', () => {
    useAppStore.setState({
      playerId: 'p1',
      hand: [{ suit: 0, rank: 14, color: 0 }],
      players: [{ id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 17, online: true }]
    });
    useAppStore.getState().handleMessage({
      type: MsgType.Landlord,
      payload: {
        player_id: 'p1',
        player_name: '青竹',
        bottom_cards: [
          { suit: 1, rank: 5, color: 1 },
          { suit: 2, rank: 11, color: 0 },
          { suit: 3, rank: 15, color: 1 }
        ],
        multiplier: 3
      }
    });
    expect(useAppStore.getState().bottomCardsRevealed).toBe(true);
    expect(useAppStore.getState().hand).toHaveLength(4);
    expect(useAppStore.getState().players[0].is_landlord).toBe(true);
  });

  it('updates bid turn and bid action history', () => {
    useAppStore.getState().handleMessage({
      type: MsgType.BidTurn,
      payload: { player_id: 'p1', timeout: 20, is_grab: false, multiplier: 1 }
    });
    expect(useAppStore.getState().phase).toBe('bidding');
    expect(useAppStore.getState().currentTurn).toBe('p1');
    expect(useAppStore.getState().isGrabTurn).toBe(false);

    useAppStore.getState().handleMessage({
      type: MsgType.BidResult,
      payload: { player_id: 'p1', player_name: '青竹', bid: true, is_grab: false, multiplier: 1 }
    });
    expect(useAppStore.getState().recentActions.at(-1)?.label).toBe('叫地主');
  });

  it('restores reconnect game state with bottom-card reveal based on phase', () => {
    useAppStore.getState().handleMessage({
      type: MsgType.Reconnected,
      payload: {
        player_id: 'p1',
        player_name: '青竹',
        room_code: '888888',
        game_state: {
          phase: 'bidding',
          players: [{ id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 17, online: true }],
          hand: [{ suit: 0, rank: 14, color: 0 }],
          bottom_cards: [{ suit: 1, rank: 5, color: 1 }],
          current_turn: 'p1',
          last_played: [],
          last_player_id: '',
          must_play: true,
          can_beat: true
        }
      }
    });
    expect(useAppStore.getState().phase).toBe('bidding');
    expect(useAppStore.getState().bottomCardsRevealed).toBe(false);

    useAppStore.getState().handleMessage({
      type: MsgType.Reconnected,
      payload: {
        player_id: 'p1',
        player_name: '青竹',
        room_code: '888888',
        game_state: {
          phase: 'playing',
          players: [{ id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: true, cards_count: 20, online: true }],
          hand: [{ suit: 0, rank: 14, color: 0 }],
          bottom_cards: [{ suit: 1, rank: 5, color: 1 }],
          current_turn: 'p1',
          last_played: [],
          last_player_id: '',
          must_play: true,
          can_beat: true
        }
      }
    });
    expect(useAppStore.getState().phase).toBe('playing');
    expect(useAppStore.getState().bottomCardsRevealed).toBe(true);
  });

  it('updates hand and action history after playing cards', () => {
    useAppStore.setState({
      playerId: 'p1',
      hand: [
        { suit: 0, rank: 14, color: 0 },
        { suit: 1, rank: 13, color: 1 }
      ],
      players: [{ id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 2, online: true }],
      cardCounter: { 14: 3, 13: 3 }
    });
    useAppStore.getState().handleMessage({
      type: MsgType.CardPlayed,
      payload: {
        player_id: 'p1',
        player_name: '青竹',
        cards: [{ suit: 0, rank: 14, color: 0 }],
        cards_left: 1,
        hand_type: '单张'
      }
    });
    expect(useAppStore.getState().hand).toEqual([{ suit: 1, rank: 13, color: 1 }]);
    expect(useAppStore.getState().players[0].cards_count).toBe(1);
    expect(useAppStore.getState().seatActions.p1?.type).toBe('play');
    expect(useAppStore.getState().recentActions).toHaveLength(1);
  });

  it('keeps seat actions for a trick and clears them on a new lead play', () => {
    useAppStore.setState({
      playerId: 'p1',
      mustPlay: false,
      players: [
        { id: 'p1', name: '青竹', seat: 0, ready: true, is_landlord: false, cards_count: 17, online: true },
        { id: 'p2', name: '山月', seat: 1, ready: true, is_landlord: false, cards_count: 17, online: true },
        { id: 'p3', name: '松风', seat: 2, ready: true, is_landlord: false, cards_count: 17, online: true }
      ]
    });

    useAppStore.getState().handleMessage({
      type: MsgType.CardPlayed,
      payload: {
        player_id: 'p2',
        player_name: '山月',
        cards: [{ suit: 1, rank: 9, color: 1 }],
        cards_left: 16,
        hand_type: '单张'
      }
    });
    useAppStore.getState().handleMessage({
      type: MsgType.PlayerPass,
      payload: { player_id: 'p3', player_name: '松风' }
    });
    expect(Object.keys(useAppStore.getState().seatActions).sort()).toEqual(['p2', 'p3']);

    useAppStore.setState({ mustPlay: true });
    useAppStore.getState().handleMessage({
      type: MsgType.CardPlayed,
      payload: {
        player_id: 'p2',
        player_name: '山月',
        cards: [{ suit: 2, rank: 11, color: 0 }],
        cards_left: 15,
        hand_type: '单张'
      }
    });
    expect(Object.keys(useAppStore.getState().seatActions)).toEqual(['p2']);
  });
});
