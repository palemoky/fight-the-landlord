package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/client"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
)

// GameModel handles game-specific logic (Waiting, Game states)
type GameModel struct {
	client *client.Client
	width  int
	height int

	input *textinput.Model

	// Game Data
	roomCode         string
	players          []protocol.PlayerInfo
	hand             []card.Card
	landlordCards    []card.Card
	currentTurn      string
	lastPlayedBy     string
	lastPlayedName   string
	lastPlayed       []card.Card
	lastHandType     string
	isLandlord       bool
	winner           string
	winnerIsLandlord bool

	// Bidding
	bidTurn string

	// State flags
	mustPlay bool
	canBeat  bool

	// Helper state
	bellPlayed     bool
	timerDuration  time.Duration
	timerStartTime time.Time

	// Features
	cardCounterEnabled bool
	remainingCards     map[card.Rank]int
	showingHelp        bool

	// Chat & Quick Messages
	chatHistory      []string
	chatInput        textinput.Model // Reuse for chat
	showQuickMsgMenu bool
}

func NewGameModel(c *client.Client, input *textinput.Model) *GameModel {
	chatInput := textinput.New()
	chatInput.Placeholder = "按 / 键聊天, T 键快捷消息..."
	chatInput.CharLimit = 50
	chatInput.Width = 30

	return &GameModel{
		client:    c,
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
