package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/config"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/encoding"
)

// handleServerMessage 处理服务器消息
// 按消息类型分发到具体的处理函数
func (m *OnlineModel) handleServerMessage(msg *protocol.Message) tea.Cmd {
	switch msg.Type {
	// 连接相关
	case protocol.MsgConnected:
		return m.handleMsgConnected(msg)
	case protocol.MsgReconnected:
		return m.handleMsgReconnected(msg)
	case protocol.MsgPong:
		return m.handleMsgPong(msg)
	case protocol.MsgError:
		return m.handleMsgError(msg)

	// 房间相关
	case protocol.MsgRoomCreated:
		return m.handleMsgRoomCreated(msg)
	case protocol.MsgRoomJoined:
		return m.handleMsgRoomJoined(msg)
	case protocol.MsgPlayerJoined:
		return m.handleMsgPlayerJoined(msg)
	case protocol.MsgPlayerLeft:
		return m.handleMsgPlayerLeft(msg)
	case protocol.MsgPlayerReady:
		return m.handleMsgPlayerReady(msg)
	case protocol.MsgPlayerOffline:
		return m.handleMsgPlayerOffline(msg)
	case protocol.MsgPlayerOnline:
		return m.handleMsgPlayerOnline(msg)
	case protocol.MsgOnlineCount:
		return m.handleMsgOnlineCount(msg)
	case protocol.MsgRoomListResult:
		return m.handleMsgRoomListResult(msg)

	// 游戏相关
	case protocol.MsgGameStart:
		return m.handleMsgGameStart(msg)
	case protocol.MsgDealCards:
		return m.handleMsgDealCards(msg)
	case protocol.MsgBidTurn:
		return m.handleMsgBidTurn(msg)
	case protocol.MsgBidResult:
		// 可以显示叫地主结果
		return nil
	case protocol.MsgLandlord:
		return m.handleMsgLandlord(msg)
	case protocol.MsgPlayTurn:
		return m.handleMsgPlayTurn(msg)
	case protocol.MsgCardPlayed:
		return m.handleMsgCardPlayed(msg)
	case protocol.MsgPlayerPass:
		// 可以显示 PASS 信息
		return nil
	case protocol.MsgGameOver:
		return m.handleMsgGameOver(msg)

	// 统计相关
	case protocol.MsgStatsResult:
		return m.handleMsgStatsResult(msg)
	case protocol.MsgLeaderboardResult:
		return m.handleMsgLeaderboardResult(msg)

	// Chat
	case protocol.MsgChat:
		return m.handleMsgChat(msg)

	// System notifications
	case protocol.MsgMaintenance:
		return m.handleMsgMaintenance(msg)
	case protocol.MsgMaintenanceStatus:
		return m.handleMsgMaintenanceStatus(msg)
	}

	return nil
}

// --- 连接相关消息处理 ---

func (m *OnlineModel) handleMsgConnected(msg *protocol.Message) tea.Cmd {
	var payload protocol.ConnectedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)

	m.playerID = payload.PlayerID
	m.playerName = payload.PlayerName
	m.client.ReconnectToken = payload.ReconnectToken

	// Request online count and maintenance status on connect
	_ = m.client.SendMessage(encoding.MustNewMessage(protocol.MsgGetOnlineCount, nil))
	_ = m.client.SendMessage(encoding.MustNewMessage(protocol.MsgGetMaintenanceStatus, nil))

	m.input.Placeholder = "输入选项 (1-5) 或房间号"
	m.input.Focus()
	m.soundManager.Play("login")
	return nil
}

func (m *OnlineModel) handleMsgReconnected(msg *protocol.Message) tea.Cmd {
	var payload protocol.ReconnectedPayload
	if err := convert.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
		return nil
	}

	m.playerID = payload.PlayerID
	// 只有当 payload 中有名字时才更新，避免被空字符串覆盖
	if payload.PlayerName != "" {
		m.playerName = payload.PlayerName
	} else if m.playerName == "" && m.client.PlayerName != "" {
		// 如果 model 名字为空，尝试从 client 恢复
		m.playerName = m.client.PlayerName
	}

	if payload.RoomCode != "" {
		m.game.state.RoomCode = payload.RoomCode
		if payload.GameState != nil {
			m.restoreGameState(payload.GameState)
		} else {
			m.phase = PhaseWaiting
		}
	} else {
		m.phase = PhaseLobby
		m.input.Placeholder = "输入选项 (1-5) 或房间号"
		m.input.Focus()
		// Note: Don't request online count here to avoid rate limiting
		// The notification will be set by ReconnectSuccessMsg, then cleared after 3s
		// At that point, if there's no other notification, it will show nothing until next update
	}
	// 注意：ReconnectSuccessMsg 已通过 OnReconnect 回调发送，这里不需要再发送
	return nil
}

