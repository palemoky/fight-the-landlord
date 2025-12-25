package ui

import (
	"encoding/json"
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
	_ = json.Unmarshal(msg.Payload, &payload)
	m.playerID = payload.PlayerID
	m.playerName = payload.PlayerName
	m.input.Placeholder = "输入选项 (1-5) 或房间号"
	m.input.Focus()
	return nil
}

func (m *OnlineModel) handleMsgReconnected(msg *protocol.Message) tea.Cmd {
	var payload protocol.ReconnectedPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.playerID = payload.PlayerID
	m.playerName = payload.PlayerName
	if payload.RoomCode != "" {
		m.roomCode = payload.RoomCode
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
	return nil
}

func (m *OnlineModel) handleMsgPong(msg *protocol.Message) tea.Cmd {
	var payload protocol.PongPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.latency = time.Now().UnixMilli() - payload.ClientTimestamp
	return nil
}

func (m *OnlineModel) handleMsgError(msg *protocol.Message) tea.Cmd {
	var payload protocol.ErrorPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.error = payload.Message
	return nil
}

// --- 房间相关消息处理 ---

func (m *OnlineModel) handleMsgRoomCreated(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomCreatedPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.roomCode = payload.RoomCode
	m.players = []protocol.PlayerInfo{payload.Player}
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	return nil
}

func (m *OnlineModel) handleMsgRoomJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.RoomJoinedPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.roomCode = payload.RoomCode
	m.players = payload.Players
	m.phase = PhaseWaiting
	m.input.Placeholder = "输入 R 准备"
	return nil
}

func (m *OnlineModel) handleMsgPlayerJoined(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerJoinedPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.players = append(m.players, payload.Player)
	return nil
}

func (m *OnlineModel) handleMsgPlayerLeft(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerLeftPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	for i, p := range m.players {
		if p.ID == payload.PlayerID {
			m.players = append(m.players[:i], m.players[i+1:]...)
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerReady(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerReadyPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	for i, p := range m.players {
		if p.ID == payload.PlayerID {
			m.players[i].Ready = payload.Ready
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

func (m *OnlineModel) handleMsgPlayerOffline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOfflinePayload
	_ = json.Unmarshal(msg.Payload, &payload)
	for i, p := range m.players {
		if p.ID == payload.PlayerID {
			m.players[i].Online = false
			break
		}
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayerOnline(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayerOnlinePayload
	_ = json.Unmarshal(msg.Payload, &payload)
	for i, p := range m.players {
		if p.ID == payload.PlayerID {
			m.players[i].Online = true
			break
		}
	}
	return nil
}

// --- 游戏相关消息处理 ---

func (m *OnlineModel) handleMsgGameStart(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameStartPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.players = payload.Players
	return nil
}

func (m *OnlineModel) handleMsgDealCards(msg *protocol.Message) tea.Cmd {
	var payload protocol.DealCardsPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.hand = protocol.InfosToCards(payload.Cards)
	m.sortHand()
	if len(payload.LandlordCards) > 0 && payload.LandlordCards[0].Rank > 0 {
		m.landlordCards = protocol.InfosToCards(payload.LandlordCards)
	}
	return nil
}

func (m *OnlineModel) handleMsgBidTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.BidTurnPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.phase = PhaseBidding
	m.bidTurn = payload.PlayerID
	m.resetBell()
	if payload.PlayerID == m.playerID {
		m.input.Placeholder = "叫地主? (Y/N)"
		m.input.Focus()
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.players {
			if p.ID == payload.PlayerID {
				m.input.Placeholder = fmt.Sprintf("等待 %s 叫地主...", p.Name)
				break
			}
		}
		m.input.Blur()
	}
	m.timerDuration = time.Duration(payload.Timeout) * time.Second
	m.timerStartTime = time.Now()
	m.timer = timer.NewWithInterval(m.timerDuration, time.Second)
	return m.timer.Start()
}

func (m *OnlineModel) handleMsgLandlord(msg *protocol.Message) tea.Cmd {
	var payload protocol.LandlordPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.landlordCards = protocol.InfosToCards(payload.LandlordCards)
	for i, p := range m.players {
		m.players[i].IsLandlord = (p.ID == payload.PlayerID)
	}
	if payload.PlayerID == m.playerID {
		m.isLandlord = true
	}
	return nil
}

func (m *OnlineModel) handleMsgPlayTurn(msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayTurnPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.phase = PhasePlaying
	m.currentTurn = payload.PlayerID
	m.mustPlay = payload.MustPlay
	m.canBeat = payload.CanBeat
	m.resetBell()
	if payload.PlayerID == m.playerID {
		switch {
		case payload.MustPlay:
			m.input.Placeholder = "你必须出牌 (如 33344)"
		case payload.CanBeat:
			m.input.Placeholder = "出牌或 PASS"
		default:
			m.input.Placeholder = "没有能打过的牌，输入 PASS"
		}
		m.input.Focus()
	} else {
		// 不是自己的回合，显示等待提示
		for _, p := range m.players {
			if p.ID == payload.PlayerID {
				m.input.Placeholder = fmt.Sprintf("等待 %s 出牌...", p.Name)
				break
			}
		}
		m.input.Blur()
	}
	m.timerDuration = time.Duration(payload.Timeout) * time.Second
	m.timerStartTime = time.Now()
	m.timer = timer.NewWithInterval(m.timerDuration, time.Second)
	return m.timer.Start()
}

func (m *OnlineModel) handleMsgCardPlayed(msg *protocol.Message) tea.Cmd {
	var payload protocol.CardPlayedPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.lastPlayedBy = payload.PlayerID
	m.lastPlayedName = payload.PlayerName
	m.lastPlayed = protocol.InfosToCards(payload.Cards)
	m.lastHandType = payload.HandType
	for i, p := range m.players {
		if p.ID == payload.PlayerID {
			m.players[i].CardsCount = payload.CardsLeft
			break
		}
	}
	if payload.PlayerID == m.playerID {
		m.hand = card.RemoveCards(m.hand, m.lastPlayed)
	}
	return nil
}

func (m *OnlineModel) handleMsgGameOver(msg *protocol.Message) tea.Cmd {
	var payload protocol.GameOverPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.phase = PhaseGameOver
	m.winner = payload.WinnerName
	m.winnerIsLandlord = payload.IsLandlord
	m.input.Placeholder = "按回车返回大厅"
	return nil
}

// --- 统计相关消息处理 ---

func (m *OnlineModel) handleMsgStatsResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.StatsResultPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.myStats = &payload
	return nil
}

func (m *OnlineModel) handleMsgLeaderboardResult(msg *protocol.Message) tea.Cmd {
	var payload protocol.LeaderboardResultPayload
	_ = json.Unmarshal(msg.Payload, &payload)
	m.leaderboard = payload.Entries
	return nil
}
