package mr

import (
	"fmt"
	"log"
)
import "net"
import "net/rpc"
import "net/http"

type Task struct {
	MapId int
	ReduceId int
	Filename string

}

type Coordinator struct {
	// Your definitions here.
	State int // 0 表示map， 1表示reduce， 2表示finish
	NumMapTask int
	NumReduceTask int
	MapTask chan Task
	ReduceTask chan Task
	MapTaskFin chan bool
	ReduceTaskFin chan bool
}

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetMapTask(args *MapTaskRequest, reply *MapTaskResponse) error {
	if c.State == 0 {
		mapTask, ok := <- c.MapTask
		if ok {
			reply.MapTask = mapTask
		}
	}
	reply.State = c.State
	reply.NumMapTask = c.NumMapTask
	reply.NumReduceTask = c.NumReduceTask
	return nil
}

func (c *Coordinator) GetMapFinTask(args *MapTaskRequest, reply *MapTaskResponse) error {
	c.MapTaskFin <- true
	if len(c.MapTaskFin) == c.NumMapTask {
		fmt.Println("mapTask has finished")
		c.State = 1
	}
	reply.State = c.State
	return nil
}

func (c *Coordinator) GetReduceTask(args *ReduceTaskRequest, reply *ReduceTaskResponse) error {
	if c.State == 1 {
		reduceTask, ok := <- c.ReduceTask
		if ok {
			reply.ReduceTask = reduceTask
		}
	}
	reply.State = c.State
	reply.NumMapTask = c.NumMapTask
	reply.NumReduceTask = c.NumReduceTask
	return nil
}

func (c *Coordinator) GetReduceFinTask(args *ReduceTaskRequest, reply *ReduceTaskResponse) error {
	c.ReduceTaskFin <- true
	if len(c.ReduceTaskFin) == c.NumReduceTask {
		fmt.Println("reduceTask has finished")
		c.State = 2
	}
	reply.State = c.State
	return nil
}
//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	//sockname := coordinatorSock()
	//os.Remove(sockname)
	//l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	if c.State == 2 {
		ret = true
	}


	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		State: 0,
		NumReduceTask: nReduce,
		NumMapTask: len(files),
		MapTask: make(chan Task, len(files)),
		ReduceTask: make(chan Task, nReduce),
		MapTaskFin: make(chan bool, len(files)),
		ReduceTaskFin: make(chan bool, nReduce),
	}

	// 初始化Task, 将任务分发到worker中Map chan，等待worke取task
	for fileId, filename := range files {
		c.MapTask <- Task{
			MapId: fileId,
			Filename: filename,
		}
	}

	for i := 0 ; i < nReduce; i++ {
		c.ReduceTask <- Task{
			ReduceId: i,
		}
	}


	// Your code here.
	c.server()
	return &c
}
