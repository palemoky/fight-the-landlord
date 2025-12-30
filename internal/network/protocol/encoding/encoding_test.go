package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestNewMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		msgType     protocol.MessageType
		payload     any
		expectError bool
	}{
		{
			name:        "nil payload",
			msgType:     protocol.MsgPing,
			payload:     nil,
			expectError: false,
		},
		{
			name:    "with PingPayload",
			msgType: protocol.MsgPing,
			payload: protocol.PingPayload{
				Timestamp: 12345,
			},
			expectError: false,
		},
		{
			name:    "with ChatPayload",
			msgType: protocol.MsgChat,
			payload: protocol.ChatPayload{
				Content: "hello",
				Scope:   "lobby",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, err := NewMessage(tt.msgType, tt.payload)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, msg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, msg)
				assert.Equal(t, tt.msgType, msg.Type)
				if tt.payload == nil {
					assert.Nil(t, msg.Payload)
				} else {
					assert.NotNil(t, msg.Payload)
				}
				PutMessage(msg)
			}
		})
	}
}

func TestMustNewMessage(t *testing.T) {
	t.Parallel()

	// Should not panic with valid payload
	msg := MustNewMessage(protocol.MsgPong, protocol.PongPayload{
		ClientTimestamp: 100,
		ServerTimestamp: 200,
	})
	require.NotNil(t, msg)
	assert.Equal(t, protocol.MsgPong, msg.Type)
	PutMessage(msg)
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msgType protocol.MessageType
		payload any
	}{
		{
			name:    "PingPayload",
			msgType: protocol.MsgPing,
			payload: protocol.PingPayload{Timestamp: 9999},
		},
		{
			name:    "ErrorPayload",
			msgType: protocol.MsgError,
			payload: protocol.ErrorPayload{Code: 1001, Message: "test error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create and encode
			original, err := NewMessage(tt.msgType, tt.payload)
			require.NoError(t, err)
			require.NotNil(t, original)

			encoded, err := Encode(original)
			require.NoError(t, err)
			assert.NotEmpty(t, encoded)

			// Decode
			decoded, err := Decode(encoded)
			require.NoError(t, err)
			require.NotNil(t, decoded)

			assert.Equal(t, original.Type, decoded.Type)
			assert.NotNil(t, decoded.Payload)

			PutMessage(original)
			PutMessage(decoded)
		})
	}
}

func TestParsePayload(t *testing.T) {
	t.Parallel()

	msg, err := NewMessage(protocol.MsgPing, protocol.PingPayload{Timestamp: 12345})
	require.NoError(t, err)
	require.NotNil(t, msg)

	payload, err := ParsePayload[protocol.PingPayload](msg)
	require.NoError(t, err)
	require.NotNil(t, payload)
	assert.Equal(t, int64(12345), payload.Timestamp)

	PutMessage(msg)
}

func TestNewErrorMessage(t *testing.T) {
	t.Parallel()

	msg := NewErrorMessage(protocol.ErrCodeRoomNotFound)
	require.NotNil(t, msg)
	assert.Equal(t, protocol.MsgError, msg.Type)

	payload, err := ParsePayload[protocol.ErrorPayload](msg)
	require.NoError(t, err)
	assert.Equal(t, protocol.ErrCodeRoomNotFound, payload.Code)
	assert.Equal(t, protocol.ErrorMessages[protocol.ErrCodeRoomNotFound], payload.Message)

	PutMessage(msg)
}

func TestNewErrorMessageWithText(t *testing.T) {
	t.Parallel()

	customText := "自定义错误信息"
	msg := NewErrorMessageWithText(protocol.ErrCodeUnknown, customText)
	require.NotNil(t, msg)
	assert.Equal(t, protocol.MsgError, msg.Type)

	payload, err := ParsePayload[protocol.ErrorPayload](msg)
	require.NoError(t, err)
	assert.Equal(t, protocol.ErrCodeUnknown, payload.Code)
	assert.Equal(t, customText, payload.Message)

	PutMessage(msg)
}
