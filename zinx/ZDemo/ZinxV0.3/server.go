package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

// ping test 自定义路由

type PingRouter struct {
	znet.BaseRouter
}

// Test PreHandle
func (this *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping"))
	if err != nil {
		fmt.Println("call back before ping err")
	}
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping ........"))
	if err != nil {
		fmt.Println("call back  ping err")
	}
}

// Test PostHandle
func (this *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call Router Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping"))
	if err != nil {
		fmt.Println("call back  after ping err")
	}
}

func main()  {
	// 创建一个Server
	s := znet.NewServer("ZinxV0.1")

	// 给当前zinx框架添加自定义的router
	s.AddRouter(&PingRouter{})

	// 启动服务器， 监听端口， 等待客户端连接
	s.Serve()
}
