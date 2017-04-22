package task_manager
import (
    "net/http"
    "strconv"
    "logger"
    "syscall"
)

/* Represent a task manager. Some task can be add and distributes over many nodes */


/* METHODS */
// Add task (resize a folder and sub)
// Add node
// Heartbeat


type ManagerServer struct {
    nodeManager Manager
}

func NewManagerServer()ManagerServer{
    return ManagerServer{NewManager()}
}

func (m ManagerServer)kill(response http.ResponseWriter, request * http.Request){
    syscall.Exit(1)
}

func (m * ManagerServer)result(response http.ResponseWriter, request * http.Request){
    m.nodeManager.ReceiveAck(request.FormValue("id"))
}

func (m * ManagerServer)join(response http.ResponseWriter, request * http.Request){
    capacity := 1
    if v,err := strconv.ParseInt(request.FormValue("capacity"),10,32) ; err == nil {
        capacity = int(v)
    }
    m.nodeManager.AddNode(request.FormValue("address"),capacity,false)
}

func (m ManagerServer)getFile(response http.ResponseWriter, request * http.Request){
    http.ServeFile(response,request,"launch_task.html")
}

func (m * ManagerServer)parseAndResize(response http.ResponseWriter, request * http.Request){
    /*if !m.nodeManager.CanTreat() {
        logger.GetLogger().Error("Impossible to treat, no node available")
    }*/
    if request.FormValue("prefix") == "" || request.FormValue("folder") == "" || request.FormValue("outputFolder") == "" {
        response.Write([]byte("Impossible to launch task, parameters are missing"))
        return
    }
    m.nodeManager.ParseAndResizeFolder(request.FormValue("prefix"),request.FormValue("folder"),request.FormValue("outputFolder"),request.FormValue("first") == "true",request.FormValue("force") == "true",nil)
}

func (m ManagerServer)testSubmit(response http.ResponseWriter, request * http.Request){
    logger.GetLogger2().Info("Launch",request.FormValue("prefix"),request.FormValue("folder"),request.FormValue("outputFolder"),request.FormValue("first"),request.FormValue("force"))
}

func (m ManagerServer)Start(){
    port := "8010"
    server := http.NewServeMux()
    server.HandleFunc("/result",m.result)
    server.HandleFunc("/join",m.join)
    server.HandleFunc("/testSubmit",m.testSubmit)
    server.HandleFunc("/parseAndResize",m.parseAndResize)
    server.HandleFunc("/",m.getFile)

    logger.GetLogger2().Info("Run server on port",port)
    logger.GetLogger2().Error("End of server",http.ListenAndServe(":" + port,server))
}