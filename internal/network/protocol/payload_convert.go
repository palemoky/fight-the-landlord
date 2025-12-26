package protocol

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// EncodePayload 将 Go struct payload 编码为 protobuf bytes
//
//nolint:gocyclo // Payload conversion function with many message types
func EncodePayload(msgType MessageType, payload any) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}

	var pbPayload proto.Message

	switch msgType {
	// 客户端请求
	case MsgReconnect:
		p := payload.(ReconnectPayload)
		pbPayload = &pb.ReconnectPayload{
			Token:    p.Token,
			PlayerId: p.PlayerID,
		}
	case MsgPing:
		p := payload.(PingPayload)
		pbPayload = &pb.PingPayload{
			Timestamp: p.Timestamp,
		}
	case MsgJoinRoom:
		p := payload.(JoinRoomPayload)
		pbPayload = &pb.JoinRoomPayload{
			RoomCode: p.RoomCode,
		}
	case MsgBid:
		p := payload.(BidPayload)
		pbPayload = &pb.BidPayload{
			Bid: p.Bid,
		}
	case MsgPlayCards:
		p := payload.(PlayCardsPayload)
		pbPayload = &pb.PlayCardsPayload{
			Cards: cardsToProto(p.Cards),
		}
	case MsgGetLeaderboard:
		p := payload.(GetLeaderboardPayload)
		pbPayload = &pb.GetLeaderboardPayload{
			Type:   p.Type,
			Offset: int32(p.Offset),
			Limit:  int32(p.Limit),
		}
	case MsgGetOnlineCount:
		// No payload needed for this message
		return nil, nil

	// 服务端响应
	case MsgConnected:
		p := payload.(ConnectedPayload)
		pbPayload = &pb.ConnectedPayload{
			PlayerId:       p.PlayerID,
			PlayerName:     p.PlayerName,
			ReconnectToken: p.ReconnectToken,
		}
	case MsgReconnected:
		p := payload.(ReconnectedPayload)
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
	case MsgPong:
		p := payload.(PongPayload)
		pbPayload = &pb.PongPayload{
			ClientTimestamp: p.ClientTimestamp,
			ServerTimestamp: p.ServerTimestamp,
		}
	case MsgPlayerOffline:
		p := payload.(PlayerOfflinePayload)
		pbPayload = &pb.PlayerOfflinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Timeout:    int32(p.Timeout),
		}
	case MsgPlayerOnline:
		p := payload.(PlayerOnlinePayload)
		pbPayload = &pb.PlayerOnlinePayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case MsgOnlineCount:
		p := payload.(OnlineCountPayload)
		pbPayload = &pb.OnlineCountPayload{
			Count: int32(p.Count),
		}
	case MsgRoomCreated:
		p := payload.(RoomCreatedPayload)
		pbPayload = &pb.RoomCreatedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
		}
	case MsgRoomJoined:
		p := payload.(RoomJoinedPayload)
		pbPayload = &pb.RoomJoinedPayload{
			RoomCode: p.RoomCode,
			Player:   playerInfoToProto(&p.Player),
			Players:  playerInfosToProto(p.Players),
		}
	case MsgPlayerJoined:
		p := payload.(PlayerJoinedPayload)
		pbPayload = &pb.PlayerJoinedPayload{
			Player: playerInfoToProto(&p.Player),
		}
	case MsgPlayerLeft:
		p := payload.(PlayerLeftPayload)
		pbPayload = &pb.PlayerLeftPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case MsgPlayerReady:
		p := payload.(PlayerReadyPayload)
		pbPayload = &pb.PlayerReadyPayload{
			PlayerId: p.PlayerID,
			Ready:    p.Ready,
		}
	case MsgGameStart:
		p := payload.(GameStartPayload)
		pbPayload = &pb.GameStartPayload{
			Players: playerInfosToProto(p.Players),
		}
	case MsgDealCards:
		p := payload.(DealCardsPayload)
		pbPayload = &pb.DealCardsPayload{
			Cards:         cardsToProto(p.Cards),
			LandlordCards: cardsToProto(p.LandlordCards),
		}
	case MsgBidTurn:
		p := payload.(BidTurnPayload)
		pbPayload = &pb.BidTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int32(p.Timeout),
		}
	case MsgBidResult:
		p := payload.(BidResultPayload)
		pbPayload = &pb.BidResultPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Bid:        p.Bid,
		}
	case MsgLandlord:
		p := payload.(LandlordPayload)
		pbPayload = &pb.LandlordPayload{
			PlayerId:      p.PlayerID,
			PlayerName:    p.PlayerName,
			LandlordCards: cardsToProto(p.LandlordCards),
		}
	case MsgPlayTurn:
		p := payload.(PlayTurnPayload)
		pbPayload = &pb.PlayTurnPayload{
			PlayerId: p.PlayerID,
			Timeout:  int32(p.Timeout),
			MustPlay: p.MustPlay,
			CanBeat:  p.CanBeat,
		}
	case MsgCardPlayed:
		p := payload.(CardPlayedPayload)
		pbPayload = &pb.CardPlayedPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
			Cards:      cardsToProto(p.Cards),
			CardsLeft:  int32(p.CardsLeft),
			HandType:   p.HandType,
		}
	case MsgPlayerPass:
		p := payload.(PlayerPassPayload)
		pbPayload = &pb.PlayerPassPayload{
			PlayerId:   p.PlayerID,
			PlayerName: p.PlayerName,
		}
	case MsgGameOver:
		p := payload.(GameOverPayload)
		pbPayload = &pb.GameOverPayload{
			WinnerId:    p.WinnerID,
			WinnerName:  p.WinnerName,
			IsLandlord:  p.IsLandlord,
			PlayerHands: playerHandsToProto(p.PlayerHands),
		}
	case MsgStatsResult:
		p := payload.(StatsResultPayload)
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
	case MsgLeaderboardResult:
		p := payload.(LeaderboardResultPayload)
		pbPayload = &pb.LeaderboardResultPayload{
			Type:    p.Type,
			Entries: leaderboardEntriesToProto(p.Entries),
		}
	case MsgRoomListResult:
		p := payload.(RoomListResultPayload)
		pbPayload = &pb.RoomListResultPayload{
			Rooms: roomListItemsToProto(p.Rooms),
		}
	case MsgError:
		p := payload.(ErrorPayload)
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

