package znet

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/lihao20110/go-zinx/ziface"
)

//ClientTest 模拟客户端
func ClientTest() {
	fmt.Println("Client Test ... Start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client Start err,exit!")
		return
	}
	for {
		_, err := conn.Write([]byte("Zinx v0.3"))
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}
		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Read buf err:", err)
			return
		}
		fmt.Printf("server call back:%sback cnt=%d\n", buf[:cnt], cnt)
		time.Sleep(1 * time.Second)
	}
}

//PingRouter ping test 自定义路由
type PingRouter struct {
	BaseRouter //一定要先基础BaseRouter
}

//PreHandle Test
func (p *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ...\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Handle Test
func (p *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("ping...ping...\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//PostHandle Test
func (p *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call PingRouter PostHandle")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("After ping...\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//TestServer 服务端模块测试函数
func TestServer(t *testing.T) {
	//1.创建一个server句柄 s
	s := NewServer("[zinx v0.3]")
	s.AddRouter(&PingRouter{})
	//客户端测试
	go ClientTest()

	//2.开启服务
	s.Serve()
}
