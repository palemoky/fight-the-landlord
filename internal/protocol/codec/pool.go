package codec

import (
	"bytes"
	"sync"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/pb"
)

// Message pools for reducing GC pressure
var (
	messagePool = sync.Pool{
		New: func() any {
			return &protocol.Message{}
		},
	}

	pbMessagePool = sync.Pool{
		New: func() any {
			return &pb.Message{}
		},
	}

	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

// GetMessage retrieves a Message from the pool
func GetMessage() *protocol.Message {
	return messagePool.Get().(*protocol.Message)
}

// PutMessage returns a Message to the pool
// The message fields are reset to prevent memory leaks
func PutMessage(msg *protocol.Message) {
	if msg == nil {
		return
	}
	// Reset fields to avoid holding references
	msg.Type = ""
	msg.Payload = nil
	messagePool.Put(msg)
}

// GetPBMessage retrieves a pb.Message from the pool
func GetPBMessage() *pb.Message {
	return pbMessagePool.Get().(*pb.Message)
}

// PutPBMessage returns a pb.Message to the pool
func PutPBMessage(msg *pb.Message) {
	if msg == nil {
		return
	}
	// Reset the message
	msg.Reset()
	pbMessagePool.Put(msg)
}

// GetBuffer retrieves a bytes.Buffer from the pool
func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// PutBuffer returns a bytes.Buffer to the pool
// The buffer is reset but capacity is preserved
func PutBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	bufferPool.Put(buf)
}
