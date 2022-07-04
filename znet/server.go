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
	Name      string         //服务器的名称
	IPVersion string         //tcp4 or other
	IP        string         //服务器绑定的地址
	Port      int            //服务器绑定的端口
	Router    ziface.IRouter //当前Server由用户绑定的回调router,也就是Server注册的链接对应的处理业务
}

//NewServer 创造一个服务器句柄
func NewServer(name string) ziface.IServer {
	//先初始化全局配置文件
	global.ServerObj.Reload()
	return &Server{
		Name:      global.ServerObj.Name,
		IPVersion: "tcp4",
		IP:        global.ServerObj.Host,
		Port:      global.ServerObj.TcpPort,
		Router:    nil,
	}
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
			//阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//TODO Server.Start() 设置服务器最大连接控制，如果超过最大连接，那么关闭此新的连接
			//处理该连接请求的业务方法，handler和conn绑定
			dealConn := NewConnection(conn, cid, s.Router)
			cid++
			//启动当前业务的连接业务处理
			go dealConn.Start()
		}
	}()
}

//Stop 停止服务器
func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server,name", s.Name)
	//TODO Server.Stop() 将其他需要清理的连接信息或其他信息，也要一并停止或清理
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
func (s *Server) AddRouter(router ziface.IRouter) {
	s.Router = router
	fmt.Println("Add Router success!")
}
