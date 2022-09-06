package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

/*
连接模块具体实现
*/

type Connect struct {
	TcpServer ziface.IServer
	Conn *net.TCPConn
	ConnID uint32
	isClosed bool // 连接状态
	ExitChan chan bool // 告知连接是否退出, (由reader告知writer退出的管道）
	MsgHandler ziface.IMsgHandle
	// 无缓冲的管道， 用与读写goroutine之间的消息通信
	msgChan chan []byte

	//设置链接的属性， 使用map， 为保护需要设置读写锁加以保护
	property map[string]interface{}
	propertyLock sync.RWMutex
}

func NewConnect(tcpServer ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connect {
	c := &Connect{
		Conn: conn,
		ConnID: connID,
		MsgHandler: msgHandler,
		isClosed: false,
		msgChan: make(chan []byte),
		ExitChan: make(chan bool, 1),
		TcpServer: tcpServer,
		property: make(map[string]interface{}),
	}

	c.TcpServer.GetConnMgr().Add(c)
	return c
}

func (c *Connect) Stop()  {
	fmt.Println("Conn stop()....ConnID", c.ConnID)

	if c.isClosed == true {
		return
	}

	c.isClosed = true

	// 告知writer关闭
	c.ExitChan <- true

	//调用开发者注册的销毁连接之前需要执行的hook函数
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket连接
	c.Conn.Close()

	//关闭连接
	c.TcpServer.GetConnMgr().Remove(c)

	// 回收资源
	close(c.ExitChan)
	close(c.msgChan)
}

func (c *Connect) Start()  {
	fmt.Println("Conn start()....ConnID", c.ConnID)

	// 启动从当前连接读数据的业务
	go c.StartReader()
	go c.StartWriter()

	//按照开发者传递进来的， 创建连接之后需要调用的的处理业务， 执行对应的hook函数
	c.TcpServer.CallOnConnStart(c)

}

func (c *Connect) GetTCPConnection() *net.TCPConn  {
	return c.Conn
}

func (c *Connect) GetConnID() uint32  {
	return c.ConnID
}

func (c *Connect) RemoteAddr() net.Addr  {
	return c.Conn.RemoteAddr()
}

func (c *Connect) SendMsg(id uint32, data []byte) error  {
	if c.isClosed {
		return errors.New("connect is closed when send msg")
	}

	// 将data封包
	dp := DataPack{}

	msg, err := dp.Pack(NewMsgPack(id, data))
	if err != nil {
		fmt.Println("pack error msg id = ", id)
		return errors.New("pack msg err")
	}

	// 将数据发送给msgChan
	c.msgChan <- msg
	return nil
}


// 读业务的处理
func (c *Connect) StartReader() {
	fmt.Println("Reader Goroutine is running")
	defer fmt.Println("connID=", c.ConnID, "Reader is exit, remote addr is", c.RemoteAddr().String())
	defer c.Stop()

	for {

		// 定义一个拆包对象
		dp := &DataPack{}

		// 1 先读取msg中前8个字节，DataID|DataLen
		headData := make([]byte, dp.GetDataHeader())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg from client err:", err)
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("msg unpack err", err)
			break
		}

		// 2 再根据datalen， 读取实际的内容
		var data []byte
		if msg.GetMsgDataLen() > 0 {
			data = make([]byte, msg.GetMsgDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msd data err", err)
				break
			}
		}
		msg.SetMsgData(data)


		//得到当前conn数据的request请求数据
		req := Request{
			conn: c,
			msg: msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经开启工作池， 将消息发送给工作池处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			// 从路由中， 找到注册绑定的COnn对应的路由
			// 根据绑定的的msgID， 找到对应的路由
			go c.MsgHandler.DoMsgHandler(&req)

		}

	}
}


//写消息gotinue， 专门发送给客户端消息的模块
func (c *Connect) StartWriter() {
	fmt.Println("[Writer goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn write exit]")

	// 不断地阻塞的等待chan的消息， 进行写给客户端
	for  {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("send data err", err)
				return
			}
		case <-c.ExitChan:
			// 代表reader已经退出， writer也要退出
			return
		}
	}
}


//设置连接属性
func (c *Connect) SetConnProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获得属性值
func (c *Connect) GetConnProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}else {
		return nil, errors.New("Get property err")
	}
}

//删除属性值
func (c *Connect) RemoveConnProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

