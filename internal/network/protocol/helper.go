package protocol

import "encoding/json"

// NewMessage 创建一个新消息
func NewMessage(msgType MessageType, payload any) (*Message, error) {
	var data json.RawMessage
	if payload != nil {
		var err error
		data, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	return &Message{
		Type:    msgType,
		Payload: data,
	}, nil
}

// MustNewMessage 创建消息，失败时 panic
func MustNewMessage(msgType MessageType, payload any) *Message {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		panic(err)
	}
	return msg
}

// Encode 将消息编码为 JSON 字节
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 从 JSON 字节解码消息
func Decode(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ParsePayload 解析消息的 Payload 到指定类型
func ParsePayload[T any](msg *Message) (*T, error) {
	var payload T
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// NewErrorMessage 创建错误消息
func NewErrorMessage(code int) *Message {
	msg, _ := NewMessage(MsgError, ErrorPayload{
		Code:    code,
		Message: ErrorMessages[code],
	})
	return msg
}

// NewErrorMessageWithText 创建带自定义文本的错误消息
func NewErrorMessageWithText(code int, text string) *Message {
	msg, _ := NewMessage(MsgError, ErrorPayload{
		Code:    code,
		Message: text,
	})
	return msg
}
