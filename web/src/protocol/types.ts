export const MsgType = {
  Reconnect: 'reconnect',
  Ping: 'ping',
  CreateRoom: 'create_room',
  JoinRoom: 'join_room',
  LeaveRoom: 'leave_room',
  QuickMatch: 'quick_match',
  PracticeMatch: 'practice_match',
  Ready: 'ready',
  CancelReady: 'cancel_ready',
  Bid: 'bid',
  PlayCards: 'play_cards',
  Pass: 'pass',
  GetStats: 'get_stats',
  GetLeaderboard: 'get_leaderboard',
  GetRoomList: 'get_room_list',
  GetOnlineCount: 'get_online_count',
  GetMaintenanceStatus: 'get_maintenance_status',
  Chat: 'chat',
  Connected: 'connected',
  Reconnected: 'reconnected',
  Pong: 'pong',
  PlayerOffline: 'player_offline',
  PlayerOnline: 'player_online',
  OnlineCount: 'online_count',
  RoomCreated: 'room_created',
  RoomJoined: 'room_joined',
  PlayerJoined: 'player_joined',
  PlayerLeft: 'player_left',
  PlayerReady: 'player_ready',
  MatchFound: 'match_found',
  GameStart: 'game_start',
  DealCards: 'deal_cards',
  BidTurn: 'bid_turn',
  BidResult: 'bid_result',
  Landlord: 'landlord',
  PlayTurn: 'play_turn',
  CardPlayed: 'card_played',
  PlayerPass: 'player_pass',
  GameOver: 'game_over',
  RoundResult: 'round_result',
  StatsResult: 'stats_result',
  LeaderboardResult: 'leaderboard_result',
  RoomListResult: 'room_list_result',
  MaintenancePush: 'maintenance_push',
  MaintenancePull: 'maintenance_pull',
  Maintenance: 'maintenance',
  MaintenanceStatus: 'maintenance_status',
  Error: 'error'
} as const;

export type MessageType = (typeof MsgType)[keyof typeof MsgType];
export type Phase = 'connecting' | 'lobby' | 'matching' | 'waiting' | 'bidding' | 'playing' | 'game_over';
export type LobbyPanel = 'home' | 'leaderboard' | 'stats' | 'rules' | 'chat';
export type UtilityDrawer = 'none' | 'chat' | 'counter' | 'history' | 'rules';

export interface CardInfo {
  suit: number;
  rank: number;
  color: number;
}

export interface PlayerInfo {
  id: string;
  name: string;
  seat: number;
  ready: boolean;
  is_landlord: boolean;
  cards_count: number;
  online: boolean;
  is_bot?: boolean;
}

export interface PlayerHand {
  player_id: string;
  player_name: string;
  cards: CardInfo[];
}

export interface PlayerScore {
  player_id: string;
  player_name: string;
  is_landlord: boolean;
  score: number;
}

export interface GameStateDTO {
  phase: 'bidding' | 'playing' | string;
  players: PlayerInfo[];
  hand: CardInfo[];
  bottom_cards: CardInfo[];
  current_turn: string;
  last_played: CardInfo[];
  last_player_id: string;
  must_play: boolean;
  can_beat: boolean;
}

export interface LeaderboardEntry {
  rank: number;
  player_id: string;
  player_name: string;
  score: number;
  wins: number;
  win_rate: number;
}

export interface RoomListItem {
  room_code: string;
  player_count: number;
  max_players: number;
}

export interface StatsResultPayload {
  player_id: string;
  player_name: string;
  total_games: number;
  wins: number;
  losses: number;
  win_rate: number;
  landlord_games: number;
  landlord_wins: number;
  farmer_games: number;
  farmer_wins: number;
  score: number;
  rank: number;
  current_streak: number;
  max_win_streak: number;
}

export interface ChatPayload {
  sender_id?: string;
  sender_name?: string;
  content: string;
  scope: 'lobby' | 'room' | string;
  time?: number;
  is_system?: boolean;
}

export interface ReconnectedPayload {
  player_id: string;
  player_name: string;
  room_code?: string;
  game_state?: GameStateDTO;
}

export interface IncomingPayloadMap {
  connected: { player_id: string; player_name: string; reconnect_token: string };
  reconnected: ReconnectedPayload;
  pong: { client_timestamp: number; server_timestamp: number };
  online_count: { count: number };
  room_created: { room_code: string; player: PlayerInfo };
  room_joined: { room_code: string; player: PlayerInfo; players: PlayerInfo[] };
  player_joined: { player: PlayerInfo };
  player_left: { player_id: string; player_name: string };
  player_ready: { player_id: string; ready: boolean };
  match_found: Record<string, never>;
  game_start: { players: PlayerInfo[] };
  deal_cards: { cards: CardInfo[]; bottom_cards: CardInfo[] };
  bid_turn: { player_id: string; timeout: number; is_grab: boolean; multiplier: number };
  bid_result: { player_id: string; player_name: string; bid: boolean; is_grab: boolean; multiplier: number };
  landlord: { player_id: string; player_name: string; bottom_cards: CardInfo[]; multiplier: number };
  play_turn: { player_id: string; timeout: number; must_play: boolean; can_beat: boolean };
  card_played: { player_id: string; player_name: string; cards: CardInfo[]; cards_left: number; hand_type: string };
  player_pass: { player_id: string; player_name: string };
  game_over: { winner_id: string; winner_name: string; is_landlord: boolean; player_hands: PlayerHand[]; multiplier: number; scores: PlayerScore[] };
  stats_result: StatsResultPayload;
  leaderboard_result: { type: string; entries: LeaderboardEntry[] };
  room_list_result: { rooms: RoomListItem[] };
  chat: ChatPayload;
  player_offline: { player_id: string; player_name: string; timeout: number };
  player_online: { player_id: string; player_name: string };
  maintenance: { maintenance: boolean };
  maintenance_status: { maintenance: boolean };
  maintenance_push: { maintenance: boolean };
  maintenance_pull: { maintenance: boolean };
  error: { code: number; message: string };
}

export type IncomingMessage = {
  [K in keyof IncomingPayloadMap]: { type: K; payload: IncomingPayloadMap[K] }
}[keyof IncomingPayloadMap] | { type: string; payload: unknown };

export type OutgoingPayload = Record<string, unknown> | undefined;
