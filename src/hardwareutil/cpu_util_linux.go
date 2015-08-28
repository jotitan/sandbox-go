package hardwareutil

import (
	"strconv"
	"strings"
	"os/exec"
	"os"
	"fmt"
	"bytes"
)


// GetCPUUsage compute CPU usage by using top
func GetCPUUsage()int{
	cmd := exec.Command("top","-p",fmt.Sprintf("%d",os.Getpid()),"-d","1","-b","-n","1")
	cmdGrep := exec.Command("grep","bernardo")
	cmdAwk := exec.Command("awk","{print $9}")

	cmdGrep.Stdin,_ = cmd.StdoutPipe()
	cmdAwk.Stdin,_ = cmdGrep.StdoutPipe()
	buf := bytes.NewBuffer(make([]byte,0))
	cmdAwk.Stdout = buf

	cmdGrep.Start()
	cmdAwk.Start()
	cmd.Run()
	cmdGrep.Wait()
	cmdAwk.Wait()

	if val,err := strconv.ParseInt(strings.Replace(string(buf.Bytes()),"\n","",-1),10,32) ; err == nil{
		return int(val)
	}
	return 0
}
