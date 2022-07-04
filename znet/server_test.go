package znet

import (
	"fmt"
	"net"
	"testing"
	"time"
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
		_, err := conn.Write([]byte("hello ZINX"))
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
		fmt.Printf("server call back:%s,cnt=%d\n", buf[:cnt], cnt)
		time.Sleep(1 * time.Second)
	}
}

//TestServer 服务端模块测试函数
func TestServer(t *testing.T) {
	//1.创建一个server句柄 s
	s := NewServer("[zinx v0.1]")

	//客户端测试
	go ClientTest()

	//2.开启服务
	s.Serve()
}
