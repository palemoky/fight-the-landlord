import { create } from 'zustand';
import { MsgType, type CardInfo, type ChatPayload, type GameStateDTO, type IncomingMessage, type LeaderboardEntry, type LobbyPanel, type Phase, type PlayerHand, type PlayerInfo, type PlayerScore, type RoomListItem, type StatsResultPayload, type UtilityDrawer } from '../protocol/types';
import { cardKey, deductCounter, initialCounter, sortCards } from '../shared/cards/cardModel';

interface ConnectionSlice {
  connected: boolean;
  playerId: string;
  playerName: string;
  reconnectToken: string;
  latency: number;
  error: string;
  maintenance: boolean;
}

interface LobbySlice {
  phase: Phase;
  roomCode: string;
  players: PlayerInfo[];
  onlineCount: number;
  lobbyPanel: LobbyPanel;
  roomCodeInput: string;
  stats: StatsResultPayload | null;
  leaderboard: LeaderboardEntry[];
  leaderboardType: string;
  roomList: RoomListItem[];
}

interface TableSlice {
  hand: CardInfo[];
  bottomCards: CardInfo[];
  bottomCardsRevealed: boolean;
  currentTurn: string;
  lastPlayed: CardInfo[];
  lastPlayedBy: string;
  lastPlayedName: string;
  lastHandType: string;
  mustPlay: boolean;
  canBeat: boolean;
  multiplier: number;
  isGrabTurn: boolean;
  timeout: number;
  timerStart: number;
  winnerId: string;
  winnerName: string;
  winnerIsLandlord: boolean;
  finalMultiplier: number;
  scores: PlayerScore[];
  playerHands: PlayerHand[];
  cardCounter: Record<number, number>;
  recentActions: TableAction[];
  seatActions: Record<string, SeatAction>;
}

interface UiSlice {
  selectedCards: Set<string>;
  drawer: UtilityDrawer;
  chatInput: string;
  tableMessage: string;
}

export interface TableAction {
  type: 'play' | 'pass' | 'bid' | 'system';
  player_id?: string;
  player_name?: string;
  cards?: CardInfo[];
  hand_type?: string;
  label?: string;
}

export interface SeatAction {
  type: 'play' | 'pass' | 'bid';
  player_id: string;
  player_name?: string;
  cards?: CardInfo[];
  hand_type?: string;
  label?: string;
}

interface AppActions {
  setConnected: (connected: boolean) => void;
  setError: (error: string) => void;
  setLobbyPanel: (panel: LobbyPanel) => void;
  setRoomCodeInput: (value: string) => void;
  setChatInput: (value: string) => void;
  setDrawer: (drawer: UtilityDrawer) => void;
  toggleCard: (key: string) => void;
  setSelection: (keys: string[]) => void;
  clearSelection: () => void;
  leaveLocalRoom: () => void;
  handleMessage: (message: IncomingMessage) => void;
}

export type AppState = ConnectionSlice & LobbySlice & TableSlice & UiSlice & AppActions;

const initialTable: TableSlice = {
  hand: [],
  bottomCards: [],
  bottomCardsRevealed: false,
  currentTurn: '',
  lastPlayed: [],
  lastPlayedBy: '',
  lastPlayedName: '',
  lastHandType: '',
  mustPlay: false,
  canBeat: false,
  multiplier: 1,
  isGrabTurn: false,
  timeout: 0,
  timerStart: 0,
  winnerId: '',
  winnerName: '',
  winnerIsLandlord: false,
  finalMultiplier: 1,
  scores: [],
  playerHands: [],
  cardCounter: {},
  recentActions: [],
  seatActions: {}
};

