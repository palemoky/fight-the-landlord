package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

// Icon constants
const (
	LandlordIcon = "👑"
	FarmerIcon   = "🧑‍🌾"

	TopBorderStart    = "┌──"
	TopBorderEnd      = "┌──┐"
	SideBorder        = "│"
	BottomBorderStart = "└──"
	BottomBorderEnd   = "└──┘"
)

// Lipgloss Styles - shared across local and online modes
var (
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#CD0000")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	blackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	grayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("#FFFFFF")).Bold(true)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true).Render
	boxStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	promptStyle  = lipgloss.NewStyle().MarginTop(1)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	displayOrder = []card.Rank{card.RankRedJoker, card.RankBlackJoker, card.Rank2, card.RankA, card.RankK, card.RankQ, card.RankJ, card.Rank10, card.Rank9, card.Rank8, card.Rank7, card.Rank6, card.Rank5, card.Rank4, card.Rank3}
)