// DecodePayload 从 protobuf bytes 解码为 Go struct
//
//nolint:gocyclo // Payload decoding function with many message types
func DecodePayload(msgType MessageType, data []byte, target any) error {
	if len(data) == 0 {
		return nil
	}

	switch msgType {
	// 客户端请求
	case MsgReconnect:
		var pb pb.ReconnectPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*ReconnectPayload) = ReconnectPayload{
			Token:    pb.Token,
			PlayerID: pb.PlayerId,
		}
	case MsgPing:
		var pb pb.PingPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*PingPayload) = PingPayload{
			Timestamp: pb.Timestamp,
		}
	case MsgJoinRoom:
		var pb pb.JoinRoomPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*JoinRoomPayload) = JoinRoomPayload{
			RoomCode: pb.RoomCode,
		}
	case MsgBid:
		var pb pb.BidPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*BidPayload) = BidPayload{
			Bid: pb.Bid,
		}
	case MsgPlayCards:
		var pb pb.PlayCardsPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*PlayCardsPayload) = PlayCardsPayload{
			Cards: protoToCards(pb.Cards),
		}
	case MsgGetLeaderboard:
		var pb pb.GetLeaderboardPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*GetLeaderboardPayload) = GetLeaderboardPayload{
			Type:   pb.Type,
			Offset: int(pb.Offset),
			Limit:  int(pb.Limit),
		}

	// 服务端响应
	case MsgConnected:
		var pb pb.ConnectedPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*ConnectedPayload) = ConnectedPayload{
			PlayerID:       pb.PlayerId,
			PlayerName:     pb.PlayerName,
			ReconnectToken: pb.ReconnectToken,
		}
	case MsgPong:
		var pb pb.PongPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*PongPayload) = PongPayload{
			ClientTimestamp: pb.ClientTimestamp,
			ServerTimestamp: pb.ServerTimestamp,
		}
	case MsgOnlineCount:
		var pb pb.OnlineCountPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*OnlineCountPayload) = OnlineCountPayload{
			Count: int(pb.Count),
		}
	case MsgError:
		var pb pb.ErrorPayload
		if err := proto.Unmarshal(data, &pb); err != nil {
			return err
		}
		*target.(*ErrorPayload) = ErrorPayload{
			Code:    int(pb.Code),
			Message: pb.Message,
		}
	case MsgReconnected:
		var pbMsg pb.ReconnectedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		var gameState *GameStateDTO
		if pbMsg.GameState != nil {
			gameState = protoToGameStateDTO(pbMsg.GameState)
		}
		*target.(*ReconnectedPayload) = ReconnectedPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			RoomCode:   pbMsg.RoomCode,
			GameState:  gameState,
		}
	case MsgPlayerOffline:
		var pbMsg pb.PlayerOfflinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerOfflinePayload) = PlayerOfflinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Timeout:    int(pbMsg.Timeout),
		}
	case MsgPlayerOnline:
		var pbMsg pb.PlayerOnlinePayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerOnlinePayload) = PlayerOnlinePayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case MsgRoomCreated:
		var pbMsg pb.RoomCreatedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*RoomCreatedPayload) = RoomCreatedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
		}
	case MsgRoomJoined:
		var pbMsg pb.RoomJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*RoomJoinedPayload) = RoomJoinedPayload{
			RoomCode: pbMsg.RoomCode,
			Player:   protoToPlayerInfo(pbMsg.Player),
			Players:  protoToPlayerInfos(pbMsg.Players),
		}
	case MsgPlayerJoined:
		var pbMsg pb.PlayerJoinedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerJoinedPayload) = PlayerJoinedPayload{
			Player: protoToPlayerInfo(pbMsg.Player),
		}
	case MsgPlayerLeft:
		var pbMsg pb.PlayerLeftPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerLeftPayload) = PlayerLeftPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case MsgPlayerReady:
		var pbMsg pb.PlayerReadyPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerReadyPayload) = PlayerReadyPayload{
			PlayerID: pbMsg.PlayerId,
			Ready:    pbMsg.Ready,
		}
	case MsgGameStart:
		var pbMsg pb.GameStartPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*GameStartPayload) = GameStartPayload{
			Players: protoToPlayerInfos(pbMsg.Players),
		}
	case MsgDealCards:
		var pbMsg pb.DealCardsPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*DealCardsPayload) = DealCardsPayload{
			Cards:         protoToCards(pbMsg.Cards),
			LandlordCards: protoToCards(pbMsg.LandlordCards),
		}
	case MsgBidTurn:
		var pbMsg pb.BidTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*BidTurnPayload) = BidTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
		}
	case MsgBidResult:
		var pbMsg pb.BidResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*BidResultPayload) = BidResultPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Bid:        pbMsg.Bid,
		}
	case MsgLandlord:
		var pbMsg pb.LandlordPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*LandlordPayload) = LandlordPayload{
			PlayerID:      pbMsg.PlayerId,
			PlayerName:    pbMsg.PlayerName,
			LandlordCards: protoToCards(pbMsg.LandlordCards),
		}
	case MsgPlayTurn:
		var pbMsg pb.PlayTurnPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayTurnPayload) = PlayTurnPayload{
			PlayerID: pbMsg.PlayerId,
			Timeout:  int(pbMsg.Timeout),
			MustPlay: pbMsg.MustPlay,
			CanBeat:  pbMsg.CanBeat,
		}
	case MsgCardPlayed:
		var pbMsg pb.CardPlayedPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*CardPlayedPayload) = CardPlayedPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
			Cards:      protoToCards(pbMsg.Cards),
			CardsLeft:  int(pbMsg.CardsLeft),
			HandType:   pbMsg.HandType,
		}
	case MsgPlayerPass:
		var pbMsg pb.PlayerPassPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*PlayerPassPayload) = PlayerPassPayload{
			PlayerID:   pbMsg.PlayerId,
			PlayerName: pbMsg.PlayerName,
		}
	case MsgGameOver:
		var pbMsg pb.GameOverPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*GameOverPayload) = GameOverPayload{
			WinnerID:    pbMsg.WinnerId,
			WinnerName:  pbMsg.WinnerName,
			IsLandlord:  pbMsg.IsLandlord,
			PlayerHands: protoToPlayerHands(pbMsg.PlayerHands),
		}
	case MsgStatsResult:
		var pbMsg pb.StatsResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*StatsResultPayload) = StatsResultPayload{
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
	case MsgLeaderboardResult:
		var pbMsg pb.LeaderboardResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*LeaderboardResultPayload) = LeaderboardResultPayload{
			Type:    pbMsg.Type,
			Entries: protoToLeaderboardEntries(pbMsg.Entries),
		}
	case MsgRoomListResult:
		var pbMsg pb.RoomListResultPayload
		if err := proto.Unmarshal(data, &pbMsg); err != nil {
			return err
		}
		*target.(*RoomListResultPayload) = RoomListResultPayload{
			Rooms: protoToRoomListItems(pbMsg.Rooms),
		}

	default:
		// 未知类型，回退到 JSON
		return json.Unmarshal(data, target)
	}

	return nil
}

