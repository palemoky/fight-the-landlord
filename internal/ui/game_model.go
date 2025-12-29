package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	gameClient "github.com/palemoky/fight-the-landlord/internal/client"
	"github.com/palemoky/fight-the-landlord/internal/network/client"
)

// GameModel handles game-specific UI state
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
	// Game timer and input updates handled by parent OnlineModel for now
	// Logic can be moved here if we delegate update loop fully
	return m, nil
}
