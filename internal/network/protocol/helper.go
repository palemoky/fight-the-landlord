package protocol

import (
	"google.golang.org/protobuf/proto"
)

// NewMessage 创建一个新消息
// 注意: 使用完毕后应调用 PutMessage 归还对象到池
func NewMessage(msgType MessageType, payload any) (*Message, error) {
	msg := GetMessage()
	msg.Type = msgType

	if payload != nil {
		var err error
		// 使用 Protobuf 编码 payload
		msg.Payload, err = EncodePayload(msgType, payload)
		if err != nil {
			PutMessage(msg) // 失败时归还
			return nil, err
		}
	}
	return msg, nil
}

// MustNewMessage 创建消息，失败时 panic
func MustNewMessage(msgType MessageType, payload any) *Message {
	msg, err := NewMessage(msgType, payload)
	if err != nil {
		panic(err)
	}
	return msg
}

// Encode 将消息编码为 Protobuf 字节
func (m *Message) Encode() ([]byte, error) {
	pbMsg := GetPBMessage()
	defer PutPBMessage(pbMsg)

	pbMsg.Type = stringToProtoMessageType(string(m.Type))
	pbMsg.Payload = m.Payload // Protobuf payload

	return proto.Marshal(pbMsg)
}

// Decode 从 Protobuf 字节解码消息
// 注意: 使用完毕后应调用 PutMessage 归还对象到池
func Decode(data []byte) (*Message, error) {
	pbMsg := GetPBMessage()
	defer PutPBMessage(pbMsg)

	if err := proto.Unmarshal(data, pbMsg); err != nil {
		return nil, err
	}

	msg := GetMessage()
	msg.Type = MessageType(protoMessageTypeToString(pbMsg.Type))
	msg.Payload = append([]byte(nil), pbMsg.Payload...) // 复制 payload 避免引用

	return msg, nil
}

// ParsePayload 解析消息的 Payload 到指定类型
func ParsePayload[T any](msg *Message) (*T, error) {
	var payload T
	if err := DecodePayload(msg.Type, msg.Payload, &payload); err != nil {
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
