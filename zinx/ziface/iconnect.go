package ziface

import (
	"net"
)

type IConnect interface {

	// 启动链接， 让当前连接开始工作
	Start()

	// 停止连接， 结束工作
	Stop()

	// 获取当前连接的绑定socket conn
	GetTCPConnection() *net.TCPConn

	// 获取当前连接模块的连接ID
	GetConnID() uint32

	// 获取远程客户端的IP Port
	RemoteAddr() net.Addr

	// 发送数据， 将数据发送给远程的客户端
	SendMsg(uint32, []byte) error

	//设置连接属性
	SetConnProperty(key string, value interface{})

	//获得属性值
	GetConnProperty(key string) (interface{}, error)

	//删除属性值
	RemoveConnProperty(key string)
}

// 定义一个处理连接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error