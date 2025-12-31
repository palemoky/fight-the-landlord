package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

func TestEncodePayload_NilPayload(t *testing.T) {
	t.Parallel()

	data, err := EncodePayload(protocol.MsgPing, nil)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestDecodePayload_EmptyData(t *testing.T) {
	t.Parallel()

	var target protocol.PingPayload
	err := DecodePayload(protocol.MsgPing, []byte{}, &target)
	require.NoError(t, err)
}

func TestPayloadRoundTrip_ClientRequests(t *testing.T) {
	t.Parallel()

	t.Run("Ping", func(t *testing.T) {
		t.Parallel()
		original := protocol.PingPayload{Timestamp: 1234567890}

		data, err := EncodePayload(protocol.MsgPing, original)
		require.NoError(t, err)
		require.NotNil(t, data)

		var result protocol.PingPayload
		err = DecodePayload(protocol.MsgPing, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Timestamp, result.Timestamp)
	})

	t.Run("Reconnect", func(t *testing.T) {
		t.Parallel()
		original := protocol.ReconnectPayload{Token: "token123", PlayerID: "p1"}

		data, err := EncodePayload(protocol.MsgReconnect, original)
		require.NoError(t, err)

		var result protocol.ReconnectPayload
		err = DecodePayload(protocol.MsgReconnect, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Token, result.Token)
		assert.Equal(t, original.PlayerID, result.PlayerID)
	})

	t.Run("JoinRoom", func(t *testing.T) {
		t.Parallel()
		original := protocol.JoinRoomPayload{RoomCode: "123456"}

		data, err := EncodePayload(protocol.MsgJoinRoom, original)
		require.NoError(t, err)

		var result protocol.JoinRoomPayload
		err = DecodePayload(protocol.MsgJoinRoom, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.RoomCode, result.RoomCode)
	})

	t.Run("Bid", func(t *testing.T) {
		t.Parallel()
		original := protocol.BidPayload{Bid: true}

		data, err := EncodePayload(protocol.MsgBid, original)
		require.NoError(t, err)

		var result protocol.BidPayload
		err = DecodePayload(protocol.MsgBid, data, &result)
		require.NoError(t, err)

		assert.True(t, result.Bid)
	})

	t.Run("PlayCards", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayCardsPayload{Cards: []protocol.CardInfo{{Suit: 0, Rank: 3, Color: 0}}}

		data, err := EncodePayload(protocol.MsgPlayCards, original)
		require.NoError(t, err)

		var result protocol.PlayCardsPayload
		err = DecodePayload(protocol.MsgPlayCards, data, &result)
		require.NoError(t, err)

		require.Len(t, result.Cards, 1)
		assert.Equal(t, 3, result.Cards[0].Rank)
	})

	t.Run("GetLeaderboard", func(t *testing.T) {
		t.Parallel()
		original := protocol.GetLeaderboardPayload{Type: "total", Offset: 0, Limit: 10}

		data, err := EncodePayload(protocol.MsgGetLeaderboard, original)
		require.NoError(t, err)

		var result protocol.GetLeaderboardPayload
		err = DecodePayload(protocol.MsgGetLeaderboard, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Type, result.Type)
		assert.Equal(t, original.Limit, result.Limit)
	})
}

