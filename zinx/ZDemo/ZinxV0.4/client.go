package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("client start.......")
	time.Sleep(1*time.Second)

	// 1 连接远程服务器得到conn
	conn, err := net.Dial("tcp", "127.0.0.1:8889")
	if err != nil {
		fmt.Println("client start err:", err)
	}
	// 2 连接调用write方法写数据
	for  {
		_, err = conn.Write([]byte("Hello zinx v0.1"))
		if err != nil {
			fmt.Println("write conn err", err)
			return
		}

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)

		if err != nil {
			fmt.Println("read buf err", err)
			return
		}

		fmt.Printf("server call back: %s, cnt=%d\n", buf, cnt)

		// cpu 阻塞
		time.Sleep(1*time.Second)
	}
}
