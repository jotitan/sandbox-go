package node

import (
	"sync"
	"fmt"
	"errors"
	"logger"
	"hardwareutil"
)


/* Object to manage differents tasks running on server */

const (
	WaitTaskType = "WAIT_TASK"
	ResizeTaskType = "RESIZE_TASK"
)

func NewSequence()Sequence{
	return Sequence{0,sync.Mutex{}}
}

type Sequence struct {
	current int
	locker sync.Mutex
}

func (sq * Sequence)Next()int{
	sq.locker.Lock()
	sq.current++
	id := sq.current
	sq.locker.Unlock()
	return id
}


// TasksManager contains all tasks by status
type TasksManager struct {
	// Number of parallel task which are accepted
	NbParallelTask int
	//
	url string
	// List of all tasks
	tasks map[int]Task

	seq * Sequence

	taskChan chan Task
	// List of node of cluster. Usefull to deport treatment on other node
	nodes map[string]NodeClient
}

func NewTaskManager(nbTask int,localUrl string)TasksManager{
	seq := NewSequence()
	tm := TasksManager{NbParallelTask:nbTask,
		url:localUrl,
		tasks:make(map[int]Task),
		seq:&seq,
		taskChan:make(chan Task),
		nodes:make(map[string]NodeClient),
	}
	taskLimiter := make(chan int,nbTask)

	// Launch task consumer
	go func(){
		for task := range tm.taskChan {
			logger.GetLogger().Info("Receive task",task.GetInfo().Id)
			tm.tasks[task.GetInfo().Id] = task
			go func(t Task){
				// To limit number
				taskLimiter <-0
				logger.GetLogger().Info("Begin task",t.GetInfo().TypeTask,t.GetInfo().Id)
				t = t.Start()
				logger.GetLogger().Info("End task",t.GetInfo().Id,t.GetInfo().Status)
				delete(tm.tasks,t.GetInfo().Id)
				<- taskLimiter
			}(task)
		}
	}()
	return tm
}

func (tm TasksManager)Info(){
	logger.GetLogger().Info("NODES",len(tm.nodes))
}

// DiscoverNetwork discover the network of king A.B.C.xxx where xxx can be between 1 and 253
func (tm * TasksManager) DiscoverNetwork(baseIP string, rangePort []int, rangeIP []int) {
	for ip4 := rangeIP[0] ; ip4 <= rangeIP[1] ; ip4++ {
		for port := rangePort[0] ; port <=rangePort[1] ; port++ {
			if url := fmt.Sprintf("http://%s.%d:%d",baseIP,ip4,port) ; url != tm.url {
				if CheckNode(url) {
					if node,err := tm.RegisterNode(url) ; err == nil {
						node.Register(tm.url)
					}
				}
			}
		}
	}
}

func (tm * TasksManager)RegisterNode(nodeURL string)(NodeClient,error){
	if _,present := tm.nodes[nodeURL] ; present {
		return NodeClient{},errors.New("Client " + nodeURL + " already exists")
	}
	node := NewNodeClient(nodeURL)
	tm.nodes[nodeURL] = node
	logger.GetLogger().Info("Node add to cluster", node.Url)
	return node,nil
}

func (tm TasksManager)GetAllStats()[]Stats{
	stats := make([]Stats,0,len(tm.nodes)+1)
	stats = append(stats,tm.GetStats())
	for _,node := range tm.nodes {
		stats = append(stats,node.GetStats())
	}
	return stats
}

func (tm TasksManager)GetStats()Stats{
	stats := Stats{}
	stats.CPU = hardwareutil.GetCPUUsage()
	stats.Memory = hardwareutil.GetCurrentMemory()
	stats.NbTaskers = tm.NbParallelTask
	stats.Load = tm.GetLoad()
	stats.NbTasks = len(tm.tasks)
	stats.Temperature = hardwareutil.GetTemperature()

	return stats
}

// Check if it's possible to run a new task
func (tm TasksManager)GetLoad()float64{
	return float64(len(tm.tasks)) / float64(tm.NbParallelTask)
}

func (tm TasksManager)findQuiteNode(load float64)(NodeClient,error){
	var minNode NodeClient
	minLoad := load
	for _,n := range tm.nodes{
		if nodeLoad := n.GetLoad() ; minLoad > nodeLoad{
			minLoad = nodeLoad
			minNode = n
		}
	}
	if minLoad == load {
		return NodeClient{},errors.New("No quiter node")
	}
	return minNode,nil
}

// @param force : if true, no load server check, just add the task
func (tm TasksManager)AddTask(task Task,force bool){
   	// Search an other one
	if !force {
		if load := tm.GetLoad() ; load > 1 {
			if node, notFound := tm.findQuiteNode(load) ; notFound == nil {
				// Add task to this node, quiter
				node.SendTask(task)
				return
			}
		}
	}
	// check ratio, ask friend and add later
	tm.taskChan <- task

}
