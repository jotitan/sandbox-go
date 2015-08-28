package server

import (
	"resize"
	"time"
	"sync"
)


/* Object to manage differents tasks running on server */

const (
	STATUS_NEW = 0
	STATUS_RUNNING = 1
	STATUS_DONE = 2
	STATUS_ERROR = 3
)

type SequenceId struct {
	currentId int
	locker sync.Mutex
}

func NewSequenceId()SequenceId{
	return SequenceId{0,sync.Mutex{}}
}

func (sq * SequenceId)next()int {
	sq.locker.Lock()
	sq.currentId++
	id := sq.currentId
	sq.locker.Unlock()
	return id
}

type Task interface {
	/* Return info about task */
	GetInfo()Info
	/* Start the function */
	Start()
}

type Info struct {
	Id int
	Status int8
	CreateTime time.Time
	StartTime time.Time
	EndTime time.Time
}

type Sequence struct {
	lock sync.Mutex
	current int
}

func (s * Sequence)Next()int{
	s.lock.Lock()
	defer func(){
		s.current++
		s.lock.Unlock()
	}()
	return s.current
}

func NewInfo(id int)Info{
	return Info{Id:id,Status:STATUS_NEW,CreateTime:time.Now}
}

type ResizeTask struct {
	info Info
	from string
	to string
	width uint
	height uint
}

func (tm TasksManager)NewResizeTask()ResizeTask{
	// TODO synchronize id (maybe use timestamp)
	return ResizeTask{NewInfo(tm.seq.next())}
}

func (task * ResizeTask)Start(){
	task.info.Status = STATUS_RUNNING
	task.info.StartTime = time.Now()
	if err := resize.Resize(task.from, task.to, task.width,task.height) ; err == nil {
		task.info.Status = STATUS_DONE
	}else{
		task.info.Status = STATUS_ERROR
	}
	task.info.EndTime = time.Now()
}

func (task ResizeTask)GetInfo(){
	return task.info
}

// TasksManager contains all tasks by status
type TasksManager struct {
	// Number of parallel task which are accepted
	NbParallelTask int
	// Used to limit number of task. Check when add in runningTasks
	locker  sync.Locker
	// List of all tasks
	tasks map[int]*Task

	seq * SequenceId

	taskChan chan Task
}

func NewTaskManager(nbTask int)TasksManager{
	tm := TasksManager{NbParallelTask:nbTask,
		locker:sync.Mutex{},
		tasks:make(map[int]*Task),
		seq:&(NewSequenceId()),
		taskChan:make(chan Task),
	}
	cch := make(chan int,nbTask)

	// Launch task consumer
	go func(){
		for task := range tm.taskChan {
			tm.tasks[task.GetInfo().Id] = task
			go func(){
				// To limit number
				cch <-
				task.Start()
				delete(tm.tasks,task.GetInfo().Id)
				<- cch
			}()
		}
	}()
	return tm
}

// Check if it's possible to run a new task
func (tm TasksManager)getRatio()float64{
	return float64(len(tm.tasks)) / tm.NbParallelTask
}

func (tm TasksManager)addTask(task Task){
	// check ratio, ask friend and add later
	tm.taskChan <- task

}

func (tm TasksManager)RunTask(id int){
	// Check if task exist and waiting run
	task,err := tm.tasks[id]
	if err != nil {
		return
	}
	// Check is enougth running slot are available
	canDoIt := false
	tm.locker.Lock()
	if canDoIt = tm.CanRunTask() ; canDoIt == true {
		runningTasks
	}
	tm.locker.Unlock()
}