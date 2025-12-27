package protocol

import (
	"bytes"
	"compress/flate"
	"io"
	"testing"
)

// BenchmarkCompressionOverhead 测试压缩开销
func BenchmarkCompressionOverhead(b *testing.B) {
	// 创建一个典型的游戏消息
	msg, _ := NewMessage(MsgPlayCards, PlayCardsPayload{
		Cards: []CardInfo{
			{Rank: 3, Suit: 0},
			{Rank: 3, Suit: 1},
			{Rank: 4, Suit: 0},
		},
	})
	defer PutMessage(msg)

	data, _ := msg.Encode()

	b.Run("NoCompression", func(b *testing.B) {
		for b.Loop() {
			_ = append([]byte(nil), data...)
		}
	})

	b.Run("WithFlateCompression", func(b *testing.B) {
		for b.Loop() {
			var buf bytes.Buffer
			w, _ := flate.NewWriter(&buf, flate.BestSpeed)
			_, _ = w.Write(data)
			_ = w.Close()
			_ = buf.Bytes()
		}
	})

	b.Run("WithFlateDecompression", func(b *testing.B) {
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flate.BestSpeed)
		_, _ = w.Write(data)
		_ = w.Close()
		compressed := buf.Bytes()

		b.ResetTimer()
		for b.Loop() {
			r := flate.NewReader(bytes.NewReader(compressed))
			_, _ = io.ReadAll(r)
			_ = r.Close()
		}
	})
}

// BenchmarkMessageSizeComparison 对比不同消息大小的压缩效果
func BenchmarkMessageSizeComparison(b *testing.B) {
	testCases := []struct {
		name    string
		msgType MessageType
		payload any
	}{
		{
			name:    "SmallMessage_Ping",
			msgType: MsgPing,
			payload: PingPayload{Timestamp: 123456},
		},
		{
			name:    "MediumMessage_PlayCards",
			msgType: MsgPlayCards,
			payload: PlayCardsPayload{
				Cards: []CardInfo{
					{Rank: 3, Suit: 0},
					{Rank: 3, Suit: 1},
					{Rank: 4, Suit: 0},
					{Rank: 4, Suit: 1},
					{Rank: 5, Suit: 0},
				},
			},
		},
		{
			name:    "LargeMessage_RoomList",
			msgType: MsgRoomListResult,
			payload: RoomListResultPayload{
				Rooms: []RoomListItem{
					{RoomCode: "ROOM1", PlayerCount: 2},
					{RoomCode: "ROOM2", PlayerCount: 3},
					{RoomCode: "ROOM3", PlayerCount: 1},
					{RoomCode: "ROOM4", PlayerCount: 2},
					{RoomCode: "ROOM5", PlayerCount: 3},
				},
			},
		},
	}

	for _, tc := range testCases {
		msg, _ := NewMessage(tc.msgType, tc.payload)
		data, _ := msg.Encode()
		PutMessage(msg)

		// 测试压缩后的大小
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flate.BestSpeed)
		_, _ = w.Write(data)
		_ = w.Close()
		compressed := buf.Bytes()

		ratio := float64(len(compressed)) / float64(len(data)) * 100

		b.Logf("%s: Original=%d bytes, Compressed=%d bytes, Ratio=%.1f%%",
			tc.name, len(data), len(compressed), ratio)
	}
}
