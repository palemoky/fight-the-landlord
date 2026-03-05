package rule

import (
	"slices"
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
)

// hasWinningBombOrRocket 检查是否有炸弹或王炸能压过对手的牌
func hasWinningBombOrRocket(analysis HandAnalysis, opponentHand ParsedHand) bool {
	// 检查王炸
	if analysis.counts[card.RankBlackJoker] >= 1 && analysis.counts[card.RankRedJoker] >= 1 {
		// 王炸压一切
		return true
	}

	// 检查炸弹
	for _, r := range analysis.fours {
		myBomb, _ := ParseHand([]card.Card{{Rank: r}, {Rank: r}, {Rank: r}, {Rank: r}})
		if CanBeat(myBomb, opponentHand) {
			return true
		}
	}
	return false
}

// findWinningSingle 检查是否有更大的单牌能压过对手
func findWinningSingle(analysis HandAnalysis, opponentHand ParsedHand) bool {
	for r := range analysis.counts {
		if r > opponentHand.KeyRank {
			return true
		}
	}
	return false
}

// findWinningPair 检查是否有更大的对子能压过对手
func findWinningPair(analysis HandAnalysis, opponentHand ParsedHand) bool {
	for r, count := range analysis.counts {
		if count >= 2 && r > opponentHand.KeyRank {
			return true
		}
	}
	return false
}

// findWinningTrio 检查是否有更大的三条（含带牌）能压过对手
// kickerType: 0=不带, 1=带单, 2=带对
func findWinningTrio(analysis HandAnalysis, opponentHand ParsedHand, kickerType int) bool {
	for r, count := range analysis.counts {
		if count >= 3 && r > opponentHand.KeyRank {
			// 找到更大的三条，检查是否有足够的带牌
			remainingCards := len(analysis.ones) + len(analysis.pairs)*2 + len(analysis.trios)*3 + len(analysis.fours)*4 - 3
			switch kickerType {
			case 0: // 不带牌
				return true
			case 1: // 带一张单牌
				if remainingCards >= 1 {
					return true
				}
			case 2: // 带一对
				if remainingCards < 2 {
					continue
				}
				// 检查剩余牌中是否有对子（其他对/三条/四条，或当前三条来自四条）
				if len(analysis.pairs) > 0 || len(analysis.trios) > 1 || len(analysis.fours) > 1 || (count == 4) {
					return true
				}
			}
		}
	}
	return false
}

// findWinningStraight 检查是否有更大的顺子能压过对手
func findWinningStraight(analysis HandAnalysis, opponentHand ParsedHand) bool {
	length := opponentHand.Length

	var availableRanks []card.Rank
	for r := range analysis.counts {
		if r < card.Rank2 { // 顺子不能包含2和王
			availableRanks = append(availableRanks, r)
		}
	}
	sort.Slice(availableRanks, func(i, j int) bool { return availableRanks[i] < availableRanks[j] })

	if len(availableRanks) < length {
		return false
	}

	for i := 0; i <= len(availableRanks)-length; i++ {
		// 检查连续序列
		isStraight := true
		for j := 1; j < length; j++ {
			if availableRanks[i+j-1]+1 != availableRanks[i+j] {
				isStraight = false
				break
			}
		}
		if isStraight && availableRanks[i] > opponentHand.KeyRank {
			return true
		}
	}
	return false
}

// findWinningPairStraight 检查是否有更大的连对能压过对手
func findWinningPairStraight(analysis HandAnalysis, opponentHand ParsedHand) bool {
	length := opponentHand.Length

	var pairRanks []card.Rank
	for r, count := range analysis.counts {
		if count >= 2 && r < card.Rank2 {
			pairRanks = append(pairRanks, r)
		}
	}
	slices.Sort(pairRanks)

	if len(pairRanks) < length {
		return false
	}

	// 使用与 findWinningStraight 相同的滑动窗口逻辑
	for i := 0; i <= len(pairRanks)-length; i++ {
		isPairStraight := true
		for j := 1; j < length; j++ {
			if pairRanks[i+j-1]+1 != pairRanks[i+j] {
				isPairStraight = false
				break
			}
		}
		if isPairStraight && pairRanks[i] > opponentHand.KeyRank {
			return true
		}
	}
	return false
}

// findWinningPlane 检查是否有更大的飞机（含带牌）能压过对手
// kickerType: 0=不带, 1=带单, 2=带对
func findWinningPlane(analysis HandAnalysis, opponentHand ParsedHand, kickerType int) bool {
	length := opponentHand.Length

	var trioRanks []card.Rank
	for r, count := range analysis.counts {
		if count >= 3 && r < card.Rank2 {
			trioRanks = append(trioRanks, r)
		}
	}
	slices.Sort(trioRanks)

	if len(trioRanks) < length {
		return false
	}

	for i := 0; i <= len(trioRanks)-length; i++ {
		// 检查连续序列（飞机主体）
		if !isContinuousSequence(trioRanks, i, length) {
			continue
		}

		// 检查点数是否更大
		if trioRanks[i] <= opponentHand.KeyRank {
			continue
		}

		// 检查带牌
		if checkKickers(analysis, trioRanks, i, length, kickerType) {
			return true
		}
	}
	return false
}

// isContinuousSequence 检查给定点数切片是否构成连续序列
func isContinuousSequence(ranks []card.Rank, startIndex, length int) bool {
	for j := 1; j < length; j++ {
		if ranks[startIndex+j-1]+1 != ranks[startIndex+j] {
			return false
		}
	}
	return true
}

// checkKickers 检查飞机是否有足够的带牌
func checkKickers(analysis HandAnalysis, trioRanks []card.Rank, startIndex, length, kickerType int) bool {
	if kickerType == 0 {
		return true
	}

	totalCardsInHand := 0
	for _, c := range analysis.counts {
		totalCardsInHand += c
	}
	remainingCardCount := totalCardsInHand - (length * 3)

	switch kickerType {
	case 1: // 需要 N 张单牌
		return remainingCardCount >= length
	case 2: // 需要 N 对
		if remainingCardCount < length*2 {
			return false
		}

		startRank := trioRanks[startIndex]
		endRank := trioRanks[startIndex+length-1]

		kickerPairs := 0
		for r, count := range analysis.counts {
			// 跳过飞机主体的点数
			if r >= startRank && r <= endRank {
				continue
			}
			kickerPairs += count / 2
		}
		return kickerPairs >= length
	}
	return false
}
