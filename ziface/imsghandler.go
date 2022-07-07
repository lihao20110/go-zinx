package ziface

//IMsgHandle 消息管理抽象层
type IMsgHandle interface {
	DoMsgHandler(request IRequest)          //马上以非阻塞方式处理消息，调用Router中具体Handle()等方法的接口
	AddRouter(msgId uint32, router IRouter) //为消息添加具体的处理逻辑，添加一个msgId和一个路由关系到Apis中
}
