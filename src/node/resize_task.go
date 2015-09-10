package node

import (
	"time"
	"resize"
	"fmt"
	"path/filepath"
	"logger"
	"regexp"
	"os"
)

type ResizeTask struct {
	info *Info
	from string
	to string
	width uint
	height uint
}

func cleanPath(value string)string{
	r,_ := regexp.Compile("(/|\\\\)")
	return r.ReplaceAllString(value,string(os.PathSeparator)+":")
}

func (tm TasksManager)NewResizeTask(from,to string,width,height uint)ResizeTask{
	return ResizeTask{info:NewInfo(tm.seq.Next(),ResizeTaskType),
		from:cleanPath(from),to:cleanPath(to),
		width:width,height:height}
}

func (task ResizeTask)Start(folder string)Task{
	task.info.StartTime = time.Now()
	if err := resize.Resize(filepath.Join(folder,task.from), filepath.Join(folder,task.to), task.width,task.height) ; err == nil {
		task.info.Status = StatusDone
	}else{
		logger.GetLogger().Error(err,folder)
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

func (task ResizeTask)ToString()string{
	return fmt.Sprintf("RESIZE => id:%s, from:%s, to:%s, width:%d, height:%d",task.info.Id,
	task.from,task.to,task.width,task.height)
}
