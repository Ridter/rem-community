package message

import (
	"fmt"
	"github.com/chainreactors/rem/x/utils"
	"google.golang.org/protobuf/proto"
)

// MsgType 定义消息类型
type MsgType int8

// 定义消息状态常量
const (
	StatusFailed  = 0
	StatusSuccess = 1
)

// 定义消息类型常量
const (
	LoginMsg MsgType = iota + 1
	AckMsg
	ControlMsg
	PingMsg
	PongMsg
	PacketMsg
	ConnStartMsg
	ConnEndMsg
	RedirectMsg
	End
)

// 定义标准错误类型
var (
	ErrEmptyMessage    = fmt.Errorf("empty message")
	ErrInvalidType     = fmt.Errorf("invalid message type")
	ErrUnknownType     = fmt.Errorf("unknown message type")
	ErrMarshal         = fmt.Errorf("marshal error")
	ErrUnmarshal       = fmt.Errorf("unmarshal error")
	ErrTypeMismatch    = fmt.Errorf("message type mismatch")
	ErrMessageLength   = fmt.Errorf("message length error")
	ErrInvalidStatus   = fmt.Errorf("invalid message status")
	ErrConnectionError = fmt.Errorf("connection error")
)

// msgRegistry 存储消息类型到消息创建函数的映射
var msgRegistry = map[MsgType]func() proto.Message{
	LoginMsg:     func() proto.Message { return &Login{} },
	AckMsg:       func() proto.Message { return &Ack{} },
	ControlMsg:   func() proto.Message { return &Control{} },
	PingMsg:      func() proto.Message { return &Ping{} },
	PongMsg:      func() proto.Message { return &Pong{} },
	PacketMsg:    func() proto.Message { return &Packet{} },
	ConnStartMsg: func() proto.Message { return &ConnStart{} },
	ConnEndMsg: func() proto.Message {
		return &ConnEnd{}
	},
	RedirectMsg: func() proto.Message { return &Redirect{} },
}

// NewMessage 根据消息类型创建新的消息实例
func NewMessage(msgType MsgType) proto.Message {
	if creator, ok := msgRegistry[msgType]; ok {
		return creator()
	}
	return nil
}

// ValidateMessageType 验证消息类型是否有效
func ValidateMessageType(msgType MsgType) bool {
	return msgType > 0 && msgType < End
}

// WrapError 包装错误信息
func WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: "+format, append([]interface{}{err}, args...)...)
}

// GetMessageType 根据消息实例获取消息类型
func GetMessageType(msg proto.Message) MsgType {
	if msg == nil {
		return 0
	}

	switch msg.(type) {
	case *Login:
		return LoginMsg
	case *Ack:
		return AckMsg
	case *Control:
		return ControlMsg
	case *Ping:
		return PingMsg
	case *Pong:
		return PongMsg
	case *Packet:
		return PacketMsg
	case *ConnStart:
		return ConnStartMsg
	case *ConnEnd:
		return ConnEndMsg
	case *Redirect:
		return RedirectMsg
	default:
		return 0
	}
}

func Wrap(src, dst string, m proto.Message) *Redirect {
	msg := &Redirect{
		Source:      src,
		Destination: dst,
	}
	switch m.(type) {
	case *Packet:
		msg.Msg = &Redirect_Packet{Packet: m.(*Packet)}
	case *ConnStart:
		msg.Msg = &Redirect_Start{Start: m.(*ConnStart)}
	case *ConnEnd:
		msg.Msg = &Redirect_End{End: m.(*ConnEnd)}
	default:
		utils.Log.Error(ErrInvalidType)
	}
	return msg
}

func Unwrap(m *Redirect) proto.Message {
	var msg proto.Message
	switch m.GetMsg().(type) {
	case *Redirect_Packet:
		msg = m.GetPacket()
	case *Redirect_Start:
		msg = m.GetStart()
	case *Redirect_End:
		msg = m.GetEnd()
	default:
		utils.Log.Error(ErrInvalidType)
	}
	return msg
}
