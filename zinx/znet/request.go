package znet

import "zinx/ziface"

type Request struct {
	// 已经和客户端建立好的连接
	conn ziface.IConnect

	// 客户端请求的数据
	msg ziface.IMessage
}

func (r *Request) GetConnection() ziface.IConnect {
	return r.conn
}


func (r *Request) GetData() []byte  {
	return r.msg.GetMsgData()
}

func (r *Request) GetMsgID() uint32  {
	return r.msg.GetMsgID()
}