export const useAppStore = create<AppState>((set, get) => ({
  connected: false,
  playerId: '',
  playerName: '',
  reconnectToken: '',
  latency: 0,
  error: '',
  maintenance: false,
  phase: 'connecting',
  roomCode: '',
  players: [],
  onlineCount: 0,
  lobbyPanel: 'home',
  roomCodeInput: '',
  stats: null,
  leaderboard: [],
  leaderboardType: 'total',
  roomList: [],
  ...initialTable,
  selectedCards: new Set<string>(),
  drawer: 'none',
  chatInput: '',
  tableMessage: '',

  setConnected: (connected) => set({ connected }),
  setError: (error) => set({ error }),
  setLobbyPanel: (lobbyPanel) => set({ lobbyPanel }),
  setRoomCodeInput: (roomCodeInput) => set({ roomCodeInput }),
  setChatInput: (chatInput) => set({ chatInput }),
  setDrawer: (drawer) => set({ drawer }),
  toggleCard: (key) => {
    const selectedCards = new Set(get().selectedCards);
    if (selectedCards.has(key)) selectedCards.delete(key);
    else selectedCards.add(key);
    set({ selectedCards, tableMessage: '' });
  },
  setSelection: (keys) => set({ selectedCards: new Set(keys) }),
  clearSelection: () => set({ selectedCards: new Set(), tableMessage: '' }),
  leaveLocalRoom: () => set({ roomCode: '', players: [], phase: 'lobby', ...initialTable, selectedCards: new Set() }),
  handleMessage: (message) => {
    const state = get();
    switch (message.type) {
      case MsgType.Connected: {
        const payload = message.payload as { player_id: string; player_name: string; reconnect_token: string };
        persistReconnect(payload.player_id, payload.reconnect_token);
        set({
          connected: true,
          error: '',
          phase: 'lobby',
          playerId: payload.player_id,
          playerName: payload.player_name,
          reconnectToken: payload.reconnect_token
        });
        break;
      }
      case MsgType.Reconnected: {
        const payload = message.payload as { player_id: string; player_name: string; room_code?: string; game_state?: GameStateDTO };
        persistReconnect(payload.player_id, state.reconnectToken);
        if (payload.game_state) {
          set({ playerId: payload.player_id, playerName: payload.player_name, roomCode: payload.room_code ?? '', ...restoreSnapshot(payload.game_state, payload.player_id) });
        } else {
          set({ playerId: payload.player_id, playerName: payload.player_name, roomCode: payload.room_code ?? '', phase: payload.room_code ? 'waiting' : 'lobby' });
        }
        break;
      }
      case MsgType.Pong: {
        const payload = message.payload as { client_timestamp: number };
        set({ latency: Date.now() - (payload.client_timestamp || Date.now()) });
        break;
      }
      case MsgType.Error: {
        const payload = message.payload as { message?: string };
        set({ error: payload.message || '未知错误', phase: state.phase === 'matching' ? 'lobby' : state.phase });
        break;
      }
      case MsgType.OnlineCount:
        set({ onlineCount: (message.payload as { count: number }).count });
        break;
      case MsgType.RoomCreated: {
        const payload = message.payload as { room_code: string; player: PlayerInfo };
        set({ roomCode: payload.room_code, players: [normalizeLobbyPlayer(payload.player)], phase: 'waiting', lobbyPanel: 'home' });
        break;
      }
      case MsgType.RoomJoined: {
        const payload = message.payload as { room_code: string; players: PlayerInfo[] };
        set({ roomCode: payload.room_code, players: normalizeLobbyPlayers(payload.players ?? []), phase: 'waiting', lobbyPanel: 'home' });
        break;
      }
      case MsgType.PlayerJoined: {
        const payload = message.payload as { player: PlayerInfo };
        set({ players: mergePlayer(state.players, normalizeLobbyPlayer(payload.player)) });
        break;
      }
      case MsgType.PlayerLeft: {
        const payload = message.payload as { player_id: string };
        if (payload.player_id === state.playerId) get().leaveLocalRoom();
        else set({ players: state.players.filter((player) => player.id !== payload.player_id) });
        break;
      }
      case MsgType.PlayerReady: {
        const payload = message.payload as { player_id: string; ready: boolean };
        set({ players: state.players.map((player) => player.id === payload.player_id ? { ...player, ready: payload.ready } : player) });
        break;
      }
      case MsgType.MatchFound:
        set({ phase: 'waiting' });
        break;
      case MsgType.GameStart: {
        const payload = message.payload as { players: PlayerInfo[] };
        set({ ...initialTable, players: normalizeGamePlayers(payload.players ?? [], state.playerId), phase: 'bidding', selectedCards: new Set() });
        break;
      }
      case MsgType.DealCards: {
        const payload = message.payload as { cards: CardInfo[]; bottom_cards: CardInfo[] };
        const hand = sortCards(payload.cards ?? []);
        set({
          hand,
          bottomCards: payload.bottom_cards ?? [],
          bottomCardsRevealed: false,
          seatActions: {},
          players: syncDealtCounts(state.players, state.playerId, hand.length),
          cardCounter: initialCounter(hand)
        });
        break;
      }
      case MsgType.BidTurn: {
        const payload = message.payload as { player_id: string; timeout: number; is_grab: boolean; multiplier: number };
        set({ phase: 'bidding', currentTurn: payload.player_id, timeout: payload.timeout, isGrabTurn: payload.is_grab, multiplier: payload.multiplier || 1, timerStart: Date.now(), tableMessage: '' });
        break;
      }
      case MsgType.BidResult: {
        const payload = message.payload as { player_id: string; player_name: string; bid: boolean; is_grab: boolean; multiplier: number };
        const seatAction = {
          type: 'bid' as const,
          player_id: payload.player_id,
          player_name: payload.player_name,
          label: payload.bid ? (payload.is_grab ? '抢地主' : '叫地主') : (payload.is_grab ? '不抢' : '不叫')
        };
        set({ seatActions: setSeatAction(state.seatActions, seatAction) });
        set({ multiplier: payload.multiplier || state.multiplier, recentActions: pushAction(state.recentActions, { type: 'bid', player_id: payload.player_id, player_name: payload.player_name, label: payload.bid ? (payload.is_grab ? '抢地主' : '叫地主') : (payload.is_grab ? '不抢' : '不叫') }) });
        break;
      }
      case MsgType.Landlord: {
        const payload = message.payload as { player_id: string; player_name: string; bottom_cards: CardInfo[]; multiplier: number };
        const isMe = payload.player_id === state.playerId;
        const nextHand = isMe ? sortCards([...state.hand, ...(payload.bottom_cards ?? [])]) : state.hand;
        set({
          players: syncLandlordCounts(state.players, state.playerId, payload.player_id, nextHand.length),
          bottomCards: payload.bottom_cards ?? [],
          bottomCardsRevealed: true,
          multiplier: payload.multiplier || 1,
          hand: nextHand,
          seatActions: {},
          recentActions: pushAction(state.recentActions, { type: 'system', label: `${payload.player_name} 成为地主` })
        });
        break;
      }
      case MsgType.PlayTurn: {
        const payload = message.payload as { player_id: string; timeout: number; must_play: boolean; can_beat: boolean };
        set({ phase: 'playing', currentTurn: payload.player_id, timeout: payload.timeout, mustPlay: payload.must_play, canBeat: payload.can_beat, timerStart: Date.now(), tableMessage: '' });
        break;
      }
      case MsgType.CardPlayed: {
        const payload = message.payload as { player_id: string; player_name: string; cards: CardInfo[]; cards_left: number; hand_type: string };
        const playedKeys = new Set((payload.cards ?? []).map(cardKey));
        const isMe = payload.player_id === state.playerId;
        const action = {
          type: 'play' as const,
          player_id: payload.player_id,
          player_name: payload.player_name,
          cards: payload.cards ?? [],
          hand_type: payload.hand_type
        };
        set({
          lastPlayed: payload.cards ?? [],
          lastPlayedBy: payload.player_id,
          lastPlayedName: payload.player_name,
          lastHandType: payload.hand_type,
          hand: isMe ? state.hand.filter((card) => !playedKeys.has(cardKey(card))) : state.hand,
          players: state.players.map((player) => player.id === payload.player_id ? { ...player, cards_count: payload.cards_left } : player),
          cardCounter: isMe ? state.cardCounter : deductCounter(state.cardCounter, payload.cards ?? []),
          recentActions: pushAction(state.recentActions, action),
          seatActions: setSeatAction(state.mustPlay ? {} : state.seatActions, action),
          selectedCards: new Set()
        });
        break;
      }
      case MsgType.PlayerPass: {
        const payload = message.payload as { player_id: string; player_name: string };
        const action = { type: 'pass' as const, player_id: payload.player_id, player_name: payload.player_name, label: '不出' };
        set({
          lastPlayedBy: payload.player_id,
          lastPlayedName: payload.player_name,
          recentActions: pushAction(state.recentActions, action),
          seatActions: setSeatAction(state.seatActions, action)
        });
        break;
      }
      case MsgType.GameOver: {
        const payload = message.payload as { winner_id: string; winner_name: string; is_landlord: boolean; multiplier: number; scores: PlayerScore[]; player_hands: PlayerHand[] };
        set({ phase: 'game_over', winnerId: payload.winner_id, winnerName: payload.winner_name, winnerIsLandlord: payload.is_landlord, finalMultiplier: payload.multiplier, scores: payload.scores ?? [], playerHands: payload.player_hands ?? [], drawer: 'none' });
        break;
      }
      case MsgType.StatsResult:
        set({ stats: message.payload as StatsResultPayload, lobbyPanel: 'stats' });
        break;
      case MsgType.LeaderboardResult: {
        const payload = message.payload as { type: string; entries: LeaderboardEntry[] };
        set({ leaderboardType: payload.type, leaderboard: payload.entries ?? [], lobbyPanel: 'leaderboard' });
        break;
      }
      case MsgType.RoomListResult:
        set({ roomList: (message.payload as { rooms: RoomListItem[] }).rooms ?? [] });
        break;
      case MsgType.Chat:
        useChatStore.getState().push(message.payload as ChatPayload);
        break;
      case MsgType.PlayerOffline: {
        const payload = message.payload as { player_id: string };
        set({ players: state.players.map((player) => player.id === payload.player_id ? { ...player, online: false } : player) });
        break;
      }
      case MsgType.PlayerOnline: {
        const payload = message.payload as { player_id: string };
        set({ players: state.players.map((player) => player.id === payload.player_id ? { ...player, online: true } : player) });
        break;
      }
      case MsgType.Maintenance:
      case MsgType.MaintenanceStatus:
      case MsgType.MaintenancePull:
      case MsgType.MaintenancePush:
        set({ maintenance: Boolean((message.payload as { maintenance: boolean }).maintenance) });
        break;
      default:
        break;
    }
  }
}));

