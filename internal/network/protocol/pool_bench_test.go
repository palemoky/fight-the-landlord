package protocol

import (
	"testing"
)

// BenchmarkNewMessage_NoPool 测试不使用对象池的性能
func BenchmarkNewMessage_NoPool(b *testing.B) {
	payload := PingPayload{Timestamp: 123456}

	for b.Loop() {
		data, _ := EncodePayload(MsgPing, payload)
		msg := &Message{
			Type:    MsgPing,
			Payload: data,
		}
		_ = msg
	}
}

// BenchmarkNewMessage_WithPool 测试使用对象池的性能
func BenchmarkNewMessage_WithPool(b *testing.B) {
	payload := PingPayload{Timestamp: 123456}

	for b.Loop() {
		msg, _ := NewMessage(MsgPing, payload)
		PutMessage(msg)
	}
}

// BenchmarkEncode_NoPool 测试编码不使用对象池
func BenchmarkEncode_NoPool(b *testing.B) {
	msg := &Message{
		Type:    MsgPing,
		Payload: []byte("test payload"),
	}

	for b.Loop() {
		_, _ = msg.Encode()
	}
}

// BenchmarkEncode_WithPool 测试编码使用对象池
func BenchmarkEncode_WithPool(b *testing.B) {
	msg, _ := NewMessage(MsgPing, PingPayload{Timestamp: 123456})
	defer PutMessage(msg)

	for b.Loop() {
		_, _ = msg.Encode()
	}
}

// BenchmarkDecode_NoPool 测试解码不使用对象池
func BenchmarkDecode_NoPool(b *testing.B) {
	msg, _ := NewMessage(MsgPing, PingPayload{Timestamp: 123456})
	data, _ := msg.Encode()
	PutMessage(msg)

	for b.Loop() {
		decodedMsg, _ := Decode(data)
		_ = decodedMsg
	}
}

// BenchmarkDecode_WithPool 测试解码使用对象池
func BenchmarkDecode_WithPool(b *testing.B) {
	msg, _ := NewMessage(MsgPing, PingPayload{Timestamp: 123456})
	data, _ := msg.Encode()
	PutMessage(msg)

	for b.Loop() {
		decodedMsg, _ := Decode(data)
		PutMessage(decodedMsg)
	}
}

// BenchmarkMessageLifecycle 测试完整的消息生命周期
func BenchmarkMessageLifecycle(b *testing.B) {
	payload := PingPayload{Timestamp: 123456}

	for b.Loop() {
		// 创建
		msg, _ := NewMessage(MsgPing, payload)

		// 编码
		data, _ := msg.Encode()

		// 解码
		decodedMsg, _ := Decode(data)

		// 归还
		PutMessage(msg)
		PutMessage(decodedMsg)
	}
}

// BenchmarkConcurrentMessageCreation 测试并发创建消息
func BenchmarkConcurrentMessageCreation(b *testing.B) {
	payload := PingPayload{Timestamp: 123456}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg, _ := NewMessage(MsgPing, payload)
			PutMessage(msg)
		}
	})
}
