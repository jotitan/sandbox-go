package node

import (
	"time"
	"resize"
	"fmt"
)

type ResizeTask struct {
	info *Info
	from string
	to string
	width uint
	height uint
}

func (tm TasksManager)NewResizeTask(from,to string,width,height uint)ResizeTask{
	return ResizeTask{info:NewInfo(tm.seq.Next(),ResizeTaskType),from:from,to:to,width:width,height:height}
}

func (task ResizeTask)Start()Task{
	task.info.Status = StatusRunning
	task.info.StartTime = time.Now()
	if err := resize.Resize(task.from, task.to, task.width,task.height) ; err == nil {
		task.info.Status = StatusDone
	}else{
		task.info.Status = StatusError
	}
	task.info.EndTime = time.Now()
	return task
}

func (task ResizeTask)GetInfo()*Info{
	return task.info
}

func (task ResizeTask)Serialize() []string {
	return []string{
		fmt.Sprintf("from=%s",task.from),
		fmt.Sprintf("to=%s",task.to),
		fmt.Sprintf("width=%s",task.width),
		fmt.Sprintf("height=%s",task.height),
	}
}
