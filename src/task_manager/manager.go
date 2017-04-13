package task_manager
import (
    "logger"
    "os"
    "strings"
    "path/filepath"
    "sync"
    "strconv"
    "net/url"
    "time"
)

/* Manage nodes and task */

type Task struct{
    // Compose with id running node and local id
    key string
    task string
    parameters string
    // used to monitor progression of big task
    idTaskControl int
}

type Manager struct {
    nodes map[int]NodeClient
    // Pending task in a chanel of task
    pendingTasks chan *Task
    currentId int
    locker sync.Mutex
    // current id of task control
    currentIdTaskControl int
    // Term manager
    term Term
}

func NewManager()Manager{
    m :=  Manager{make(map[int]NodeClient),make(chan *Task,100),0,sync.Mutex{},0,NewTerm()}
    go m.CheckNodes()

    update:=make(chan GaugeData)

    m.term.CreateGauge(update)
    update <- GaugeData{2,5}

    return m
}

// Run every 10 seconds to manage nodes
func (m * Manager)CheckNodes(){
    toRemove := make([]int,0,len(m.nodes))
    for key,node := range m.nodes {
        if !node.Heartbeat() {
            node.Stop()
            toRemove = append(toRemove,key)

        }
    }
    if len(toRemove) > 0 {
        // Nodes to delete
        for _,key:= range toRemove {
            m.RemoveNode(key)
        }
    }

    time.Sleep(time.Second*time.Duration(10))
    m.CheckNodes()
}

//AddNode : add a node which executes tasks.
func (m * Manager)AddNode(url string, capacity int){
    m.locker.Lock()
    id := m.currentId
    m.currentId++
    m.locker.Unlock()

    logger.GetLogger().Info("Add new node",url,"with capacity",capacity)
    node := NewNode(id, url,capacity)
    m.nodes[id] = node
    go node.Run(m.pendingTasks)
}

func (m * Manager)RemoveNode(id int){
    node := m.nodes[id]
    // running tasks will be canceled automaticaly and inject into pendingTasks by calling Stop method
    node.Stop()
    logger.GetLogger().Info("Remove node",id,":",node.url)
    // Delete node
    delete(m.nodes,id)
}

func (m Manager)CanTreat()bool{
    return len(m.nodes) > 0
}

// key is composed by idNode_idTask
func (m * Manager)ReceiveAck(key string){
    if !strings.Contains(key,"_") {
        // error
    }
    // Increase progression of big task
    idTaskControl,_ := strconv.ParseInt(strings.Split(key,"_")[0],10,32)
    taskControls[int(idTaskControl)].progress()

    idNode,_ := strconv.ParseInt(strings.Split(key,"_")[1],10,32)
    if node,ok := m.nodes[int(idNode)] ; ok {
        node.SetTreat(key)
    }
}

// Store for each task key, the task control linked
var taskControls = make(map[int]*TaskControl)

type TaskControl struct{
    treated int
    total int
    id int
    update chan GaugeData
}

func (tc * TaskControl)progress()  {
    tc.treated++
    tc.update <- GaugeData{tc.treated,tc.total}
}

// Prefix is the part to remove to transfer, like c:\, specific to mount
// input folder to parse and found photo
// output folder where to produce resize photos
func (m * Manager)ParseAndResizeFolder(prefix ,inputFolder, outputFolder string,taskControl *TaskControl){
    if taskControl == nil {
        // Create task control and gauge
        taskControl = &TaskControl{update:make(chan GaugeData)}
        m.locker.Lock()
        m.currentIdTaskControl++
        taskControl.id = m.currentIdTaskControl
        taskControls[taskControl.id] = taskControl
        m.locker.Unlock()

        m.term.CreateGauge(taskControl.update)
    }
    root := strings.Replace(inputFolder,prefix,"",-1)
    output := strings.Replace(outputFolder,prefix,"",-1)
    logger.GetLogger().Info("Parse folder to resize",inputFolder)
    if dir,err := os.Open(inputFolder) ; err == nil {
        // Parse all files, dig into folder, no limit
        files,_ := dir.Readdir(-1)
        for _,file := range files {
            if file.IsDir() {
                m.ParseAndResizeFolder(prefix,filepath.Join(inputFolder,file.Name()),filepath.Join(outputFolder,file.Name()),taskControl)
            }else{
                if strings.HasSuffix(strings.ToLower(file.Name()),"jpg") || strings.HasSuffix(strings.ToLower(file.Name()),"jpeg") {
                    taskControl.total++

                    parameters := "in=" + url.QueryEscape(filepath.Join(root, file.Name())) + "&out=" + url.QueryEscape(filepath.Join(output, file.Name()))
                    m.pendingTasks <- &Task{task:"resize",idTaskControl:taskControl.id,parameters:parameters}
                }
            }
        }
    }
}