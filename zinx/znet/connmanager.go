package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnect
	connLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnect),
	}
}

// 添加链接
func (cm *ConnManager) Add(connect ziface.IConnect) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	cm.connections[connect.GetConnID()] = connect
	fmt.Println("connID = ", connect.GetConnID(), "add to connections len = ", cm.Len())
}
// 删除链接
func (cm *ConnManager) Remove(connect ziface.IConnect) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	delete(cm.connections, connect.GetConnID())
	fmt.Println("connID = ", connect.GetConnID(), "remove from connections len = ", cm.Len())

}
// 获取链接的数量
func (cm *ConnManager) Len() int {
	return len(cm.connections)
}
//清除所有的链接
func (cm *ConnManager) ClearALlConn() {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	for connID, conn := range cm.connections {
		conn.Stop()
		delete(cm.connections, connID)
	}
	fmt.Println("clear all connection len = ", cm.Len())
}
//根据ConnID获取链接
func (cm *ConnManager) Get(connID uint32) (ziface.IConnect, error) {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not FOUND!")
	}
}
