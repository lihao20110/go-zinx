package ziface

import (
	"net"
)

//IConnection 定义连接模块的抽象层
type IConnection interface {
	Start()                                  //启动连接，让当前的连接准备开始工作
	Stop()                                   //停止连接，结束当前连接的工作
	GetTCPConnection() *net.TCPConn          //获取当前连接的绑定socket conn
	GetConnID() uint32                       //获取当前连接模块的连接ID
	RemoteAddr() net.Addr                    //获取远程客户端的TCP状态 IP port
	SendMsg(msgId uint32, data []byte) error //直接将Message数据发送数据给远程的TCP客户端
}

//HandleFunc 定义一个处理连接业务的方法,函数类型
type HandleFunc func(*net.TCPConn, []byte, int) error
