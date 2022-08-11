package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

/*
	消息处理模块实现
*/

type MsgHandler struct {

	// 存放每个MsgId对应的处理方法
	Apis map[uint32]ziface.IRouter
	// 负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//负责业务工作workers池的worker数量
	WorkerPoolSize uint32
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis: make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

func (mh *MsgHandler) DoMsgHandler(request ziface.IRequest)  {
	// 从request中找到msgID
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID=", request.GetMsgID(), "is NOT FOUND! NEED REGISTER!")
	}

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)

}

func (mh *MsgHandler) AddRouter(msgID uint32, router ziface.IRouter)  {
	// 当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		// id已经注册
		panic("repeat api, msgID=" + strconv.Itoa(int(msgID)))
	}
	mh.Apis[msgID] = router
	fmt.Println("add api MsgID=", msgID, "success!")
}

// 启动一个worker工作池
func (mh *MsgHandler) StartWorkerPool() {
	// 每个worker使用go
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 当前的worker对应的channel消息队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerPoolSize)
		// 启动当前的worker， 阻塞的等待消息从channel传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个worker工作流程
func (mh *MsgHandler) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("worker ID = ", workerID, "is started")

	// 不断阻塞的等待消息队列的消息
	for  {
		select {
		//如果有消息过来，客户端的一个request， 执行request绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}


// 将消息交给TaskQueue， 由worker处理
func (mh *MsgHandler) SendMsgToTaskQueue(request ziface.IRequest) {
	// 将消息平均分配worker
	// 更具客户端建立的ConnId进行分配
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("add ConnId = ", request.GetConnection().GetConnID(), "request MsgID = ", request.GetMsgID(), "to WorkerID = ", workerID)

	// 将消息发送给worker对应的TaskQueue
	mh.TaskQueue[workerID] <- request
}
