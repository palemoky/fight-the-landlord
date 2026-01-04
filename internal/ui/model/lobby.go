// Package model contains the UI model implementations.
package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/transport"
)

const (
	// chatBoxWidth is the width of the chat box.
	chatBoxWidth = 50
	// chatInputWidth is the width of the chat input.
	chatInputWidth = chatBoxWidth - 5
)

// LobbyModel handles the lobby interface.
type LobbyModel struct {
	client *transport.Client
	width  int
	height int

	// Navigation
	selectedIndex int

	// Data
	onlineCount     int
	availableRooms  []protocol.RoomListItem
	selectedRoomIdx int
	leaderboard     []protocol.LeaderboardEntry
	myStats         *protocol.StatsResultPayload

	// Chat
	chatHistory []string
	chatInput   textinput.Model

	// Input reference
	input *textinput.Model
}

// NewLobbyModel creates a new LobbyModel.
func NewLobbyModel(c *transport.Client, input *textinput.Model) *LobbyModel {
	chatInput := textinput.New()
	chatInput.Placeholder = "按 / 键聊天..."
	chatInput.CharLimit = 50
	chatInput.Width = chatInputWidth

	return &LobbyModel{
		client:    c,
		input:     input,
		chatInput: chatInput,
	}
}

func (m *LobbyModel) Init() tea.Cmd {
	return nil
}

func (m *LobbyModel) View() string {
	return "" // Not used directly, managed by OnlineModel
}

func (m *LobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// --- LobbyAccessor implementation ---

func (m *LobbyModel) OnlineCount() int                        { return m.onlineCount }
func (m *LobbyModel) SetOnlineCount(count int)                { m.onlineCount = count }
func (m *LobbyModel) AvailableRooms() []protocol.RoomListItem { return m.availableRooms }
func (m *LobbyModel) SetAvailableRooms(rooms []protocol.RoomListItem) {
	m.availableRooms = rooms
	m.selectedRoomIdx = 0
}
func (m *LobbyModel) SelectedRoomIdx() int                               { return m.selectedRoomIdx }
func (m *LobbyModel) SetSelectedRoomIdx(idx int)                         { m.selectedRoomIdx = idx }
func (m *LobbyModel) Leaderboard() []protocol.LeaderboardEntry           { return m.leaderboard }
func (m *LobbyModel) SetLeaderboard(entries []protocol.LeaderboardEntry) { m.leaderboard = entries }
func (m *LobbyModel) MyStats() *protocol.StatsResultPayload              { return m.myStats }
func (m *LobbyModel) SetMyStats(stats *protocol.StatsResultPayload)      { m.myStats = stats }

func (m *LobbyModel) ChatHistory() []string { return m.chatHistory }
func (m *LobbyModel) AddChatMessage(msg string) {
	m.chatHistory = append(m.chatHistory, msg)
	if len(m.chatHistory) > 50 {
		m.chatHistory = m.chatHistory[len(m.chatHistory)-50:]
	}
}
func (m *LobbyModel) ChatInput() *textinput.Model { return &m.chatInput }

func (m *LobbyModel) HandleUpKey(phase GamePhase) {
	if phase == PhaseRoomList && len(m.availableRooms) > 0 {
		m.selectedRoomIdx--
		if m.selectedRoomIdx < 0 {
			m.selectedRoomIdx = len(m.availableRooms) - 1
		}
	} else if phase == PhaseLobby {
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = 5
		}
	}
}

func (m *LobbyModel) HandleDownKey(phase GamePhase) {
	if phase == PhaseRoomList && len(m.availableRooms) > 0 {
		m.selectedRoomIdx++
		if m.selectedRoomIdx >= len(m.availableRooms) {
			m.selectedRoomIdx = 0
		}
	} else if phase == PhaseLobby {
		m.selectedIndex++
		if m.selectedIndex > 5 {
			m.selectedIndex = 0
		}
	}
}

func (m *LobbyModel) Width() int  { return m.width }
func (m *LobbyModel) Height() int { return m.height }
func (m *LobbyModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
func (m *LobbyModel) Input() *textinput.Model   { return m.input }
func (m *LobbyModel) SelectedIndex() int        { return m.selectedIndex }
func (m *LobbyModel) SetSelectedIndex(idx int)  { m.selectedIndex = idx }
func (m *LobbyModel) Client() *transport.Client { return m.client }
