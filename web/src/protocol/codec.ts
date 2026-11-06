import protobuf from 'protobufjs';
import messageProto from './proto/message.proto?raw';
import commonProto from './proto/common.proto?raw';
import clientProto from './proto/client.proto?raw';
import serverProto from './proto/server.proto?raw';
import gameProto from './proto/game.proto?raw';
import type { IncomingMessage, MessageType, OutgoingPayload } from './types';

const schemaSource = buildSchema([messageProto, commonProto, clientProto, serverProto, gameProto]);
const parseRoot = new protobuf.Root();
const root = protobuf.parse(schemaSource, parseRoot, { keepCase: true }).root;
const Message = root.lookupType('protocol.Message');
const MessageTypeEnum = root.lookupEnum('protocol.MessageType');

const enumToType = new Map<number, string>();
const typeToEnum = new Map<string, number>();

for (const [name, value] of Object.entries(MessageTypeEnum.values)) {
  const wireName = name.replace(/^MSG_/, '').toLowerCase();
  enumToType.set(value, wireName);
  typeToEnum.set(wireName, value);
}

typeToEnum.set('maintenance_pull', MessageTypeEnum.values.MSG_MAINTENANCE_STATUS);
typeToEnum.set('maintenance_push', MessageTypeEnum.values.MSG_MAINTENANCE);
enumToType.set(MessageTypeEnum.values.MSG_MAINTENANCE_STATUS, 'maintenance_pull');
enumToType.set(MessageTypeEnum.values.MSG_MAINTENANCE, 'maintenance_push');

const payloadNames: Record<string, string> = {
  reconnect: 'ReconnectPayload',
  ping: 'PingPayload',
  join_room: 'JoinRoomPayload',
  bid: 'BidPayload',
  play_cards: 'PlayCardsPayload',
  get_leaderboard: 'GetLeaderboardPayload',
  connected: 'ConnectedPayload',
  reconnected: 'ReconnectedPayload',
  pong: 'PongPayload',
  player_offline: 'PlayerOfflinePayload',
  player_online: 'PlayerOnlinePayload',
  online_count: 'OnlineCountPayload',
  room_created: 'RoomCreatedPayload',
  room_joined: 'RoomJoinedPayload',
  player_joined: 'PlayerJoinedPayload',
  player_left: 'PlayerLeftPayload',
  player_ready: 'PlayerReadyPayload',
  game_start: 'GameStartPayload',
  deal_cards: 'DealCardsPayload',
  bid_turn: 'BidTurnPayload',
  bid_result: 'BidResultPayload',
  landlord: 'LandlordPayload',
  play_turn: 'PlayTurnPayload',
  card_played: 'CardPlayedPayload',
  player_pass: 'PlayerPassPayload',
  game_over: 'GameOverPayload',
  stats_result: 'StatsResultPayload',
  leaderboard_result: 'LeaderboardResultPayload',
  room_list_result: 'RoomListResultPayload',
  maintenance: 'MaintenancePayload',
  maintenance_push: 'MaintenancePayload',
  maintenance_status: 'MaintenanceStatusPayload',
  maintenance_pull: 'MaintenanceStatusPayload',
  error: 'ErrorPayload',
  chat: 'ChatPayload'
};

const payloadTypes = new Map<string, protobuf.Type>();
for (const [type, name] of Object.entries(payloadNames)) {
  payloadTypes.set(type, root.lookupType(`protocol.${name}`));
}

const jsonPayloadTypes = new Set(['chat']);

export function encodeMessage(type: MessageType | string, payload?: OutgoingPayload): Uint8Array<ArrayBufferLike> {
  const enumValue = typeToEnum.get(type) ?? MessageTypeEnum.values.MSG_UNKNOWN;
  const PayloadType = payloadTypes.get(type);
  let payloadBytes: Uint8Array<ArrayBufferLike> = new Uint8Array();

  if (payload && PayloadType && !jsonPayloadTypes.has(type)) {
    payloadBytes = PayloadType.encode(PayloadType.fromObject(payload)).finish();
  } else if (payload) {
    payloadBytes = new TextEncoder().encode(JSON.stringify(payload));
  }

  return Message.encode(Message.fromObject({ type: enumValue, payload: payloadBytes })).finish();
}

export function decodeMessage(data: ArrayBuffer | Uint8Array): IncomingMessage {
  const buffer = data instanceof Uint8Array ? data : new Uint8Array(data);
  const decoded = Message.decode(buffer) as unknown as { type: number; payload?: Uint8Array };
  const type = enumToType.get(decoded.type) ?? 'unknown';
  const payload = decodePayload(type, decoded.payload);
  return { type, payload } as IncomingMessage;
}

function decodePayload(type: string, bytes?: Uint8Array): unknown {
  if (!bytes || bytes.length === 0) return {};
  if (jsonPayloadTypes.has(type)) {
    try {
      return JSON.parse(new TextDecoder().decode(bytes));
    } catch {
      const PayloadType = payloadTypes.get(type);
      if (PayloadType) {
        const message = PayloadType.decode(bytes);
        return PayloadType.toObject(message, {
          longs: Number,
          enums: Number,
          defaults: true,
          arrays: true,
          objects: true
        });
      }
    }
  }
  const PayloadType = payloadTypes.get(type);
  if (PayloadType) {
    const message = PayloadType.decode(bytes);
    return PayloadType.toObject(message, {
      longs: Number,
      enums: Number,
      defaults: true,
      arrays: true,
      objects: true
    });
  }

  try {
    return JSON.parse(new TextDecoder().decode(bytes));
  } catch {
    return { _raw: Array.from(bytes) };
  }
}

function buildSchema(parts: string[]): string {
  return parts
    .join('\n')
    .replace(/^syntax\s*=.*?;\s*$/gm, '')
    .replace(/^package\s+protocol;\s*$/gm, '')
    .replace(/^option\s+.*?;\s*$/gm, '')
    .replace(/^import\s+.*?;\s*$/gm, '')
    .replace(/^\s*$/gm, '')
    .replace(/^/, 'syntax = "proto3";\npackage protocol;\n');
}
