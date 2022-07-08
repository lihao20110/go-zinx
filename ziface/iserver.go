package ziface

//IServer 定义服务器接口
type IServer interface {
	Start()                                 //启动服务器方法
	Stop()                                  //停止服务器方法
	Serve()                                 //开启业务服务方法
	AddRouter(msgId uint32, router IRouter) //路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用,可以自定一个Router处理业务方法。
	GetConnMgr() IConnManager               //得到连接管理方法
	// SetOnConnStart 给Zinx增添两个链接创建后和断开前时机的回调函数，一般也称作Hook(钩子)函数。
	SetOnConnStart(func(IConnection)) //设置该Server的连接创建时Hook函数
	SetOnConnStop(func(IConnection))  //设置该Server的连接断开时的Hook函数
	CallOnConnStart(conn IConnection) //调用连接OnConnStart Hook函数
	CallOnConnStop(conn IConnection)  //调用连接OnConnStop Hook函数
}
