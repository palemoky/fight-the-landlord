package card

import (
	"fmt"
	"slices"
	"strings"
)

// FindCardsInHand 从手牌中根据输入字符串找出对应的牌
func FindCardsInHand(hand []Card, input string) ([]Card, error) {
	if input == "JOKER" {
		var black, red *Card
		for i := range hand {
			if hand[i].Rank == RankBlackJoker {
				black = &hand[i]
			}
			if hand[i].Rank == RankRedJoker {
				red = &hand[i]
			}
		}
		if black != nil && red != nil {
			return []Card{*black, *red}, nil
		}
		return nil, fmt.Errorf("你没有王炸")
	}

	inputRanks := make(map[Rank]int)
	cleanInput := strings.ReplaceAll(input, "10", "T")

	for _, char := range cleanInput {
		rank, err := RankFromChar(char)
		if err != nil {
			return nil, err
		}
		inputRanks[rank]++
	}

	var result []Card
	handCopy := make([]Card, len(hand))
	copy(handCopy, hand)

	// 先检查手牌是否足够
	handCounts := make(map[Rank]int)
	for _, c := range hand {
		handCounts[c.Rank]++
	}
	for r, count := range inputRanks {
		if handCounts[r] < count {
			return nil, fmt.Errorf("你的 %s 不够", r.String())
		}
	}

	// 提取牌
	for rank, count := range inputRanks {
		found := 0
		for i := len(handCopy) - 1; i >= 0; i-- {
			if handCopy[i].Rank == rank {
				result = append(result, handCopy[i])
				handCopy = slices.Delete(handCopy, i, i+1)
				found++
				if found == count {
					break
				}
			}
		}
	}
	return result, nil
}

// RemoveCards 从手牌中移除指定的牌
func RemoveCards(hand []Card, toRemove []Card) []Card {
	var result []Card
	for _, hCard := range hand {
		if !slices.Contains(toRemove, hCard) {
			result = append(result, hCard)
		}
	}
	return result
}
