package global

import (
	"encoding/json"
	"io/ioutil"

	"github.com/lihao20110/go-zinx/ziface"
)

//ServerObject 存储一切有关Zinx框架的全局参数，供其他模块使用 一些参数也可以通过 用户根据 config/zinx.json来配置
type ServerObject struct {
	TcpServer        ziface.IServer //当前Zinx的全局Server对象
	Host             string         //当前服务器主机IP
	TcpPort          int            //当前服务器主机监听端口号
	Name             string         //当前服务器名称
	Version          string         //当前Zinx版本号
	MaxPacketSize    uint32         //传输数据包的最大值
	MaxConn          int            //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32         //业务工作Worker池的数量
	MaxWorkerTaskLen uint32         //业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32         //读、写两个goroutine之间的消息通信管道大小
	ConfigFilePath   string         //config file path
}

//ServerObj 定义一个全部对象
var ServerObj = &ServerObject{}

//Reload 读取用户配置文件
func (s *ServerObject) Reload() {
	data, err := ioutil.ReadFile("../config/zinx.json")
	if err != nil {
		panic(err)
	}
	//将json数据解析到struct中
	//fmt.Printf("json:%s\n",data)
	err = json.Unmarshal(data, ServerObj)
	if err != nil {
		panic(err)
	}
}

//提供init方法，默认加载
func init() {
	//从配置文件config/zinx.json中加载一些用户配置的参数
	ServerObj.Reload()
}