func TestPayloadRoundTrip_ServerResponses(t *testing.T) {
	t.Parallel()

	t.Run("Connected", func(t *testing.T) {
		t.Parallel()
		original := protocol.ConnectedPayload{
			PlayerID:       "p1",
			PlayerName:     "Player1",
			ReconnectToken: "token123",
		}

		data, err := EncodePayload(protocol.MsgConnected, original)
		require.NoError(t, err)

		var result protocol.ConnectedPayload
		err = DecodePayload(protocol.MsgConnected, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.Equal(t, original.PlayerName, result.PlayerName)
		assert.Equal(t, original.ReconnectToken, result.ReconnectToken)
	})

	t.Run("Pong", func(t *testing.T) {
		t.Parallel()
		original := protocol.PongPayload{
			ClientTimestamp: 1234567890,
			ServerTimestamp: 1234567891,
		}

		data, err := EncodePayload(protocol.MsgPong, original)
		require.NoError(t, err)

		var result protocol.PongPayload
		err = DecodePayload(protocol.MsgPong, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.ClientTimestamp, result.ClientTimestamp)
		assert.Equal(t, original.ServerTimestamp, result.ServerTimestamp)
	})

	t.Run("OnlineCount", func(t *testing.T) {
		t.Parallel()
		original := protocol.OnlineCountPayload{Count: 42}

		data, err := EncodePayload(protocol.MsgOnlineCount, original)
		require.NoError(t, err)

		var result protocol.OnlineCountPayload
		err = DecodePayload(protocol.MsgOnlineCount, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Count, result.Count)
	})

	t.Run("RoomCreated", func(t *testing.T) {
		t.Parallel()
		original := protocol.RoomCreatedPayload{
			RoomCode: "123456",
			Player: protocol.PlayerInfo{
				ID:   "p1",
				Name: "Player1",
				Seat: 0,
			},
		}

		data, err := EncodePayload(protocol.MsgRoomCreated, original)
		require.NoError(t, err)

		var result protocol.RoomCreatedPayload
		err = DecodePayload(protocol.MsgRoomCreated, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.RoomCode, result.RoomCode)
		assert.Equal(t, original.Player.ID, result.Player.ID)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		original := protocol.ErrorPayload{
			Code:    404,
			Message: "Not found",
		}

		data, err := EncodePayload(protocol.MsgError, original)
		require.NoError(t, err)

		var result protocol.ErrorPayload
		err = DecodePayload(protocol.MsgError, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Code, result.Code)
		assert.Equal(t, original.Message, result.Message)
	})
}

func TestPayloadRoundTrip_GameMessages(t *testing.T) {
	t.Parallel()

	t.Run("GameStart", func(t *testing.T) {
		t.Parallel()
		original := protocol.GameStartPayload{
			Players: []protocol.PlayerInfo{
				{ID: "p1", Name: "Player1", Seat: 0},
				{ID: "p2", Name: "Player2", Seat: 1},
				{ID: "p3", Name: "Player3", Seat: 2},
			},
		}

		data, err := EncodePayload(protocol.MsgGameStart, original)
		require.NoError(t, err)

		var result protocol.GameStartPayload
		err = DecodePayload(protocol.MsgGameStart, data, &result)
		require.NoError(t, err)

		require.Len(t, result.Players, 3)
		assert.Equal(t, "p1", result.Players[0].ID)
	})

	t.Run("DealCards", func(t *testing.T) {
		t.Parallel()
		original := protocol.DealCardsPayload{
			Cards: []protocol.CardInfo{
				{Suit: 0, Rank: 3, Color: 0},
				{Suit: 1, Rank: 14, Color: 1},
			},
			BottomCards: []protocol.CardInfo{
				{Suit: 2, Rank: 5, Color: 0},
			},
		}

		data, err := EncodePayload(protocol.MsgDealCards, original)
		require.NoError(t, err)

		var result protocol.DealCardsPayload
		err = DecodePayload(protocol.MsgDealCards, data, &result)
		require.NoError(t, err)

		assert.Len(t, result.Cards, 2)
		assert.Len(t, result.BottomCards, 1)
	})

	t.Run("BidTurn", func(t *testing.T) {
		t.Parallel()
		original := protocol.BidTurnPayload{
			PlayerID: "p1",
			Timeout:  30,
		}

		data, err := EncodePayload(protocol.MsgBidTurn, original)
		require.NoError(t, err)

		var result protocol.BidTurnPayload
		err = DecodePayload(protocol.MsgBidTurn, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.Equal(t, original.Timeout, result.Timeout)
	})

	t.Run("BidResult", func(t *testing.T) {
		t.Parallel()
		original := protocol.BidResultPayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
			Bid:        true,
		}

		data, err := EncodePayload(protocol.MsgBidResult, original)
		require.NoError(t, err)

		var result protocol.BidResultPayload
		err = DecodePayload(protocol.MsgBidResult, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.True(t, result.Bid)
	})

	t.Run("Landlord", func(t *testing.T) {
		t.Parallel()
		original := protocol.LandlordPayload{
			PlayerID:    "p1",
			PlayerName:  "Player1",
			BottomCards: []protocol.CardInfo{{Suit: 0, Rank: 3, Color: 0}},
		}

		data, err := EncodePayload(protocol.MsgLandlord, original)
		require.NoError(t, err)

		var result protocol.LandlordPayload
		err = DecodePayload(protocol.MsgLandlord, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.Len(t, result.BottomCards, 1)
	})

	t.Run("PlayTurn", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayTurnPayload{
			PlayerID: "p1",
			Timeout:  30,
			MustPlay: true,
			CanBeat:  false,
		}

		data, err := EncodePayload(protocol.MsgPlayTurn, original)
		require.NoError(t, err)

		var result protocol.PlayTurnPayload
		err = DecodePayload(protocol.MsgPlayTurn, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.True(t, result.MustPlay)
		assert.False(t, result.CanBeat)
	})

	t.Run("CardPlayed", func(t *testing.T) {
		t.Parallel()
		original := protocol.CardPlayedPayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
			Cards:      []protocol.CardInfo{{Suit: 0, Rank: 3, Color: 0}},
			CardsLeft:  16,
			HandType:   "Single",
		}

		data, err := EncodePayload(protocol.MsgCardPlayed, original)
		require.NoError(t, err)

		var result protocol.CardPlayedPayload
		err = DecodePayload(protocol.MsgCardPlayed, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.Equal(t, 16, result.CardsLeft)
		assert.Equal(t, "Single", result.HandType)
	})

	t.Run("PlayerPass", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerPassPayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
		}

		data, err := EncodePayload(protocol.MsgPlayerPass, original)
		require.NoError(t, err)

		var result protocol.PlayerPassPayload
		err = DecodePayload(protocol.MsgPlayerPass, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
	})

	t.Run("GameOver", func(t *testing.T) {
		t.Parallel()
		original := protocol.GameOverPayload{
			WinnerID:   "p1",
			WinnerName: "Player1",
			IsLandlord: true,
			PlayerHands: []protocol.PlayerHand{
				{PlayerID: "p1", PlayerName: "Player1", Cards: []protocol.CardInfo{}},
			},
		}

		data, err := EncodePayload(protocol.MsgGameOver, original)
		require.NoError(t, err)

		var result protocol.GameOverPayload
		err = DecodePayload(protocol.MsgGameOver, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.WinnerID, result.WinnerID)
		assert.True(t, result.IsLandlord)
		assert.Len(t, result.PlayerHands, 1)
	})
}

func TestPayloadRoundTrip_RoomMessages(t *testing.T) {
	t.Parallel()

	t.Run("RoomJoined", func(t *testing.T) {
		t.Parallel()
		original := protocol.RoomJoinedPayload{
			RoomCode: "123456",
			Player:   protocol.PlayerInfo{ID: "p1", Name: "Player1"},
			Players: []protocol.PlayerInfo{
				{ID: "p1", Name: "Player1"},
				{ID: "p2", Name: "Player2"},
			},
		}

		data, err := EncodePayload(protocol.MsgRoomJoined, original)
		require.NoError(t, err)

		var result protocol.RoomJoinedPayload
		err = DecodePayload(protocol.MsgRoomJoined, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.RoomCode, result.RoomCode)
		assert.Len(t, result.Players, 2)
	})

	t.Run("PlayerJoined", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerJoinedPayload{
			Player: protocol.PlayerInfo{ID: "p2", Name: "Player2", Seat: 1},
		}

		data, err := EncodePayload(protocol.MsgPlayerJoined, original)
		require.NoError(t, err)

		var result protocol.PlayerJoinedPayload
		err = DecodePayload(protocol.MsgPlayerJoined, data, &result)
		require.NoError(t, err)

		assert.Equal(t, "p2", result.Player.ID)
	})

	t.Run("PlayerLeft", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerLeftPayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
		}

		data, err := EncodePayload(protocol.MsgPlayerLeft, original)
		require.NoError(t, err)

		var result protocol.PlayerLeftPayload
		err = DecodePayload(protocol.MsgPlayerLeft, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
	})

	t.Run("PlayerReady", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerReadyPayload{
			PlayerID: "p1",
			Ready:    true,
		}

		data, err := EncodePayload(protocol.MsgPlayerReady, original)
		require.NoError(t, err)

		var result protocol.PlayerReadyPayload
		err = DecodePayload(protocol.MsgPlayerReady, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.True(t, result.Ready)
	})
}

