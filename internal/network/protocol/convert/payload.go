package convert

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// protocol.EncodePayload 将 Go struct payload 编码为 protobuf bytes
//
//nolint:gocyclo // Payload conversion function with many message types
func EncodePayload(msgType protocol.MessageType, payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}

	var pbPayload proto.Message

	switch msgType {
	// 客户端请求
	case protocol.MsgReconnect:
		p := payload.(protocol.ReconnectPayload)
		pbPayload = &pb.ReconnectPayload{
			Token:    p.Token,
			PlayerId: p.PlayerID,
		}
	case protocol.MsgPing:
		p := payload.(protocol.PingPayload)
		pbPayload = &pb.PingPayload{
			Timestamp: p.Timestamp,
		}
	case protocol.MsgJoinRoom:
		p := payload.(protocol.JoinRoomPayload)
		pbPayload = &pb.JoinRoomPayload{
			RoomCode: p.RoomCode,
		}
	case protocol.MsgBid:
		p := payload.(protocol.BidPayload)
		pbPayload = &pb.BidPayload{
			Bid: p.Bid,
		}
	case protocol.MsgPlayCards:
		p := payload.(protocol.PlayCardsPayload)
		pbPayload = &pb.PlayCardsPayload{
			Cards: cardsToProto(p.Cards),
		}
	case protocol.MsgGetLeaderboard:
		p := payload.(protocol.GetLeaderboardPayload)
		pbPayload = &pb.GetLeaderboardPayload{
			Type:   p.Type,
			Offset: int32(p.Offset),
			Limit:  int32(p.Limit),
		}
	case protocol.MsgGetOnlineCount, protocol.MsgGetMaintenanceStatus:
		// No payload needed for these messages
		return nil, nil

	// 服务端响应
	case protocol.MsgConnected:
		p := payload.(protocol.ConnectedPayload)
		pbPayload = &pb.ConnectedPayload{
			PlayerId:       p.PlayerID,
			PlayerName:     p.PlayerName,
			ReconnectToken: p.ReconnectToken,
		}
	case protocol.MsgReconnected:
		p := payload.(protocol.ReconnectedPayload)
		var gameState *pb.GameStateDTO
		if p.GameState != nil {
			gameState = gameStateDTOToProto(p.GameState)
		}
		pbPayload = &pb.ReconnectedPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			RoomCode:   p.RoomCode,
			GameState:  gameState,
		}
	case protocol.MsgPong:
		p := payload.(protocol.PongPayload)
		pbPayload = &pb.PongPayload{
			ClientTimestamp: p.ClientTimestamp,
			ServerTimestamp: p.ServerTimestamp,
		}
	case protocol.MsgPlayerOffline:
		p := payload.(protocol.PlayerOfflinePayload)
		pbPayload = &pb.PlayerOfflinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Timeout:    int32(p.Timeout),
		}
	case protocol.MsgPlayerOnline:
		p := payload.(protocol.PlayerOnlinePayload)
		pbPayload = &pb.PlayerOnlinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case protocol.MsgOnlineCount:
		p := payload.(protocol.OnlineCountPayload)
		pbPayload = &pb.OnlineCountPayload{
			Count: int32(p.Count),
		}
	case protocol.MsgMaintenanceStatus:
		p := payload.(protocol.MaintenanceStatusPayload)
		pbPayload = &pb.MaintenanceStatusPayload{
			Maintenance: p.Maintenance,
		}
	case protocol.MsgMaintenance:
		p := payload.(protocol.MaintenancePayload)
		pbPayload = &pb.MaintenancePayload{
			Maintenance: p.Maintenance,
		}
	case protocol.MsgRoomCreated:
		p := payload.(protocol.RoomCreatedPayload)
		pbPayload = &pb.RoomCreatedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
		}
	case protocol.MsgRoomJoined:
		p := payload.(protocol.RoomJoinedPayload)
		pbPayload = &pb.RoomJoinedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
			Players:  playerInfosToProto(p.Players),
		}
	case protocol.MsgPlayerJoined:
		p := payload.(protocol.PlayerJoinedPayload)
		pbPayload = &pb.PlayerJoinedPayload{
			Player: playerInfoToProto(&p.Player),
		}
	case protocol.MsgPlayerLeft:
		p := payload.(protocol.PlayerLeftPayload)
		pbPayload = &pb.PlayerLeftPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case protocol.MsgPlayerReady:
		p := payload.(protocol.PlayerReadyPayload)
		pbPayload = &pb.PlayerReadyPayload{
			PlayerId: p.PlayerID,
			Ready:    p.Ready,
		}
	case protocol.MsgGameStart:
		p := payload.(protocol.GameStartPayload)
		pbPayload = &pb.GameStartPayload{
			Players: playerInfosToProto(p.Players),
		}
	case protocol.MsgDealCards:
		p := payload.(protocol.DealCardsPayload)
		pbPayload = &pb.DealCardsPayload{
			Cards:         cardsToProto(p.Cards),
			LandlordCards: cardsToProto(p.LandlordCards),
		}
	case protocol.MsgBidTurn:
		p := payload.(protocol.BidTurnPayload)
		pbPayload = &pb.BidTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int32(p.Timeout),
		}
	case protocol.MsgBidResult:
		p := payload.(protocol.BidResultPayload)
		pbPayload = &pb.BidResultPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Bid:        p.Bid,
		}
	case protocol.MsgLandlord:
		p := payload.(protocol.LandlordPayload)
		pbPayload = &pb.LandlordPayload{
			PlayerId:      p.PlayerID,
			PlayerName:    p.PlayerName,
			LandlordCards: cardsToProto(p.LandlordCards),
		}
	case protocol.MsgPlayTurn:
		p := payload.(protocol.PlayTurnPayload)
		pbPayload = &pb.PlayTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int32(p.Timeout),
			MustPlay: p.MustPlay,
			CanBeat:  p.CanBeat,
		}
	case protocol.MsgCardPlayed:
		p := payload.(protocol.CardPlayedPayload)
		pbPayload = &pb.CardPlayedPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Cards:      cardsToProto(p.Cards),
			CardsLeft:  int32(p.CardsLeft),
			HandType:   p.HandType,
		}
	case protocol.MsgPlayerPass:
		p := payload.(protocol.PlayerPassPayload)
		pbPayload = &pb.PlayerPassPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case protocol.MsgGameOver:
		p := payload.(protocol.GameOverPayload)
		pbPayload = &pb.GameOverPayload{
			WinnerId:    p.WinnerID,
			WinnerName:  p.WinnerName,
			IsLandlord:  p.IsLandlord,
			PlayerHands: playerHandsToProto(p.PlayerHands),
		}
	case protocol.MsgStatsResult:
		p := payload.(protocol.StatsResultPayload)
		pbPayload = &pb.StatsResultPayload{
			PlayerId:      p.PlayerID,
			PlayerName:    p.PlayerName,
			TotalGames:    int32(p.TotalGames),
			Wins:          int32(p.Wins),
			Losses:        int32(p.Losses),
			WinRate:       p.WinRate,
			LandlordGames: int32(p.LandlordGames),
			LandlordWins:  int32(p.LandlordWins),
			FarmerGames:   int32(p.FarmerGames),
			FarmerWins:    int32(p.FarmerWins),
			Score:         int32(p.Score),
			Rank:          int32(p.Rank),
			CurrentStreak: int32(p.CurrentStreak),
			MaxWinStreak:  int32(p.MaxWinStreak),
		}
	case protocol.MsgLeaderboardResult:
		p := payload.(protocol.LeaderboardResultPayload)
		pbPayload = &pb.LeaderboardResultPayload{
			Type:    p.Type,
			Entries: leaderboardEntriesToProto(p.Entries),
		}
	case protocol.MsgRoomListResult:
		p := payload.(protocol.RoomListResultPayload)
		pbPayload = &pb.RoomListResultPayload{
			Rooms: roomListItemsToProto(p.Rooms),
		}
	case protocol.MsgError:
		p := payload.(protocol.ErrorPayload)
		pbPayload = &pb.ErrorPayload{
			Code:    int32(p.Code),
			Message: p.Message,
		}

	default:
		// 未知类型，回退到 JSON
		return json.Marshal(payload)
	}

	return proto.Marshal(pbPayload)
}

