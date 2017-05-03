package log_reader

import (
	"os/exec"
	"net/http"
	"fmt"
	"io"
	"time"
	"strings"
)

var logProcesses = make(map[string]*exec.Cmd)

// Read log in file and push in server sent event
func LaunchLogReader(w http.ResponseWriter, logFilename string){
	w.Header().Set("Content-Type","text/event-stream")
	w.Header().Set("Cache-Control","no-cache")
	w.Header().Set("Connection","keep-alive")
	w.Header().Set("Access-Control-Allow-Origin","*")

	cmd := exec.Command("tail","-500f",logFilename)
	idLog := fmt.Sprintf("log_%d",len(logProcesses))
	// check communication reset, stop
	go func(){
		<-w.(http.CloseNotifier).CloseNotify()
		killLogProc(idLog)
	}()

	val := ""
	lw := LogFlusher{w.(http.Flusher),w,&val}
	cmd.Stdout = lw

	lw.Write([]byte("_id_log_" + idLog + "\n"))
	go sendLogCheck(lw,cmd)
	logProcesses[idLog] = cmd
	cmd.Run()
}

func killLogProc(idLog string){
	if cmd,ok := logProcesses[idLog]; ok {
		cmd.Process.Kill()
		delete(logProcesses,idLog)
	}
}

// LogFlusher provide tool to write tool in SSE
type LogFlusher struct{
	f http.Flusher
	w io.Writer
	previous *string
}

func sendLogCheck(lw LogFlusher,cmd * exec.Cmd){
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return
	}
	//logger.GetLogger().Info("Send ping")
	lw.Ping()
	time.Sleep(time.Duration(60)*time.Second)
	sendLogCheck(lw,cmd)
}

// Ping make a ping to the client of SSE (Server Side Event)
func (l LogFlusher)Ping(){
	defer func(){
		if e := recover() ; e!=nil{}
	}()
	if l.f == nil {
		return
	}
	l.w.Write([]byte("event: ping\n"))
	l.w.Write([]byte("data:  \n\n"))
	l.f.Flush()
}

// Write is the implementation of Writer
func (l LogFlusher)Write(p []byte) (n int, err error){
	l.w.Write(append([]byte("retry:1000\n"),p...))
	values := strings.Split(string(p),"\n")
	skipPrevious := false
	if values[len(values)-1] != ""{
		// Last value is not complete, save for later
		*(l.previous) = values[len(values)-1]
		values = values[:len(values)-1]
		skipPrevious = true
	}
	for _,s := range values{
		if !skipPrevious && *(l.previous) != ""{
			s=*(l.previous)+s
			*(l.previous) = ""
		}
		if s != "" {
			l.w.Write([]byte("data:" + s + "\n\n"))
		}
	}
	l.f.Flush()
	return len(p),nil
}