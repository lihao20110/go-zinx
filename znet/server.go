package znet

import (
	"fmt"
	"net"
	"time"

	"github.com/lihao20110/go-zinx/ziface"
)

// Server IServer 接口实现，定义一个Server服务类
type Server struct {
	Name      string //服务器的名称
	IPVersion string //tcp4 or other
	IP        string //服务器绑定的地址
	Port      int    //服务器绑定的端口
}

//NewServer 创造一个服务器句柄
func NewServer(name string) ziface.IServer {
	return &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "0.0.0.0",
		Port:      7777,
	}
}

// ==实现ziface.IServer里的全部接口方法==

// Start 启动服务器
func (s *Server) Start() {
	fmt.Printf("[START] Server listener at %s:%d is starting\n", s.IP, s.Port)
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

		//3.启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}
			//3.2 TODO Server.Start() 设置服务器最大连接控制，如果超过最大连接，那么关闭此新的连接

			//3.3 TODO Server.Start() 处理该连接请求的业务方法，此时应该有handler和conn绑定
			//这里暂时做一个最大512字节的回显服务
			go func() {
				for { //不断的循环从客户端获取数据
					buf := make([]byte, 512)
					cnt, err := conn.Read(buf)
					if err != nil {
						fmt.Println("receive buf err", err)
						continue
					}
					//回显
					if _, err := conn.Write(buf[:cnt]); err != nil {
						fmt.Println("Write back buf err", err)
						continue
					}
				}
			}()
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