export const useChatStore = create<{ messages: ChatPayload[]; push: (message: ChatPayload) => void; clear: () => void }>((set, get) => ({
  messages: [],
  push: (message) => set({ messages: [...get().messages, message].slice(-80) }),
  clear: () => set({ messages: [] })
}));

export function loadReconnect(): { id: string; token: string } | null {
  try {
    const raw = localStorage.getItem('ddz_next_reconnect');
    return raw ? JSON.parse(raw) as { id: string; token: string } : null;
  } catch {
    return null;
  }
}

function persistReconnect(id: string, token: string): void {
  if (!id || !token) return;
  localStorage.setItem('ddz_next_reconnect', JSON.stringify({ id, token }));
}

function mergePlayer(players: PlayerInfo[], next: PlayerInfo): PlayerInfo[] {
  return players.some((player) => player.id === next.id)
    ? players.map((player) => player.id === next.id ? next : player)
    : [...players, next].sort((a, b) => a.seat - b.seat);
}

function normalizeLobbyPlayer(player: PlayerInfo): PlayerInfo {
  return { ...player, online: true };
}

function normalizeLobbyPlayers(players: PlayerInfo[]): PlayerInfo[] {
  return players.map(normalizeLobbyPlayer).sort((a, b) => a.seat - b.seat);
}

