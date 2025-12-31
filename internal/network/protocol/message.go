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

	// 排行榜
	MsgGetStats             MessageType = "get_stats"              // 获取个人统计
	MsgGetLeaderboard       MessageType = "get_leaderboard"        // 获取排行榜
	MsgGetRoomList          MessageType = "get_room_list"          // 获取房间列表
	MsgGetOnlineCount       MessageType = "get_online_count"       // 获取在线人数
	MsgGetMaintenanceStatus MessageType = "get_maintenance_status" // 获取维护状态
	MsgChat                 MessageType = "chat"                   // 聊天消息
)

// 服务端 → 客户端 消息类型
const (
	// 连接相关
	MsgConnected     MessageType = "connected"      // 连接成功
	MsgReconnected   MessageType = "reconnected"    // 重连成功
	MsgPong          MessageType = "pong"           // 心跳 pong
	MsgPlayerOffline MessageType = "player_offline" // 玩家掉线通知
	MsgPlayerOnline  MessageType = "player_online"  // 玩家上线通知
	MsgOnlineCount   MessageType = "online_count"   // 在线人数更新

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

	// 排行榜
	MsgStatsResult       MessageType = "stats_result"       // 个人统计结果
	MsgLeaderboardResult MessageType = "leaderboard_result" // 排行榜结果
	MsgRoomListResult    MessageType = "room_list_result"   // 房间列表结果

	// 系统通知
	MsgMaintenancePush MessageType = "maintenance_push" // 主动推送
	MsgMaintenancePull MessageType = "maintenance_pull" // 被动拉取

	// 错误
	MsgError MessageType = "error" // 错误消息
)
