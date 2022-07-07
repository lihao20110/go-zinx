package znet

import (
	"fmt"
	"strconv"

	"github.com/lihao20110/go-zinx/ziface"
)

type MsgHandle struct {
	Apis map[uint32]ziface.IRouter //存放每个MsgId 所对应的处理方法的map属性
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis: make(map[uint32]ziface.IRouter),
	}
}

//DoMsgHandler 马上以非阻塞方式处理消息,调用Router中具体Handle()等方法的接口
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), "is not found!")
		return
	}
	//执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

//AddRouter 添加一个msgId和一个路由关系到Apis中
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//判断当前msg绑定的API方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeat api,msgId = " + strconv.Itoa(int(msgId)))
	}
	//添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}
