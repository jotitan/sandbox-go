package task_manager
import (
    "fmt"
    "net/http"
    "path/filepath"
    "logger"
    "net/url"
    "resize"
    "os"
)

type Node struct{
    urlManager string
    // folder where photos are
    folder string
    capacity int
}

func (n Node)ConnectManager(port string){
    http.Get(fmt.Sprintf("%s/join?capacity=%d&address=%s:%s",n.urlManager,n.capacity,"http://localhost",port))
}

func (n Node)treatTask(response http.ResponseWriter,request * http.Request) {
    // get in and out parameter
    key := request.FormValue("key")
    task:= request.FormValue("task")
    n.Treat(key,task,request.Form)
}

func (n Node)Treat(idTask,typeTask string, parameters url.Values) {
    switch typeTask{
        case "resize":
        n.resize(filepath.Join(n.folder,parameters.Get("in")), filepath.Join(n.folder,parameters.Get("out")),parameters.Get("force") == "true")
        n.setAckToManager(idTask)
    }
}

// Check before if image already exist, except if force = true
func (n Node)resize(in,out string, force bool){
    if !force {
        if _,err := os.Open(out) ; err == nil{
            // File already exist, don't create a new one
            return
        }
    }
    logger.GetLogger2().Info("Resize image",in,"to",out,resize.GetResizer().ToString())
    // Create complete folder path if necessary
    os.MkdirAll(filepath.Dir(out),os.ModePerm)
    if err := resize.GetResizer().Resize(in ,out,0,400) ; err != nil {
        logger.GetLogger2().Error("Impossible to resize img",err)
    }
}

func (n Node)setAckToManager(idTask string){
    http.Get(fmt.Sprintf("%s/%s?id=%s",n.urlManager,"result",idTask))
}

// capacity : treatment capacity of node
func LaunchServer(urlManager, folder string, capacity int){
    node := Node{urlManager,folder,capacity}
    port := "8012"

    server := http.NewServeMux()
    server.HandleFunc("/treat",node.treatTask)
    // Connect to Manager
    logger.GetLogger2().Info("Launch node on port",port)
    node.ConnectManager(port)

    http.ListenAndServe(":" + port,server)
}