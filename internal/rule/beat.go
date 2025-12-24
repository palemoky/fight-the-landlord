package rule

import (
	"slices"
	"sort"

	"github.com/palemoky/fight-the-landlord/internal/card"
)

// hasWinningBombOrRocket checks for any bomb or rocket that can beat the opponent's hand.
func hasWinningBombOrRocket(analysis HandAnalysis, opponentHand ParsedHand) bool {
	// Check for a winning Rocket.
	if analysis.counts[card.RankBlackJoker] >= 1 && analysis.counts[card.RankRedJoker] >= 1 {
		// A Rocket beats anything.
		return true
	}

	// Check for a winning Bomb.
	for _, r := range analysis.fours {
		myBomb, _ := ParseHand([]card.Card{{Rank: r}, {Rank: r}, {Rank: r}, {Rank: r}})
		if CanBeat(myBomb, opponentHand) {
			return true
		}
	}
	return false
}

// findWinningSingle checks for any single card that can win.
func findWinningSingle(analysis HandAnalysis, opponentHand ParsedHand) bool {
	for r := range analysis.counts {
		if r > opponentHand.KeyRank {
			return true // Found a higher card.
		}
	}
	return false
}

// findWinningPair checks for any pair that can win.
func findWinningPair(analysis HandAnalysis, opponentHand ParsedHand) bool {
	for r, count := range analysis.counts {
		if count >= 2 && r > opponentHand.KeyRank {
			return true // Found a higher pair.
		}
	}
	return false
}

// findWinningTrio checks for trios with or without kickers.
// kickerType: 0=none, 1=single, 2=pair.
func findWinningTrio(analysis HandAnalysis, opponentHand ParsedHand, kickerType int) bool {
	for r, count := range analysis.counts {
		if count >= 3 && r > opponentHand.KeyRank {
			// Found a higher trio. Now check if we have enough cards for kickers.
			remainingCards := len(analysis.ones) + len(analysis.pairs)*2 + len(analysis.trios)*3 + len(analysis.fours)*4 - 3
			switch kickerType {
			case 0: // No kicker needed
				return true
			case 1: // Need one single
				if remainingCards >= 1 {
					return true
				}
			case 2: // Need one pair
				if remainingCards < 2 {
					continue
				}
				// Check if the remaining cards contain a pair.
				// This is true if there's any other pair/trio/four, or if the current trio came from a four.
				if len(analysis.pairs) > 0 || len(analysis.trios) > 1 || len(analysis.fours) > 1 || (count == 4) {
					return true
				}
			}
		}
	}
	return false
}

// findWinningStraight checks for a winning straight of a specific length.
func findWinningStraight(analysis HandAnalysis, opponentHand ParsedHand) bool {
	length := opponentHand.Length

	var availableRanks []card.Rank
	for r := range analysis.counts {
		if r < card.Rank2 { // Straights cannot include 2 or Jokers
			availableRanks = append(availableRanks, r)
		}
	}
	sort.Slice(availableRanks, func(i, j int) bool { return availableRanks[i] < availableRanks[j] })

	if len(availableRanks) < length {
		return false
	}

	for i := 0; i <= len(availableRanks)-length; i++ {
		// Check for a continuous sequence
		isStraight := true
		for j := 1; j < length; j++ {
			if availableRanks[i+j-1]+1 != availableRanks[i+j] {
				isStraight = false
				break
			}
		}
		if isStraight && availableRanks[i] > opponentHand.KeyRank {
			return true // Found a higher straight.
		}
	}
	return false
}

// findWinningPairStraight checks for a winning pair straight.
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

	// Use the same sliding window logic as findWinningStraight
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

// findWinningPlane checks for a winning plane with or without kickers.
// kickerType: 0=none, 1=singles, 2=pairs.
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
		isPlane := true
		for j := 1; j < length; j++ {
			if trioRanks[i+j-1]+1 != trioRanks[i+j] {
				isPlane = false
				break
			}
		}

		if isPlane && trioRanks[i] > opponentHand.KeyRank {
			// Found a higher plane. Now check for kickers.
			totalCardsInHand := 0
			for _, c := range analysis.counts {
				totalCardsInHand += c
			}
			remainingCardCount := totalCardsInHand - (length * 3)

			switch kickerType {
			case 0: // No kickers
				return true
			case 1: // Need N singles
				if remainingCardCount >= length {
					return true
				}
			case 2: // Need N pairs
				if remainingCardCount < length*2 {
					continue
				}
				// This is a complex check. A simplified but effective heuristic:
				// Count how many pairs can be formed from the rest of the hand.
				kickerPairs := 0
				for r, count := range analysis.counts {
					// Is this rank part of the plane we just found?
					isPlaneRank := false
					for k := range length {
						if trioRanks[i+k] == r {
							isPlaneRank = true
							break
						}
					}
					if !isPlaneRank {
						kickerPairs += count / 2
					}
				}
				if kickerPairs >= length {
					return true
				}
			}
		}
	}
	return false
}
