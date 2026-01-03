package convert

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// protocol.DecodePayload 从 protobuf bytes 解码为 Go struct
func DecodePayload(msgType protocol.MessageType, data []byte, target any) error {
	if len(data) == 0 {
		return nil
	}

	// 尝试作为客户端消息解码
	if ok, err := decodeClientPayload(msgType, data, target); ok {
		return err
	}

	// 尝试作为服务端消息解码
	if ok, err := decodeServerPayload(msgType, data, target); ok {
		return err
	}

	// 未知类型，回退到 JSON
	return json.Unmarshal(data, target)
}

// decodeClientPayload 解码客户端发送的消息
// 返回 (是否处理了该消息类型, 错误)
func decodeClientPayload(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgReconnect:
		var pbMsg pb.ReconnectPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.ReconnectPayload) = protocol.ReconnectPayload{
			Token:    pbMsg.Token,
			PlayerID: pbMsg.PlayerId,
		}
		return true, nil
	case protocol.MsgPing:
		var pbMsg pb.PingPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PingPayload) = protocol.PingPayload{
			Timestamp: pbMsg.Timestamp,
		}
		return true, nil
	case protocol.MsgJoinRoom:
		var pbMsg pb.JoinRoomPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.JoinRoomPayload) = protocol.JoinRoomPayload{
			RoomCode: pbMsg.RoomCode,
		}
		return true, nil
	case protocol.MsgBid:
		var pbMsg pb.BidPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.BidPayload) = protocol.BidPayload{
			Bid: pbMsg.Bid,
		}
		return true, nil
	case protocol.MsgPlayCards:
		var pbMsg pb.PlayCardsPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayCardsPayload) = protocol.PlayCardsPayload{
			Cards: protoToCards(pbMsg.Cards),
		}
		return true, nil
	case protocol.MsgGetLeaderboard:
		var pbMsg pb.GetLeaderboardPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.GetLeaderboardPayload) = protocol.GetLeaderboardPayload{
			Type:   pbMsg.Type,
			Offset: int(pbMsg.Offset),
			Limit:  int(pbMsg.Limit),
		}
		return true, nil
	}
	return false, nil
}

// decodeServerPayload 解码服务端发送的消息
// 返回 (是否处理了该消息类型, 错误)
func decodeServerPayload(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	// 尝试作为系统消息解码
	if ok, err := decodeSystemMessages(msgType, data, target); ok {
		return true, err
	}

	// 尝试作为房间消息解码
	if ok, err := decodeRoomMessages(msgType, data, target); ok {
		return true, err
	}

	// 尝试作为游戏消息解码
	if ok, err := decodeGameMessages(msgType, data, target); ok {
		return true, err
	}

	return false, nil
}

// decodeConnectionMessages 解码连接相关消息
func decodeConnectionMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgConnected:
		var pbMsg pb.ConnectedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.ConnectedPayload) = protocol.ConnectedPayload{
			PlayerID:       pbMsg.PlayerId,
			PlayerName:     pbMsg.PlayerName,
			ReconnectToken: pbMsg.ReconnectToken,
		}
		return true, nil
	case protocol.MsgPong:
		var pbMsg pb.PongPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PongPayload) = protocol.PongPayload{
			ClientTimestamp: pbMsg.ClientTimestamp,
			ServerTimestamp: pbMsg.ServerTimestamp,
		}
		return true, nil
	case protocol.MsgReconnected:
		var pbMsg pb.ReconnectedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		var gameState *protocol.GameStateDTO
		if pbMsg.GameState != nil {
			gameState = protoToGameStateDTO(pbMsg.GameState)
		}
		*target.(*protocol.ReconnectedPayload) = protocol.ReconnectedPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			RoomCode:   pbMsg.RoomCode,
			GameState:  gameState,
		}
		return true, nil
	case protocol.MsgError:
		var pbMsg pb.ErrorPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.ErrorPayload) = protocol.ErrorPayload{
			Code:    int(pbMsg.Code),
			Message: pbMsg.Message,
		}
		return true, nil
	}
	return false, nil
}

// decodeInfoMessages 解码信息查询相关消息
func decodeInfoMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgOnlineCount:
		var pbMsg pb.OnlineCountPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.OnlineCountPayload) = protocol.OnlineCountPayload{
			Count: int(pbMsg.Count),
		}
		return true, nil
	case protocol.MsgMaintenancePull:
		var pbMsg pb.MaintenanceStatusPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.MaintenanceStatusPayload) = protocol.MaintenanceStatusPayload{
			Maintenance: pbMsg.Maintenance,
		}
		return true, nil
	case protocol.MsgMaintenancePush:
		var pbMsg pb.MaintenancePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.MaintenancePayload) = protocol.MaintenancePayload{
			Maintenance: pbMsg.Maintenance,
		}
		return true, nil
	case protocol.MsgStatsResult:
		var pbMsg pb.StatsResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.StatsResultPayload) = protocol.StatsResultPayload{
			PlayerID:      pbMsg.PlayerId,
			PlayerName:    pbMsg.PlayerName,
			TotalGames:    int(pbMsg.TotalGames),
			Wins:          int(pbMsg.Wins),
			Losses:        int(pbMsg.Losses),
			WinRate:       pbMsg.WinRate,
			LandlordGames: int(pbMsg.LandlordGames),
			LandlordWins:  int(pbMsg.LandlordWins),
			FarmerGames:   int(pbMsg.FarmerGames),
			FarmerWins:    int(pbMsg.FarmerWins),
			Score:         int(pbMsg.Score),
			Rank:          int(pbMsg.Rank),
			CurrentStreak: int(pbMsg.CurrentStreak),
			MaxWinStreak:  int(pbMsg.MaxWinStreak),
		}
		return true, nil
	case protocol.MsgLeaderboardResult:
		var pbMsg pb.LeaderboardResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.LeaderboardResultPayload) = protocol.LeaderboardResultPayload{
			Type:    pbMsg.Type,
			Entries: protoToLeaderboardEntries(pbMsg.Entries),
		}
		return true, nil
	}
	return false, nil
}

