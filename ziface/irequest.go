package ziface

//IRequest 实际上是把客户端请求的连接信息和请求数据封装到Request中
type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetData() []byte            //获取请求消息的数据
}
