// Package ui provides the main entry point for the UI.
package ui

import (
	"github.com/palemoky/fight-the-landlord/internal/ui/handler"
	"github.com/palemoky/fight-the-landlord/internal/ui/input"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
	"github.com/palemoky/fight-the-landlord/internal/ui/view"
)

// NewOnlineModel creates a new OnlineModel for online game mode.
func NewOnlineModel(serverURL string) *model.OnlineModel {
	m := model.NewOnlineModel(serverURL)
	m.SetViewRenderer(view.CreateViewRenderer())
	m.SetKeyHandler(input.HandleKeyPress)
	m.SetServerMessageHandler(handler.HandleServerMessage)
	return m
}
