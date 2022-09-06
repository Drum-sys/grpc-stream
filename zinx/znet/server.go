package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// Iserver 的接口实现， 定义一个Server的服务器模块

type Server struct {
	Name string // 服务器名称
	IPVersion string // 服务器绑定的ip版本
	IP string // 服务器监听的ip
	Port int// 服务器监听的端口
	//Router ziface.IRouter// 给当前的server添加一个Router, server注册的连接对应的处理业务

	// server 的消息管理模块， 用来绑定msgID和对应的业务处理关系
	MsgHandler ziface.IMsgHandle
	ConnMg ziface.IConnManager

	//该server创建链接之后调用Hook函数
	OnConnStart func(connect ziface.IConnect)
	//该server销毁链接之后调用Hook函数
	OnConnStop func(connect ziface.IConnect)
}

func NewServer(name string) ziface.IServer  {
	s := &Server{
		Name: utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP: utils.GlobalObject.Host,
		Port: utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandler(),
		ConnMg: NewConnManager(),
	}
	return s
}


func (s *Server) Start()  {
	fmt.Printf("[zinx server name : %s, listener at ip : %s, Port: %d]\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	//fmt.Printf("Start Server Listener at IP :%s, Port: %d, is starting\n", s.IP, s.Port)

	// 开启消息队列，以及工作池
	s.MsgHandler.StartWorkerPool()

	// 1 获取一个TCP addr
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("resolve tcp addr err", err)
		return
	}
	// 监听服务器地址
	listener, err := net.ListenTCP(s.IPVersion, addr)
	if err != nil {
		fmt.Println("listening", s.IPVersion, "err", err)
		return
	}

	fmt.Println("start Zinx server succ", s.Name)

	var cid uint32
	cid = 0

	// 阻塞的等待客户端连接， 处理客户端连接业务
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("Accept err", err)
			continue
		}

		if s.ConnMg.Len() > utils.GlobalObject.MaxConn {
			conn.Close()
			continue
		}

		// 将处理新连接业务的方法和 conn 绑定， 得到新的conn
		dealConn := NewConnect(s, conn, cid, s.MsgHandler)
		cid++

		//启动当前业务处理
		go dealConn.Start()

	}

}

func (s *Server) Stop()  {

	// 将服务器的资源， 状态回收
	s.ConnMg.ClearALlConn()
	fmt.Println("Stop Zinx server name", s.Name)

}

func (s *Server) Serve()  {
	go s.Start()

	//TODO 做启动服务器之后的业务

	//select 阻塞状态
	select {

	}
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter)  {
	s.MsgHandler.AddRouter(msgID, router)
}


func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMg
}

func (s *Server) SetOnConnStart(hookFunc func(connect ziface.IConnect)) {
	s.OnConnStart = hookFunc
}
func (s *Server) SetOnConnStop(hookFunc func(connect ziface.IConnect)) {
	s.OnConnStop = hookFunc
}
func (s *Server) CallOnConnStart(connect ziface.IConnect){
	if s.OnConnStart != nil {
		fmt.Println("-----<<< Call OnCOnnStart")
		s.OnConnStart(connect)
	}
}
func (s *Server) CallOnConnStop(connect ziface.IConnect){
	if s.OnConnStop != nil {
		fmt.Println("-----<<< Call OnCOnnStop")
		s.OnConnStop(connect)
	}
}
