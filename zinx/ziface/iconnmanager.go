package ziface

type IConnManager interface {
	// 添加链接
	Add(connect IConnect)
	// 删除链接
	Remove(connect IConnect)
	// 获取链接的数量
	Len() int
	//清除所有的链接
	ClearALlConn()
	//根据ConnID获取链接
	Get(connID uint32) (IConnect, error)
}
