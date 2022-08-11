package main

import "zinx/znet"

func main()  {
	// 创建一个Server
	s := znet.NewServer("ZinxV0.1")

	// 启动服务器， 监听端口， 等待客户端连接
	s.Serve()
}
