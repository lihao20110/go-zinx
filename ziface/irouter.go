package ziface

//IRouter 路由接口，这里面路由是使用框架这给该连接自定的处理业务方法；路由里的IRequest 则包含该连接的连接信息和请求数据信息
type IRouter interface {
	PreHandle(request IRequest)  //在处理conn业务之前的钩子方法,有前置业务，可以重写这个方法
	Handle(request IRequest)     //处理conn业务的方法,是处理当前链接的主业务函数
	PostHandle(request IRequest) //处理conn业务之后的钩子方法,有后置业务，可以重写这个方法
}