// 辅助转换函数
func cardToProto(c CardInfo) *pb.CardInfo {
	return &pb.CardInfo{
		Suit:  int32(c.Suit),
		Rank:  int32(c.Rank),
		Color: int32(c.Color),
	}
}

func cardsToProto(cards []CardInfo) []*pb.CardInfo {
	result := make([]*pb.CardInfo, len(cards))
	for i, c := range cards {
		result[i] = cardToProto(c)
	}
	return result
}

func protoToCard(pb *pb.CardInfo) CardInfo {
	return CardInfo{
		Suit:  int(pb.Suit),
		Rank:  int(pb.Rank),
		Color: int(pb.Color),
	}
}

func protoToCards(pbs []*pb.CardInfo) []CardInfo {
	result := make([]CardInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = protoToCard(pb)
	}
	return result
}

func playerInfoToProto(p *PlayerInfo) *pb.PlayerInfo {
	return &pb.PlayerInfo{
		Id:         p.ID,
		Name:       p.Name,
		Seat:       int32(p.Seat),
		Ready:      p.Ready,
		IsLandlord: p.IsLandlord,
		CardsCount: int32(p.CardsCount),
		Online:     p.Online,
	}
}

func playerInfosToProto(players []PlayerInfo) []*pb.PlayerInfo {
	result := make([]*pb.PlayerInfo, len(players))
	for i, p := range players {
		result[i] = playerInfoToProto(&p)
	}
	return result
}

