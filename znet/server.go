package znet

import (
	"fmt"
	"net"
	"time"

	"github.com/lihao20110/go-zinx/global"
	"github.com/lihao20110/go-zinx/ziface"
)

// Server IServer 接口实现，定义一个Server服务类
type Server struct {
	Name       string              //服务器的名称
	IPVersion  string              //tcp4 or other
	IP         string              //服务器绑定的地址
	Port       int                 //服务器绑定的端口
	msgHandler ziface.IMsgHandle   //当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	ConnMgr    ziface.IConnManager //当前Server的连接管理器
	//新增两个hook函数原型
	OnConnStart func(conn ziface.IConnection) //该Server的连接创建时Hook函数
	OnConnStop  func(conn ziface.IConnection) //该Server的连接断开时的Hook函数
}

//NewServer 创造一个服务器句柄
func NewServer() ziface.IServer {
	//请先初始化全局配置文件config/zinx.json
	global.ServerObj.Reload()
	s := &Server{
		Name:       global.ServerObj.Name,
		IPVersion:  "tcp4",
		IP:         global.ServerObj.Host,
		Port:       global.ServerObj.TcpPort,
		msgHandler: NewMsgHandle(),   //msgHandler 初始化
		ConnMgr:    NewConnManager(), //创建ConnManager
	}
	global.ServerObj.TcpServer = s
	return s
}

// Start 启动服务器
func (s *Server) Start() {
	fmt.Printf("[START] Server listener at %s:%d is starting\n", s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		global.ServerObj.Version,
		global.ServerObj.MaxConn,
		global.ServerObj.MaxPacketSize)
	//开启一个goroutine去做服务端的Listen监听服务
	go func() {
		//0.启动worker工作池机制
		s.msgHandler.StartWorkerPool()
		//1.获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err:", err)
			return
		}
		//2.监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}
		fmt.Println("start Zinx server", s.Name, "success, now listening...") //已经监听成功
		var cid uint32
		cid = 0
		//3.启动server网络连接业务,阻塞等待客户端建立连接请求,处理客户端连接业务(读写)
		for {
			//3.1阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//3.2 设置服务器最大连接控制，如果超过最大连接数，那么关闭此新的连接
			if s.ConnMgr.Len() >= global.ServerObj.MaxConn {
				conn.Close()
				continue
			}
			//3.3处理该连接请求的业务方法，handler和conn绑定
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++
			//3.4启动当前连接的业务处理
			go dealConn.Start()
		}
	}()
}

//Stop 停止服务器
func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server,name", s.Name)
	//将其他需要清理的连接信息或其他信息，也要一并停止或清理
	s.ConnMgr.ClearConn()
}

//Serve 开启业务服务
func (s *Server) Serve() {
	s.Start()
	//TODO Server.Serve() 是否在启动服务的时候，还要处理其他的事情，可以在这里添加
	//阻塞，否则主Go退出，listener的go将会退出
	for {
		time.Sleep(10 * time.Second)
	}
}

//AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
	fmt.Println("Add Router success!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

//SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

//SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

//CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

//CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
