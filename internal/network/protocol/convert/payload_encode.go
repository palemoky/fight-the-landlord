package convert

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// protocol.EncodePayload 将 Go struct payload 编码为 protobuf bytes
func EncodePayload(msgType protocol.MessageType, payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}

	// 尝试作为客户端请求编码
	if pb, handled := encodeClientRequests(msgType, payload); handled {
		if pb == nil {
			return nil, nil
		}
		return proto.Marshal(pb)
	}

	// 尝试作为服务端系统消息编码
	if pb, handled := encodeServerSystemMessages(msgType, payload); handled {
		return proto.Marshal(pb)
	}

	// 尝试作为服务端房间消息编码
	if pb, handled := encodeServerRoomMessages(msgType, payload); handled {
		return proto.Marshal(pb)
	}

	// 尝试作为服务端游戏消息编码
	if pb, handled := encodeServerGameMessages(msgType, payload); handled {
		return proto.Marshal(pb)
	}

	// 未知类型，回退到 JSON
	return json.Marshal(payload)
}

// encodeClientRequests 编码客户端请求
func encodeClientRequests(msgType protocol.MessageType, payload any) (proto.Message, bool) {
	switch msgType {
	case protocol.MsgReconnect:
		p := payload.(protocol.ReconnectPayload)
		return &pb.ReconnectPayload{
			Token:    p.Token,
			PlayerId: p.PlayerID,
		}, true
	case protocol.MsgPing:
		p := payload.(protocol.PingPayload)
		return &pb.PingPayload{
			Timestamp: p.Timestamp,
		}, true
	case protocol.MsgJoinRoom:
		p := payload.(protocol.JoinRoomPayload)
		return &pb.JoinRoomPayload{
			RoomCode: p.RoomCode,
		}, true
	case protocol.MsgBid:
		p := payload.(protocol.BidPayload)
		return &pb.BidPayload{
			Bid: p.Bid,
		}, true
	case protocol.MsgPlayCards:
		p := payload.(protocol.PlayCardsPayload)
		return &pb.PlayCardsPayload{
			Cards: cardsToProto(p.Cards),
		}, true
	case protocol.MsgGetLeaderboard:
		p := payload.(protocol.GetLeaderboardPayload)
		return &pb.GetLeaderboardPayload{
			Type:   p.Type,
			Offset: int64(p.Offset),
			Limit:  int64(p.Limit),
		}, true
	case protocol.MsgGetOnlineCount, protocol.MsgGetMaintenanceStatus:
		// No payload needed for these messages
		return nil, true
	}
	return nil, false
}

// encodeServerSystemMessages 编码系统相关消息
func encodeServerSystemMessages(msgType protocol.MessageType, payload any) (proto.Message, bool) {
	switch msgType {
	case protocol.MsgConnected:
		p := payload.(protocol.ConnectedPayload)
		return &pb.ConnectedPayload{
			PlayerId:       p.PlayerID,
			PlayerName:     p.PlayerName,
			ReconnectToken: p.ReconnectToken,
		}, true
	case protocol.MsgReconnected:
		p := payload.(protocol.ReconnectedPayload)
		var gameState *pb.GameStateDTO
		if p.GameState != nil {
			gameState = gameStateDTOToProto(p.GameState)
		}
		return &pb.ReconnectedPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			RoomCode:   p.RoomCode,
			GameState:  gameState,
		}, true
	case protocol.MsgPong:
		p := payload.(protocol.PongPayload)
		return &pb.PongPayload{
			ClientTimestamp: p.ClientTimestamp,
			ServerTimestamp: p.ServerTimestamp,
		}, true
	case protocol.MsgOnlineCount:
		p := payload.(protocol.OnlineCountPayload)
		return &pb.OnlineCountPayload{
			Count: int64(p.Count),
		}, true
	case protocol.MsgMaintenancePull:
		p := payload.(protocol.MaintenanceStatusPayload)
		return &pb.MaintenanceStatusPayload{
			Maintenance: p.Maintenance,
		}, true
	case protocol.MsgMaintenancePush:
		p := payload.(protocol.MaintenancePayload)
		return &pb.MaintenancePayload{
			Maintenance: p.Maintenance,
		}, true
	case protocol.MsgStatsResult:
		p := payload.(protocol.StatsResultPayload)
		return &pb.StatsResultPayload{
			PlayerId:      p.PlayerID,
			PlayerName:    p.PlayerName,
			TotalGames:    int64(p.TotalGames),
			Wins:          int64(p.Wins),
			Losses:        int64(p.Losses),
			WinRate:       p.WinRate,
			LandlordGames: int64(p.LandlordGames),
			LandlordWins:  int64(p.LandlordWins),
			FarmerGames:   int64(p.FarmerGames),
			FarmerWins:    int64(p.FarmerWins),
			Score:         int64(p.Score),
			Rank:          int64(p.Rank),
			CurrentStreak: int64(p.CurrentStreak),
			MaxWinStreak:  int64(p.MaxWinStreak),
		}, true
	case protocol.MsgLeaderboardResult:
		p := payload.(protocol.LeaderboardResultPayload)
		return &pb.LeaderboardResultPayload{
			Type:    p.Type,
			Entries: leaderboardEntriesToProto(p.Entries),
		}, true
	case protocol.MsgError:
		p := payload.(protocol.ErrorPayload)
		return &pb.ErrorPayload{
			Code:    int64(p.Code),
			Message: p.Message,
		}, true
	}
	return nil, false
}

