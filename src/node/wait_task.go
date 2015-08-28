package node

import (
	"time"
	"fmt"
)


// WaitTask represent a task which wait some time
type WaitTask struct {
	info *Info
	// wait time in millis
	wait int
}

func (tm TasksManager)NewWaitTask(wait int)WaitTask{
	return WaitTask{info:NewInfo(tm.seq.Next(),WaitTaskType),wait:wait}
}

func (task WaitTask)Start()Task{
	time.Sleep(time.Duration(task.wait) * time.Millisecond)
	task.info.Status = StatusDone
	return task
}

func (task WaitTask)GetInfo()*Info{
	return task.info
}

func (task WaitTask)Serialize() []string {
	return []string{fmt.Sprintf("wait=%d",task.wait)}
}
