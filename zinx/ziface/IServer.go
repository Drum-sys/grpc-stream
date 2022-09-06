package ziface


// 定义一个服务器接口

type IServer interface {

	Start() // 启动服务器

	Stop() // 关闭服务器

	Serve() // 运行服务器

	AddRouter(msgID uint32, router IRouter) //给当前服务注册路由， 供客户端连接使用

	GetConnMgr() IConnManager //获取当前server的链接管理器

	SetOnConnStart(func(connect IConnect))
	SetOnConnStop(func(connect IConnect))
	CallOnConnStart(connect IConnect)
	CallOnConnStop(connect IConnect)


}