// decodeSystemMessages 解码系统相关消息
func decodeSystemMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	if ok, err := decodeConnectionMessages(msgType, data, target); ok {
		return true, err
	}
	return decodeInfoMessages(msgType, data, target)
}

// decodePlayerStateMessages 解码玩家状态消息
func decodePlayerStateMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgPlayerOffline:
		var pbMsg pb.PlayerOfflinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerOfflinePayload) = protocol.PlayerOfflinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Timeout:    int(pbMsg.Timeout),
		}
		return true, nil
	case protocol.MsgPlayerOnline:
		var pbMsg pb.PlayerOnlinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerOnlinePayload) = protocol.PlayerOnlinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
		return true, nil
	case protocol.MsgPlayerJoined:
		var pbMsg pb.PlayerJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerJoinedPayload) = protocol.PlayerJoinedPayload{
			Player: protoToPlayerInfo(pbMsg.Player),
		}
		return true, nil
	case protocol.MsgPlayerLeft:
		var pbMsg pb.PlayerLeftPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerLeftPayload) = protocol.PlayerLeftPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
		return true, nil
	case protocol.MsgPlayerReady:
		var pbMsg pb.PlayerReadyPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerReadyPayload) = protocol.PlayerReadyPayload{
			PlayerID: pbMsg.PlayerId,
			Ready:    pbMsg.Ready,
		}
		return true, nil
	}
	return false, nil
}

// decodeRoomStateMessages 解码房间状态消息
func decodeRoomStateMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgRoomCreated:
		var pbMsg pb.RoomCreatedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.RoomCreatedPayload) = protocol.RoomCreatedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
		}
		return true, nil
	case protocol.MsgRoomJoined:
		var pbMsg pb.RoomJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.RoomJoinedPayload) = protocol.RoomJoinedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
			Players:  protoToPlayerInfos(pbMsg.Players),
		}
		return true, nil
	case protocol.MsgRoomListResult:
		var pbMsg pb.RoomListResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.RoomListResultPayload) = protocol.RoomListResultPayload{
			Rooms: protoToRoomListItems(pbMsg.Rooms),
		}
		return true, nil
	}
	return false, nil
}

// decodeRoomMessages 解码房间及玩家状态消息
func decodeRoomMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	if ok, err := decodePlayerStateMessages(msgType, data, target); ok {
		return true, err
	}
	return decodeRoomStateMessages(msgType, data, target)
}

// decodeBiddingMessages 解码叫地主阶段消息
func decodeBiddingMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgGameStart:
		var pbMsg pb.GameStartPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.GameStartPayload) = protocol.GameStartPayload{
			Players: protoToPlayerInfos(pbMsg.Players),
		}
		return true, nil
	case protocol.MsgDealCards:
		var pbMsg pb.DealCardsPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.DealCardsPayload) = protocol.DealCardsPayload{
			Cards:       protoToCards(pbMsg.Cards),
			BottomCards: protoToCards(pbMsg.BottomCards),
		}
		return true, nil
	case protocol.MsgBidTurn:
		var pbMsg pb.BidTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.BidTurnPayload) = protocol.BidTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
		}
		return true, nil
	case protocol.MsgBidResult:
		var pbMsg pb.BidResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.BidResultPayload) = protocol.BidResultPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Bid:        pbMsg.Bid,
		}
		return true, nil
	case protocol.MsgLandlord:
		var pbMsg pb.LandlordPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.LandlordPayload) = protocol.LandlordPayload{
			PlayerID:    pbMsg.PlayerId,
			PlayerName:  pbMsg.PlayerName,
			BottomCards: protoToCards(pbMsg.BottomCards),
		}
		return true, nil
	}
	return false, nil
}

// decodePlayingMessages 解码出牌阶段消息
func decodePlayingMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	switch msgType {
	case protocol.MsgPlayTurn:
		var pbMsg pb.PlayTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayTurnPayload) = protocol.PlayTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
			MustPlay: pbMsg.MustPlay,
			CanBeat:  pbMsg.CanBeat,
		}
		return true, nil
	case protocol.MsgCardPlayed:
		var pbMsg pb.CardPlayedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.CardPlayedPayload) = protocol.CardPlayedPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Cards:      protoToCards(pbMsg.Cards),
			CardsLeft:  int(pbMsg.CardsLeft),
			HandType:   pbMsg.HandType,
		}
		return true, nil
	case protocol.MsgPlayerPass:
		var pbMsg pb.PlayerPassPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.PlayerPassPayload) = protocol.PlayerPassPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
		return true, nil
	case protocol.MsgGameOver:
		var pbMsg pb.GameOverPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return true, err
		}
		*target.(*protocol.GameOverPayload) = protocol.GameOverPayload{
			WinnerID:    pbMsg.WinnerId,
			WinnerName:  pbMsg.WinnerName,
			IsLandlord:  pbMsg.IsLandlord,
			PlayerHands: protoToPlayerHands(pbMsg.PlayerHands),
		}
		return true, nil
	}
	return false, nil
}

// decodeGameMessages 解码游戏逻辑相关消息
func decodeGameMessages(msgType protocol.MessageType, data []byte, target any) (bool, error) {
	if ok, err := decodeBiddingMessages(msgType, data, target); ok {
		return true, err
	}
	return decodePlayingMessages(msgType, data, target)
}
