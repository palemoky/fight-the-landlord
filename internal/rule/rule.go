package rule

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
)

// HandType 定义牌型
type HandType int

const (
	Invalid        HandType = iota
	Single                  // 单张
	Pair                    // 对子
	Trio                    // 三张不带
	TrioWithSingle          // 三带一
	TrioWithPair            // 三带二

	Straight         // 顺子（5张或以上连续单张）
	PairStraight     // 连对（3对或以上）
	Plane            // 飞机不带翅膀（2个或以上连续三张）
	PlaneWithSingles // 飞机带单
	PlaneWithPairs   // 飞机带对

	Bomb             // 炸弹（四张相同）
	FourWithTwo      // 四带二（带两张相同或不同的单牌）
	FourWithTwoPairs // 四带两对（带两对）

	Rocket // 王炸（双王）
)

// ParsedHand 解析后的手牌，用于比较
type ParsedHand struct {
	Type    HandType
	KeyRank card.Rank   // 决定大小的关键牌的点数 (例如 3334 中的 3, 或 34567 中的 3)
	Length  int         // 牌型的长度，主要用于顺子、连对、飞机
	Cards   []card.Card // 这手牌包含的卡牌
}

func (p ParsedHand) IsEmpty() bool {
	return len(p.Cards) == 0
}

// HandAnalysis 对一手牌进行预分析，统计不同点数的牌出现了几次
type HandAnalysis struct {
	counts map[card.Rank]int // 每种点数牌的数量
	// 为了方便，提前将不同数量的牌分组
	fours []card.Rank
	trios []card.Rank
	pairs []card.Rank
	ones  []card.Rank
}

// analyzeCards 分析手牌，返回一个包含所有统计信息的结构
func analyzeCards(cards []card.Card) HandAnalysis {
	analysis := HandAnalysis{
		counts: make(map[card.Rank]int),
	}
	for _, c := range cards {
		analysis.counts[c.Rank]++
	}

	for r, count := range analysis.counts {
		switch count {
		case 4:
			analysis.fours = append(analysis.fours, r)
		case 3:
			analysis.trios = append(analysis.trios, r)
		case 2:
			analysis.pairs = append(analysis.pairs, r)
		case 1:
			analysis.ones = append(analysis.ones, r)
		}
	}

	// 对结果进行排序，方便后续判断连续性
	sortRanks := func(ranks []card.Rank) {
		sort.Slice(ranks, func(i, j int) bool { return ranks[i] < ranks[j] })
	}
	sortRanks(analysis.fours)
	sortRanks(analysis.trios)
	sortRanks(analysis.pairs)
	sortRanks(analysis.ones)

	return analysis
}

// isContinuous 检查给定的点数切片是否连续，并且不能包含 2 和大小王
func isContinuous(ranks []card.Rank) bool {
	if len(ranks) == 0 {
		return false
	}
	for i, r := range ranks {
		if r >= card.Rank2 { // 顺子、连对、飞机不能包含2和王
			return false
		}
		if i > 0 && ranks[i-1]+1 != r {
			return false
		}
	}
	return true
}

