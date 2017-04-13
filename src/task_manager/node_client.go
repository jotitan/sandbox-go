package task_manager
import (
    "net/http"
    "fmt"
    "logger"
)


type NodeClient struct{
    // local id of node
    localId int
    url string
    capacity int
    // Use to limit parallel running task for node
    limit chan struct{}
    RunningTasks map[string]*Task
    // counter of treated task
    localCounter int
    // manager which manage node
    manager Manager
    // used to stop in emergency
    emergencyStop bool
}

func NewNode(id int, url string, capacity int) NodeClient {
    return NodeClient{url:url,capacity:capacity,limit:make(chan struct{},capacity),RunningTasks:make(map[string]*Task),localCounter:0,localId:id}
}

//Run : launch thread which consumes task in manager chanel
func (n * NodeClient)Run(pendingTasks chan *Task){
    // Reserve a slot on node (chanel), get next task and treat
    for {
        n.limit <- struct{}{}
       if n.emergencyStop {
           break
       }
        task := <- pendingTasks
        task.key = fmt.Sprintf("%d_%d_%d",task.idTaskControl,n.localId,n.localCounter)
        n.localCounter++
        go n.Treat(task)
    }
    // End of loop (surely cause by error), reinject running tasks into main chanel
    for _,task := range n.RunningTasks {
        pendingTasks <- task
    }
}

func (n * NodeClient)Stop(){
    //logger.GetLogger().Info("Stop tasks treatment")
    n.emergencyStop = true
}

//SetTreat : when server respond for async request, set end treat
func (n * NodeClient)SetTreat(keyTask string){
    if _,ok := n.RunningTasks[keyTask] ; ok {
        delete(n.RunningTasks,keyTask)
        <- n.limit
    }
}

func (n * NodeClient)Treat(task * Task){
    n.RunningTasks[task.key] = task
    // Set task is recorded in list
    if _,err := http.Get(fmt.Sprintf("%s/treat?key=%s&task=%s&%s",n.url,task.key,task.task,task.parameters)) ; err != nil {
        logger.GetLogger().Error("When sending request",err)
        n.Stop()
    }
}

func (n NodeClient)Heartbeat()bool{
    if resp,err := http.Get(n.url) ; err == nil && resp != nil{
        return resp.StatusCode != 200
    }
    return false
}
