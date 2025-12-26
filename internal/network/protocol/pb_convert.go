package protocol

import (
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// stringToProtoMessageType 字符串消息类型转 protobuf 枚举
//
//nolint:gocyclo // Simple mapping function with many cases
func stringToProtoMessageType(s string) pb.MessageType {
	switch s {
	case "reconnect":
		return pb.MessageType_MSG_RECONNECT
	case "ping":
		return pb.MessageType_MSG_PING
	case "create_room":
		return pb.MessageType_MSG_CREATE_ROOM
	case "join_room":
		return pb.MessageType_MSG_JOIN_ROOM
	case "leave_room":
		return pb.MessageType_MSG_LEAVE_ROOM
	case "quick_match":
		return pb.MessageType_MSG_QUICK_MATCH
	case "ready":
		return pb.MessageType_MSG_READY
	case "cancel_ready":
		return pb.MessageType_MSG_CANCEL_READY
	case "bid":
		return pb.MessageType_MSG_BID
	case "play_cards":
		return pb.MessageType_MSG_PLAY_CARDS
	case "pass":
		return pb.MessageType_MSG_PASS
	case "get_stats":
		return pb.MessageType_MSG_GET_STATS
	case "get_leaderboard":
		return pb.MessageType_MSG_GET_LEADERBOARD
	case "get_room_list":
		return pb.MessageType_MSG_GET_ROOM_LIST
	case "get_online_count":
		return pb.MessageType_MSG_GET_ONLINE_COUNT
	case "chat":
		return pb.MessageType_MSG_CHAT
	case "connected":
		return pb.MessageType_MSG_CONNECTED
	case "reconnected":
		return pb.MessageType_MSG_RECONNECTED
	case "pong":
		return pb.MessageType_MSG_PONG
	case "player_offline":
		return pb.MessageType_MSG_PLAYER_OFFLINE
	case "player_online":
		return pb.MessageType_MSG_PLAYER_ONLINE
	case "online_count":
		return pb.MessageType_MSG_ONLINE_COUNT
	case "room_created":
		return pb.MessageType_MSG_ROOM_CREATED
	case "room_joined":
		return pb.MessageType_MSG_ROOM_JOINED
	case "player_joined":
		return pb.MessageType_MSG_PLAYER_JOINED
	case "player_left":
		return pb.MessageType_MSG_PLAYER_LEFT
	case "player_ready":
		return pb.MessageType_MSG_PLAYER_READY
	case "match_found":
		return pb.MessageType_MSG_MATCH_FOUND
	case "game_start":
		return pb.MessageType_MSG_GAME_START
	case "deal_cards":
		return pb.MessageType_MSG_DEAL_CARDS
	case "bid_turn":
		return pb.MessageType_MSG_BID_TURN
	case "bid_result":
		return pb.MessageType_MSG_BID_RESULT
	case "landlord":
		return pb.MessageType_MSG_LANDLORD
	case "play_turn":
		return pb.MessageType_MSG_PLAY_TURN
	case "card_played":
		return pb.MessageType_MSG_CARD_PLAYED
	case "player_pass":
		return pb.MessageType_MSG_PLAYER_PASS
	case "game_over":
		return pb.MessageType_MSG_GAME_OVER
	case "round_result":
		return pb.MessageType_MSG_ROUND_RESULT
	case "stats_result":
		return pb.MessageType_MSG_STATS_RESULT
	case "leaderboard_result":
		return pb.MessageType_MSG_LEADERBOARD_RESULT
	case "room_list_result":
		return pb.MessageType_MSG_ROOM_LIST_RESULT
	case "error":
		return pb.MessageType_MSG_ERROR
	default:
		return pb.MessageType_MSG_UNKNOWN
	}
}

// protoMessageTypeToString protobuf 枚举转字符串消息类型
//
//nolint:gocyclo // Simple mapping function with many cases
func protoMessageTypeToString(t pb.MessageType) string {
	switch t {
	case pb.MessageType_MSG_RECONNECT:
		return "reconnect"
	case pb.MessageType_MSG_PING:
		return "ping"
	case pb.MessageType_MSG_CREATE_ROOM:
		return "create_room"
	case pb.MessageType_MSG_JOIN_ROOM:
		return "join_room"
	case pb.MessageType_MSG_LEAVE_ROOM:
		return "leave_room"
	case pb.MessageType_MSG_QUICK_MATCH:
		return "quick_match"
	case pb.MessageType_MSG_READY:
		return "ready"
	case pb.MessageType_MSG_CANCEL_READY:
		return "cancel_ready"
	case pb.MessageType_MSG_BID:
		return "bid"
	case pb.MessageType_MSG_PLAY_CARDS:
		return "play_cards"
	case pb.MessageType_MSG_PASS:
		return "pass"
	case pb.MessageType_MSG_GET_STATS:
		return "get_stats"
	case pb.MessageType_MSG_GET_LEADERBOARD:
		return "get_leaderboard"
	case pb.MessageType_MSG_GET_ROOM_LIST:
		return "get_room_list"
	case pb.MessageType_MSG_GET_ONLINE_COUNT:
		return "get_online_count"
	case pb.MessageType_MSG_CHAT:
		return "chat"
	case pb.MessageType_MSG_CONNECTED:
		return "connected"
	case pb.MessageType_MSG_RECONNECTED:
		return "reconnected"
	case pb.MessageType_MSG_PONG:
		return "pong"
	case pb.MessageType_MSG_PLAYER_OFFLINE:
		return "player_offline"
	case pb.MessageType_MSG_PLAYER_ONLINE:
		return "player_online"
	case pb.MessageType_MSG_ONLINE_COUNT:
		return "online_count"
	case pb.MessageType_MSG_ROOM_CREATED:
		return "room_created"
	case pb.MessageType_MSG_ROOM_JOINED:
		return "room_joined"
	case pb.MessageType_MSG_PLAYER_JOINED:
		return "player_joined"
	case pb.MessageType_MSG_PLAYER_LEFT:
		return "player_left"
	case pb.MessageType_MSG_PLAYER_READY:
		return "player_ready"
	case pb.MessageType_MSG_MATCH_FOUND:
		return "match_found"
	case pb.MessageType_MSG_GAME_START:
		return "game_start"
	case pb.MessageType_MSG_DEAL_CARDS:
		return "deal_cards"
	case pb.MessageType_MSG_BID_TURN:
		return "bid_turn"
	case pb.MessageType_MSG_BID_RESULT:
		return "bid_result"
	case pb.MessageType_MSG_LANDLORD:
		return "landlord"
	case pb.MessageType_MSG_PLAY_TURN:
		return "play_turn"
	case pb.MessageType_MSG_CARD_PLAYED:
		return "card_played"
	case pb.MessageType_MSG_PLAYER_PASS:
		return "player_pass"
	case pb.MessageType_MSG_GAME_OVER:
		return "game_over"
	case pb.MessageType_MSG_ROUND_RESULT:
		return "round_result"
	case pb.MessageType_MSG_STATS_RESULT:
		return "stats_result"
	case pb.MessageType_MSG_LEADERBOARD_RESULT:
		return "leaderboard_result"
	case pb.MessageType_MSG_ROOM_LIST_RESULT:
		return "room_list_result"
	case pb.MessageType_MSG_ERROR:
		return "error"
	default:
		return "unknown"
	}
}
