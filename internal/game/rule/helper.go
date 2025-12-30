package rule

import "github.com/palemoky/fight-the-landlord/internal/game/card"

// FindSmallestBeatingCards 找到能打过 opponentHand 的最小牌组
// 如果找不到，返回 nil
func FindSmallestBeatingCards(playerHand []card.Card, opponentHand ParsedHand) []card.Card {
	// 如果是新一轮，出最小的单牌
	if opponentHand.IsEmpty() {
		if len(playerHand) > 0 {
			return []card.Card{playerHand[len(playerHand)-1]}
		}
		return nil
	}

	analysis := analyzeCards(playerHand)

	// 优先尝试找同类型的最小牌
	var result []card.Card

	switch opponentHand.Type {
	case Single:
		result = findSmallestBeatingSingle(playerHand, analysis, opponentHand)
	case Pair:
		result = findSmallestBeatingPair(playerHand, analysis, opponentHand)
	case Trio:
		result = findSmallestBeatingTrio(playerHand, analysis, opponentHand, 0)
	case TrioWithSingle:
		result = findSmallestBeatingTrio(playerHand, analysis, opponentHand, 1)
	case TrioWithPair:
		result = findSmallestBeatingTrio(playerHand, analysis, opponentHand, 2)
	}

	// 如果找到了同类型的牌，返回
	if result != nil {
		return result
	}

	// 否则尝试用最小的炸弹
	result = findSmallestBomb(playerHand, analysis, opponentHand)
	if result != nil {
		return result
	}

	// 最后尝试王炸（一般不会用）
	if hasRocket(analysis) && opponentHand.Type != Rocket {
		return findRocket(playerHand)
	}

	return nil
}

// findSmallestBeatingSingle 找到能打过的最小单牌
func findSmallestBeatingSingle(playerHand []card.Card, analysis HandAnalysis, opponentHand ParsedHand) []card.Card {
	for _, r := range analysis.ones {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 1)
		}
	}
	for _, r := range analysis.pairs {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 1)
		}
	}
	for _, r := range analysis.trios {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 1)
		}
	}
	for _, r := range analysis.fours {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 1)
		}
	}
	return nil
}

// findSmallestBeatingPair 找到能打过的最小对子
func findSmallestBeatingPair(playerHand []card.Card, analysis HandAnalysis, opponentHand ParsedHand) []card.Card {
	for _, r := range analysis.pairs {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 2)
		}
	}
	for _, r := range analysis.trios {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 2)
		}
	}
	for _, r := range analysis.fours {
		if r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 2)
		}
	}
	return nil
}

// findSmallestBeatingTrio 找到能打过的最小三张（带或不带）
func findSmallestBeatingTrio(playerHand []card.Card, analysis HandAnalysis, opponentHand ParsedHand, kickerType int) []card.Card {
	for _, r := range analysis.trios {
		if r > opponentHand.KeyRank {
			result := findCardsWithRank(playerHand, r, 3)
			if kickerType == 0 {
				return result
			}
			// 需要带牌
			kickers := findSmallestKickers(playerHand, analysis, r, kickerType)
			if kickers != nil {
				return append(result, kickers...)
			}
		}
	}
	for _, r := range analysis.fours {
		if r > opponentHand.KeyRank {
			result := findCardsWithRank(playerHand, r, 3)
			if kickerType == 0 {
				return result
			}
			kickers := findSmallestKickers(playerHand, analysis, r, kickerType)
			if kickers != nil {
				return append(result, kickers...)
			}
		}
	}
	return nil
}

// findSmallestBomb 找到最小的炸弹
func findSmallestBomb(playerHand []card.Card, analysis HandAnalysis, opponentHand ParsedHand) []card.Card {
	for _, r := range analysis.fours {
		if opponentHand.Type != Bomb || r > opponentHand.KeyRank {
			return findCardsWithRank(playerHand, r, 4)
		}
	}
	return nil
}

// findSmallestKickers 找到最小的带牌
// kickerType: 1=带单张, 2=带对子
func findSmallestKickers(playerHand []card.Card, analysis HandAnalysis, excludeRank card.Rank, kickerType int) []card.Card {
	var kickers []card.Card
	neededCards := kickerType // 1张单牌或2张(1对)

	// collectFromRanks 从给定的点数列表中收集 kicker 牌
	collectFromRanks := func(ranks []card.Rank, countPerRank int) bool {
		for _, r := range ranks {
			if r != excludeRank {
				kickers = append(kickers, findCardsWithRank(playerHand, r, countPerRank)...)
				if len(kickers) >= neededCards {
					kickers = kickers[:neededCards]
					return true
				}
			}
		}
		return false
	}

	if kickerType == 1 {
		// 带单张：优先从单牌、对子中取
		if collectFromRanks(analysis.ones, 1) || collectFromRanks(analysis.pairs, 1) {
			return kickers
		}
	} else {
		// 带对子：从对子、三张、四张中取
		if collectFromRanks(analysis.pairs, 2) ||
			collectFromRanks(analysis.trios, 2) ||
			collectFromRanks(analysis.fours, 2) {
			return kickers
		}
	}
	return nil
}

// findCardsWithRank 从手牌中找到指定点数的牌
func findCardsWithRank(playerHand []card.Card, rank card.Rank, count int) []card.Card {
	var result []card.Card
	for _, c := range playerHand {
		if c.Rank == rank {
			result = append(result, c)
			if len(result) >= count {
				return result
			}
		}
	}
	return result
}

// hasRocket 检查是否有王炸
func hasRocket(analysis HandAnalysis) bool {
	return analysis.counts[card.RankBlackJoker] > 0 && analysis.counts[card.RankRedJoker] > 0
}

// findRocket 找到王炸
func findRocket(playerHand []card.Card) []card.Card {
	var result []card.Card
	for _, c := range playerHand {
		if c.Rank == card.RankBlackJoker || c.Rank == card.RankRedJoker {
			result = append(result, c)
		}
	}
	return result
}
