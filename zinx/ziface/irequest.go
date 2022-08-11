package ziface

/*
实际上是把客户端请求的连接信息 和 请求数据包装到一个request中
*/

type IRequest interface {
	// 获取当前连接
	GetConnection() IConnect

	// 得到请求消息的数据
	GetData() []byte

	// 得到请求消息的ID
	GetMsgID() uint32
}
