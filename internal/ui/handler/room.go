package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgRoomCreated(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomCreatedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().RoomCode = payload.RoomCode
	m.Game().State().Players = []protocol.PlayerInfo{payload.Player}
	m.SetPhase(model.PhaseWaiting)
	m.Input().Placeholder = "输入 R 准备"
	return nil
}

func handleMsgRoomJoined(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomJoinedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().RoomCode = payload.RoomCode
	m.Game().State().Players = payload.Players
	m.SetPhase(model.PhaseWaiting)
	m.Input().Placeholder = "输入 R 准备"
	m.PlaySound("join")
	return nil
}

func handleMsgPlayerJoined(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerJoinedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().Players = append(m.Game().State().Players, payload.Player)
	return nil
}

func handleMsgPlayerLeft(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerLeftPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	players := m.Game().State().Players
	for i, p := range players {
		if p.ID == payload.PlayerID {
			players = append(players[:i], players[i+1:]...)
			m.Game().State().Players = players
			break
		}
	}
	return nil
}

func handleMsgPlayerReady(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerReadyPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.Game().State().Players {
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].Ready = payload.Ready
			if payload.PlayerID == m.PlayerID() {
				if payload.Ready {
					m.Input().Placeholder = "等待其他玩家准备..."
					m.Input().Blur()
				} else {
					m.Input().Placeholder = "输入 R 准备"
					m.Input().Focus()
				}
			}
			break
		}
	}
	return nil
}

func handleMsgRoomListResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomListResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetAvailableRooms(payload.Rooms)
	return nil
}

func handleMsgPlayerOffline(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOfflinePayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.Game().State().Players {
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].Online = false
			break
		}
	}
	return nil
}

func handleMsgPlayerOnline(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOnlinePayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.Game().State().Players {
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].Online = true
			break
		}
	}
	return nil
}
