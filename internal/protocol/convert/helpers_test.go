package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/pb"
)

// --- Card conversion tests ---

func TestCardProtoRoundTrip(t *testing.T) {
	t.Parallel()

	cards := []protocol.CardInfo{
		{Suit: 0, Rank: 3, Color: 0},
		{Suit: 1, Rank: 14, Color: 1},
		{Suit: 4, Rank: 17, Color: 1},
	}

	protos := CardsToProto(cards)
	result := ProtoToCards(protos)

	require.Len(t, result, len(cards))
	for i, card := range cards {
		assert.Equal(t, card, result[i])
	}
}

// --- PlayerInfo conversion tests ---

func TestPlayerInfosRoundTrip(t *testing.T) {
	t.Parallel()

	players := []protocol.PlayerInfo{
		{ID: "p1", Name: "Player1", Seat: 0, Ready: true, IsLandlord: false, CardsCount: 17, Online: true},
		{ID: "p2", Name: "Player2", Seat: 1, Ready: true, IsLandlord: true, CardsCount: 20, Online: true},
		{ID: "p3", Name: "Player3", Seat: 2, Ready: true, IsLandlord: false, CardsCount: 17, Online: false},
	}

	protos := PlayerInfosToProto(players)
	result := ProtoToPlayerInfos(protos)

	require.Len(t, result, len(players))
	for i, player := range players {
		assert.Equal(t, player, result[i])
	}
}

// --- GameStateDTO conversion tests ---

func TestGameStateDTORoundTrip(t *testing.T) {
	t.Parallel()

	gs := &protocol.GameStateDTO{
		Phase: "playing",
		Players: []protocol.PlayerInfo{
			{ID: "p1", Name: "Player1", Seat: 0, IsLandlord: true, CardsCount: 18, Online: true},
			{ID: "p2", Name: "Player2", Seat: 1, IsLandlord: false, CardsCount: 17, Online: true},
		},
		Hand: []protocol.CardInfo{
			{Suit: 0, Rank: 3, Color: 0},
			{Suit: 1, Rank: 14, Color: 1},
		},
		BottomCards: []protocol.CardInfo{
			{Suit: 2, Rank: 5, Color: 0},
		},
		CurrentTurn:  "p1",
		LastPlayed:   []protocol.CardInfo{{Suit: 0, Rank: 10, Color: 0}},
		LastPlayerID: "p2",
		MustPlay:     true,
		CanBeat:      true,
	}

	proto := GameStateDTOToProto(gs)
	result := ProtoToGameStateDTO(proto)

	assert.Equal(t, gs.Phase, result.Phase)
	assert.Equal(t, gs.CurrentTurn, result.CurrentTurn)
	assert.Equal(t, gs.LastPlayerID, result.LastPlayerID)
	assert.Equal(t, gs.MustPlay, result.MustPlay)
	assert.Equal(t, gs.CanBeat, result.CanBeat)
	assert.Len(t, result.Players, len(gs.Players))
	assert.Len(t, result.Hand, len(gs.Hand))
	assert.Len(t, result.BottomCards, len(gs.BottomCards))
	assert.Len(t, result.LastPlayed, len(gs.LastPlayed))
}

// --- PlayerHand conversion tests ---

func TestPlayerHandsRoundTrip(t *testing.T) {
	t.Parallel()

	hands := []protocol.PlayerHand{
		{
			PlayerID:   "p1",
			PlayerName: "Player1",
			Cards: []protocol.CardInfo{
				{Suit: 0, Rank: 3, Color: 0},
				{Suit: 1, Rank: 5, Color: 1},
			},
		},
		{
			PlayerID:   "p2",
			PlayerName: "Player2",
			Cards:      []protocol.CardInfo{},
		},
	}

	protos := PlayerHandsToProto(hands)
	result := ProtoToPlayerHands(protos)

	require.Len(t, result, len(hands))
	for i, hand := range hands {
		assert.Equal(t, hand.PlayerID, result[i].PlayerID)
		assert.Equal(t, hand.PlayerName, result[i].PlayerName)
		assert.Len(t, result[i].Cards, len(hand.Cards))
	}
}

// --- Leaderboard conversion tests ---

func TestLeaderboardEntriesRoundTrip(t *testing.T) {
	t.Parallel()

	entries := []protocol.LeaderboardEntry{
		{Rank: 1, PlayerID: "p1", PlayerName: "Champion", Score: 1000, Wins: 50, WinRate: 0.75},
		{Rank: 2, PlayerID: "p2", PlayerName: "Runner", Score: 800, Wins: 40, WinRate: 0.65},
		{Rank: 3, PlayerID: "p3", PlayerName: "Third", Score: 600, Wins: 30, WinRate: 0.55},
	}

	protos := LeaderboardEntriesToProto(entries)
	result := ProtoToLeaderboardEntries(protos)

	require.Len(t, result, len(entries))
	for i, entry := range entries {
		assert.Equal(t, entry.Rank, result[i].Rank)
		assert.Equal(t, entry.PlayerID, result[i].PlayerID)
		assert.Equal(t, entry.PlayerName, result[i].PlayerName)
		assert.Equal(t, entry.Score, result[i].Score)
		assert.Equal(t, entry.Wins, result[i].Wins)
		assert.Equal(t, entry.WinRate, result[i].WinRate)
	}
}

func TestLeaderboardEntriesEmpty(t *testing.T) {
	t.Parallel()

	protos := LeaderboardEntriesToProto([]protocol.LeaderboardEntry{})
	assert.Empty(t, protos)

	result := ProtoToLeaderboardEntries([]*pb.LeaderboardEntry{})
	assert.Empty(t, result)
}

// --- RoomList conversion tests ---

func TestRoomListItemsRoundTrip(t *testing.T) {
	t.Parallel()

	rooms := []protocol.RoomListItem{
		{RoomCode: "123456", PlayerCount: 2, MaxPlayers: 3},
		{RoomCode: "654321", PlayerCount: 1, MaxPlayers: 3},
	}

	protos := RoomListItemsToProto(rooms)
	result := ProtoToRoomListItems(protos)

	require.Len(t, result, len(rooms))
	for i, room := range rooms {
		assert.Equal(t, room.RoomCode, result[i].RoomCode)
		assert.Equal(t, room.PlayerCount, result[i].PlayerCount)
		assert.Equal(t, room.MaxPlayers, result[i].MaxPlayers)
	}
}

func TestRoomListItemsEmpty(t *testing.T) {
	t.Parallel()

	protos := RoomListItemsToProto([]protocol.RoomListItem{})
	assert.Empty(t, protos)

	result := ProtoToRoomListItems([]*pb.RoomListItem{})
	assert.Empty(t, result)
}
