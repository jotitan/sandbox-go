package main
import (
    "net/http"
    "fmt"
    "strconv"
    "resize"
    "runtime"
	"node"
	"arguments"
	"logger"
	"encoding/json"
	"net"
	"strings"
	"time"
	"bufio"
)

/* Launch a server to treat resize image */

var tasksManager * node.TasksManager

// Return the id of task to track it
func addTask(response http.ResponseWriter, request *http.Request){
	var task node.Task
	force := false
	if forceValue := request.FormValue("force") ; forceValue == "true" || forceValue == "1"{
		force = true
	}
	logger.GetLogger().Info("=>RECEIVE EVENT", request.FormValue("type"))
	switch request.FormValue("type") {
		case node.WaitTaskType :
		waitTime,_ := strconv.ParseInt(request.FormValue("wait"),10,0)
		task = tasksManager.NewWaitTask(int(waitTime))

		case node.ResizeTaskType :
		width,_ := strconv.ParseInt(request.FormValue("width"),10,0)
		height,_ := strconv.ParseInt(request.FormValue("height"),10,0)
		from:=request.FormValue("from")
		to:=request.FormValue("to")
		task = tasksManager.NewResizeTask(from,to,uint(width),uint(height))
	}
	if task != nil {
		realID := tasksManager.AddTask(task,force)
		response.Write([]byte(realID))
	}
}

// Return the status of a task
func getStatusTask(response http.ResponseWriter,request *http.Request){
	status := tasksManager.GetStatusTask(request.FormValue("id"))
	response.Write([]byte(fmt.Sprintf("%d",status)))
}

// Return all tasks
func getTasks(response http.ResponseWriter,request *http.Request){
	infos := tasksManager.GetInfoTasks()
	data,_ := json.Marshal(infos)
	response.Write(data)
}

// Return all tasks
func getAllTasks(response http.ResponseWriter,request *http.Request){
	infos := tasksManager.GetAllInfoTasks()
	data,_ := json.Marshal(infos)
	response.Write(data)
}

// getTasksAsSSE return in a server sent event. All new tasks are notified
func getTasksAsSSE(response http.ResponseWriter,request *http.Request){
	infos := tasksManager.GetInfoTasks()
	dataInfos,_ := json.Marshal(infos)
	// Send all init tasks, then send only updated tasks
	o,stream := node.NewTaskObserver()
	tasksManager.Register("1",o)

	createSSEHeader(response)
	logger.GetLogger().Info("Get request tasks SSE")
	response.Write([]byte("data: " + string(dataInfos) + "\n\n"))
	response.(http.Flusher).Flush()
	for data := range stream {
		response.Write([]byte("data: " + string(data) + "\n\n"))
		response.(http.Flusher).Flush()
	}
}

// getAllTasksAsSSE return all tasks in sse (use client sse)
func getAllTasksAsSSE(response http.ResponseWriter,request *http.Request){
	stream := make(chan []byte)

	// TODO add end event

	// Get infos of nodes
	for _,url := range tasksManager.GetNodes(){
		go func(u string) {
			resp, _ := http.DefaultClient.Get(u + "/tasksAsSSE")
			reader := bufio.NewReader(resp.Body)
			for {
				data, _, err := reader.ReadLine()
				if err != nil {
					return
				}
				if string(data) != "" {
					stream <- []byte(strings.Replace(string(data),"data: ","",-1))
				}
			}
		}(url)
	}

	go func() {
		infos := tasksManager.GetInfoTasks()
		dataInfos, _ := json.Marshal(infos)
		stream <- dataInfos
		tasksManager.Register("1",node.NewTaskObserverFromStream(stream))
	}()

	createSSEHeader(response)

	for data := range stream {
		response.Write([]byte("data: " + string(data) + "\n\n"))
		response.(http.Flusher).Flush()
	}
}


func toto(response http.ResponseWriter, request * http.Request){
	url := "http://10.30.24.110:9011/statsAsSSE"
	resp,_ := http.DefaultClient.Get(url)

	reader := bufio.NewReader(resp.Body)
	for {
		data,_,err := reader.ReadLine()
		if err != nil {
		    logger.GetLogger().Info("ERR",err)
				return
		}
		if string(data)!="" {
			logger.GetLogger().Info("DAT=>", string(data))
		}
	}
}


func root(response http.ResponseWriter, request *http.Request){
	url := request.RequestURI
	http.ServeFile(response,request,"resources/" + url[1:])
}

// Use to find node with very short timeout
func status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
}

func createSSEHeader(response http.ResponseWriter){
	response.Header().Set("Content-Type","text/event-stream")
	response.Header().Set("Cache-Control","no-cache")
	response.Header().Set("Connection","keep-alive")
	response.Header().Set("Access-Control-Allow-Origin","*")
}

// Return stats by server side event
func statsAsSSE(response http.ResponseWriter, request *http.Request){
	createSSEHeader(response)
	sendStats(response)
}

