// Package ui provides the main entry point for the UI.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/ui/handler"
	"github.com/palemoky/fight-the-landlord/internal/ui/input"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
	"github.com/palemoky/fight-the-landlord/internal/ui/view"
)

// NewOnlineModel creates a new OnlineModel for online game mode.
func NewOnlineModel(serverURL string) *model.OnlineModel {
	m := model.NewOnlineModel(serverURL)
	m.SetViewRenderer(view.CreateViewRenderer())
	m.SetKeyHandler(func(mdl model.Model, msg tea.KeyMsg) (bool, tea.Cmd) {
		return input.HandleKeyPress(mdl, msg)
	})
	m.SetServerMessageHandler(func(mdl model.Model, msg *protocol.Message) tea.Cmd {
		return handler.HandleServerMessage(mdl, msg)
	})
	return m
}
