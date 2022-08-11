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

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handle")
	//1 先读取客户端的数据， 在回写
	fmt.Println("recv from client: msgId=", request.GetMsgID(),
	"data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(200, []byte("ping ...  ping ...."))
	if err != nil {
		fmt.Println(err)
	}
}

type HelloZinxRouter struct {
	znet.BaseRouter
}

// Test Handle
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handle")
	//1 先读取客户端的数据， 在回写
	fmt.Println("recv from client: msgId=", request.GetMsgID(),
		"data = ", string(request.GetData()))
	err := request.GetConnection().SendMsg(201, []byte("Welcome to Zinx"))
	if err != nil {
		fmt.Println(err)
	}
}


func main()  {
	// 创建一个Server
	s := znet.NewServer("ZinxV0.1")

	// 给当前zinx框架添加自定义的router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	// 启动服务器， 监听端口， 等待客户端连接
	s.Serve()
}
