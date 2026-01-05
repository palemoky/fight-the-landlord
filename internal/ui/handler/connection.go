package handler

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/codec"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgConnected(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ConnectedPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)

	m.SetPlayerInfo(payload.PlayerID, payload.PlayerName)
	m.Client().ReconnectToken = payload.ReconnectToken

	_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetOnlineCount, nil))
	_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	m.Input().Placeholder = "输入选项 (1-5) 或房间号"
	m.Input().Focus()
	m.PlaySound("login")
	return nil
}

func handleMsgReconnected(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.ReconnectedPayload
	if err := payloadconv.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	m.SetPlayerInfo(payload.PlayerID, payload.PlayerName)

	// 清除旧的维护通知，避免显示过期状态
	m.ClearNotification(model.NotifyMaintenance)
	m.SetMaintenanceMode(false)

	// 从服务器获取最新状态
	_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetOnlineCount, nil))
	_ = m.Client().SendMessage(codec.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

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

func handleMsgPong(msg *protocol.Message) tea.Cmd {
	var payload protocol.PongPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	return nil
}

func handleMsgError(m model.Model, msg *protocol.Message) tea.Cmd {
	payload, err := codec.ParsePayload[protocol.ErrorPayload](msg)
	if err != nil {
		return nil
	}

	// 维护模式通知 - 持久显示
	if payload.Code == protocol.ErrCodeServerMaintenance {
		m.SetMaintenanceMode(true)
		m.SetNotification(model.NotifyMaintenance, payload.Message, false)
		return nil
	}

	// 游戏中的错误显示在输入框
	if m.Phase() == model.PhaseBidding || m.Phase() == model.PhasePlaying {
		m.Input().Placeholder = payload.Message
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return model.ClearInputErrorMsg{}
		})
	}

	// 其他错误显示为临时通知
	m.SetNotification(model.NotifyError, fmt.Sprintf("⚠️ %s", payload.Message), true)
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return model.ClearSystemNotificationMsg{}
	})
}

func handleMsgOnlineCount(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.OnlineCountPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Lobby().SetOnlineCount(payload.Count)
	m.SetNotification(model.NotifyOnlineCount, fmt.Sprintf("🌐 在线玩家: %d 人", payload.Count), false)
	return nil
}
