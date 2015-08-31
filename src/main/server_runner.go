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
)

/* Launch a server to treat resize image */

var tasksManager node.TasksManager

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


func root(response http.ResponseWriter, request *http.Request){
	url := request.RequestURI
	fmt.Println("../../resources/html/" + url[1:])
	http.ServeFile(response,request,"../resources/html/" + url[1:])
}

// Use to find node with very short timeout
func status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
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

func createServer(port string,baseIP string,rangeIP []int,rangePort []int,nbTaskers int){
	if port == ""{
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := findExposedURL()
	tasksManager = node.NewTaskManager(nbTaskers,fmt.Sprintf("http://%s:%s",localIP,port))

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
	mux.HandleFunc("/allStats",allStats)
	mux.HandleFunc("/taskStatus",getStatusTask)
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

	createServer(port,baseIP,rangeIP,rangePort,2)
}
