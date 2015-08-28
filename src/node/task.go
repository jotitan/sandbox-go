package node

import "time"

const (
	StatusNew = 0
	StatusRunning = 1
	StatusDone = 2
	StatusError = 3
)


type Task interface {
	/* Return info about task */
	GetInfo()*Info
	/* Start the function */
	Start() Task
	/* Serialize task to be sent to another node */
	Serialize() []string
}

type Info struct {
	Id int
	TypeTask string
	Status int8
	CreateTime time.Time
	StartTime time.Time
	EndTime time.Time
}

func NewInfo(id int,typeTask string)*Info{
	return &Info{Id:id,Status:StatusNew,CreateTime:time.Now(),TypeTask:typeTask}
}
