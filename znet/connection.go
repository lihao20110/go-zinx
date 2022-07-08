package znet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/lihao20110/go-zinx/global"
	"github.com/lihao20110/go-zinx/ziface"
)

//Connection 连接模块
type Connection struct {
	//当前Conn属于哪个Server:Server和Connection建立能够互相索引的关系
	TcpServer    ziface.IServer    //当前conn属于哪个server，在conn初始化的时候添加即可
	Conn         *net.TCPConn      //当前连接的socket TCP套接字
	ConnID       uint32            //连接的ID
	isClosed     bool              //当前的连接状态
	MsgHandler   ziface.IMsgHandle //消息管理MsgId和对应处理方法的消息管理模块
	ExitBuffChan chan bool         //告知当前连接已退出、停止 channel
	msgBuffChan  chan []byte       //有缓冲管道，用于读、写两个goroutine之间的消息通信
}

//NewConnection 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		server,
		conn,
		connID,
		false,
		msgHandler,
		make(chan bool, 1),
		make(chan []byte, global.ServerObj.MaxMsgChanLen), //有缓冲管道，用于读、写两个goroutine之间的消息通信
	}
	//将新创建的Conn添加到连接管理中
	c.TcpServer.GetConnMgr().Add(c) //将当前新创建的连接添加到ConnManager中
	return c
}

//StartWriter 写消息goroutine，用户将数据发送给客户端
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit")
	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if ok { //有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:", err, "conn Writer exit")
					return
				}
			} else {
				break
				fmt.Println("msgBuffChan is Closed")
			}
		case <-c.ExitBuffChan:
			//conn 已经关闭
			return
		}
	}
}

//StartReader 连接的读消息业务方法
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
			c.ExitBuffChan <- true
			continue
		}
		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.ExitBuffChan <- true
			continue
		}
		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)
		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}
		//判断用户配置WorkerPoolSize的个数，如果大于0，那么我就启动多任务机制处理链接请求消息，如果=0或者<0那么，我们依然只是之前的开启一个临时的Goroutine处理客户端请求消息。
		if global.ServerObj.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

//Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start()... ConnID=", c.ConnID)
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)
	for {
		select {
		case <-c.ExitBuffChan:
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

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)

	c.Conn.Close() //关闭socket连接

	c.ExitBuffChan <- true //关闭Writer Goroutine
	//将连接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c) //删除conn从ConnManager中

	//关闭该连接全部管道,回收资源
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
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

//SendBuffMsg 发送客户端的数据改为发送至msgBuffChan,交给写goroutine执行发送
func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
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
	c.msgBuffChan <- msg //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取
	return nil
}
