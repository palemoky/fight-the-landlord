package model

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/stretchr/testify/assert"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// --- GameModel Tests ---

func TestNewGameModel(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewGameModel(nil, &input)

	assert.NotNil(t, m)
	assert.NotNil(t, m.State())
	assert.Equal(t, "", m.BidTurn())
	assert.False(t, m.MustPlay())
	assert.False(t, m.CanBeat())
	assert.False(t, m.ShowingHelp())
	assert.False(t, m.CardCounterEnabled())
	assert.Empty(t, m.ChatHistory())
}

func TestGameModel_BidTurn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		bidTurn  string
		expected string
	}{
		{"set player1", "player1", "player1"},
		{"set player2", "player2", "player2"},
		{"set empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)
			m.SetBidTurn(tt.bidTurn)
			assert.Equal(t, tt.expected, m.BidTurn())
		})
	}
}

func TestGameModel_PlayFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mustPlay bool
		canBeat  bool
	}{
		{"both true", true, true},
		{"both false", false, false},
		{"must only", true, false},
		{"beat only", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)
			m.SetMustPlay(tt.mustPlay)
			m.SetCanBeat(tt.canBeat)
			assert.Equal(t, tt.mustPlay, m.MustPlay())
			assert.Equal(t, tt.canBeat, m.CanBeat())
		})
	}
}

func TestGameModel_Timer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"30 seconds", 30 * time.Second},
		{"1 minute", 1 * time.Minute},
		{"0 duration", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)
			startTime := time.Now()

			m.SetTimerDuration(tt.duration)
			m.SetTimerStartTime(startTime)

			assert.Equal(t, tt.duration, m.TimerDuration())
			assert.Equal(t, startTime.Unix(), m.TimerStartTime().Unix())
		})
	}
}

func TestGameModel_Features(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		counter    bool
		help       bool
		bellPlayed bool
	}{
		{"all enabled", true, true, true},
		{"all disabled", false, false, false},
		{"counter only", true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)

			m.SetCardCounterEnabled(tt.counter)
			m.SetShowingHelp(tt.help)
			m.SetBellPlayed(tt.bellPlayed)

			assert.Equal(t, tt.counter, m.CardCounterEnabled())
			assert.Equal(t, tt.help, m.ShowingHelp())
			assert.Equal(t, tt.bellPlayed, m.BellPlayed())
		})
	}
}

func TestGameModel_ChatHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		messages      []string
		expectedLen   int
		expectedFirst string
	}{
		{"empty", []string{}, 0, ""},
		{"single message", []string{"Hello"}, 1, "Hello"},
		{"multiple messages", []string{"A", "B", "C"}, 3, "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)

			for _, msg := range tt.messages {
				m.AddChatMessage(msg)
			}

			assert.Len(t, m.ChatHistory(), tt.expectedLen)
			if tt.expectedLen > 0 {
				assert.Equal(t, tt.expectedFirst, m.ChatHistory()[0])
			}
		})
	}
}

func TestGameModel_ChatHistory_MaxLimit(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewGameModel(nil, &input)

	// Add 60 messages
	for i := 0; i < 60; i++ {
		m.AddChatMessage("msg")
	}

	// Should only keep last 50
	assert.Len(t, m.ChatHistory(), 50)
}

func TestGameModel_ClearChatHistory(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewGameModel(nil, &input)

	m.AddChatMessage("test1")
	m.AddChatMessage("test2")
	assert.Len(t, m.ChatHistory(), 2)

	m.ClearChatHistory()
	assert.Empty(t, m.ChatHistory())
}

func TestGameModel_Size(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", 800, 600},
		{"large", 1920, 1080},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewGameModel(nil, &input)

			m.SetSize(tt.width, tt.height)
			assert.Equal(t, tt.width, m.Width())
			assert.Equal(t, tt.height, m.Height())
		})
	}
}

// --- LobbyModel Tests ---

func TestNewLobbyModel(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewLobbyModel(nil, &input)

	assert.NotNil(t, m)
	assert.Equal(t, 0, m.OnlineCount())
	assert.Equal(t, 0, m.SelectedIndex())
	assert.Empty(t, m.ChatHistory())
	assert.Empty(t, m.AvailableRooms())
}

func TestLobbyModel_OnlineCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"zero", 0, 0},
		{"positive", 42, 42},
		{"large", 10000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewLobbyModel(nil, &input)

			m.SetOnlineCount(tt.count)
			assert.Equal(t, tt.expected, m.OnlineCount())
		})
	}
}