func TestPayloadRoundTrip_StatsAndLeaderboard(t *testing.T) {
	t.Parallel()

	t.Run("StatsResult", func(t *testing.T) {
		t.Parallel()
		original := protocol.StatsResultPayload{
			PlayerID:      "p1",
			PlayerName:    "Player1",
			TotalGames:    100,
			Wins:          60,
			Losses:        40,
			WinRate:       0.6,
			LandlordGames: 50,
			LandlordWins:  30,
			FarmerGames:   50,
			FarmerWins:    30,
			Score:         1000,
			Rank:          5,
			CurrentStreak: 3,
			MaxWinStreak:  10,
		}

		data, err := EncodePayload(protocol.MsgStatsResult, original)
		require.NoError(t, err)

		var result protocol.StatsResultPayload
		err = DecodePayload(protocol.MsgStatsResult, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.TotalGames, result.TotalGames)
		assert.Equal(t, original.WinRate, result.WinRate)
		assert.Equal(t, original.Score, result.Score)
	})

	t.Run("LeaderboardResult", func(t *testing.T) {
		t.Parallel()
		original := protocol.LeaderboardResultPayload{
			Type: "total",
			Entries: []protocol.LeaderboardEntry{
				{Rank: 1, PlayerID: "p1", PlayerName: "Champion", Score: 1000},
				{Rank: 2, PlayerID: "p2", PlayerName: "Runner", Score: 900},
			},
		}

		data, err := EncodePayload(protocol.MsgLeaderboardResult, original)
		require.NoError(t, err)

		var result protocol.LeaderboardResultPayload
		err = DecodePayload(protocol.MsgLeaderboardResult, data, &result)
		require.NoError(t, err)

		assert.Equal(t, "total", result.Type)
		assert.Len(t, result.Entries, 2)
	})

	t.Run("RoomListResult", func(t *testing.T) {
		t.Parallel()
		original := protocol.RoomListResultPayload{
			Rooms: []protocol.RoomListItem{
				{RoomCode: "123456", PlayerCount: 2, MaxPlayers: 3},
			},
		}

		data, err := EncodePayload(protocol.MsgRoomListResult, original)
		require.NoError(t, err)

		var result protocol.RoomListResultPayload
		err = DecodePayload(protocol.MsgRoomListResult, data, &result)
		require.NoError(t, err)

		assert.Len(t, result.Rooms, 1)
		assert.Equal(t, "123456", result.Rooms[0].RoomCode)
	})
}

