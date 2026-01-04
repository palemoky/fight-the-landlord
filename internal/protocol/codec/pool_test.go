package codec

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/pb"
)

func TestMessagePool_GetPut(t *testing.T) {
	t.Parallel()

	// Get message from pool
	msg := GetMessage()
	assert.NotNil(t, msg)

	// Use the message
	msg.Type = "test"
	msg.Payload = []byte("data")

	// Put back to pool
	PutMessage(msg)

	// Get again - should be reset
	msg2 := GetMessage()
	assert.NotNil(t, msg2)
	assert.Empty(t, msg2.Type)
	assert.Nil(t, msg2.Payload)
}

func TestMessagePool_PutNil(t *testing.T) {
	t.Parallel()

	// Should not panic
	assert.NotPanics(t, func() {
		PutMessage(nil)
	})
}

func TestPBMessagePool_GetPut(t *testing.T) {
	t.Parallel()

	// Get pb message from pool
	msg := GetPBMessage()
	assert.NotNil(t, msg)

	// Use the message
	msg.Type = pb.MessageType_MSG_PING
	msg.Payload = []byte("test")

	// Put back to pool
	PutPBMessage(msg)

	// Get again - should be reset
	msg2 := GetPBMessage()
	assert.NotNil(t, msg2)
	assert.Equal(t, pb.MessageType_MSG_UNKNOWN, msg2.Type)
	assert.Empty(t, msg2.Payload)
}

func TestPBMessagePool_PutNil(t *testing.T) {
	t.Parallel()

	// Should not panic
	assert.NotPanics(t, func() {
		PutPBMessage(nil)
	})
}

func TestBufferPool_GetPut(t *testing.T) {
	t.Parallel()

	// Get buffer from pool
	buf := GetBuffer()
	assert.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len())

	// Use the buffer
	buf.WriteString("test data")
	assert.Equal(t, 9, buf.Len())

	// Put back to pool
	PutBuffer(buf)

	// Get again - should be reset
	buf2 := GetBuffer()
	assert.NotNil(t, buf2)
	assert.Equal(t, 0, buf2.Len())
}

func TestBufferPool_PutNil(t *testing.T) {
	t.Parallel()

	// Should not panic
	assert.NotPanics(t, func() {
		PutBuffer(nil)
	})
}

func TestMessagePool_Concurrency(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent get/put
	for range iterations {
		wg.Go(func() {
			msg := GetMessage()
			msg.Type = "concurrent"
			msg.Payload = []byte("test")
			PutMessage(msg)
		})
	}

	wg.Wait()
	// If we get here without panic, concurrency is safe
}

func TestPBMessagePool_Concurrency(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent get/put
	for range iterations {
		wg.Go(func() {
			msg := GetPBMessage()
			msg.Type = pb.MessageType_MSG_PING
			msg.Payload = []byte("test")
			PutPBMessage(msg)
		})
	}

	wg.Wait()
	// If we get here without panic, concurrency is safe
}

func TestBufferPool_Concurrency(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent get/put
	for range iterations {
		wg.Go(func() {
			buf := GetBuffer()
			buf.WriteString("concurrent test")
			PutBuffer(buf)
		})
	}

	wg.Wait()
	// If we get here without panic, concurrency is safe
}

func TestMessagePool_Reuse(t *testing.T) {
	t.Parallel()

	// Get and put multiple times
	for range 10 {
		msg := GetMessage()
		msg.Type = "reuse"
		msg.Payload = []byte("data")
		PutMessage(msg)
	}

	// Verify pool is working (messages are being reused)
	msg := GetMessage()
	assert.NotNil(t, msg)
	assert.Empty(t, msg.Type) // Should be reset
}

func TestBufferPool_CapacityPreserved(t *testing.T) {
	t.Parallel()

	// Get buffer and write large data
	buf := GetBuffer()
	largeData := make([]byte, 1024)
	buf.Write(largeData)

	capacity := buf.Cap()
	assert.GreaterOrEqual(t, capacity, 1024)

	// Put back
	PutBuffer(buf)

	// Get again - capacity should be preserved
	buf2 := GetBuffer()
	assert.GreaterOrEqual(t, buf2.Cap(), capacity)
	assert.Equal(t, 0, buf2.Len()) // But length should be 0
}

func BenchmarkMessagePool_GetPut(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := GetMessage()
			msg.Type = "benchmark"
			msg.Payload = []byte("test")
			PutMessage(msg)
		}
	})
}

func BenchmarkMessagePool_NoPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := &protocol.Message{}
			// No pool - just let GC handle it
			_ = msg
		}
	})
}

func BenchmarkBufferPool_GetPut(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := GetBuffer()
			buf.WriteString("benchmark test data")
			PutBuffer(buf)
		}
	})
}