// ParseHand 解析牌型
func ParseHand(cards []card.Card) (ParsedHand, error) {
	if len(cards) == 0 {
		return ParsedHand{}, fmt.Errorf("不能出空牌")
	}

	analysis := analyzeCards(cards)
	cardLen := len(cards)

	// 1. 王炸
	if cardLen == 2 && analysis.counts[card.RankBlackJoker] == 1 && analysis.counts[card.RankRedJoker] == 1 {
		return ParsedHand{Type: Rocket, KeyRank: card.RankRedJoker, Cards: cards}, nil
	}

	// 2. 炸弹
	if len(analysis.fours) == 1 && cardLen == 4 {
		return ParsedHand{Type: Bomb, KeyRank: analysis.fours[0], Cards: cards}, nil
	}

	// 3. 简单牌型: 单、对、三
	if len(analysis.counts) == 1 {
		switch cardLen {
		case 1:
			return ParsedHand{Type: Single, KeyRank: analysis.ones[0], Cards: cards}, nil
		case 2:
			return ParsedHand{Type: Pair, KeyRank: analysis.pairs[0], Cards: cards}, nil
		case 3:
			return ParsedHand{Type: Trio, KeyRank: analysis.trios[0], Cards: cards}, nil
		}
	}

	// 4. 三带X
	if len(analysis.trios) == 1 {
		if cardLen == 4 && len(analysis.ones) == 1 { // AAAB
			return ParsedHand{Type: TrioWithSingle, KeyRank: analysis.trios[0], Cards: cards}, nil
		}
		if cardLen == 5 && len(analysis.pairs) == 1 { // AAABB
			return ParsedHand{Type: TrioWithPair, KeyRank: analysis.trios[0], Cards: cards}, nil
		}
	}

	// 5. 四带二
	if len(analysis.fours) == 1 {
		if cardLen == 6 && (len(analysis.ones) == 2 || len(analysis.pairs) == 1) { // AAAABB、AAAABC
			// 四带二，可以带两张单牌，也可以带一个对子(不算四带两对)
			return ParsedHand{Type: FourWithTwo, KeyRank: analysis.fours[0], Cards: cards}, nil
		}
		if cardLen == 8 && len(analysis.pairs) == 2 { // AAAABBCC、AAAABBBB
			return ParsedHand{Type: FourWithTwoPairs, KeyRank: analysis.fours[0], Cards: cards}, nil
		}
	}

	// 6. 顺子类型 (顺子, 连对, 飞机)
	// 6.1 顺子
	if isContinuous(analysis.ones) && len(analysis.ones) == cardLen && cardLen >= 5 { // ABCDE+
		return ParsedHand{Type: Straight, KeyRank: analysis.ones[0], Length: cardLen, Cards: cards}, nil
	}

	// 6.2 连对
	if isContinuous(analysis.pairs) && len(analysis.pairs)*2 == cardLen && len(analysis.pairs) >= 3 { // AABBCC+
		return ParsedHand{Type: PairStraight, KeyRank: analysis.pairs[0], Length: len(analysis.pairs), Cards: cards}, nil
	}

	// 6.3 飞机
	planeLen := len(analysis.trios)
	if isContinuous(analysis.trios) && planeLen >= 2 {
		// 飞机不带翅膀
		if planeLen*3 == cardLen { // AAABBB+
			return ParsedHand{Type: Plane, KeyRank: analysis.trios[0], Length: planeLen, Cards: cards}, nil
		}
		// 飞机带单
		if planeLen*4 == cardLen && len(analysis.ones) == planeLen { // AAABBBCD+、AAABBAC+、AAABBBCC+
			return ParsedHand{Type: PlaneWithSingles, KeyRank: analysis.trios[0], Length: planeLen, Cards: cards}, nil
		}
		// 飞机带对
		if planeLen*5 == cardLen && len(analysis.pairs) == planeLen { // AAABBBCCDD+
			return ParsedHand{Type: PlaneWithPairs, KeyRank: analysis.trios[0], Length: planeLen, Cards: cards}, nil
		}
	}

	return ParsedHand{}, fmt.Errorf("不支持的牌型: %v", cards)
}

// CanBeat 判断 newHand 是否能大过 lastHand
func CanBeat(newHand, lastHand ParsedHand) bool {
	// 王炸最大
	if newHand.Type == Rocket {
		return true
	}
	if lastHand.Type == Rocket {
		return false
	}

	// 炸弹可以大过任何非炸弹和非王炸的牌
	if newHand.Type == Bomb && lastHand.Type != Bomb {
		return true
	}

	// 如果牌型不同 (且我不是炸弹)，不能出
	if newHand.Type != lastHand.Type {
		return false
	}

	// 对于顺子、连对、飞机，长度必须一致
	if newHand.Length != lastHand.Length && (newHand.Type == Straight || newHand.Type == PairStraight || newHand.Type == Plane || newHand.Type == PlaneWithSingles || newHand.Type == PlaneWithPairs) {
		return false
	}

	// 如果牌型相同或者是炸弹盖炸弹
	return newHand.KeyRank > lastHand.KeyRank
}

