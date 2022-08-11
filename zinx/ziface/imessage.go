package ziface

type IMessage interface {
	GetMsgID() uint32
	GetMsgDataLen() uint32
	GetMsgData() []byte

	SetMsgID(uint32)
	SetMsgDataLen(uint32)
	SetMsgData([]byte)
}
