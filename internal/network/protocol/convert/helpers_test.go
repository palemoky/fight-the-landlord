package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/pb"
)

// --- Card conversion tests ---

func TestCardToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		card     protocol.CardInfo
		expected *pb.CardInfo
	}{
		{
			name:     "basic card",
			card:     protocol.CardInfo{Suit: 0, Rank: 3, Color: 0},
			expected: &pb.CardInfo{Suit: 0, Rank: 3, Color: 0},
		},
		{
			name:     "red joker",
			card:     protocol.CardInfo{Suit: 4, Rank: 17, Color: 1},
			expected: &pb.CardInfo{Suit: 4, Rank: 17, Color: 1},
		},
		{
			name:     "heart ace",
			card:     protocol.CardInfo{Suit: 1, Rank: 14, Color: 1},
			expected: &pb.CardInfo{Suit: 1, Rank: 14, Color: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := cardToProto(tt.card)
			assert.Equal(t, tt.expected.Suit, result.Suit)
			assert.Equal(t, tt.expected.Rank, result.Rank)
			assert.Equal(t, tt.expected.Color, result.Color)
		})
	}
}

func TestProtoToCard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		proto    *pb.CardInfo
		expected protocol.CardInfo
	}{
		{
			name:     "basic card",
			proto:    &pb.CardInfo{Suit: 0, Rank: 3, Color: 0},
			expected: protocol.CardInfo{Suit: 0, Rank: 3, Color: 0},
		},
		{
			name:     "red joker",
			proto:    &pb.CardInfo{Suit: 4, Rank: 17, Color: 1},
			expected: protocol.CardInfo{Suit: 4, Rank: 17, Color: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := protoToCard(tt.proto)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCardsToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cards    []protocol.CardInfo
		expected int
	}{
		{"empty slice", []protocol.CardInfo{}, 0},
		{"single card", []protocol.CardInfo{{Suit: 0, Rank: 3, Color: 0}}, 1},
		{"multiple cards", []protocol.CardInfo{
			{Suit: 0, Rank: 3, Color: 0},
			{Suit: 1, Rank: 14, Color: 1},
		}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := cardsToProto(tt.cards)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestProtoToCards(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		protos   []*pb.CardInfo
		expected int
	}{
		{"empty slice", []*pb.CardInfo{}, 0},
		{"single card", []*pb.CardInfo{{Suit: 0, Rank: 3, Color: 0}}, 1},
		{"multiple cards", []*pb.CardInfo{
			{Suit: 0, Rank: 3},
			{Suit: 1, Rank: 14},
		}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := protoToCards(tt.protos)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestCardProtoRoundTrip(t *testing.T) {
	t.Parallel()

	cards := []protocol.CardInfo{
		{Suit: 0, Rank: 3, Color: 0},
		{Suit: 1, Rank: 14, Color: 1},
		{Suit: 4, Rank: 17, Color: 1},
	}

	protos := cardsToProto(cards)
	result := protoToCards(protos)

	require.Len(t, result, len(cards))
	for i, card := range cards {
		assert.Equal(t, card, result[i])
	}
}

// --- PlayerInfo conversion tests ---

func TestPlayerInfoToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		player protocol.PlayerInfo
	}{
		{
			name: "basic player",
			player: protocol.PlayerInfo{
				ID:         "p1",
				Name:       "Player1",
				Seat:       0,
				Ready:      true,
				IsLandlord: false,
				CardsCount: 17,
				Online:     true,
			},
		},
		{
			name: "landlord player",
			player: protocol.PlayerInfo{
				ID:         "p2",
				Name:       "Landlord",
				Seat:       1,
				Ready:      true,
				IsLandlord: true,
				CardsCount: 20,
				Online:     true,
			},
		},
		{
			name: "offline player",
			player: protocol.PlayerInfo{
				ID:         "p3",
				Name:       "Offline",
				Seat:       2,
				Ready:      false,
				IsLandlord: false,
				CardsCount: 10,
				Online:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := playerInfoToProto(&tt.player)
			assert.Equal(t, tt.player.ID, result.Id)
			assert.Equal(t, tt.player.Name, result.Name)
			assert.Equal(t, int64(tt.player.Seat), result.Seat)
			assert.Equal(t, tt.player.Ready, result.Ready)
			assert.Equal(t, tt.player.IsLandlord, result.IsLandlord)
			assert.Equal(t, int64(tt.player.CardsCount), result.CardsCount)
			assert.Equal(t, tt.player.Online, result.Online)
		})
	}
}

func TestProtoToPlayerInfo(t *testing.T) {
	t.Parallel()

	proto := &pb.PlayerInfo{
		Id:         "p1",
		Name:       "TestPlayer",
		Seat:       1,
		Ready:      true,
		IsLandlord: true,
		CardsCount: 15,
		Online:     true,
	}

	result := protoToPlayerInfo(proto)

	assert.Equal(t, "p1", result.ID)
	assert.Equal(t, "TestPlayer", result.Name)
	assert.Equal(t, 1, result.Seat)
	assert.True(t, result.Ready)
	assert.True(t, result.IsLandlord)
	assert.Equal(t, 15, result.CardsCount)
	assert.True(t, result.Online)
}

func TestPlayerInfosRoundTrip(t *testing.T) {
	t.Parallel()

	players := []protocol.PlayerInfo{
		{ID: "p1", Name: "Player1", Seat: 0, Ready: true, IsLandlord: false, CardsCount: 17, Online: true},
		{ID: "p2", Name: "Player2", Seat: 1, Ready: true, IsLandlord: true, CardsCount: 20, Online: true},
		{ID: "p3", Name: "Player3", Seat: 2, Ready: true, IsLandlord: false, CardsCount: 17, Online: false},
	}

	protos := playerInfosToProto(players)
	result := protoToPlayerInfos(protos)

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

	proto := gameStateDTOToProto(gs)
	result := protoToGameStateDTO(proto)

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

	protos := playerHandsToProto(hands)
	result := protoToPlayerHands(protos)

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

	protos := leaderboardEntriesToProto(entries)
	result := protoToLeaderboardEntries(protos)

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

	protos := leaderboardEntriesToProto([]protocol.LeaderboardEntry{})
	assert.Empty(t, protos)

	result := protoToLeaderboardEntries([]*pb.LeaderboardEntry{})
	assert.Empty(t, result)
}

// --- RoomList conversion tests ---

func TestRoomListItemsRoundTrip(t *testing.T) {
	t.Parallel()

	rooms := []protocol.RoomListItem{
		{RoomCode: "123456", PlayerCount: 2, MaxPlayers: 3},
		{RoomCode: "654321", PlayerCount: 1, MaxPlayers: 3},
	}

	protos := roomListItemsToProto(rooms)
	result := protoToRoomListItems(protos)

	require.Len(t, result, len(rooms))
	for i, room := range rooms {
		assert.Equal(t, room.RoomCode, result[i].RoomCode)
		assert.Equal(t, room.PlayerCount, result[i].PlayerCount)
		assert.Equal(t, room.MaxPlayers, result[i].MaxPlayers)
	}
}

func TestRoomListItemsEmpty(t *testing.T) {
	t.Parallel()

	protos := roomListItemsToProto([]protocol.RoomListItem{})
	assert.Empty(t, protos)

	result := protoToRoomListItems([]*pb.RoomListItem{})
	assert.Empty(t, result)
}
