package protocol

import "encoding/json"

// Message 基础消息结构
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// MessageType 消息类型
type MessageType string

// 客户端 → 服务端 消息类型
const (
	// 连接操作
	MsgReconnect MessageType = "reconnect" // 断线重连
	MsgPing      MessageType = "ping"      // 心跳 ping

	// 房间操作
	MsgCreateRoom  MessageType = "create_room"  // 创建房间
	MsgJoinRoom    MessageType = "join_room"    // 加入房间
	MsgLeaveRoom   MessageType = "leave_room"   // 离开房间
	MsgQuickMatch  MessageType = "quick_match"  // 快速匹配
	MsgReady       MessageType = "ready"        // 准备就绪
	MsgCancelReady MessageType = "cancel_ready" // 取消准备

	// 游戏操作
	MsgBid       MessageType = "bid"        // 叫地主
	MsgPlayCards MessageType = "play_cards" // 出牌
	MsgPass      MessageType = "pass"       // 不出
)

// 服务端 → 客户端 消息类型
const (
	// 连接相关
	MsgConnected     MessageType = "connected"      // 连接成功
	MsgReconnected   MessageType = "reconnected"    // 重连成功
	MsgPong          MessageType = "pong"           // 心跳 pong
	MsgPlayerOffline MessageType = "player_offline" // 玩家掉线通知
	MsgPlayerOnline  MessageType = "player_online"  // 玩家上线通知

	// 房间相关
	MsgRoomCreated  MessageType = "room_created"  // 房间创建成功
	MsgRoomJoined   MessageType = "room_joined"   // 加入房间成功
	MsgPlayerJoined MessageType = "player_joined" // 其他玩家加入
	MsgPlayerLeft   MessageType = "player_left"   // 玩家离开
	MsgPlayerReady  MessageType = "player_ready"  // 玩家准备
	MsgMatchFound   MessageType = "match_found"   // 匹配成功

	// 游戏流程
	MsgGameStart   MessageType = "game_start"   // 游戏开始
	MsgDealCards   MessageType = "deal_cards"   // 发牌
	MsgBidTurn     MessageType = "bid_turn"     // 轮到叫地主
	MsgBidResult   MessageType = "bid_result"   // 叫地主结果
	MsgLandlord    MessageType = "landlord"     // 地主确定
	MsgPlayTurn    MessageType = "play_turn"    // 轮到出牌
	MsgCardPlayed  MessageType = "card_played"  // 有人出牌
	MsgPlayerPass  MessageType = "player_pass"  // 有人不出
	MsgGameOver    MessageType = "game_over"    // 游戏结束
	MsgRoundResult MessageType = "round_result" // 本轮结果

	// 错误
	MsgError MessageType = "error" // 错误消息
)

// --- 客户端请求 Payloads ---

// ReconnectPayload 断线重连请求
type ReconnectPayload struct {
	Token    string `json:"token"`     // 重连令牌
	PlayerID string `json:"player_id"` // 玩家 ID
}

// PingPayload 心跳请求
type PingPayload struct {
	Timestamp int64 `json:"timestamp"` // 客户端时间戳（毫秒）
}

// JoinRoomPayload 加入房间请求
type JoinRoomPayload struct {
	RoomCode string `json:"room_code"`
}

// BidPayload 叫地主请求
type BidPayload struct {
	Bid bool `json:"bid"` // true = 叫地主, false = 不叫
}

// PlayCardsPayload 出牌请求
type PlayCardsPayload struct {
	Cards []CardInfo `json:"cards"`
}

// --- 服务端响应 Payloads ---

// ConnectedPayload 连接成功响应
type ConnectedPayload struct {
	PlayerID       string `json:"player_id"`
	PlayerName     string `json:"player_name"`
	ReconnectToken string `json:"reconnect_token"` // 重连令牌
}

// ReconnectedPayload 重连成功响应
type ReconnectedPayload struct {
	PlayerID   string        `json:"player_id"`
	PlayerName string        `json:"player_name"`
	RoomCode   string        `json:"room_code,omitempty"`  // 如果在房间中
	GameState  *GameStateDTO `json:"game_state,omitempty"` // 如果在游戏中
}

// GameStateDTO 游戏状态数据传输对象（用于重连恢复）
type GameStateDTO struct {
	Phase         string       `json:"phase"`          // bidding/playing
	Players       []PlayerInfo `json:"players"`        // 所有玩家信息
	Hand          []CardInfo   `json:"hand"`           // 自己的手牌
	LandlordCards []CardInfo   `json:"landlord_cards"` // 底牌
	CurrentTurn   string       `json:"current_turn"`   // 当前回合玩家 ID
	LastPlayed    []CardInfo   `json:"last_played"`    // 上家出的牌
	LastPlayerID  string       `json:"last_player_id"` // 上家 ID
	MustPlay      bool         `json:"must_play"`      // 是否必须出牌
	CanBeat       bool         `json:"can_beat"`       // 是否能打过
}

// PongPayload 心跳响应
type PongPayload struct {
	ClientTimestamp int64 `json:"client_timestamp"` // 客户端发送的时间戳
	ServerTimestamp int64 `json:"server_timestamp"` // 服务器时间戳（毫秒）
}

// PlayerOfflinePayload 玩家掉线通知
type PlayerOfflinePayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Timeout    int    `json:"timeout"` // 等待重连超时（秒）
}

