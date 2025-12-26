package protocol

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	// Test creating a simple message
	payload := JoinRoomPayload{RoomCode: "1234"}
	msg, err := NewMessage(MsgJoinRoom, payload)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, MsgJoinRoom, msg.Type)
	assert.NotEmpty(t, msg.Payload)
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
	// Test round trip for a specific payload type (Protobuf based)
	original := RoomCreatedPayload{
		RoomCode: "9999",
		Player: PlayerInfo{
			ID: "p1", Name: "Player 1", Seat: 0,
		},
	}

	// Encode
	encodedBytes, err := EncodePayload(MsgRoomCreated, original)
	assert.NoError(t, err)
	assert.NotEmpty(t, encodedBytes)

	// Decode
	var decoded RoomCreatedPayload
	err = DecodePayload(MsgRoomCreated, encodedBytes, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, original.RoomCode, decoded.RoomCode)
	assert.Equal(t, original.Player.ID, decoded.Player.ID)
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
