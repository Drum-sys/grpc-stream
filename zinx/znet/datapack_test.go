package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

func TestDataPack_Pack(t *testing.T) {
	/*
	模拟的服务器
	*/

	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err", err)
		return
	}

	// 创建go负责处理客户端业务
	go func() {
		for {
			// 从客户端读数据， 拆包处理
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept error", err)
				return
			}
			go func(conn net.Conn) {
				//处理客户端请求
				// 定义一个拆包对象
				dp := &DataPack{}
				for {
					// 第一次从conn读， 把报的head读出来
					headData := make([]byte, dp.GetDataHeader())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("read head err", err)
						break
					}

					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("server unpack err", err)
						return
					}

					if msgHead.GetMsgDataLen() > 0 {
						// msg有数据， 开始第二次
						// 第二次从conn读， 根据header的datalen读取data
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgDataLen())

						// 再根据datalen长度再次从io流读取
						_, err = io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err", err)
							return
						}

						// 完整的一个消息读取完毕
						fmt.Println("-------》》》Recv MSgID=", msg.ID, ", datalen=", msg.DataLen, ", data=", string(msg.Data))
					}

				}
			}(conn)
		}
	}()


	/*
	模拟客户端
	*/

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client dial err", err)
		return
	}

	// 创建一个封包对象
	dp := NewDataPack()

	// 模拟占包过程， 封装两个msg
	// 封装第一个msg
	msg1 := &Message{
		ID: 1,
		DataLen: 4,
		Data: []byte{'z', 'i', 'n', 'x'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack1 err", err)
		return
	}

	// 封装第二个msg
	msg2 := &Message{
		ID: 2,
		DataLen: 3,
		Data: []byte{'l', 'j', 'w'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack2 err", err)
		return
	}
	sendData1 = append(sendData1, sendData2...)

	conn.Write(sendData1)

	// 阻塞
	select {
	}

}

