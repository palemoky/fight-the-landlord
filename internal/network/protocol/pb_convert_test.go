package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

func TestMessageTypeConversion(t *testing.T) {
	tests := []struct {
		str    string
		pbEnum pb.MessageType
	}{
		{"join_room", pb.MessageType_MSG_JOIN_ROOM},
		{"play_cards", pb.MessageType_MSG_PLAY_CARDS},
		{"room_created", pb.MessageType_MSG_ROOM_CREATED},
		{"game_start", pb.MessageType_MSG_GAME_START},
		{"unknown_random_string", pb.MessageType_MSG_UNKNOWN},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			t.Parallel()
			// Test String -> Proto
			pbEnum := stringToProtoMessageType(tt.str)
			assert.Equal(t, tt.pbEnum, pbEnum)

			// Test Proto -> String (only if not unknown, as unknown maps to empty or "unknown" depending on impl)
			if tt.str != "unknown_random_string" {
				str := protoMessageTypeToString(tt.pbEnum)
				assert.Equal(t, tt.str, str)
			}
		})
	}
}
