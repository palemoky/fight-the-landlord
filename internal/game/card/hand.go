package card

import (
	"fmt"
	"slices"
	"strings"
)

// findRocketInHand 查找手牌中的王炸
func findRocketInHand(hand []Card) ([]Card, bool) {
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
		return []Card{*black, *red}, true
	}
	return nil, false
}

// parseInputRanks 解析输入字符串为 Rank 计数
func parseInputRanks(input string) (map[Rank]int, error) {
	inputRanks := make(map[Rank]int)
	cleanInput := strings.ReplaceAll(input, "10", "T")

	for _, char := range cleanInput {
		rank, err := RankFromChar(char)
		if err != nil {
			return nil, err
		}
		inputRanks[rank]++
	}
	return inputRanks, nil
}

// countHandRanks 统计手牌中各 Rank 的数量
func countHandRanks(hand []Card) map[Rank]int {
	counts := make(map[Rank]int)
	for _, c := range hand {
		counts[c.Rank]++
	}
	return counts
}

// extractCards 从手牌中提取指定数量的指定 Rank 的牌
func extractCards(handCopy []Card, inputRanks map[Rank]int) []Card {
	var result []Card
	for rank, count := range inputRanks {
		found := 0
		for i := len(handCopy) - 1; i >= 0 && found < count; i-- {
			if handCopy[i].Rank == rank {
				result = append(result, handCopy[i])
				handCopy = slices.Delete(handCopy, i, i+1)
				found++
			}
		}
	}
	return result
}

// FindCardsInHand 从手牌中根据输入字符串找出对应的牌
func FindCardsInHand(hand []Card, input string) ([]Card, error) {
	// 处理王炸特殊情况
	if input == "JOKER" {
		if cards, ok := findRocketInHand(hand); ok {
			return cards, nil
		}
		return nil, fmt.Errorf("你没有王炸")
	}

	// 解析输入
	inputRanks, err := parseInputRanks(input)
	if err != nil {
		return nil, err
	}

	// 检查手牌是否足够
	handCounts := countHandRanks(hand)
	for r, count := range inputRanks {
		if handCounts[r] < count {
			return nil, fmt.Errorf("你的 %s 不够", r.String())
		}
	}

	// 提取牌
	handCopy := make([]Card, len(hand))
	copy(handCopy, hand)
	return extractCards(handCopy, inputRanks), nil
}

// RemoveCards 从手牌中移除指定的牌
func RemoveCards(hand, toRemove []Card) []Card {
	var result []Card
	for _, hCard := range hand {
		if !slices.Contains(toRemove, hCard) {
			result = append(result, hCard)
		}
	}
	return result
}
