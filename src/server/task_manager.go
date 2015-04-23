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

func NewResizeTask()ResizeTask{
	return ResizeTask{NewInfo()}
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