func TestPayloadRoundTrip_ReconnectionMessages(t *testing.T) {
	t.Parallel()

	t.Run("Reconnected with GameState", func(t *testing.T) {
		t.Parallel()
		original := protocol.ReconnectedPayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
			RoomCode:   "123456",
			GameState: &protocol.GameStateDTO{
				Phase: "playing",
				Players: []protocol.PlayerInfo{
					{ID: "p1", Name: "Player1", Seat: 0, IsLandlord: true},
				},
				Hand:        []protocol.CardInfo{{Suit: 0, Rank: 3}},
				BottomCards: []protocol.CardInfo{{Suit: 1, Rank: 5}},
				CurrentTurn: "p1",
				MustPlay:    true,
			},
		}

		data, err := EncodePayload(protocol.MsgReconnected, original)
		require.NoError(t, err)

		var result protocol.ReconnectedPayload
		err = DecodePayload(protocol.MsgReconnected, data, &result)
		require.NoError(t, err)

		assert.Equal(t, "p1", result.PlayerID)
		assert.Equal(t, "123456", result.RoomCode)
		require.NotNil(t, result.GameState)
		assert.Equal(t, "playing", result.GameState.Phase)
		assert.True(t, result.GameState.MustPlay)
	})

	t.Run("PlayerOffline", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerOfflinePayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
			Timeout:    120,
		}

		data, err := EncodePayload(protocol.MsgPlayerOffline, original)
		require.NoError(t, err)

		var result protocol.PlayerOfflinePayload
		err = DecodePayload(protocol.MsgPlayerOffline, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
		assert.Equal(t, 120, result.Timeout)
	})

	t.Run("PlayerOnline", func(t *testing.T) {
		t.Parallel()
		original := protocol.PlayerOnlinePayload{
			PlayerID:   "p1",
			PlayerName: "Player1",
		}

		data, err := EncodePayload(protocol.MsgPlayerOnline, original)
		require.NoError(t, err)

		var result protocol.PlayerOnlinePayload
		err = DecodePayload(protocol.MsgPlayerOnline, data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.PlayerID, result.PlayerID)
	})
}

func TestPayloadRoundTrip_MaintenanceMessages(t *testing.T) {
	t.Parallel()

	t.Run("Maintenance", func(t *testing.T) {
		t.Parallel()
		original := protocol.MaintenancePayload{Maintenance: true}

		data, err := EncodePayload(protocol.MsgMaintenancePush, original)
		require.NoError(t, err)

		var result protocol.MaintenancePayload
		err = DecodePayload(protocol.MsgMaintenancePush, data, &result)
		require.NoError(t, err)

		assert.True(t, result.Maintenance)
	})

	t.Run("MaintenanceStatus", func(t *testing.T) {
		t.Parallel()
		original := protocol.MaintenanceStatusPayload{Maintenance: false}

		data, err := EncodePayload(protocol.MsgMaintenancePull, original)
		require.NoError(t, err)

		var result protocol.MaintenanceStatusPayload
		err = DecodePayload(protocol.MsgMaintenancePull, data, &result)
		require.NoError(t, err)

		assert.False(t, result.Maintenance)
	})
}
