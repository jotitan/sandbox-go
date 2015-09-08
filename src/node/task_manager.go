package node

import (
	"sync"
	"fmt"
	"errors"
	"logger"
	"strings"
	"time"
	"encoding/json"
	"strconv"
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
	stats Stats
	eventObserver EventObserver
}


func (tm * TasksManager)BuildTask(typeTask string,values map[string][]string,force bool)(string,error) {
	switch typeTask {
		case WaitTaskType :
			waitTime, _ := strconv.ParseInt(values["wait"][0], 10, 0)
			return tm.AddTask(tm.NewWaitTask(int(waitTime)),force),nil

		case ResizeTaskType :
			width, _ := strconv.ParseInt(values["width"][0], 10, 0)
			height, _ := strconv.ParseInt(values["height"][0], 10, 0)
			from := values["from"][0]
			to := values["to"][0]
			return tm.AddTask(tm.NewResizeTask(from, to, uint(width), uint(height)),force),nil
		default :return "",errors.New("Impossible to define task with name " + typeTask)
	}

	return "",errors.New("Impossible case")
}

type Observer interface{
	NewTask(task Task)
	UpdateTask(task Task)
}

// EventObserver manage observer of tasks activities
type EventObserver struct {
	observers map[int]Observer
	currentId int
	locker sync.Mutex
}

func NewEventObserver()EventObserver{
	return EventObserver{observers:make(map[int]Observer),
		currentId:0,
		locker:sync.Mutex{},
	}
}

func (obs * EventObserver)Remove(id int){
	delete(obs.observers,id)
}

func (obs * EventObserver)Register(observer Observer)int{
	obs.locker.Lock()
	obs.currentId++
	id :=obs.currentId
	obs.locker.Unlock()
	obs.observers[id] = observer
	return id
}

func (obs EventObserver)NewTask(task Task){
	for _,observer := range obs.observers {
		go observer.NewTask(task)
	}
}

func (obs EventObserver)UpdateTask(task Task){
	for _,observer := range obs.observers {
		go observer.UpdateTask(task)
	}
}

// TaskObserver stream implementation of observer
type TaskObserver struct{
	stream chan []byte
}

func NewTaskObserverFromStream(channel chan []byte)(Observer){
	return TaskObserver{channel}
}

func NewTaskObserver()(Observer,chan []byte){
	channel := make(chan []byte)
	return TaskObserver{channel},channel
}

func (to TaskObserver)NewTask(task Task){
	data,_ := json.Marshal([]Info{*task.GetInfo()})
	to.stream <- data
}

func (to TaskObserver)UpdateTask(task Task){
	data,_ := json.Marshal([]Info{*task.GetInfo()})
	to.stream <- data
}

func (tm * TasksManager)Register(observer Observer)int{
	return tm.eventObserver.Register(observer)
}

func (tm * TasksManager)RemoveObserver(id int){
	tm.eventObserver.Remove(id)
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
		eventObserver:NewEventObserver(),
	}

	// Launch task consumer
	tm.runTasksConsumer(nbTask)

	// Launch stats mechanism : works alone, data are returned when necessary
	tm.runStatsGetter()
	return &tm
}

func (tm * TasksManager)runTasksConsumer(nbTask int){
	taskLimiter := make(chan int,nbTask)
	go func(){
		for task := range tm.taskChan {
			tm.tasks[task.GetInfo().Id] = task
			tm.eventObserver.NewTask(task)
			go func(t Task){
				// To limit number
				taskLimiter <-0
				tm.launchTask(t)
				<- taskLimiter
			}(task)
		}
	}()
}

func (tm * TasksManager)launchTask(task Task){
	info := task.GetInfo()
	info.Status = StatusRunning
	tm.eventObserver.UpdateTask(task)
	logger.GetLogger().Info("Begin task",task.ToString())

	task = task.Start(tm.folder)

	logger.GetLogger().Info("End task",info.Id,task.GetInfo().Status)
	tm.eventObserver.UpdateTask(task)
	tm.finishedTasks[info.Id] = task
	delete(tm.tasks,info.Id)
}

func (tm * TasksManager)runStatsGetter(){
	go func(){
		for {
			tm.stats = getStats(*tm)
			time.Sleep(time.Second * 1)
		}
	}()
}

func (tm * TasksManager)SetFolder(folder string){
	tm.folder = folder
}

func (tm TasksManager)GetInfoTasks()[]Info{
	infos := make([]Info,0,len(tm.tasks))
	for _,t := range tm.tasks {
		infos = append(infos,*t.GetInfo())
	}
	return infos
}

// GetInfoTasks return info about tasks running or to be running over all nodes
func (tm TasksManager)GetAllInfoTasks()[]Info{
	infos := make([]Info,0,len(tm.tasks)*(len(tm.nodes)+1))
	infoChan := make(chan Info)
	wait := sync.WaitGroup{}
	wait.Add(len(tm.nodes)+1)

	// Treat task infos
	go func(){
		for info := range infoChan {
			infos = append(infos,info)
			wait.Done()
		}
	}()

	// Tasks of nodes
	for _,n := range tm.nodes {
		go func(nc NodeClient){
			tasks := nc.GetTasks()
			wait.Add(len(tasks))
			for _,info := range tasks{
				infoChan <- info
			}
			wait.Done()
		}(n)
	}

	// tasks of local node
	for _,t := range tm.tasks {
		wait.Add(1)
		infoChan <- *t.GetInfo()
	}
	wait.Done()
	wait.Wait()

	return infos
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

// GetNodes return connected nodes (not the local)
func (tm TasksManager)GetNodes()[]string{
	nodes := make([]string,0,len(tm.nodes)-1)
	for url,_ := range tm.nodes {
		if url != tm.url {
			nodes = append(nodes,url)
		}
	}
	return nodes
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
	return tm.stats
}

// Check if it's possible to run a new task
func (tm TasksManager)GetLoad()float64{
	return float64(len(tm.tasks)) / float64(tm.NbParallelTask)
}

type loadNode struct {
	node NodeClient
	load float64
}

func (tm TasksManager)findQuiteNode(load float64)(NodeClient,error){
	var minNode NodeClient
	minLoad := load
	chanLoad := make(chan loadNode)
	waiter := make(chan int,1)
	go func(){
		for _ = range tm.nodes {
			 if ln := <- chanLoad ; minLoad > ln.load {
				 minLoad = ln.load
				 minNode = ln.node
			 }
		}
		waiter <- 1
	}()
	for _,n := range tm.nodes{
		go func(nc NodeClient){
			 chanLoad <- loadNode{nc,nc.GetLoad()}
		}(n)
	}
	<-waiter
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
	<- limiter
	tm.taskChan <- task
	return task.GetInfo().Id
}
