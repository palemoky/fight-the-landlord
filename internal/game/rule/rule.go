package rule

import (
	"fmt"
	"slices"

	"github.com/palemoky/fight-the-landlord/internal/game/card"
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

// String 返回牌型的中文名称
// handTypeNames 牌型名称映射表
var handTypeNames = map[HandType]string{
	Single:           "单张",
	Pair:             "对子",
	Trio:             "三张",
	TrioWithSingle:   "三带一",
	TrioWithPair:     "三带二",
	Straight:         "顺子",
	PairStraight:     "连对",
	Plane:            "飞机",
	PlaneWithSingles: "飞机带单",
	PlaneWithPairs:   "飞机带对",
	Bomb:             "炸弹",
	FourWithTwo:      "四带二",
	FourWithTwoPairs: "四带两对",
	Rocket:           "王炸",
}

// handChecker 牌型检查函数类型
type handChecker func(HandAnalysis, ParsedHand) bool

// handCheckers 牌型检查函数映射表
var handCheckers = map[HandType]handChecker{
	Single:           func(a HandAnalysis, h ParsedHand) bool { return findWinningSingle(a, h) },
	Pair:             func(a HandAnalysis, h ParsedHand) bool { return findWinningPair(a, h) },
	Trio:             func(a HandAnalysis, h ParsedHand) bool { return findWinningTrio(a, h, 0) },
	TrioWithSingle:   func(a HandAnalysis, h ParsedHand) bool { return findWinningTrio(a, h, 1) },
	TrioWithPair:     func(a HandAnalysis, h ParsedHand) bool { return findWinningTrio(a, h, 2) },
	Straight:         func(a HandAnalysis, h ParsedHand) bool { return findWinningStraight(a, h) },
	PairStraight:     func(a HandAnalysis, h ParsedHand) bool { return findWinningPairStraight(a, h) },
	Plane:            func(a HandAnalysis, h ParsedHand) bool { return findWinningPlane(a, h, 0) },
	PlaneWithSingles: func(a HandAnalysis, h ParsedHand) bool { return findWinningPlane(a, h, 1) },
	PlaneWithPairs:   func(a HandAnalysis, h ParsedHand) bool { return findWinningPlane(a, h, 2) },
}

func (h HandType) String() string {
	if name, ok := handTypeNames[h]; ok {
		return name
	}
	return "无效"
}

// ParsedHand 解析后的手牌，用于比较
type ParsedHand struct {
	Type    HandType
	KeyRank card.Rank   // 决定大小的关键牌的点数 (例如 3334 中的 3, 或 34567 中的 3)
	Length  int         // 牌型的长度，主要用于顺子、连对、飞机
	Cards   []card.Card // 这手牌包含的卡牌
}

func (p ParsedHand) IsEmpty() bool {
	return p.Type == Invalid
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
		slices.Sort(ranks)
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

	// 按优先级检查各种牌型
	checks := []func(HandAnalysis, []card.Card) (ParsedHand, bool){
		isRocket,          // 王炸
		isBomb,            // 炸弹
		isFourWithKickers, // 四带二
		isTrioWithKickers, // 三带X
		isPlane,           // 飞机
		isStraight,        // 顺子
		isPairStraight,    // 连对
		isSimpleType,      // 简单牌型（单、对、三）
	}

	for _, check := range checks {
		if hand, ok := check(analysis, cards); ok {
			return hand, nil
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

// CanBeatWithHand 检查一个玩家的整手牌中是否存在任何可以打过 opponentHand 的组合
func CanBeatWithHand(playerHand []card.Card, opponentHand ParsedHand) bool {
	// 1. 如果是新一轮，总是有牌可出
	if opponentHand.IsEmpty() {
		return true
	}

	analysis := analyzeCards(playerHand)

	// 2. 检查是否有炸弹或王炸 (它们几乎可以打任何牌)
	if hasWinningBombOrRocket(analysis, opponentHand) {
		return true
	}

	if opponentHand.Type == Bomb || opponentHand.Type == Rocket {
		return false
	}

	// 3. 检查是否有同类型的、更大的牌
	if checker, ok := handCheckers[opponentHand.Type]; ok {
		return checker(analysis, opponentHand)
	}
	return false
}
