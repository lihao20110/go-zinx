package znet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/lihao20110/go-zinx/ziface"
)

//ConnManager 连接管理模块:用一个map来承载全部的连接信息，key是连接ID，value则是连接本身。其中有一个读写锁connLock主要是针对map做多任务修改时的保护作用。
type ConnManager struct {
	connections map[uint32]ziface.IConnection //管理的连接信息
	connLock    sync.RWMutex                  //读写连接的读写锁
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

//Len 获取当前连接数量
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

//Add 添加连接
func (connMgr *ConnManager) Add(conn ziface.IConnection) {
	//保护共享资源Map 加读写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	//将conn来连接添加到ConnManager中
	connMgr.connections[conn.GetConnID()] = conn
	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}

//Remove 删除连接,只是单纯的将conn从map中摘掉
func (connMgr *ConnManager) Remove(conn ziface.IConnection) {
	//保护共享资源Map，加读写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())
	fmt.Println("connection Remove ConnID=", conn.GetConnID(), "successfully:conn num = ", connMgr.Len())
}

//Get 利用ConnID获取连接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	//保护共享资源Map,加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

//ClearConn 停止并删除全部的连接信息：先停止连接业务，c.Stop()，然后再从map中摘除。
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源Map，加读写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		conn.Stop()                         //停止
		delete(connMgr.connections, connID) //删除
	}
	fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
}
