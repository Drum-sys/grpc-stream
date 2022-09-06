package utils

import (
	"encoding/json"
	"io/ioutil"
	"zinx/ziface"
)

/*
存储Zinx框架的全局参数， 供其他模块使用
*/

type GlobalObj struct {
	TcpServer ziface.IServer // 当前Zinx全局的Server对象
	Host string
	TcpPort int
	Name string // 服务器名称


	Version string
	MaxConn int // 最大连接数
	MaxPackageSize uint32 // 发送数据包的最大值

	WorkerPoolSize uint32	// 池的大小
	MaxWorkerPoolSize uint32
}

// 定义一个全局对外的GlobalObj
var GlobalObject *GlobalObj

/*
提供一个初始化方法， 配置参数

*/

func init() {
	GlobalObject = &GlobalObj{
		Name: "Zinx Server",
		Version: "V0.6",
		TcpPort: 8999,
		Host: "0.0.0.0",
		MaxConn: 1000,
		MaxPackageSize: 4096,
		WorkerPoolSize: 10,
		MaxWorkerPoolSize: 1024,
	}
	//更新配置文件
	GlobalObject.Reload()
}

// 加载配置文件
func (g *GlobalObj) Reload()  {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}



