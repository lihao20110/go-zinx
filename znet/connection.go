package znet

import (
	"fmt"
	"net"

	"github.com/lihao20110/go-zinx/ziface"
)

//Connection 连接模块
type Connection struct {
	Conn     *net.TCPConn   //当前连接的socket TCP套接字
	ConnID   uint32         //连接的ID
	isClosed bool           //当前的连接状态
	Router   ziface.IRouter //该连接的处理方法router
	ExitChan chan bool      //告知当前连接已退出、停止 channel
}

//NewConnection 创建连接的方法
func NewConnection(conn *net.TCPConn, connID uint32, router ziface.IRouter) *Connection {
	return &Connection{
		conn,
		connID,
		false,
		router,
		make(chan bool, 1),
	}
}

//StartReader 连接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID=", c.ConnID, "Reader is exit,remote addr is", c.RemoteAddr().String())
	defer c.Stop()
	for {
		//读取客户端的数据到buf中
		buf := make([]byte, 512)
		_, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("receive buf err", err)
			c.ExitChan <- true
			continue
		}
		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			data: buf,
		}
		//从路由Router中找到注册绑定Conn的对应Handle
		go func(request ziface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&req)
	}
}

//Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start()... ConnID=", c.ConnID)
	//启动从当前连接的读数据业务
	//TODO 启动从当前连接的读数据业务
	go c.StartReader()

	for {
		select {
		case <-c.ExitChan:
			//得到退出消息，不在阻塞
			return
		}
	}
}

//Stop 停止连接，结束当前连接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()... ConnID=", c.ConnID)
	if c.isClosed == true { //如果当前连接已经关闭
		return
	}
	c.isClosed = true
	//TODO Connection Stop() 如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.Conn.Close() //关闭socket连接
	//通知从缓冲队列读数据的业务，该连接已经关闭、
	c.ExitChan <- true
	close(c.ExitChan) //回收资源
}

//GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

//GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

//RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//Send 发送数据给远程客户端
func (c *Connection) Send(data []byte) error {
	return nil
}