func (m *OnlineModel) handleMsgPong(msg *protocol.Message) tea.Cmd {
	var payload protocol.PongPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.latency = time.Now().UnixMilli() - payload.ClientTimestamp
	return nil
}

func (m *OnlineModel) handleMsgError(msg *protocol.Message) tea.Cmd {
	payload, err := encoding.ParsePayload[protocol.ErrorPayload](msg)
	if err != nil {
		return nil
	}

	// 检测维护模式错误码
	if payload.Code == protocol.ErrCodeServerMaintenance {
		m.maintenanceMode = true
		// 设置维护通知（持久显示）
		m.setNotification(NotifyMaintenance, "⚠️ 服务器维护中，暂停接受新连接", false)
	}

	// 在游戏阶段（叫地主、出牌），将错误显示在 placeholder 中
	if m.phase == PhaseBidding || m.phase == PhasePlaying {
		m.input.Placeholder = payload.Message
		// 3秒后恢复原 placeholder
		return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearInputErrorMsg{}
		})
	}

	// 在大厅阶段，显示在系统通知区域（临时通知）
	// 但是如果当前正在显示重连成功消息，不要覆盖它
	if notification := m.getCurrentNotification(); notification != nil && notification.Type == NotifyReconnectSuccess {
		// 正在显示重连成功消息，忽略此错误
		return nil
	}

	m.setNotification(NotifyError, fmt.Sprintf("⚠️ %s", payload.Message), true)

	// 3秒后自动消失
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ClearSystemNotificationMsg{}
	})
}

func (m *OnlineModel) handleMsgOnlineCount(msg *protocol.Message) tea.Cmd {
	var payload protocol.OnlineCountPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	// m.onlineCount moved to LobbyModel
	m.lobby.onlineCount = payload.Count
	// 设置在线人数通知（持久显示）
	m.setNotification(NotifyOnlineCount, fmt.Sprintf("🌐 在线玩家: %d 人", payload.Count), false)
	return nil
}

// --- 房间相关消息处理 ---

func (m *OnlineModel) handleMsgRoomCreated(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomCreatedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.RoomCode = payload.RoomCode
	m.game.state.Players = []protocol.PlayerInfo{payload.Player}
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	return nil
}

func (m *OnlineModel) handleMsgRoomJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomJoinedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.RoomCode = payload.RoomCode
	m.game.state.Players = payload.Players
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	m.soundManager.Play("join")
	return nil
}

func (m *OnlineModel) handleMsgPlayerJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerJoinedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.Players = append(m.game.state.Players, payload.Player)
	return nil
}

