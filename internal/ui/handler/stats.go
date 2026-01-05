package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgStatsResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.StatsResultPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetMyStats(&payload)
	return nil
}

func handleMsgLeaderboardResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.LeaderboardResultPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetLeaderboard(payload.Entries)
	return nil
}
