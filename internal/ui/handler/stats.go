package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgStatsResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.StatsResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetMyStats(&payload)
	return nil
}

func handleMsgLeaderboardResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.LeaderboardResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetLeaderboard(payload.Entries)
	return nil
}