// protocol.DecodePayload 从 protobuf bytes 解码为 Go struct
//
//nolint:gocyclo // Payload decoding function with many message types
func DecodePayload(msgType protocol.MessageType, data []byte, target any) error {
	if len(data) == 0 {
		return nil
	}

	switch msgType {
	// 客户端请求
	case protocol.MsgReconnect:
		var pbMsg pb.ReconnectPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.ReconnectPayload) = protocol.ReconnectPayload{
			Token:    pbMsg.Token,
			PlayerID: pbMsg.PlayerId,
		}
	case protocol.MsgPing:
		var pbMsg pb.PingPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PingPayload) = protocol.PingPayload{
			Timestamp: pbMsg.Timestamp,
		}
	case protocol.MsgJoinRoom:
		var pbMsg pb.JoinRoomPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.JoinRoomPayload) = protocol.JoinRoomPayload{
			RoomCode: pbMsg.RoomCode,
		}
	case protocol.MsgBid:
		var pbMsg pb.BidPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.BidPayload) = protocol.BidPayload{
			Bid: pbMsg.Bid,
		}
	case protocol.MsgPlayCards:
		var pbMsg pb.PlayCardsPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayCardsPayload) = protocol.PlayCardsPayload{
			Cards: protoToCards(pbMsg.Cards),
		}
	case protocol.MsgGetLeaderboard:
		var pbMsg pb.GetLeaderboardPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.GetLeaderboardPayload) = protocol.GetLeaderboardPayload{
			Type:   pbMsg.Type,
			Offset: int(pbMsg.Offset),
			Limit:  int(pbMsg.Limit),
		}

	// 服务端响应
	case protocol.MsgConnected:
		var pbMsg pb.ConnectedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.ConnectedPayload) = protocol.ConnectedPayload{
			PlayerID:       pbMsg.PlayerId,
			PlayerName:     pbMsg.PlayerName,
			ReconnectToken: pbMsg.ReconnectToken,
		}
	case protocol.MsgPong:
		var pbMsg pb.PongPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PongPayload) = protocol.PongPayload{
			ClientTimestamp: pbMsg.ClientTimestamp,
			ServerTimestamp: pbMsg.ServerTimestamp,
		}
	case protocol.MsgOnlineCount:
		var pbMsg pb.OnlineCountPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.OnlineCountPayload) = protocol.OnlineCountPayload{
			Count: int(pbMsg.Count),
		}
	case protocol.MsgMaintenanceStatus:
		var pbMsg pb.MaintenanceStatusPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.MaintenanceStatusPayload) = protocol.MaintenanceStatusPayload{
			Maintenance: pbMsg.Maintenance,
		}
	case protocol.MsgMaintenance:
		var pbMsg pb.MaintenancePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.MaintenancePayload) = protocol.MaintenancePayload{
			Maintenance: pbMsg.Maintenance,
		}
	case protocol.MsgError:
		var pbMsg pb.ErrorPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.ErrorPayload) = protocol.ErrorPayload{
			Code:    int(pbMsg.Code),
			Message: pbMsg.Message,
		}
	case protocol.MsgReconnected:
		var pbMsg pb.ReconnectedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
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
	case protocol.MsgPlayerOffline:
		var pbMsg pb.PlayerOfflinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerOfflinePayload) = protocol.PlayerOfflinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Timeout:    int(pbMsg.Timeout),
		}
	case protocol.MsgPlayerOnline:
		var pbMsg pb.PlayerOnlinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerOnlinePayload) = protocol.PlayerOnlinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case protocol.MsgRoomCreated:
		var pbMsg pb.RoomCreatedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.RoomCreatedPayload) = protocol.RoomCreatedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
		}
	case protocol.MsgRoomJoined:
		var pbMsg pb.RoomJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.RoomJoinedPayload) = protocol.RoomJoinedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
			Players:  protoToPlayerInfos(pbMsg.Players),
		}
	case protocol.MsgPlayerJoined:
		var pbMsg pb.PlayerJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerJoinedPayload) = protocol.PlayerJoinedPayload{
			Player: protoToPlayerInfo(pbMsg.Player),
		}
	case protocol.MsgPlayerLeft:
		var pbMsg pb.PlayerLeftPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerLeftPayload) = protocol.PlayerLeftPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case protocol.MsgPlayerReady:
		var pbMsg pb.PlayerReadyPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerReadyPayload) = protocol.PlayerReadyPayload{
			PlayerID: pbMsg.PlayerId,
			Ready:    pbMsg.Ready,
		}
	case protocol.MsgGameStart:
		var pbMsg pb.GameStartPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.GameStartPayload) = protocol.GameStartPayload{
			Players: protoToPlayerInfos(pbMsg.Players),
		}
	case protocol.MsgDealCards:
		var pbMsg pb.DealCardsPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.DealCardsPayload) = protocol.DealCardsPayload{
			Cards:         protoToCards(pbMsg.Cards),
			LandlordCards: protoToCards(pbMsg.LandlordCards),
		}
	case protocol.MsgBidTurn:
		var pbMsg pb.BidTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.BidTurnPayload) = protocol.BidTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
		}
	case protocol.MsgBidResult:
		var pbMsg pb.BidResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.BidResultPayload) = protocol.BidResultPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Bid:        pbMsg.Bid,
		}
	case protocol.MsgLandlord:
		var pbMsg pb.LandlordPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.LandlordPayload) = protocol.LandlordPayload{
			PlayerID:      pbMsg.PlayerId,
			PlayerName:    pbMsg.PlayerName,
			LandlordCards: protoToCards(pbMsg.LandlordCards),
		}
	case protocol.MsgPlayTurn:
		var pbMsg pb.PlayTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayTurnPayload) = protocol.PlayTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
			MustPlay: pbMsg.MustPlay,
			CanBeat:  pbMsg.CanBeat,
		}
	case protocol.MsgCardPlayed:
		var pbMsg pb.CardPlayedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.CardPlayedPayload) = protocol.CardPlayedPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Cards:      protoToCards(pbMsg.Cards),
			CardsLeft:  int(pbMsg.CardsLeft),
			HandType:   pbMsg.HandType,
		}
	case protocol.MsgPlayerPass:
		var pbMsg pb.PlayerPassPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.PlayerPassPayload) = protocol.PlayerPassPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case protocol.MsgGameOver:
		var pbMsg pb.GameOverPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.GameOverPayload) = protocol.GameOverPayload{
			WinnerID:    pbMsg.WinnerId,
			WinnerName:  pbMsg.WinnerName,
			IsLandlord:  pbMsg.IsLandlord,
			PlayerHands: protoToPlayerHands(pbMsg.PlayerHands),
		}
	case protocol.MsgStatsResult:
		var pbMsg pb.StatsResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
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
	case protocol.MsgLeaderboardResult:
		var pbMsg pb.LeaderboardResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.LeaderboardResultPayload) = protocol.LeaderboardResultPayload{
			Type:    pbMsg.Type,
			Entries: protoToLeaderboardEntries(pbMsg.Entries),
		}
	case protocol.MsgRoomListResult:
		var pbMsg pb.RoomListResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*protocol.RoomListResultPayload) = protocol.RoomListResultPayload{
			Rooms: protoToRoomListItems(pbMsg.Rooms),
		}

	default:
		// 未知类型，回退到 JSON
		return json.Unmarshal(data, target)
	}

	return nil
}
