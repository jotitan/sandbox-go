package task_manager
import (
    "fmt"
    "net/http"
    "path/filepath"
    "logger"
    "net/url"
    "resize"
    "os"
    "net"
    "strings"
    "regexp"
)

type Node struct{
    urlManager string
    // folder where photos are
    folder string
    capacity int
}

func findExposedURL()string{
    adr,_ := net.InterfaceAddrs()

    for _,a := range adr {
        if a.String() != "0.0.0.0" && !strings.Contains(a.String(),"127.0.0.1") && strings.HasPrefix(a.String(),"192.168.0"){
            if idx := strings.Index(a.String(),"/"); idx != -1 {
                return a.String()[:idx]
            }
            return a.String()
        }
    }
    return "localhost"
}

func (n Node)ConnectManager(port string)string{
    url := findExposedURL()
    http.Get(fmt.Sprintf("%s/join?capacity=%d&address=http://%s:%s",n.urlManager,n.capacity,url,port))
    return url
}

func (n Node)treatTask(response http.ResponseWriter,request * http.Request) {
    // get in and out parameter
    key := request.FormValue("key")
    task:= request.FormValue("task")
    n.Treat(key,task,request.Form)
}

var cleanPath,_ = regexp.Compile("[/\\\\]")
func clean(path string)string{
    return cleanPath.ReplaceAllString(path,string(os.PathSeparator))
}

func (n Node)Treat(idTask,typeTask string, parameters url.Values) {
    switch typeTask{
        case "resize":
        n.resize(clean(filepath.Join(n.folder,parameters.Get("in"))), clean(filepath.Join(n.folder,parameters.Get("out")) ),parameters.Get("force") == "true")
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
    if err := resize.GetResizer().Resize(in ,out,0,1080) ; err != nil {
        logger.GetLogger2().Error("Impossible to resize img",err)
    }
}

func (n Node)setAckToManager(idTask string){
    http.Get(fmt.Sprintf("%s/%s?id=%s",n.urlManager,"result",idTask))
}

//LaunchNodeServer
// capacity : treatment capacity of node
func LaunchNodeServer(urlManager, folder string, capacity int){
    node := Node{urlManager,folder,capacity}
    port := "8012"

    server := http.NewServeMux()
    server.HandleFunc("/treat",node.treatTask)
    // Connect to Manager
    url := node.ConnectManager(port)
    logger.GetLogger2().Info("Launch node on port",port,"with address",url)

    http.ListenAndServe(":" + port,server)
}