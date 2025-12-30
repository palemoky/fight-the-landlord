// Package model contains the UI model implementations.
package model

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	gameClient "github.com/palemoky/fight-the-landlord/internal/client"
	"github.com/palemoky/fight-the-landlord/internal/network/client"
)

// GameModel handles game-specific UI state.
type GameModel struct {
	// Network client
	client *client.Client

	// Game state (business logic)
	state *gameClient.GameState

	// UI dimensions
	width  int
	height int

	// Input handling
	input *textinput.Model

	// Bidding UI state
	bidTurn string

	// UI state flags
	mustPlay bool
	canBeat  bool

	// UI helper state
	bellPlayed     bool
	timerDuration  time.Duration
	timerStartTime time.Time

	// Features
	cardCounterEnabled bool
	showingHelp        bool

	// Chat UI
	chatHistory      []string
	chatInput        textinput.Model
	showQuickMsgMenu bool
}

// NewGameModel creates a new GameModel.
func NewGameModel(c *client.Client, input *textinput.Model) *GameModel {
	chatInput := textinput.New()
	chatInput.Placeholder = "按 / 键聊天, T 键快捷消息..."
	chatInput.CharLimit = 50
	chatInput.Width = 30

	return &GameModel{
		client:    c,
		state:     gameClient.NewGameState(),
		input:     input,
		chatInput: chatInput,
	}
}

func (m *GameModel) Init() tea.Cmd {
	return nil
}

func (m *GameModel) View() string {
	return "" // Not used directly, managed by OnlineModel
}

func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// --- GameAccessor implementation ---

func (m *GameModel) State() *gameClient.GameState { return m.state }

func (m *GameModel) BidTurn() string        { return m.bidTurn }
func (m *GameModel) SetBidTurn(turn string) { m.bidTurn = turn }

func (m *GameModel) MustPlay() bool        { return m.mustPlay }
func (m *GameModel) SetMustPlay(must bool) { m.mustPlay = must }
func (m *GameModel) CanBeat() bool         { return m.canBeat }
func (m *GameModel) SetCanBeat(can bool)   { m.canBeat = can }

func (m *GameModel) TimerDuration() time.Duration     { return m.timerDuration }
func (m *GameModel) SetTimerDuration(d time.Duration) { m.timerDuration = d }
func (m *GameModel) TimerStartTime() time.Time        { return m.timerStartTime }
func (m *GameModel) SetTimerStartTime(t time.Time)    { m.timerStartTime = t }

func (m *GameModel) BellPlayed() bool          { return m.bellPlayed }
func (m *GameModel) SetBellPlayed(played bool) { m.bellPlayed = played }

func (m *GameModel) CardCounterEnabled() bool           { return m.cardCounterEnabled }
func (m *GameModel) SetCardCounterEnabled(enabled bool) { m.cardCounterEnabled = enabled }
func (m *GameModel) ShowingHelp() bool                  { return m.showingHelp }
func (m *GameModel) SetShowingHelp(showing bool)        { m.showingHelp = showing }

func (m *GameModel) ChatHistory() []string { return m.chatHistory }
func (m *GameModel) AddChatMessage(msg string) {
	m.chatHistory = append(m.chatHistory, msg)
	if len(m.chatHistory) > 50 {
		m.chatHistory = m.chatHistory[len(m.chatHistory)-50:]
	}
}
func (m *GameModel) ShowQuickMsgMenu() bool        { return m.showQuickMsgMenu }
func (m *GameModel) SetShowQuickMsgMenu(show bool) { m.showQuickMsgMenu = show }

func (m *GameModel) Width() int  { return m.width }
func (m *GameModel) Height() int { return m.height }
func (m *GameModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
func (m *GameModel) Input() *textinput.Model     { return m.input }
func (m *GameModel) ChatInput() *textinput.Model { return &m.chatInput }
func (m *GameModel) Client() *client.Client      { return m.client }
