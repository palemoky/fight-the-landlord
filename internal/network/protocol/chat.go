package protocol

// ChatPayload 聊天消息
type ChatPayload struct {
	SenderID   string `json:"sender_id,omitempty"`   // 发送者 ID (服务端填充)
	SenderName string `json:"sender_name,omitempty"` // 发送者名字 (服务端填充)
	Content    string `json:"content"`               // 消息内容
	Scope      string `json:"scope"`                 // "lobby" or "room"
	Time       int64  `json:"time,omitempty"`        // 发送时间 (服务端填充)
	IsSystem   bool   `json:"is_system,omitempty"`   // 是否是系统消息
}