func gameStateDTOToProto(gs *GameStateDTO) *pb.GameStateDTO {
	return &pb.GameStateDTO{
		Phase:         gs.Phase,
		Players:       playerInfosToProto(gs.Players),
		Hand:          cardsToProto(gs.Hand),
		LandlordCards: cardsToProto(gs.LandlordCards),
		CurrentTurn:   gs.CurrentTurn,
		LastPlayed:    cardsToProto(gs.LastPlayed),
		LastPlayerId:  gs.LastPlayerID,
		MustPlay:      gs.MustPlay,
		CanBeat:       gs.CanBeat,
	}
}

func playerHandsToProto(hands []PlayerHand) []*pb.PlayerHand {
	result := make([]*pb.PlayerHand, len(hands))
	for i, h := range hands {
		result[i] = &pb.PlayerHand{
			PlayerId:   h.PlayerID,
			PlayerName: h.PlayerName,
			Cards:      cardsToProto(h.Cards),
		}
	}
	return result
}

func leaderboardEntriesToProto(entries []LeaderboardEntry) []*pb.LeaderboardEntry {
	result := make([]*pb.LeaderboardEntry, len(entries))
	for i, e := range entries {
		result[i] = &pb.LeaderboardEntry{
			Rank:       int32(e.Rank),
			PlayerId:   e.PlayerID,
			PlayerName: e.PlayerName,
			Score:      int32(e.Score),
			Wins:       int32(e.Wins),
			WinRate:    e.WinRate,
		}
	}
	return result
}

