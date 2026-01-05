package convert

import (
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/pb"
)

// --- Card conversion ---

func CardToProto(c protocol.CardInfo) *pb.CardInfo {
	return &pb.CardInfo{
		Suit:  int64(c.Suit),
		Rank:  int64(c.Rank),
		Color: int64(c.Color),
	}
}

func CardsToProto(cards []protocol.CardInfo) []*pb.CardInfo {
	result := make([]*pb.CardInfo, len(cards))
	for i, c := range cards {
		result[i] = CardToProto(c)
	}
	return result
}

func ProtoToCard(pb *pb.CardInfo) protocol.CardInfo {
	return protocol.CardInfo{
		Suit:  int(pb.Suit),
		Rank:  int(pb.Rank),
		Color: int(pb.Color),
	}
}

func ProtoToCards(pbs []*pb.CardInfo) []protocol.CardInfo {
	result := make([]protocol.CardInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = ProtoToCard(pb)
	}
	return result
}

// --- PlayerInfo conversion ---

func PlayerInfoToProto(p *protocol.PlayerInfo) *pb.PlayerInfo {
	return &pb.PlayerInfo{
		Id:         p.ID,
		Name:       p.Name,
		Seat:       int64(p.Seat),
		Ready:      p.Ready,
		IsLandlord: p.IsLandlord,
		CardsCount: int64(p.CardsCount),
		Online:     p.Online,
	}
}

func PlayerInfosToProto(players []protocol.PlayerInfo) []*pb.PlayerInfo {
	result := make([]*pb.PlayerInfo, len(players))
	for i, p := range players {
		result[i] = PlayerInfoToProto(&p)
	}
	return result
}

func ProtoToPlayerInfo(pb *pb.PlayerInfo) protocol.PlayerInfo {
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

func ProtoToPlayerInfos(pbs []*pb.PlayerInfo) []protocol.PlayerInfo {
	result := make([]protocol.PlayerInfo, len(pbs))
	for i, pb := range pbs {
		result[i] = ProtoToPlayerInfo(pb)
	}
	return result
}

// --- GameState conversion ---

func GameStateDTOToProto(gs *protocol.GameStateDTO) *pb.GameStateDTO {
	return &pb.GameStateDTO{
		Phase:        gs.Phase,
		Players:      PlayerInfosToProto(gs.Players),
		Hand:         CardsToProto(gs.Hand),
		BottomCards:  CardsToProto(gs.BottomCards),
		CurrentTurn:  gs.CurrentTurn,
		LastPlayed:   CardsToProto(gs.LastPlayed),
		LastPlayerId: gs.LastPlayerID,
		MustPlay:     gs.MustPlay,
		CanBeat:      gs.CanBeat,
	}
}

func ProtoToGameStateDTO(pb *pb.GameStateDTO) *protocol.GameStateDTO {
	return &protocol.GameStateDTO{
		Phase:        pb.Phase,
		Players:      ProtoToPlayerInfos(pb.Players),
		Hand:         ProtoToCards(pb.Hand),
		BottomCards:  ProtoToCards(pb.BottomCards),
		CurrentTurn:  pb.CurrentTurn,
		LastPlayed:   ProtoToCards(pb.LastPlayed),
		LastPlayerID: pb.LastPlayerId,
		MustPlay:     pb.MustPlay,
		CanBeat:      pb.CanBeat,
	}
}

// --- PlayerHand conversion ---

func PlayerHandsToProto(hands []protocol.PlayerHand) []*pb.PlayerHand {
	result := make([]*pb.PlayerHand, len(hands))
	for i, h := range hands {
		result[i] = &pb.PlayerHand{
			PlayerId:   h.PlayerID,
			PlayerName: h.PlayerName,
			Cards:      CardsToProto(h.Cards),
		}
	}
	return result
}

func ProtoToPlayerHands(pbs []*pb.PlayerHand) []protocol.PlayerHand {
	result := make([]protocol.PlayerHand, len(pbs))
	for i, pb := range pbs {
		result[i] = protocol.PlayerHand{
			PlayerID:   pb.PlayerId,
			PlayerName: pb.PlayerName,
			Cards:      ProtoToCards(pb.Cards),
		}
	}
	return result
}

// --- Leaderboard conversion ---

func LeaderboardEntriesToProto(entries []protocol.LeaderboardEntry) []*pb.LeaderboardEntry {
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

func ProtoToLeaderboardEntries(pbs []*pb.LeaderboardEntry) []protocol.LeaderboardEntry {
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

func RoomListItemsToProto(rooms []protocol.RoomListItem) []*pb.RoomListItem {
	result := make([]*pb.RoomListItem, len(rooms))
	for i, r := range rooms {
		result[i] = &pb.RoomListItem{
			RoomCode:    r.RoomCode,
			PlayerCount: int64(r.PlayerCount),
			MaxPlayers:  int64(r.MaxPlayers),
		}
	}
	return result
}

func ProtoToRoomListItems(pbs []*pb.RoomListItem) []protocol.RoomListItem {
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
