package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

func main() {
	fmt.Println("client1 start.......")
	time.Sleep(1*time.Second)

	// 1 连接远程服务器得到conn
	conn, err := net.Dial("tcp", "127.0.0.1:8889")
	if err != nil {
		fmt.Println("client start err:", err)
	}
	// 2 连接调用write方法写数据
	for  {
		dp := znet.DataPack{}

		binaryMsg, err := dp.Pack(znet.NewMsgPack(1, []byte("ZinxV0.6 client1 Test Message")))
		if err != nil {
			fmt.Println("pack err", err)
			return
		}

		_ ,err = conn.Write(binaryMsg)
		if err != nil {
			fmt.Println("write err", err)
			return
		}

		// 服务器给我们回复一个message MsgId：1 pingping

		// 先读取流中的head部分， 得到ID 和 DataLen’
		binaryHead := make([]byte, dp.GetDataHeader())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head err", err)
			break
		}

		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack msgHead errr", err)
			break
		}

		if msgHead.GetMsgDataLen() > 0 {
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgDataLen())

			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data err", err)
				return
			}

			// 完整的一个消息读取完毕
			fmt.Println("-------》》》Recv MSgID=", msg.ID, ", datalen=", msg.DataLen, ", data=", string(msg.Data))
		}



		// cpu 阻塞
		time.Sleep(1*time.Second)
	}
}
