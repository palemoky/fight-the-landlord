import { describe, expect, it } from 'vitest';
import { decodeMessage, encodeMessage } from '../src/protocol/codec';
import { MsgType } from '../src/protocol/types';

describe('protocol codec', () => {
  it('round trips room join payloads', () => {
    const decoded = decodeMessage(encodeMessage(MsgType.JoinRoom, { room_code: '123456' }));
    expect(decoded.type).toBe(MsgType.JoinRoom);
    expect(decoded.payload).toEqual({ room_code: '123456' });
  });

  it('round trips play cards payloads', () => {
    const decoded = decodeMessage(encodeMessage(MsgType.PlayCards, {
      cards: [
        { suit: 1, rank: 14, color: 1 },
        { suit: 3, rank: 14, color: 1 }
      ]
    }));
    expect(decoded.type).toBe(MsgType.PlayCards);
    expect(decoded.payload).toEqual({
      cards: [
        { suit: 1, rank: 14, color: 1 },
        { suit: 3, rank: 14, color: 1 }
      ]
    });
  });

  it('decodes maintenance aliases using Go protocol names', () => {
    const decoded = decodeMessage(encodeMessage(MsgType.MaintenancePush, { maintenance: true }));
    expect(decoded.type).toBe(MsgType.MaintenancePush);
    expect(decoded.payload).toEqual({ maintenance: true });
  });

  it('round trips chat payloads as JSON for the current Go fallback', () => {
    const decoded = decodeMessage(encodeMessage(MsgType.Chat, {
      sender_id: 'p1',
      sender_name: '青竹',
      content: '你好',
      scope: 'lobby',
      time: 1
    }));
    expect(decoded.type).toBe(MsgType.Chat);
    expect(decoded.payload).toEqual({
      sender_id: 'p1',
      sender_name: '青竹',
      content: '你好',
      scope: 'lobby',
      time: 1
    });
  });
});