func (m *OnlineModel) handleMsgPlayerLeft(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerLeftPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.state.Players {
		if p.ID == payload.PlayerID {
			m.game.state.Players = append(m.game.state.Players[:i], m.game.state.Players[i+1:]...)
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerReady(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerReadyPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.state.Players {
		if p.ID == payload.PlayerID {
			m.game.state.Players[i].Ready = payload.Ready
			// 如果是自己的准备状态变化，更新 placeholder
			if payload.PlayerID == m.playerID {
				if payload.Ready {
					m.input.Placeholder = "等待其他玩家准备..."
					m.input.Blur()
				} else {
					m.input.Placeholder = "输入 R 准备"
					m.input.Focus()
				}
			}
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgRoomListResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomListResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.availableRooms = payload.Rooms
	m.lobby.selectedRoomIdx = 0
	return nil
}

func (m *OnlineModel) handleMsgPlayerOffline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOfflinePayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.state.Players {
		if p.ID == payload.PlayerID {
			m.game.state.Players[i].Online = false
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerOnline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOnlinePayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.state.Players {
		if p.ID == payload.PlayerID {
			m.game.state.Players[i].Online = true
			break
		}
	}
	return nil
}

// --- 游戏相关消息处理 ---

func (m *OnlineModel) handleMsgGameStart(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameStartPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.Players = payload.Players
	return nil
}

func (m *OnlineModel) handleMsgDealCards(msg *protocol.Message) tea.Cmd {
	var payload protocol.DealCardsPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.Hand = convert.InfosToCards(payload.Cards)
	m.game.state.SortHand()
	if len(payload.LandlordCards) > 0 && payload.LandlordCards[0].Rank > 0 {
		m.game.state.BottomCards = convert.InfosToCards(payload.LandlordCards)
	}

	// 初始化所有玩家的牌数为 17
	for i := range m.game.state.Players {
		m.game.state.Players[i].CardsCount = 17
	}

	// 初始化记牌器并扣除自己的手牌
	m.game.state.CardCounter.Reset()
	m.game.state.CardCounter.DeductCards(m.game.state.Hand)

	m.soundManager.Play("deal")
	return nil
}

func (m *OnlineModel) handleMsgBidTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.BidTurnPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhaseBidding
	m.game.bidTurn = payload.PlayerID
	m.resetBell()
	if payload.PlayerID == m.playerID {
		m.input.Placeholder = "叫地主? (Y/N)"
		m.input.Focus()
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.game.state.Players {
			if p.ID == payload.PlayerID {
				m.input.Placeholder = fmt.Sprintf("等待 %s 叫地主...", p.Name)
				break
			}
		}
		m.input.Blur()
	}
	m.game.timerDuration = time.Duration(payload.Timeout) * time.Second
	m.game.timerStartTime = time.Now()
	m.timer = timer.NewWithInterval(m.game.timerDuration, time.Second)
	return m.timer.Start()
}

func (m *OnlineModel) handleMsgLandlord(msg *protocol.Message) tea.Cmd {
	var payload protocol.LandlordPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.BottomCards = convert.InfosToCards(payload.LandlordCards)
	for i, p := range m.game.state.Players {
		m.game.state.Players[i].IsLandlord = (p.ID == payload.PlayerID)
		// 地主拿到底牌，牌数变为 20
		if p.ID == payload.PlayerID {
			m.game.state.Players[i].CardsCount = 20
		}
	}
	if payload.PlayerID == m.playerID {
		m.game.state.IsLandlord = true
	}

	// 更新记牌器：根据底牌是否公开来决定如何扣除
	// 底牌公开：所有玩家都扣除；底牌不公开：只有地主扣除
	shouldDeductBottomCards := config.BottomCardsPublic || payload.PlayerID == m.playerID
	if shouldDeductBottomCards {
		m.game.state.CardCounter.DeductCards(m.game.state.BottomCards)
	}

	m.soundManager.Play("landlord")
	return nil
}

func (m *OnlineModel) handleMsgPlayTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayTurnPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhasePlaying
	m.game.state.CurrentTurn = payload.PlayerID
	m.game.mustPlay = payload.MustPlay
	m.game.canBeat = payload.CanBeat
	m.resetBell()
	if payload.PlayerID == m.playerID {
		switch {
		case payload.MustPlay:
			m.input.Placeholder = "你必须出牌 (如 33344)"
		case payload.CanBeat:
			m.input.Placeholder = "出牌或 PASS"
		default:
			m.input.Placeholder = "没有能大过上家的牌，输入 PASS"
		}
		m.input.Focus()
		m.soundManager.Play("turn")
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.game.state.Players {
			if p.ID == payload.PlayerID {
				m.input.Placeholder = fmt.Sprintf("等待 %s 出牌...", p.Name)
				break
			}
		}
		m.input.Blur()
	}
	m.game.timerDuration = time.Duration(payload.Timeout) * time.Second
	m.game.timerStartTime = time.Now()
	m.timer = timer.NewWithInterval(m.game.timerDuration, time.Second)
	return m.timer.Start()
}

func (m *OnlineModel) handleMsgCardPlayed(msg *protocol.Message) tea.Cmd {
	var payload protocol.CardPlayedPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.state.LastPlayedBy = payload.PlayerID
	m.game.state.LastPlayedName = payload.PlayerName
	m.game.state.LastPlayed = convert.InfosToCards(payload.Cards)
	m.game.state.LastHandType = payload.HandType
	for i, p := range m.game.state.Players {
		if p.ID == payload.PlayerID {
			m.game.state.Players[i].CardsCount = payload.CardsLeft
			break
		}
	}
	if payload.PlayerID == m.playerID {
		m.game.state.Hand = card.RemoveCards(m.game.state.Hand, m.game.state.LastPlayed)
	}

	// 更新记牌器
	m.game.state.CardCounter.DeductCards(m.game.state.LastPlayed)

	m.soundManager.Play("play")
	return nil
}

func (m *OnlineModel) handleMsgGameOver(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameOverPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhaseGameOver
	m.game.state.Winner = payload.WinnerName
	m.game.state.WinnerIsLandlord = payload.IsLandlord
	m.input.Placeholder = "按回车返回大厅"

	isWinner := false
	if m.game.state.IsLandlord && m.game.state.WinnerIsLandlord {
		isWinner = true
	} else if !m.game.state.IsLandlord && !m.game.state.WinnerIsLandlord {
		isWinner = true
	}

	if isWinner {
		m.soundManager.Play("win")
	} else {
		m.soundManager.Play("lose")
	}

	return nil
}

// --- 统计相关消息处理 ---

func (m *OnlineModel) handleMsgStatsResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.StatsResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.myStats = &payload
	return nil
}

func (m *OnlineModel) handleMsgLeaderboardResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.LeaderboardResultPayload
	_ = convert.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.leaderboard = payload.Entries
	return nil
}
