package handler

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgConnected(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ConnectedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)

	m.SetPlayerInfo(payload.PlayerID, payload.PlayerName)
	m.Client().ReconnectToken = payload.ReconnectToken

	_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetOnlineCount, nil))
	_ = m.Client().SendMessage(encoding.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	m.Input().Placeholder = "输入选项 (1-5) 或房间号"
	m.Input().Focus()
	m.PlaySound("login")
	return nil
}

func handleMsgReconnected(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ReconnectedPayload
	if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	m.SetPlayerInfo(payload.PlayerID, payload.PlayerName)

	if payload.RoomCode != "" {
		m.Game().State().RoomCode = payload.RoomCode
		if payload.GameState != nil {
			m.SetPhase(model.PhasePlaying)
		} else {
			m.SetPhase(model.PhaseWaiting)
		}
	} else {
		m.SetPhase(model.PhaseLobby)
		m.Input().Placeholder = "输入选项 (1-5) 或房间号"
		m.Input().Focus()
	}
	return nil
}

func handleMsgPong(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PongPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	return nil
}

func handleMsgError(m model.Model, msg *protocol.Message) tea.Cmd {
	payload, err := encoding.ParsePayload[protocol.ErrorPayload](msg)
	if err != nil {
		return nil
	}

	if payload.Code == protocol.ErrCodeServerMaintenance {
		m.SetMaintenanceMode(true)
		m.SetNotification(model.NotifyMaintenance, "⚠️ 服务器维护中，暂停接受新连接", false)
	}

	if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
		m.Input().Placeholder = payload.Message
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return model.ClearInputErrorMsg{}
		})
	}

	m.SetNotification(model.NotifyError, fmt.Sprintf("⚠️ %s", payload.Message), true)
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return model.ClearSystemNotificationMsg{}
	})
}

func handleMsgOnlineCount(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.OnlineCountPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetOnlineCount(payload.Count)
	m.SetNotification(model.NotifyOnlineCount, fmt.Sprintf("🌐 在线玩家: %d 人", payload.Count), false)
	return nil
}
