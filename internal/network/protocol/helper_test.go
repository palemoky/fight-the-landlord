package protocol

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
		payload any
		wantErr bool
	}{
		{
			name:    "Valid JoinRoom",
			msgType: MsgJoinRoom,
			payload: JoinRoomPayload{RoomCode: "1234"},
			wantErr: false,
		},
		{
			name:    "Valid PlayCards",
			msgType: MsgPlayCards,
			payload: PlayCardsPayload{},
			wantErr: false, // Should default to empty struct payload if valid
		},
		{
			name:    "Unknown Type",
			msgType: MessageType("unknown"),
			payload: nil,
			wantErr: false, // NewMessage implementation allows unknown types usually, let's verify if not
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg, err := NewMessage(tt.msgType, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.Equal(t, tt.msgType, msg.Type)
			}
		})
	}
}

func TestEncodeDecode(t *testing.T) {
	// Setup original message
	payload := JoinRoomPayload{RoomCode: "1234"}
	originalMsg, err := NewMessage(MsgJoinRoom, payload)
	assert.NoError(t, err)

	// Encode
	bytes, err := originalMsg.Encode()
	assert.NoError(t, err)
	assert.NotEmpty(t, bytes)

	// Decode
	decodedMsg, err := Decode(bytes)
	assert.NoError(t, err)
	assert.NotNil(t, decodedMsg)

	// Verify
	assert.Equal(t, originalMsg.Type, decodedMsg.Type)
	assert.Equal(t, originalMsg.Payload, decodedMsg.Payload)
}

func TestPayloadEncoding_Protobuf(t *testing.T) {
	tests := []struct {
		name    string
		msgType MessageType
		payload any
	}{
		{
			name:    "MsgRoomCreated",
			msgType: MsgRoomCreated,
			payload: RoomCreatedPayload{RoomCode: "9999", Player: PlayerInfo{ID: "p1"}},
		},
		{
			name:    "MsgJoinRoom",
			msgType: MsgJoinRoom,
			payload: JoinRoomPayload{RoomCode: "8888"},
		},
		{
			name:    "MsgGameStart",
			msgType: MsgGameStart,
			payload: GameStartPayload{Players: []PlayerInfo{{ID: "p1"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Encode
			encodedBytes, err := EncodePayload(tt.msgType, tt.payload)
			assert.NoError(t, err)
			assert.NotEmpty(t, encodedBytes)

			// Decode
			// We need to know the type to decode into.
			// In a real generic test we would use reflection or type switch,
			// but for simple table driven we can just decode back to a fresh instance of the same type
			// by creating a new pointer to the payload type.

			// However that's hard in a simple loop without reflection.
			// Let's just decode manually for the specific types we carefully selected or use reflection.

			// Simplified approach: Re-encode and compare bytes (if deterministic), or just Ensure Decode doesn't error.
			// Better: switch on type to creates vars.

			switch tt.msgType {
			case MsgRoomCreated:
				var decoded RoomCreatedPayload
				err = DecodePayload(tt.msgType, encodedBytes, &decoded)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload.(RoomCreatedPayload).RoomCode, decoded.RoomCode)
			case MsgJoinRoom:
				var decoded JoinRoomPayload
				err = DecodePayload(tt.msgType, encodedBytes, &decoded)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload.(JoinRoomPayload).RoomCode, decoded.RoomCode)
			case MsgGameStart:
				var decoded GameStartPayload
				err = DecodePayload(tt.msgType, encodedBytes, &decoded)
				assert.NoError(t, err)
				assert.Len(t, decoded.Players, 1)
			}
		})
	}
}

func TestPayloadEncoding_JSON_Fallback(t *testing.T) {
	// Test fallback for unknown type (should use JSON)
	// Assuming "unknown_type" falls back to JSON in EncodePayload/DecodePayload default case
	type CustomPayload struct {
		Foo string `json:"foo"`
	}
	original := CustomPayload{Foo: "Bar"}
	unknownType := MessageType("unknown_custom_type")

	// Encode
	encodedBytes, err := EncodePayload(unknownType, original)
	// The implementation might return nil if switch defaults to nil for "EncodePayload"
	// UNLESS it falls back to JSON. Let's check implementation.
	// Implementation says: default: return json.Marshal(payload)
	// So checking for JSON fallback.

	// Note: If implementation default returns nil (which it does NOT in view_file 168),
	// line 234: return json.Marshal(payload).
	assert.NoError(t, err)

	// Decode
	var decoded CustomPayload
	err = DecodePayload(unknownType, encodedBytes, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, original.Foo, decoded.Foo)

	// Verify it's actually JSON
	var jsonMap map[string]interface{}
	_ = json.Unmarshal(encodedBytes, &jsonMap)
	assert.Equal(t, "Bar", jsonMap["foo"])
}
