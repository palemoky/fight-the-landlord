package convert

import (
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// --- Card conversion ---

func cardToProto(c protocol.CardInfo) *pb.CardInfo {
	return &pb.CardInfo{
		Suit:  int32(c.Suit),
		Rank:  int32(c.Rank),
		Color: int32(c.Color),
	}
}

func cardsToProto(cards []protocol.CardInfo) []*pb.CardInfo {
	result := make([]*pb.CardInfo, len(cards))
	for i, c := range cards {
		result[i] = cardToProto(c)
	}
	return result
}

func protoToCard(pb *pb.CardInfo) protocol.CardInfo {
	return protocol.CardInfo{
		Suit:  int(pb.Suit),
		Rank:  int(pb.Rank),
		Color: int(pb.Color),
	}
}

func protoToCards(pbs []*pb.CardInfo) []protocol.CardInfo {
	result := make([]protocol.CardInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = protoToCard(pb)
	}
	return result
}

// --- PlayerInfo conversion ---

func playerInfoToProto(p *protocol.PlayerInfo) *pb.PlayerInfo {
	return &pb.PlayerInfo{
		Id:         p.ID,
		Name:       p.Name,
		Seat:       int32(p.Seat),
		Ready:      p.Ready,
		IsLandlord: p.IsLandlord,
		CardsCount: int64(p.CardsCount),
		Online:     p.Online,
	}
}

func playerInfosToProto(players []protocol.PlayerInfo) []*pb.PlayerInfo {
	result := make([]*pb.PlayerInfo, len(players))
	for i, p := range players {
		result[i] = playerInfoToProto(&p)
	}
	return result
}

func protoToPlayerInfo(pb *pb.PlayerInfo) protocol.PlayerInfo {
	return protocol.PlayerInfo{
		ID:         pb.Id,
		Name:       pb.Name,
		Seat:       int(pb.Seat),
		Ready:      pb.Ready,
		IsLandlord: pb.IsLandlord,
		CardsCount: int(pb.CardsCount),
		Online:     pb.Online,
	}
}

func protoToPlayerInfos(pbs []*pb.PlayerInfo) []protocol.PlayerInfo {
	result := make([]protocol.PlayerInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = protoToPlayerInfo(pb)
	}
	return result
}

// --- GameState conversion ---

func gameStateDTOToProto(gs *protocol.GameStateDTO) *pb.GameStateDTO {
	return &pb.GameStateDTO{
		Phase:        gs.Phase,
		Players:      playerInfosToProto(gs.Players),
		Hand:         cardsToProto(gs.Hand),
		BottomCards:  cardsToProto(gs.BottomCards),
		CurrentTurn:  gs.CurrentTurn,
		LastPlayed:   cardsToProto(gs.LastPlayed),
		LastPlayerId: gs.LastPlayerID,
		MustPlay:     gs.MustPlay,
		CanBeat:      gs.CanBeat,
	}
}

func protoToGameStateDTO(pb *pb.GameStateDTO) *protocol.GameStateDTO {
	return &protocol.GameStateDTO{
		Phase:        pb.Phase,
		Players:      protoToPlayerInfos(pb.Players),
		Hand:         protoToCards(pb.Hand),
		BottomCards:  protoToCards(pb.BottomCards),
		CurrentTurn:  pb.CurrentTurn,
		LastPlayed:   protoToCards(pb.LastPlayed),
		LastPlayerID: pb.LastPlayerId,
		MustPlay:     pb.MustPlay,
		CanBeat:      pb.CanBeat,
	}
}

// --- PlayerHand conversion ---

func playerHandsToProto(hands []protocol.PlayerHand) []*pb.PlayerHand {
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

func protoToPlayerHands(pbs []*pb.PlayerHand) []protocol.PlayerHand {
	result := make([]protocol.PlayerHand, len(pbs))
	for i, pb := range pbs {
		result[i] = protocol.PlayerHand{
			PlayerID:   pb.PlayerId,
			PlayerName: pb.PlayerName,
			Cards:      protoToCards(pb.Cards),
		}
	}
	return result
}

// --- Leaderboard conversion ---

func leaderboardEntriesToProto(entries []protocol.LeaderboardEntry) []*pb.LeaderboardEntry {
	result := make([]*pb.LeaderboardEntry, len(entries))
	for i, e := range entries {
		result[i] = &pb.LeaderboardEntry{
			Rank:       int64(e.Rank),
			PlayerId:   e.PlayerID,
			PlayerName: e.PlayerName,
			Score:      int64(e.Score),
			Wins:       int64(e.Wins),
			WinRate:    e.WinRate,
		}
	}
	return result
}

func protoToLeaderboardEntries(pbs []*pb.LeaderboardEntry) []protocol.LeaderboardEntry {
	result := make([]protocol.LeaderboardEntry, len(pbs))
	for i, pb := range pbs {
		result[i] = protocol.LeaderboardEntry{
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

// --- RoomList conversion ---

func roomListItemsToProto(rooms []protocol.RoomListItem) []*pb.RoomListItem {
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

func protoToRoomListItems(pbs []*pb.RoomListItem) []protocol.RoomListItem {
	result := make([]protocol.RoomListItem, len(pbs))
	for i, pb := range pbs {
		result[i] = protocol.RoomListItem{
			RoomCode:    pb.RoomCode,
			PlayerCount: int(pb.PlayerCount),
			MaxPlayers:  int(pb.MaxPlayers),
		}
	}
	return result
}
