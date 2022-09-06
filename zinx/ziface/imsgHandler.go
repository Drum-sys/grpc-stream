package ziface

type IMsgHandle interface {
	// 调度执行对应的Router消息处理方法
	DoMsgHandler(request IRequest)

	//为消息添加具体的处理逻辑
	AddRouter(msgID uint32, router IRouter)

	// 启动worker工作池
	StartWorkerPool()

	// 将消息交给TaskQueue， 由worker处理
	SendMsgToTaskQueue(request IRequest)
}
