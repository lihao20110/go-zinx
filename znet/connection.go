package znet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/lihao20110/go-zinx/ziface"
)

//Connection 连接模块
type Connection struct {
	Conn       *net.TCPConn      //当前连接的socket TCP套接字
	ConnID     uint32            //连接的ID
	isClosed   bool              //当前的连接状态
	MsgHandler ziface.IMsgHandle //消息管理MsgId和对应处理方法的消息管理模块
	ExitChan   chan bool         //告知当前连接已退出、停止 channel
}

//NewConnection 创建连接的方法
func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	return &Connection{
		conn,
		connID,
		false,
		msgHandler,
		make(chan bool, 1),
	}
}

//StartReader 连接的读业务方法
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID=", c.ConnID, "Reader is exit,remote addr is", c.RemoteAddr().String())
	defer c.Stop()
	for {
		// 创建拆包解包的对象
		dp := NewDataPack()
		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.ExitChan <- true
			continue
		}
		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitChan <- true
			continue
		}
		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitChan <- true
				continue
			}
		}
		msg.SetData(data)
		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		//从绑定好的消息和对应的处理方法中执行对应的Handle方法
		go c.MsgHandler.DoMsgHandler(&req)
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

//SendMsg 发送数据给远程客户端,提供一个封包的接口，供Zinx发包使用。
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg ")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}
	//写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error ")
		c.ExitChan <- true
		return errors.New("conn Write error")
	}
	return nil
}