func rankFromChar(char rune) (card.Rank, error) {
	switch char {
	case '3':
		return card.Rank3, nil
	case '4':
		return card.Rank4, nil
	case '5':
		return card.Rank5, nil
	case '6':
		return card.Rank6, nil
	case '7':
		return card.Rank7, nil
	case '8':
		return card.Rank8, nil
	case '9':
		return card.Rank9, nil
	case 'T':
		return card.Rank10, nil
	case 'J':
		return card.RankJ, nil
	case 'Q':
		return card.RankQ, nil
	case 'K':
		return card.RankK, nil
	case 'A':
		return card.RankA, nil
	case '2':
		return card.Rank2, nil
	case 'B':
		return card.RankBlackJoker, nil
	case 'R':
		return card.RankRedJoker, nil
	default:
		return -1, fmt.Errorf("无法识别的点数: %c", char)
	}
}

// FindCardsInHand 从手牌中根据输入字符串找出对应的牌
func FindCardsInHand(hand []card.Card, input string) ([]card.Card, error) {
	if input == "JOKER" { // 特殊处理王炸
		var black, red *card.Card
		for i := range hand {
			if hand[i].Rank == card.RankBlackJoker {
				black = &hand[i]
			}
			if hand[i].Rank == card.RankRedJoker {
				red = &hand[i]
			}
		}
		if black != nil && red != nil {
			return []card.Card{*black, *red}, nil
		}
		return nil, fmt.Errorf("你没有王炸")
	}

	inputRanks := make(map[card.Rank]int)
	cleanInput := strings.ReplaceAll(input, "10", "T")

	for _, char := range cleanInput {
		rank, err := rankFromChar(char)
		if err != nil {
			return nil, err
		}
		inputRanks[rank]++
	}

	var result []card.Card
	handCopy := make([]card.Card, len(hand))
	copy(handCopy, hand)

	// 先检查手牌是否足够
	handCounts := make(map[card.Rank]int)
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
func RemoveCards(hand []card.Card, toRemove []card.Card) []card.Card {
	var result []card.Card
	for _, hCard := range hand {
		if !slices.Contains(toRemove, hCard) {
			result = append(result, hCard)
		}
	}
	return result
}

// CanBeatWithHand 检查一个玩家的整手牌中是否存在任何可以打过 opponentHand 的组合
func CanBeatWithHand(playerHand []card.Card, opponentHand ParsedHand) bool {
	// 1. 如果是新一轮，总是有牌可出
	if opponentHand.IsEmpty() {
		return true
	}

	// 2. 检查是否有炸弹或王炸 (它们几乎可以打任何牌)
	// (为了简化，这里只检查一个基础的炸弹逻辑)
	rankCounts := make(map[card.Rank]int)
	for _, c := range playerHand {
		rankCounts[c.Rank]++
	}
	
	hasRocket := false
	if rankCounts[card.RankBlackJoker] > 0 && rankCounts[card.RankRedJoker] > 0 {
		hasRocket = true
	}

	for rank, count := range rankCounts {
		if count == 4 { // 找到一个炸弹
			// 如果对手不是王炸或更大的炸弹，那么我方的炸弹总能打
			if opponentHand.Type != Rocket && (opponentHand.Type != Bomb || rank > opponentHand.KeyRank) {
				return true
			}
		}
	}
	if hasRocket {
		return true // 王炸最大
	}

	// 3. 检查是否有同类型的、更大的牌
	// (这是一个非常复杂的步骤，涉及到从手牌中找出所有特定类型的组合)
	// 这是一个简化的示例，仅用于演示逻辑，实际应用需要更完善的组合查找
	// TODO: 在此实现更完整的牌型查找逻辑 (如顺子、飞机等)
	// for _, possibleMove := range findAllMoves(playerHand, opponentHand.Type) {
	//     if CanBeat(possibleMove, opponentHand) {
	//         return true
	//     }
	// }

	// 4. 如果以上都找不到，则认为没有牌可以打
	// 在一个完整的实现中，如果上面的TODO部分没有返回true，才执行这里
	return false // 这是一个临时的简化返回值
}