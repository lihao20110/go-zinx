package znet

import (
	"fmt"
	"strconv"

	"github.com/lihao20110/go-zinx/global"
	"github.com/lihao20110/go-zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                    //业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: global.ServerObj.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, global.ServerObj.WorkerPoolSize), //一个worker对应一个queue
	}
}

//StartOneWorker 启动一个Worker工作流程,每个worker是不会退出的(目前没有设定worker的停止工作机制)，会永久的从对应的TaskQueue中等待消息，并处理。
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker Id = ", workerID, "is started ")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request,并执行绑定业务的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

//StartWorkerPool 启动worker工作池,根据用户配置好的WorkerPoolSize的数量来启动，然后分别给每个Worker分配一个TaskQueue，然后用一个goroutine来承载一个Worker的工作业务。
func (mh *MsgHandle) StartWorkerPool() {
	//遍历需要启动worker的数量，依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, global.ServerObj.MaxWorkerTaskLen)
		//启动当前worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

//SendMsgToTaskQueue 将消息交给TaskQueue，由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该有那个worker负责处理
	//轮询的平均分配规则，一个简单的求模运算。用余数和workerID的匹配来进行分配。
	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(),
		" request msgID=", request.GetMsgID(), "to workerID=", workerID)
	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
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
