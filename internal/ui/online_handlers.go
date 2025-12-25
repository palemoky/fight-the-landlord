package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/card"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
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
	}

	return nil
}

// --- 连接相关消息处理 ---

func (m *OnlineModel) handleMsgConnected(msg *protocol.Message) tea.Cmd {
	var payload protocol.ConnectedPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.playerID = payload.PlayerID
	m.playerName = payload.PlayerName
	m.input.Placeholder = "输入选项 (1-5) 或房间号"
	m.input.Focus()
	return nil
}

func (m *OnlineModel) handleMsgReconnected(msg *protocol.Message) tea.Cmd {
	var payload protocol.ReconnectedPayload
	if err := protocol.DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
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
		m.game.roomCode = payload.RoomCode
		if payload.GameState != nil {
			m.restoreGameState(payload.GameState)
		} else {
			m.phase = PhaseWaiting
		}
	} else {
		m.phase = PhaseLobby
		m.input.Placeholder = "输入选项 (1-5) 或房间号"
		m.input.Focus()
	}
	// 注意：ReconnectSuccessMsg 已通过 OnReconnect 回调发送，这里不需要再发送
	return nil
}

func (m *OnlineModel) handleMsgPong(msg *protocol.Message) tea.Cmd {
	var payload protocol.PongPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.latency = time.Now().UnixMilli() - payload.ClientTimestamp
	return nil
}

func (m *OnlineModel) handleMsgError(msg *protocol.Message) tea.Cmd {
	var payload protocol.ErrorPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.error = payload.Message
	return nil
}

func (m *OnlineModel) handleMsgOnlineCount(msg *protocol.Message) tea.Cmd {
	var payload protocol.OnlineCountPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	// m.onlineCount moved to LobbyModel
	m.lobby.onlineCount = payload.Count
	return nil
}

// --- 房间相关消息处理 ---

func (m *OnlineModel) handleMsgRoomCreated(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomCreatedPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.roomCode = payload.RoomCode
	m.game.players = []protocol.PlayerInfo{payload.Player}
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	return nil
}

func (m *OnlineModel) handleMsgRoomJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomJoinedPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.roomCode = payload.RoomCode
	m.game.players = payload.Players
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	return nil
}

func (m *OnlineModel) handleMsgPlayerJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerJoinedPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.players = append(m.game.players, payload.Player)
	return nil
}

func (m *OnlineModel) handleMsgPlayerLeft(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerLeftPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.players {
		if p.ID == payload.PlayerID {
			m.game.players = append(m.game.players[:i], m.game.players[i+1:]...)
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerReady(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerReadyPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.players {
		if p.ID == payload.PlayerID {
			m.game.players[i].Ready = payload.Ready
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
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.availableRooms = payload.Rooms
	m.lobby.selectedRoomIdx = 0
	return nil
}

func (m *OnlineModel) handleMsgPlayerOffline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOfflinePayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.players {
		if p.ID == payload.PlayerID {
			m.game.players[i].Online = false
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerOnline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOnlinePayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	for i, p := range m.game.players {
		if p.ID == payload.PlayerID {
			m.game.players[i].Online = true
			break
		}
	}
	return nil
}

// --- 游戏相关消息处理 ---

func (m *OnlineModel) handleMsgGameStart(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameStartPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.players = payload.Players
	return nil
}

func (m *OnlineModel) handleMsgDealCards(msg *protocol.Message) tea.Cmd {
	var payload protocol.DealCardsPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.hand = protocol.InfosToCards(payload.Cards)
	m.sortHand()
	if len(payload.LandlordCards) > 0 && payload.LandlordCards[0].Rank > 0 {
		m.game.landlordCards = protocol.InfosToCards(payload.LandlordCards)
	}

	// 初始化所有玩家的牌数为 17
	for i := range m.game.players {
		m.game.players[i].CardsCount = 17
	}

	// 初始化记牌器
	m.game.remainingCards = make(map[card.Rank]int)
	// 3-A 和 2 各 4 张
	for rank := card.Rank3; rank <= card.RankA; rank++ {
		m.game.remainingCards[rank] = 4
	}
	m.game.remainingCards[card.Rank2] = 4
	// 两个王各 1 张
	m.game.remainingCards[card.RankBlackJoker] = 1
	m.game.remainingCards[card.RankRedJoker] = 1

	// 扣除自己的手牌
	for _, c := range m.game.hand {
		if m.game.remainingCards[c.Rank] > 0 {
			m.game.remainingCards[c.Rank]--
		}
	}

	return nil
}

func (m *OnlineModel) handleMsgBidTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.BidTurnPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhaseBidding
	m.game.bidTurn = payload.PlayerID
	m.resetBell()
	if payload.PlayerID == m.playerID {
		m.input.Placeholder = "叫地主? (Y/N)"
		m.input.Focus()
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.game.players {
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
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.landlordCards = protocol.InfosToCards(payload.LandlordCards)
	for i, p := range m.game.players {
		m.game.players[i].IsLandlord = (p.ID == payload.PlayerID)
		// 地主拿到底牌，牌数变为 20
		if p.ID == payload.PlayerID {
			m.game.players[i].CardsCount = 20
		}
	}
	if payload.PlayerID == m.playerID {
		m.game.isLandlord = true
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayTurnPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhasePlaying
	m.game.currentTurn = payload.PlayerID
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
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.game.players {
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
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.game.lastPlayedBy = payload.PlayerID
	m.game.lastPlayedName = payload.PlayerName
	m.game.lastPlayed = protocol.InfosToCards(payload.Cards)
	m.game.lastHandType = payload.HandType
	for i, p := range m.game.players {
		if p.ID == payload.PlayerID {
			m.game.players[i].CardsCount = payload.CardsLeft
			break
		}
	}
	if payload.PlayerID == m.playerID {
		m.game.hand = card.RemoveCards(m.game.hand, m.game.lastPlayed)
	}

	// 更新记牌器（扣除已出的牌）
	for _, c := range m.game.lastPlayed {
		if m.game.remainingCards[c.Rank] > 0 {
			m.game.remainingCards[c.Rank]--
		}
	}

	return nil
}

func (m *OnlineModel) handleMsgGameOver(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameOverPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.phase = PhaseGameOver
	m.game.winner = payload.WinnerName
	m.game.winnerIsLandlord = payload.IsLandlord
	m.input.Placeholder = "按回车返回大厅"
	return nil
}

// --- 统计相关消息处理 ---

func (m *OnlineModel) handleMsgStatsResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.StatsResultPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.myStats = &payload
	return nil
}

func (m *OnlineModel) handleMsgLeaderboardResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.LeaderboardResultPayload
	_ = protocol.DecodePayload(msg.Type, msg.Payload, &payload)
	m.lobby.leaderboard = payload.Entries
	return nil
}
