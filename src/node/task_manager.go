package node

import (
	"sync"
	"fmt"
	"errors"
	"logger"
	"hardwareutil"
	"strings"
)


/* Object to manage differents tasks running on server */

const (
	WaitTaskType = "WAIT_TASK"
	ResizeTaskType = "RESIZE_TASK"
)

func NewSequence(url string)Sequence{
	return Sequence{0,sync.Mutex{},url}
}

type Sequence struct {
	current int
	locker sync.Mutex
	// Prefix of id, composed with ip:port to have a unique id
	baseId string
}

func (sq * Sequence)Next()string{
	sq.locker.Lock()
	sq.current++
	id := sq.current
	sq.locker.Unlock()
	return fmt.Sprintf("%s:%d",sq.baseId,id)
}


// TasksManager contains all tasks by status
type TasksManager struct {
	// Number of parallel task which are accepted
	NbParallelTask int
	//
	url string
	// List of all tasks
	tasks map[string]Task
	finishedTasks map[string]Task

	seq * Sequence

	taskChan chan Task
	// List of node of cluster. Usefull to deport treatment on other node
	nodes map[string]NodeClient
	// FOlder where photos and cache can be found
	folder string
}

func NewTaskManager(nbTask int,localUrl string)*TasksManager{
	seq := NewSequence(localUrl)
	tm := TasksManager{NbParallelTask:nbTask,
		url:localUrl,
		tasks:make(map[string]Task),
		finishedTasks:make(map[string]Task),
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
				logger.GetLogger().Info("Begin task",t.ToString())
				t = t.Start(tm.folder)
				logger.GetLogger().Info("End task",t.GetInfo().Id,t.GetInfo().Status)
				// Move task in finished list
				tm.finishedTasks[t.GetInfo().Id] = t
				delete(tm.tasks,t.GetInfo().Id)
				<- taskLimiter
			}(task)
		}
	}()
	return &tm
}
func (tm * TasksManager)SetFolder(folder string){
	tm.folder = folder
}


// GetStatusTask return the status of the task. Real id is after the last :, before it's the server address
func (tm TasksManager)GetStatusTask(id string)int8{
	urlNode := id[:strings.LastIndex(id,":")]
	// case of local task
	if tm.url == urlNode {
		if task, ok := tm.tasks[id]; ok {
			return task.GetInfo().Status
		}
		if task, ok := tm.finishedTasks[id]; ok {
			return task.GetInfo().Status
		}
		return StatusNotFound
	}
	//ask distant server
	node := tm.nodes[urlNode]
	return node.GetStatusTask(id)
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
	cs := make(chan Stats)

	for _,node := range tm.nodes {
		go func(n NodeClient){
			cs <- n.GetStats()
		}(node)
	}
	for _ = range tm.nodes {
		stats = append(stats,<-cs)
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
	stats.ID = tm.url
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
		// TODO goroutine
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

var limiter = make(chan int,1)

// @param force : if true, no load server check, just add the task
func (tm * TasksManager)AddTask(task Task,force bool)string{
   	// To avoid manage all request on only one node when many requests arrived at same time, used a channel to limit
	limiter <- 1
	// Search an other one
	if !force {
		logger.GetLogger().Info("LOAD",tm.GetLoad(),task.GetInfo().Id)
		if load := tm.GetLoad() ; load > 1 {
			if node, notFound := tm.findQuiteNode(load) ; notFound == nil {
				// Add task to this node, quiter, and return distant id
				if id,err := node.SendTask(task) ; err == nil {
					<- limiter
					return id
				}
				return ""
			}
		}
	}
	logger.GetLogger().Info("Treat here",task.GetInfo().Id)
	// check ratio, ask friend and add later
	<- limiter
	tm.taskChan <- task
	return task.GetInfo().Id
}
