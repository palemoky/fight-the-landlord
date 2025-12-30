// Package ui provides the main entry point for the UI.
package ui

import (
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

// NewOnlineModel creates a new OnlineModel for online game mode.
func NewOnlineModel(serverURL string) *model.OnlineModel {
	return model.NewOnlineModel(serverURL)
}
