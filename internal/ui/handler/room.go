package handler

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgRoomCreated(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomCreatedPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().RoomCode = payload.RoomCode
	m.Game().State().Players = []protocol.PlayerInfo{payload.Player}
	m.SetPhase(model.PhaseWaiting)
	m.Input().Placeholder = "输入 R 准备"
	return nil
}

func handleMsgRoomJoined(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomJoinedPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().RoomCode = payload.RoomCode
	m.Game().State().Players = payload.Players
	m.SetPhase(model.PhaseWaiting)
	m.Input().Placeholder = "输入 R 准备"
	m.PlaySound("join")
	return nil
}

func handleMsgPlayerJoined(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerJoinedPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().Players = append(m.Game().State().Players, payload.Player)
	return nil
}

func handleMsgPlayerLeft(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerLeftPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
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
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)

	// 更新玩家状态
	for i, p := range m.Game().State().Players {
		if p.ID != payload.PlayerID {
			continue
		}
		m.Game().State().Players[i].Ready = payload.Ready

		// 只更新自己的输入状态
		if payload.PlayerID != m.PlayerID() {
			break
		}
		if payload.Ready {
			m.Input().Placeholder = "等待其他玩家准备..."
			m.Input().Blur()
		} else {
			m.Input().Placeholder = "输入 R 准备"
			m.Input().Focus()
		}
		break
	}
	return nil
}

func handleMsgRoomListResult(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomListResultPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetAvailableRooms(payload.Rooms)
	return nil
}

func handleMsgPlayerOffline(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOfflinePayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
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
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.Game().State().Players {
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].Online = true
			break
		}
	}
	return nil
}
