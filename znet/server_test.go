package znet

import (
	"fmt"
	"io"
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
		//发封包message消息
		dp := NewDataPack()
		msg, _ := dp.Pack(NewMsgPackage(0, []byte("Zinx V0.5 Client Test Message")))
		_, err = conn.Write(msg)
		if err != nil {
			fmt.Println("Write error:", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			//break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}
			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		time.Sleep(1 * time.Second)
	}
}

//PingRouter ping test 自定义路由
type PingRouter struct {
	BaseRouter //一定要先基础BaseRouter
}

//PreHandle Test
func (p *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call PingRouter PreHandle")
	err := request.GetConnection().SendMsg(1, []byte("before ping ...\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//Handle Test
func (p *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
	//回写数据
	err := request.GetConnection().SendMsg(2, []byte("ping...ping...\n"))
	if err != nil {
		fmt.Println(err)
	}
}

//PostHandle Test
func (p *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call PingRouter PostHandle")
	err := request.GetConnection().SendMsg(3, []byte("After ping...\n"))
	if err != nil {
		fmt.Println("call back ping ping ping error")
	}
}

//TestServer 服务端模块测试函数
func TestServer(t *testing.T) {
	//1.创建一个server句柄 s
	s := NewServer()
	s.AddRouter(&PingRouter{})
	//客户端测试
	go ClientTest()

	//2.开启服务
	s.Serve()
}
