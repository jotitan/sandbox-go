package node

import "time"

const (
	StatusNew = int8(0)
	StatusRunning = int8(1)
	StatusDone = int8(2)
	StatusError = int8(3)
	StatusNotFound = int8(4)
)


type Task interface {
	/* Return info about task */
	GetInfo()*Info
	/* Start the function */
	Start(folder string) Task
	/* Serialize task to be sent to another node */
	Serialize() []string
	// ToString return a string representation of task
	ToString() string
}

type Info struct {
	// Id is composed with machine id (ip:port) and unique id
	Id string
	TypeTask string
	Status int8
	CreateTime time.Time
	StartTime time.Time
	EndTime time.Time
}

func NewInfo(id string,typeTask string)*Info{
	return &Info{Id:id,Status:StatusNew,CreateTime:time.Now(),TypeTask:typeTask}
}
