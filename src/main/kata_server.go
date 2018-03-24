package main

import (
	"logger"
	"net/http"
	"strconv"
	"fmt"
	"sync/atomic"
	"time"
	"math/rand"
	"errors"
)

type MonitorThroughput struct {
	counter *int32
	counter429 *int32
	// Parallel limit
	throughput int32
	limit429 int32
	waitPunish int32
	slowDownFactor int32
}

func (mt MonitorThroughput)tick()(bool,error){
	if *mt.counter > mt.throughput {
		atomic.AddInt32(mt.counter429,1)
		if *mt.counter429 > mt.limit429 {
			return true,errors.New("Limit 429 reach")
		}
		return true,nil
	}
	atomic.AddInt32(mt.counter,1)
	return false,nil
}

func (mt MonitorThroughput)end(){
	atomic.AddInt32(mt.counter,-1)
}

func createMonitor(throughput,limit429, waitPunish int32)MonitorThroughput {
	counter := int32(0)
	counter429 := int32(0)
	return MonitorThroughput{&counter,&counter429,throughput,limit429,waitPunish,1}
}

var monitor = createMonitor(200,300,20)

func waitRandomTime()int32{
	waitTime := (20 + rand.Int31()%400) * monitor.slowDownFactor
	time.Sleep(time.Duration(waitTime) * time.Millisecond)
	return waitTime
}

func setTempFactor()bool{
	if monitor.slowDownFactor != 1 {
		return false
	}
	monitor.slowDownFactor = 10
	go func() {
		time.Sleep(time.Duration(monitor.waitPunish)*time.Second)
		monitor.slowDownFactor = 1
	}()
	return true
}

func setHeader(w http.ResponseWriter){
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func Request(response http.ResponseWriter,request *http.Request){
	setHeader(response)
	if value,err := strconv.ParseInt(request.FormValue("value"),10,32) ; err == nil {
		if ok,err := monitor.tick() ;ok {
			// Too many 429, stop server
			if err != nil {
				http.Error(response,"Too many requests, server will slow down",509)
				if setTempFactor() {
					logger.GetLogger().Error(fmt.Sprintf("Server will slow during %ds cause :",monitor.waitPunish),err.Error())
				}
				return
			}
			http.Error(response,"Too many requests, please slow down",429)
			logger.GetLogger().Error("Too many requests, 429 sended")
		}else {
			// Wait some random time
			wait := waitRandomTime()
			response.Write([]byte(fmt.Sprintf("{\"value\":%d,\"wait\":%d}",computeValue(value),wait)))
			monitor.end()
		}
	}else{
		http.Error(response,"Badd request, set parameter value as integer",http.StatusBadRequest)
	}
}

func computeValue(value int64)int64{
	return value*-1
}

func createServer(){
	server := http.NewServeMux()
	server.HandleFunc("/request",Request)
	http.ListenAndServe(":8081",server)
}

// Launch server
func main(){
	logger.GetLogger().Info("Launch server on port 8081, you can request me !")
	createServer()
}

