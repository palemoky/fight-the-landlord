package session

import (
	"log"
	"time"

	"github.com/palemoky/fight-the-landlord/internal/network/protocol"
	"github.com/palemoky/fight-the-landlord/internal/network/protocol/convert"
	"github.com/palemoky/fight-the-landlord/internal/rule"
)

const (
	// 玩家离线等待时间（秒）
	offlineWaitTimeout = 30 * time.Second
	// 出牌/叫地主超时时间
	turnTimeout = 30 * time.Second
)

// --- 超时控制 ---

func (gs *GameSession) startBidTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	gs.timerStartTime = time.Now()
	gs.remainingTime = turnTimeout
	gs.turnTimer = time.AfterFunc(turnTimeout, func() {
		// 超时自动不叫
		currentPlayer := gs.players[gs.currentBidder]
		_ = gs.HandleBid(currentPlayer.ID, false)
	})
}

func (gs *GameSession) startPlayTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	gs.timerStartTime = time.Now()
	gs.remainingTime = turnTimeout
	gs.turnTimer = time.AfterFunc(turnTimeout, func() {
		gs.handlePlayTimeout()
	})
}

func (gs *GameSession) handlePlayTimeout() {
	gs.mu.Lock()

	if gs.state != GameStatePlaying {
		gs.mu.Unlock()
		return
	}

	currentPlayer := gs.players[gs.currentPlayer]

	// 尝试找到最小能打过的牌
	cardsToPlay := rule.FindSmallestBeatingCards(currentPlayer.Hand, gs.lastPlayedHand)

	if cardsToPlay != nil {
		// 找到了能打的牌，出牌
		playerID := currentPlayer.ID
		cardInfos := convert.CardsToInfos(cardsToPlay)
		gs.mu.Unlock()
		_ = gs.HandlePlayCards(playerID, cardInfos)
		return
	}

	// 没有能打的牌，自动 PASS
	playerID := currentPlayer.ID
	gs.mu.Unlock()
	_ = gs.HandlePass(playerID)
}

func (gs *GameSession) stopTimer() {
	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	if gs.turnTimer != nil {
		gs.turnTimer.Stop()
		gs.turnTimer = nil
	}
	if gs.offlineWaitTimer != nil {
		gs.offlineWaitTimer.Stop()
		gs.offlineWaitTimer = nil
	}
}

// --- 离线处理 ---

// PlayerOffline 玩家离线
func (gs *GameSession) PlayerOffline(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			p.IsOffline = true
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		return
	}

	// 检查是否是当前回合玩家
	isBidding := gs.state == GameStateBidding && gs.currentBidder == playerIdx
	isPlaying := gs.state == GameStatePlaying && gs.currentPlayer == playerIdx

	if !isBidding && !isPlaying {
		return // 不是当前回合，无需暂停
	}

	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	// 暂停计时器，计算剩余时间
	if gs.turnTimer != nil {
		gs.turnTimer.Stop()
		gs.remainingTime = time.Until(gs.timerStartTime.Add(gs.remainingTime))
		if gs.remainingTime < 0 {
			gs.remainingTime = 0
		}
		gs.turnTimer = nil
	}

	// 启动离线等待计时器
	gs.offlineWaitTimer = time.AfterFunc(offlineWaitTimeout, func() {
		gs.handleOfflineTimeout(playerID)
	})

	log.Printf("⏸️ 玩家 %s 离线，暂停计时等待重连 (%v)", gs.players[playerIdx].Name, offlineWaitTimeout)
}

// PlayerOnline 玩家上线
func (gs *GameSession) PlayerOnline(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			p.IsOffline = false
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		return
	}

	gs.timerMu.Lock()
	defer gs.timerMu.Unlock()

	// 取消离线等待计时器
	if gs.offlineWaitTimer != nil {
		gs.offlineWaitTimer.Stop()
		gs.offlineWaitTimer = nil
	}

	// 检查是否是当前回合玩家，如果是则恢复计时器
	isBidding := gs.state == GameStateBidding && gs.currentBidder == playerIdx
	isPlaying := gs.state == GameStatePlaying && gs.currentPlayer == playerIdx

	if !isBidding && !isPlaying {
		return
	}

	// 恢复计时器
	if gs.remainingTime > 0 {
		gs.timerStartTime = time.Now()
		if isBidding {
			gs.turnTimer = time.AfterFunc(gs.remainingTime, func() {
				currentPlayer := gs.players[gs.currentBidder]
				_ = gs.HandleBid(currentPlayer.ID, false)
			})
		} else {
			gs.turnTimer = time.AfterFunc(gs.remainingTime, func() {
				gs.handlePlayTimeout()
			})
		}
		log.Printf("▶️ 玩家 %s 重连，恢复计时 (剩余 %v)", gs.players[playerIdx].Name, gs.remainingTime)
	}
}

// handleOfflineTimeout 离线超时处理
func (gs *GameSession) handleOfflineTimeout(playerID string) {
	gs.mu.Lock()

	// 找到玩家
	playerIdx := -1
	for i, p := range gs.players {
		if p.ID == playerID {
			playerIdx = i
			break
		}
	}

	if playerIdx == -1 {
		gs.mu.Unlock()
		return
	}

	log.Printf("⏰ 玩家 %s 离线超时，自动执行操作", gs.players[playerIdx].Name)

	// 根据当前状态执行自动操作
	if gs.state == GameStateBidding && gs.currentBidder == playerIdx {
		gs.mu.Unlock()
		_ = gs.HandleBid(playerID, false)
		return
	}

	if gs.state == GameStatePlaying && gs.currentPlayer == playerIdx {
		currentPlayer := gs.players[playerIdx]
		mustPlay := gs.lastPlayerIdx == gs.currentPlayer || gs.lastPlayedHand.IsEmpty()

		if mustPlay && len(currentPlayer.Hand) > 0 {
			// 出最小的牌
			minCard := currentPlayer.Hand[len(currentPlayer.Hand)-1]
			gs.mu.Unlock()
			_ = gs.HandlePlayCards(playerID, []protocol.CardInfo{convert.CardToInfo(minCard)})
			return
		}
		gs.mu.Unlock()
		_ = gs.HandlePass(playerID)
		return
	}

	gs.mu.Unlock()
}