// PlayerOnlinePayload 玩家上线通知
type PlayerOnlinePayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

// RoomCreatedPayload 房间创建成功响应
type RoomCreatedPayload struct {
	RoomCode string     `json:"room_code"`
	Player   PlayerInfo `json:"player"`
}

// RoomJoinedPayload 加入房间成功响应
type RoomJoinedPayload struct {
	RoomCode string       `json:"room_code"`
	Player   PlayerInfo   `json:"player"`
	Players  []PlayerInfo `json:"players"` // 房间内所有玩家
}

// PlayerJoinedPayload 其他玩家加入通知
type PlayerJoinedPayload struct {
	Player PlayerInfo `json:"player"`
}

// PlayerLeftPayload 玩家离开通知
type PlayerLeftPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

// PlayerReadyPayload 玩家准备通知
type PlayerReadyPayload struct {
	PlayerID string `json:"player_id"`
	Ready    bool   `json:"ready"`
}

// GameStartPayload 游戏开始通知
type GameStartPayload struct {
	Players []PlayerInfo `json:"players"` // 按座位顺序排列
}

// DealCardsPayload 发牌通知
type DealCardsPayload struct {
	Cards         []CardInfo `json:"cards"`          // 玩家自己的手牌
	LandlordCards []CardInfo `json:"landlord_cards"` // 底牌（地主确定后才显示具体内容）
}

// BidTurnPayload 轮到叫地主通知
type BidTurnPayload struct {
	PlayerID string `json:"player_id"`
	Timeout  int    `json:"timeout"` // 超时时间（秒）
}

// BidResultPayload 叫地主结果通知
type BidResultPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Bid        bool   `json:"bid"`
}

// LandlordPayload 地主确定通知
type LandlordPayload struct {
	PlayerID      string     `json:"player_id"`
	PlayerName    string     `json:"player_name"`
	LandlordCards []CardInfo `json:"landlord_cards"` // 底牌
}

// PlayTurnPayload 轮到出牌通知
type PlayTurnPayload struct {
	PlayerID string `json:"player_id"`
	Timeout  int    `json:"timeout"`   // 超时时间（秒）
	MustPlay bool   `json:"must_play"` // 是否必须出牌（新一轮开始时为 true）
	CanBeat  bool   `json:"can_beat"`  // 是否有牌能打过上家
}

// CardPlayedPayload 出牌通知
type CardPlayedPayload struct {
	PlayerID   string     `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Cards      []CardInfo `json:"cards"`
	CardsLeft  int        `json:"cards_left"` // 剩余手牌数
	HandType   string     `json:"hand_type"`  // 牌型名称
}

// PlayerPassPayload 不出通知
type PlayerPassPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

// GameOverPayload 游戏结束通知
type GameOverPayload struct {
	WinnerID    string       `json:"winner_id"`
	WinnerName  string       `json:"winner_name"`
	IsLandlord  bool         `json:"is_landlord"`  // 获胜者是否是地主
	PlayerHands []PlayerHand `json:"player_hands"` // 所有玩家剩余手牌
}

// PlayerHand 玩家手牌信息（用于游戏结束展示）
type PlayerHand struct {
	PlayerID   string     `json:"player_id"`
	PlayerName string     `json:"player_name"`
	Cards      []CardInfo `json:"cards"`
}

// ErrorPayload 错误响应
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- 通用数据结构 ---

// PlayerInfo 玩家信息
type PlayerInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Seat       int    `json:"seat"`        // 座位号 0-2
	Ready      bool   `json:"ready"`       // 是否准备
	IsLandlord bool   `json:"is_landlord"` // 是否是地主
	CardsCount int    `json:"cards_count"` // 手牌数量
	Online     bool   `json:"online"`      // 是否在线
}

// CardInfo 牌信息
type CardInfo struct {
	Suit  int `json:"suit"`  // 花色: 0=黑桃, 1=红心, 2=梅花, 3=方块, 4=王
	Rank  int `json:"rank"`  // 点数: 3-17 (3-2, 小王=16, 大王=17)
	Color int `json:"color"` // 颜色: 0=黑, 1=红
}

// --- 错误码 ---
const (
	ErrCodeUnknown      = 1000
	ErrCodeInvalidMsg   = 1001
	ErrCodeRoomNotFound = 2001
	ErrCodeRoomFull     = 2002
	ErrCodeNotInRoom    = 2003
	ErrCodeGameNotStart = 3001
	ErrCodeNotYourTurn  = 3002
	ErrCodeInvalidCards = 3003
	ErrCodeCannotBeat   = 3004
	ErrCodeMustPlay     = 3005
)

// ErrorMessages 错误码对应的消息
var ErrorMessages = map[int]string{
	ErrCodeUnknown:      "未知错误",
	ErrCodeInvalidMsg:   "无效的消息格式",
	ErrCodeRoomNotFound: "房间不存在",
	ErrCodeRoomFull:     "房间已满",
	ErrCodeNotInRoom:    "您不在房间中",
	ErrCodeGameNotStart: "游戏尚未开始",
	ErrCodeNotYourTurn:  "还没轮到您",
	ErrCodeInvalidCards: "无效的牌型",
	ErrCodeCannotBeat:   "您的牌打不过上家",
	ErrCodeMustPlay:     "您必须出牌",
}
