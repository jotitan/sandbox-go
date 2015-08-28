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
)

/* Launch a server to treat resize image */

var tasksManager node.TasksManager

func addTask(_ http.ResponseWriter, request *http.Request){
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
		tasksManager.AddTask(task,force)
	}
}


func root(response http.ResponseWriter, request *http.Request){
    response.Write([]byte("Page d'accueil"))
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


func createServer(port string,nbTaskers int){
	tasksManager = node.NewTaskManager(nbTaskers,fmt.Sprintf("http://localhost:%s",port))
	tasksManager.DiscoverNetwork("127.0.0",[]int{9008,9015},[]int{1,1})
	tasksManager.Info()

    mux := http.NewServeMux()
    mux.HandleFunc("/resize",resizeReq)
    mux.HandleFunc("/load",load)
    mux.HandleFunc("/status",status)
    mux.HandleFunc("/register",registerNode)
    mux.HandleFunc("/add",addTask)
    mux.HandleFunc("/stats",stats)
    mux.HandleFunc("/",root)

	logger.GetLogger().Info("Runner ok on :",port)

    http.ListenAndServe(":" + port,mux)
	logger.GetLogger().Error("Runner ko")
}

func main(){
    runtime.GOMAXPROCS(runtime.NumCPU())
	createServer(arguments.ParseArgs()["port"],2)
}