// encodeServerRoomMessages 编码房间及玩家状态消息
func encodeServerRoomMessages(msgType protocol.MessageType, payload any) (proto.Message, bool) {
	switch msgType {
	case protocol.MsgPlayerOffline:
		p := payload.(protocol.PlayerOfflinePayload)
		return &pb.PlayerOfflinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Timeout:    int64(p.Timeout),
		}, true
	case protocol.MsgPlayerOnline:
		p := payload.(protocol.PlayerOnlinePayload)
		return &pb.PlayerOnlinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}, true
	case protocol.MsgRoomCreated:
		p := payload.(protocol.RoomCreatedPayload)
		return &pb.RoomCreatedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
		}, true
	case protocol.MsgRoomJoined:
		p := payload.(protocol.RoomJoinedPayload)
		return &pb.RoomJoinedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
			Players:  playerInfosToProto(p.Players),
		}, true
	case protocol.MsgPlayerJoined:
		p := payload.(protocol.PlayerJoinedPayload)
		return &pb.PlayerJoinedPayload{
			Player: playerInfoToProto(&p.Player),
		}, true
	case protocol.MsgPlayerLeft:
		p := payload.(protocol.PlayerLeftPayload)
		return &pb.PlayerLeftPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}, true
	case protocol.MsgPlayerReady:
		p := payload.(protocol.PlayerReadyPayload)
		return &pb.PlayerReadyPayload{
			PlayerId: p.PlayerID,
			Ready:    p.Ready,
		}, true
	case protocol.MsgRoomListResult:
		p := payload.(protocol.RoomListResultPayload)
		return &pb.RoomListResultPayload{
			Rooms: roomListItemsToProto(p.Rooms),
		}, true
	}
	return nil, false
}

// encodeServerGameMessages 编码游戏逻辑相关消息
func encodeServerGameMessages(msgType protocol.MessageType, payload any) (proto.Message, bool) {
	switch msgType {
	case protocol.MsgGameStart:
		p := payload.(protocol.GameStartPayload)
		return &pb.GameStartPayload{
			Players: playerInfosToProto(p.Players),
		}, true
	case protocol.MsgDealCards:
		p := payload.(protocol.DealCardsPayload)
		return &pb.DealCardsPayload{
			Cards:       cardsToProto(p.Cards),
			BottomCards: cardsToProto(p.BottomCards),
		}, true
	case protocol.MsgBidTurn:
		p := payload.(protocol.BidTurnPayload)
		return &pb.BidTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int64(p.Timeout),
		}, true
	case protocol.MsgBidResult:
		p := payload.(protocol.BidResultPayload)
		return &pb.BidResultPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Bid:        p.Bid,
		}, true
	case protocol.MsgLandlord:
		p := payload.(protocol.LandlordPayload)
		return &pb.LandlordPayload{
			PlayerId:    p.PlayerID,
			PlayerName:  p.PlayerName,
			BottomCards: cardsToProto(p.BottomCards),
		}, true
	case protocol.MsgPlayTurn:
		p := payload.(protocol.PlayTurnPayload)
		return &pb.PlayTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int64(p.Timeout),
			MustPlay: p.MustPlay,
			CanBeat:  p.CanBeat,
		}, true
	case protocol.MsgCardPlayed:
		p := payload.(protocol.CardPlayedPayload)
		return &pb.CardPlayedPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Cards:      cardsToProto(p.Cards),
			CardsLeft:  int64(p.CardsLeft),
			HandType:   p.HandType,
		}, true
	case protocol.MsgPlayerPass:
		p := payload.(protocol.PlayerPassPayload)
		return &pb.PlayerPassPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}, true
	case protocol.MsgGameOver:
		p := payload.(protocol.GameOverPayload)
		return &pb.GameOverPayload{
			WinnerId:    p.WinnerID,
			WinnerName:  p.WinnerName,
			IsLandlord:  p.IsLandlord,
			PlayerHands: playerHandsToProto(p.PlayerHands),
		}, true
	}
	return nil, false
}