func sendStats(r http.ResponseWriter){
	defer func(){
		if err := recover() ; err != nil {}
	}()
	stop := false
	go func(){
		<-r.(http.CloseNotifier).CloseNotify()
		stop=true
	}()

	for {
		stats := tasksManager.GetAllStats()
		data, _ := json.Marshal(stats)
		r.Write([]byte("data: " + string(data) + "\n\n"))
		if stop == true{
			break
		}
		r.(http.Flusher).Flush()
		time.Sleep(1 * time.Second)
	}
}

// Use to find node with very short timeout
func stats(response http.ResponseWriter, request *http.Request){
	stats := tasksManager.GetStats()
	data,_:= json.Marshal(stats)
	response.Header().Set("Content-type","application/json")
	response.Write(data)
}

// Return stats of whole cluster
func allStats(response http.ResponseWriter, request *http.Request){
	stats := tasksManager.GetAllStats()
	data,_:= json.Marshal(stats)
	response.Header().Set("Content-type","application/json")
	response.Write(data)
}

// Register a new node
func registerNode(response http.ResponseWriter, request *http.Request){
	url := request.FormValue("url")
	tasksManager.RegisterNode(url)
}

func extractInt(r *http.Request,name string)int{
	if intValue := r.FormValue(name) ; intValue!="" {
		if value,err := strconv.ParseInt(intValue,10,0) ; err == nil {
			return int(value)
		}
	}
	return 0
}

func load(response http.ResponseWriter, r *http.Request){
	load := tasksManager.GetLoad()
	str := fmt.Sprintf("{\"load\":%f}",load)
	response.Header().Set("Content-type","application/json")
	response.Write([]byte(str))
}

func resizeReq(_ http.ResponseWriter, r *http.Request){
    from := r.FormValue("from")
    to := r.FormValue("to")
	height,width := extractInt(r,"height"),extractInt(r,"width")

    resize.ResizeMany(from,to,uint(width),uint(height))
}

func findExposedURL()string{
	adr,_ := net.InterfaceAddrs()
	for _,a := range adr {
		if a.String() != "0.0.0.0" && !strings.Contains(a.String(),"127.0.0.1"){
			if idx := strings.Index(a.String(),"/"); idx != -1 {
				return a.String()[:idx]
			}
			return a.String()
		}
	}
	return "localhost"
}

func createServer(port string,baseIP string,rangeIP []int,rangePort []int,nbTaskers int,folder string){
	if port == ""{
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := findExposedURL()
	tasksManager = node.NewTaskManager(nbTaskers,fmt.Sprintf("http://%s:%s",localIP,port))
	tasksManager.SetFolder(folder)

	if baseIP != "" && len(rangePort) == 2 && len(rangeIP) == 2 {
	    logger.GetLogger().Info("Discover network",baseIP,rangeIP,rangePort)
		tasksManager.DiscoverNetwork(baseIP, rangePort, rangeIP)
		tasksManager.Info()
	}

    mux := createRoutes()
	logger.GetLogger().Info("Runner ok on :",localIP,port)
    http.ListenAndServe(":" + port,mux)

	logger.GetLogger().Error("Runner ko")
}

func createRoutes()*http.ServeMux{
	mux := http.NewServeMux()
	mux.HandleFunc("/resize",resizeReq)
	mux.HandleFunc("/load",load)
	mux.HandleFunc("/status",status)
	mux.HandleFunc("/register",registerNode)
	mux.HandleFunc("/add",addTask)
	mux.HandleFunc("/stats",stats)
	mux.HandleFunc("/statsAsSSE",statsAsSSE)
	mux.HandleFunc("/allStats",allStats)
	mux.HandleFunc("/taskStatus",getStatusTask)
	mux.HandleFunc("/tasks",getTasks)
	mux.HandleFunc("/allTasks",getAllTasks)
	mux.HandleFunc("/tasksAsSSE",getTasksAsSSE)
	mux.HandleFunc("/allTasksAsSSE",getAllTasksAsSSE)
	mux.HandleFunc("/toto",toto)
	mux.HandleFunc("/",root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]
	baseIP := args["baseIP"]
	rangeIP := make([]int,0,2)
	rangePort := make([]int,0,2)

	if port,err := strconv.ParseInt(args["ipMin"],10,0) ; err == nil {
		rangeIP = append(rangeIP,int(port))
	}

	if port,err := strconv.ParseInt(args["ipMax"],10,0) ; err == nil {
		rangeIP = append(rangeIP,int(port))
	}

	if port,err := strconv.ParseInt(args["portMin"],10,0) ; err == nil {
		rangePort = append(rangePort,int(port))
	}

	if port,err := strconv.ParseInt(args["portMax"],10,0) ; err == nil {
		rangePort = append(rangePort,int(port))
	}

	createServer(port,baseIP,rangeIP,rangePort,2,args["folder"])
}
