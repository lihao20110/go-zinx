package ziface

//IServer 定义服务器接口
type IServer interface {
	Start()                   //启动服务器方法
	Stop()                    //停止服务器方法
	Serve()                   //开启业务服务方法
	AddRouter(router IRouter) //路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用,可以自定一个Router处理业务方法。
}
