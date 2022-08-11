package mr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)
import "log"
import "net/rpc"
import "hash/fnv"

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the coordinator.
	// Map 实现过程， 通过rpc调用获取到内容保存在临时文件当中
	req := MapTaskRequest{}
	rsp := MapTaskResponse{}

	req1 := ReduceTaskRequest{}
	rsp1 := ReduceTaskResponse{}

	for  {
		CallGetMapTask(&req, &rsp)
		task := rsp.MapTask
		fmt.Println("Start map task .........")
		if rsp.State == 0 {

			nReduce := rsp.NumReduceTask
			filename := task.Filename
			id := strconv.Itoa(task.MapId)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()
			kva := mapf(filename, string(content))

			bucket := make([][]KeyValue, nReduce)
			for _, w := range kva {
				num := ihash(w.Key) % nReduce
				bucket[num] = append(bucket[num], w)
			}
			for no, listKeyValue := range bucket{
				tmpMapFile, err := ioutil.TempFile("", "mr-map-*")
				if err != nil {
					log.Fatalf("cannot opem tempfile %v", tmpMapFile)
				}
				enc := json.NewEncoder(tmpMapFile)
				err = enc.Encode(listKeyValue)
				if err != nil {
					log.Fatalf("encode bucker err %v", err)
				}
				tmpMapFile.Close()

				outMapFile := "mr-" + id + "-" +strconv.Itoa(no)
				os.Rename(tmpMapFile.Name(), outMapFile)


			}
			CallGetMapFinTask(&req, &rsp)
		}else if rsp.State == 1 {
			// reduce实现，读取map保存的临时文件，进行计算

			CallGetReduceTask(&req1, &rsp1)
			fmt.Println("Start reduce task .........")
			mapTasks := rsp1.NumMapTask
			id := strconv.Itoa(rsp1.ReduceTask.ReduceId)
			intermediate := []KeyValue{}
			for i := 0; i < mapTasks; i++ {
				fmt.Println("read map file success")
				mapFilename := "mr-" + strconv.Itoa(i) + "-" +id
				inputFile, err := os.OpenFile(mapFilename, os.O_RDONLY, 0777)
				if err != nil {
					log.Fatalf("cannot open reduceTask %v", err)
				}
				dec := json.NewDecoder(inputFile)
				for {
					var kv []KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					intermediate = append(intermediate, kv...)
				}
			}

			sort.Sort(ByKey(intermediate))
			ReduceName := "mr-out-" + id
			tmpReduceFile, err := ioutil.TempFile("", "mr-reduce-*")
			if err != nil {
				log.Fatalf("cannot opem tempfile %v", tmpReduceFile)
			}

			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(tmpReduceFile, "%v %v\n", intermediate[i].Key, output)
				i = j
			}
			tmpReduceFile.Close()
			os.Rename(tmpReduceFile.Name(), ReduceName)
			CallGetReduceFinTask(&req1, &rsp1)
		}else  {
			fmt.Println("task has all finished!!!")
			break
		}
	}
}

func CallGetMapFinTask(req *MapTaskRequest, rsp *MapTaskResponse) {

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetMapFinTask", &req, &rsp)
	if ok {
		// return the file name
		fmt.Printf("finish %d mapTask", rsp.MapTask.MapId)

	} else {
		fmt.Printf("call failed!\n")
	}

}

func CallGetMapTask(req *MapTaskRequest, rsp *MapTaskResponse) {

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetMapTask", &req, &rsp)
	if ok {
		// return the file name
		fmt.Printf("mapTask id %d, filename %s\n", rsp.MapTask.MapId,
			rsp.MapTask.Filename)

	} else {
		fmt.Printf("call failed!\n")
	}

}


func CallGetReduceTask(req *ReduceTaskRequest, rsp *ReduceTaskResponse) {

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetReduceTask", &req, &rsp)
	if ok {
		// return the file name
		fmt.Printf("reduceTask id %d\n", rsp.ReduceTask.ReduceId)

	} else {
		fmt.Printf("call failed!\n")
	}

}

func CallGetReduceFinTask(req *ReduceTaskRequest, rsp *ReduceTaskResponse) {

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.GetReduceFinTask", &req, &rsp)
	if ok {
		// return the file name
		fmt.Printf("finish %d reduceTask", rsp.ReduceTask.ReduceId)

	} else {
		fmt.Printf("call failed!\n")
	}

}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	//sockname := coordinatorSock()
	//c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