func roomListItemsToProto(rooms []RoomListItem) []*pb.RoomListItem {
	result := make([]*pb.RoomListItem, len(rooms))
	for i, r := range rooms {
		result[i] = &pb.RoomListItem{
			RoomCode:    r.RoomCode,
			PlayerCount: int32(r.PlayerCount),
			MaxPlayers:  int32(r.MaxPlayers),
		}
	}
	return result
}

// protoToPlayerInfo converts protobuf PlayerInfo to Go struct
func protoToPlayerInfo(pb *pb.PlayerInfo) PlayerInfo {
	return PlayerInfo{
		ID:         pb.Id,
		Name:       pb.Name,
		Seat:       int(pb.Seat),
		Ready:      pb.Ready,
		IsLandlord: pb.IsLandlord,
		CardsCount: int(pb.CardsCount),
		Online:     pb.Online,
	}
}

// protoToPlayerInfos converts protobuf PlayerInfo slice to Go struct slice
func protoToPlayerInfos(pbs []*pb.PlayerInfo) []PlayerInfo {
	result := make([]PlayerInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = protoToPlayerInfo(pb)
	}
	return result
}

// protoToGameStateDTO converts protobuf GameStateDTO to Go struct
func protoToGameStateDTO(pb *pb.GameStateDTO) *GameStateDTO {
	return &GameStateDTO{
		Phase:         pb.Phase,
		Players:       protoToPlayerInfos(pb.Players),
		Hand:          protoToCards(pb.Hand),
		LandlordCards: protoToCards(pb.LandlordCards),
		CurrentTurn:   pb.CurrentTurn,
		LastPlayed:    protoToCards(pb.LastPlayed),
		LastPlayerID:  pb.LastPlayerId,
		MustPlay:      pb.MustPlay,
		CanBeat:       pb.CanBeat,
	}
}

// protoToPlayerHands converts protobuf PlayerHand slice to Go struct slice
func protoToPlayerHands(pbs []*pb.PlayerHand) []PlayerHand {
	result := make([]PlayerHand, len(pbs))
	for i, pb := range pbs {
		result[i] = PlayerHand{
			PlayerID:   pb.PlayerId,
			PlayerName: pb.PlayerName,
			Cards:      protoToCards(pb.Cards),
		}
	}
	return result
}

// protoToLeaderboardEntries converts protobuf LeaderboardEntry slice to Go struct slice
func protoToLeaderboardEntries(pbs []*pb.LeaderboardEntry) []LeaderboardEntry {
	result := make([]LeaderboardEntry, len(pbs))
	for i, pb := range pbs {
		result[i] = LeaderboardEntry{
			Rank:       int(pb.Rank),
			PlayerID:   pb.PlayerId,
			PlayerName: pb.PlayerName,
			Score:      int(pb.Score),
			Wins:       int(pb.Wins),
			WinRate:    pb.WinRate,
		}
	}
	return result
}

// protoToRoomListItems converts protobuf RoomListItem slice to Go struct slice
func protoToRoomListItems(pbs []*pb.RoomListItem) []RoomListItem {
	result := make([]RoomListItem, len(pbs))
	for i, pb := range pbs {
		result[i] = RoomListItem{
			RoomCode:    pb.RoomCode,
			PlayerCount: int(pb.PlayerCount),
			MaxPlayers:  int(pb.MaxPlayers),
		}
	}
	return result
}
