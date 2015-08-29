package node

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"strings"
)


type Stats struct {
	Memory int
	CPU int
	NbTaskers int
	NbTasks int
	Load float64
	Temperature float32
}

// NodeClient
type NodeClient struct {
	Url string
}

func NewNodeClient(url string)NodeClient{
	return NodeClient{url}
}

// CheckNode return true if node exist with service
func CheckNode(url string)bool{
	client := &http.Client{}
	client.Timeout = time.Duration(100)*time.Millisecond
	if resp,err := client.Get(fmt.Sprintf("%s/status",url)) ; err == nil {
		if data,err :=ioutil.ReadAll(resp.Body) ; err == nil {
			return strings.EqualFold(string(data),"Up")
		}
	}
	return false
}

// Register the current node to distant node
func (n NodeClient)Register(url string){
	http.DefaultClient.Get(fmt.Sprintf("%s/register?url=%s",n.Url,url))
}

// getLoad return the load of the node. All task are viewed with the same cost
func (n NodeClient)GetLoad()float64{
	if resp,err := http.DefaultClient.Get(fmt.Sprintf("%s/load",n.Url)) ; err == nil {
		 data,_ :=ioutil.ReadAll(resp.Body)
		 ret := make(map[string]interface{})
		 json.Unmarshal(data,&ret)
		 return ret["load"].(float64)
	 }
	return 0
}


// GetStats return stats about node : mem, cpu, load, nb tasks, available task thread
func (n NodeClient)GetStats()Stats{
 	if resp,err := http.DefaultClient.Get(fmt.Sprintf("%s/stats",n.Url)) ; err == nil && resp!=nil {
		data,_ := ioutil.ReadAll(resp.Body)
		stats := Stats{}
		json.Unmarshal(data,&stats)
		return stats
	}
	return Stats{}
}

func (n NodeClient)SendTask(task Task){
	base := fmt.Sprintf("type=%s&force=true",task.GetInfo().TypeTask)
	str:=strings.Join(append(task.Serialize(),base),"&")
	http.DefaultClient.Get(fmt.Sprintf("%s/add?%s",n.Url,str))
}