func TestLobbyModel_AvailableRooms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		rooms             []protocol.RoomListItem
		selectIdx         int
		expectedRoomCount int
	}{
		{"empty rooms", []protocol.RoomListItem{}, 0, 0},
		{"single room", []protocol.RoomListItem{{RoomCode: "123456", PlayerCount: 2}}, 0, 1},
		{"multiple rooms", []protocol.RoomListItem{
			{RoomCode: "111111", PlayerCount: 1},
			{RoomCode: "222222", PlayerCount: 2},
		}, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewLobbyModel(nil, &input)

			m.SetAvailableRooms(tt.rooms)
			assert.Len(t, m.AvailableRooms(), tt.expectedRoomCount)
			assert.Equal(t, 0, m.SelectedRoomIdx()) // Reset to 0 on set

			if len(tt.rooms) > 0 {
				m.SetSelectedRoomIdx(tt.selectIdx)
				assert.Equal(t, tt.selectIdx, m.SelectedRoomIdx())
			}
		})
	}
}

func TestLobbyModel_Leaderboard(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewLobbyModel(nil, &input)

	entries := []protocol.LeaderboardEntry{
		{Rank: 1, PlayerName: "Player1", Score: 1000},
		{Rank: 2, PlayerName: "Player2", Score: 900},
	}

	m.SetLeaderboard(entries)
	assert.Equal(t, entries, m.Leaderboard())
}

func TestLobbyModel_MyStats(t *testing.T) {
	t.Parallel()

	input := textinput.New()
	m := NewLobbyModel(nil, &input)

	assert.Nil(t, m.MyStats())

	stats := &protocol.StatsResultPayload{
		TotalGames: 100,
		Wins:       60,
		Losses:     40,
		WinRate:    60.0,
		Score:      500,
	}

	m.SetMyStats(stats)
	assert.Equal(t, stats, m.MyStats())
}

func TestLobbyModel_HandleUpKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phase       GamePhase
		startIdx    int
		rooms       []protocol.RoomListItem
		expectedIdx int
	}{
		{"lobby wrap around from 0", PhaseLobby, 0, nil, 5},
		{"lobby normal decrement", PhaseLobby, 3, nil, 2},
		{"room list wrap around", PhaseRoomList, 0, []protocol.RoomListItem{{}, {}, {}}, 2},
		{"room list normal decrement", PhaseRoomList, 2, []protocol.RoomListItem{{}, {}, {}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewLobbyModel(nil, &input)

			if tt.rooms != nil {
				m.SetAvailableRooms(tt.rooms)
				m.SetSelectedRoomIdx(tt.startIdx)
			} else {
				m.SetSelectedIndex(tt.startIdx)
			}

			m.HandleUpKey(tt.phase)

			if tt.phase == PhaseRoomList {
				assert.Equal(t, tt.expectedIdx, m.SelectedRoomIdx())
			} else {
				assert.Equal(t, tt.expectedIdx, m.SelectedIndex())
			}
		})
	}
}

func TestLobbyModel_HandleDownKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phase       GamePhase
		startIdx    int
		rooms       []protocol.RoomListItem
		expectedIdx int
	}{
		{"lobby wrap around from 5", PhaseLobby, 5, nil, 0},
		{"lobby normal increment", PhaseLobby, 3, nil, 4},
		{"room list wrap around", PhaseRoomList, 2, []protocol.RoomListItem{{}, {}, {}}, 0},
		{"room list normal increment", PhaseRoomList, 0, []protocol.RoomListItem{{}, {}, {}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewLobbyModel(nil, &input)

			if tt.rooms != nil {
				m.SetAvailableRooms(tt.rooms)
				m.SetSelectedRoomIdx(tt.startIdx)
			} else {
				m.SetSelectedIndex(tt.startIdx)
			}

			m.HandleDownKey(tt.phase)

			if tt.phase == PhaseRoomList {
				assert.Equal(t, tt.expectedIdx, m.SelectedRoomIdx())
			} else {
				assert.Equal(t, tt.expectedIdx, m.SelectedIndex())
			}
		})
	}
}

func TestLobbyModel_Size(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", 1920, 1080},
		{"small", 800, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := textinput.New()
			m := NewLobbyModel(nil, &input)

			m.SetSize(tt.width, tt.height)
			assert.Equal(t, tt.width, m.Width())
			assert.Equal(t, tt.height, m.Height())
		})
	}
}