function normalizeGamePlayers(players: PlayerInfo[], currentPlayerId: string, currentHandCount = 17): PlayerInfo[] {
  return players
    .map((player) => ({
      ...player,
      online: true,
      cards_count: player.cards_count || (player.id === currentPlayerId ? currentHandCount : 17)
    }))
    .sort((a, b) => a.seat - b.seat);
}

function syncDealtCounts(players: PlayerInfo[], currentPlayerId: string, handCount: number): PlayerInfo[] {
  return players.map((player) => ({
    ...player,
    online: true,
    cards_count: player.id === currentPlayerId ? handCount : (player.cards_count || 17)
  }));
}

function syncLandlordCounts(players: PlayerInfo[], currentPlayerId: string, landlordId: string, currentHandCount: number): PlayerInfo[] {
  return players.map((player) => {
    const isLandlord = player.id === landlordId;
    const fallbackCount = isLandlord ? 20 : 17;
    return {
      ...player,
      online: true,
      is_landlord: isLandlord,
      cards_count: player.id === currentPlayerId ? currentHandCount : (player.cards_count || fallbackCount)
    };
  });
}

function restoreSnapshot(dto: GameStateDTO, currentPlayerId: string): Partial<AppState> {
  const isPlaying = dto.phase !== 'bidding';
  const hasLandlord = (dto.players ?? []).some((player) => player.is_landlord);
  const hand = sortCards(dto.hand ?? []);
  const seatActions = dto.last_played?.length && dto.last_player_id
    ? { [dto.last_player_id]: { type: 'play' as const, player_id: dto.last_player_id, cards: dto.last_played } }
    : {};
  return {
    phase: dto.phase === 'bidding' ? 'bidding' : 'playing',
    players: normalizeGamePlayers(dto.players ?? [], currentPlayerId, hand.length || 17),
    hand,
    bottomCards: dto.bottom_cards ?? [],
    bottomCardsRevealed: isPlaying || hasLandlord,
    currentTurn: dto.current_turn ?? '',
    lastPlayed: dto.last_played ?? [],
    lastPlayedBy: dto.last_player_id ?? '',
    seatActions,
    mustPlay: dto.must_play,
    canBeat: dto.can_beat,
    cardCounter: initialCounter(dto.hand ?? []),
    recentActions: dto.last_played?.length ? [{ type: 'play', player_id: dto.last_player_id, cards: dto.last_played, label: '上一手' }] : []
  };
}

function pushAction(actions: TableAction[], action: TableAction): TableAction[] {
  return [...actions, action].slice(-10);
}

function setSeatAction(actions: Record<string, SeatAction>, action: SeatAction): Record<string, SeatAction> {
  return { ...actions, [action.player_id]: action };
}
