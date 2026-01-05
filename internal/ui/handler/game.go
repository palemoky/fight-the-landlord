package handler

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
	"github.com/palemoky/fight-the-landlord/internal/protocol"
	"github.com/palemoky/fight-the-landlord/internal/protocol/convert"
	payloadconv "github.com/palemoky/fight-the-landlord/internal/protocol/convert/payload"
	"github.com/palemoky/fight-the-landlord/internal/ui/model"
)

func handleMsgGameStart(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.GameStartPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().Players = payload.Players
	return nil
}

func handleMsgDealCards(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.DealCardsPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().Hand = convert.InfosToCards(payload.Cards)
	m.Game().State().SortHand()
	if len(payload.BottomCards) > 0 && payload.BottomCards[0].Rank > 0 {
		m.Game().State().BottomCards = convert.InfosToCards(payload.BottomCards)
	}

	for i := range m.Game().State().Players {
		m.Game().State().Players[i].CardsCount = 17
	}

	m.Game().State().CardCounter.Reset()
	m.Game().State().CardCounter.DeductCards(m.Game().State().Hand)

	m.PlaySound("deal")
	return nil
}

func handleMsgBidTurn(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.BidTurnPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.SetPhase(model.PhaseBidding)
	m.Game().SetBidTurn(payload.PlayerID)
	m.Game().SetBellPlayed(false)
	if payload.PlayerID == m.PlayerID() {
		m.Input().Placeholder = "叫地主? (Y/N)"
		m.Input().Focus()
	} else {
		for _, p := range m.Game().State().Players {
			if p.ID == payload.PlayerID {
				m.Input().Placeholder = fmt.Sprintf("等待 %s 叫地主...", p.Name)
				break
			}
		}
		m.Input().Blur()
	}
	m.Game().SetTimerDuration(time.Duration(payload.Timeout) * time.Second)
	m.Game().SetTimerStartTime(time.Now())
	t := timer.NewWithInterval(m.Game().TimerDuration(), time.Second)
	m.SetTimer(t)
	return t.Start()
}

func handleMsgLandlord(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.LandlordPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().BottomCards = convert.InfosToCards(payload.BottomCards)
	for i, p := range m.Game().State().Players {
		m.Game().State().Players[i].IsLandlord = (p.ID == payload.PlayerID)
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].CardsCount = 20
		}
	}
	if payload.PlayerID == m.PlayerID() {
		m.Game().State().IsLandlord = true
		m.Game().State().CardCounter.DeductCards(m.Game().State().BottomCards)
	}

	m.PlaySound("landlord")
	return nil
}

func handleMsgPlayTurn(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.PlayTurnPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.SetPhase(model.PhasePlaying)
	m.Game().State().CurrentTurn = payload.PlayerID
	m.Game().SetMustPlay(payload.MustPlay)
	m.Game().SetCanBeat(payload.CanBeat)
	m.Game().SetBellPlayed(false)
	if payload.PlayerID == m.PlayerID() {
		switch {
		case payload.MustPlay:
			m.Input().Placeholder = "你必须出牌 (如 33344)"
		case payload.CanBeat:
			m.Input().Placeholder = "出牌或 PASS"
		default:
			m.Input().Placeholder = "没有能大过上家的牌，输入 PASS"
		}
		m.Input().Focus()
		m.PlaySound("turn")
	} else {
		for _, p := range m.Game().State().Players {
			if p.ID == payload.PlayerID {
				m.Input().Placeholder = fmt.Sprintf("等待 %s 出牌...", p.Name)
				break
			}
		}
		m.Input().Blur()
	}
	m.Game().SetTimerDuration(time.Duration(payload.Timeout) * time.Second)
	m.Game().SetTimerStartTime(time.Now())
	t := timer.NewWithInterval(m.Game().TimerDuration(), time.Second)
	m.SetTimer(t)
	return t.Start()
}

func handleMsgCardPlayed(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.CardPlayedPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.Game().State().LastPlayedBy = payload.PlayerID
	m.Game().State().LastPlayedName = payload.PlayerName
	m.Game().State().LastPlayed = convert.InfosToCards(payload.Cards)
	m.Game().State().LastHandType = payload.HandType
	for i, p := range m.Game().State().Players {
		if p.ID == payload.PlayerID {
			m.Game().State().Players[i].CardsCount = payload.CardsLeft
			break
		}
	}
	if payload.PlayerID == m.PlayerID() {
		m.Game().State().Hand = card.RemoveCards(m.Game().State().Hand, m.Game().State().LastPlayed)
	} else {
		// 只记录其他玩家出的牌
		m.Game().State().CardCounter.DeductCards(m.Game().State().LastPlayed)
	}

	m.PlaySound("play")
	return nil
}

func handleMsgGameOver(m model.Model, msg *protocol.Message) tea.Cmd {
	var payload protocol.GameOverPayload
	_ = payloadconv.DecodePayload(msg.Type, msg.Payload, &payload)
	m.SetPhase(model.PhaseGameOver)
	m.Game().State().Winner = payload.WinnerName
	m.Game().State().WinnerIsLandlord = payload.IsLandlord
	m.Input().Placeholder = "按回车返回大厅"

	// 判断是否获胜：玩家身份和赢家身份一致即为胜利
	if m.Game().State().IsLandlord == m.Game().State().WinnerIsLandlord {
		m.PlaySound("win")
	} else {
		m.PlaySound("lose")
	}

	return nil
}
