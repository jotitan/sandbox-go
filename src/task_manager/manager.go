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
    "fmt"
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

func test(m * Manager){
    for i := 0 ; i < 20 ; i++ {
        c := make(chan GaugeData)
        m.term.CreateGauge(c,fmt.Sprintf("Title %d",i))
        c <- GaugeData{i+1,i+4}
        go func(cc chan GaugeData){
            time.Sleep(time.Second*time.Duration(2))
            cc <- GaugeData{-1,-1}
        }(c)
    }
}

func NewManager()Manager{
    // Create term logger
    t1,t2 := NewTermLogger()
    logger.InitComplexLogger(t1,t2,false)
    m :=  Manager{make(map[int]NodeClient),make(chan *Task,100),0,sync.Mutex{},0,NewTerm()}
    //test(&m)
    go m.CheckNodes()

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

    logger.GetLogger2().Info("Add new node",url,"with capacity",capacity)
    node := NewNode(id, url,capacity)
    m.nodes[id] = node
    m.term.ShowNodes(len(m.nodes))
    go node.Run(m.pendingTasks)
}

func (m * Manager)RemoveNode(id int){
    node := m.nodes[id]
    // running tasks will be canceled automaticaly and inject into pendingTasks by calling Stop method
    node.Stop()
    logger.GetLogger2().Info("Remove node",id,":",node.url)
    // Delete node
    delete(m.nodes,id)
    m.term.ShowNodes(len(m.nodes))
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
    if tc.treated == tc.total {
        tc.update <- GaugeData{-1,-1}
    }
}

func (m * Manager)createTaskControl(title string)*TaskControl{
    // Create task control and gauge
    taskControl := &TaskControl{update:make(chan GaugeData)}
    m.locker.Lock()
    m.currentIdTaskControl++
    taskControl.id = m.currentIdTaskControl
    taskControls[taskControl.id] = taskControl
    m.locker.Unlock()

    m.term.CreateGauge(taskControl.update,title)
    return taskControl
}

// Prefix is the part to remove to transfer, like c:\, specific to mount
// input folder to parse and found photo
// output folder where to produce resize photos
// force if picture must be resize again
// includeFirst : add folder in output name
func (m * Manager)ParseAndResizeFolder(prefix ,inputFolder, outputFolder string,includeFirst,force bool, taskControl *TaskControl){
    root := strings.Replace(inputFolder,prefix,"",-1)
    output := strings.Replace(outputFolder,prefix,"",-1)

    // Compare number of files into root and output
    sizeInput := getImageFilesNumber(inputFolder)
    sizeOutput := getImageFilesNumber(outputFolder)
    if sizeInput > 0 && sizeInput == sizeOutput && !force{
        // Already treated, return
        return
    }

    logger.GetLogger2().Info("Parse folder to resize",inputFolder,outputFolder)
    if dir,err := os.Open(inputFolder) ; err == nil {
        // Parse all files, dig into folder, no limit
        files,_ := dir.Readdir(-1)
        for _,file := range files {
            outputName := filepath.Join(output, file.Name())
            if includeFirst {
                outputName = filepath.Join(output, filepath.Base(inputFolder),file.Name())
            }
            if file.IsDir() {
                m.ParseAndResizeFolder(prefix,filepath.Join(inputFolder,file.Name()),outputName,false,force,taskControl)
            }else{
                if strings.HasSuffix(strings.ToLower(file.Name()),"jpg") || strings.HasSuffix(strings.ToLower(file.Name()),"jpeg") {
                    if taskControl == nil {
                        // Create task control and gauge
                        taskControl = m.createTaskControl("Resize task")
                    }
                    taskControl.total++
                    parameters := fmt.Sprintf("force=%v&in=%s&out=%s",force,url.QueryEscape(filepath.Join(root, file.Name())),url.QueryEscape(outputName))
                    m.pendingTasks <- &Task{task:"resize",idTaskControl:taskControl.id,parameters:parameters}
                }
            }
        }
    }
}

func getFilesNumber(folder string)int{
    if f,err := os.Open(folder) ; err == nil {
        defer f.Close()
        if files,err := f.Readdirnames(-1) ; err == nil {
            return len(files)
        }
    }
    return 0
}

// return number of image in folder (jpg or jpeg)
func getImageFilesNumber(folder string)int{
    if f,err := os.Open(folder) ; err == nil {
        defer f.Close()
        if files,err := f.Readdirnames(-1) ; err == nil {
            count := 0
            for _,file := range files {
                if strings.HasSuffix(strings.ToLower(file),".jpg") || strings.HasSuffix(strings.ToLower(file),".jpeg") {
                    count++
                }
            }
            return count
        }
    }
    return 0
}