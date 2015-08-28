package hardwareutil

import (
	"os/exec"
	"time"
	"strings"
	"strconv"
	"os"
	"fmt"
	"runtime"
)

var currentUsage = 0
var previous = 0.0
var cpuStatsRun = false

const (
	waitTime = 3
)

// GetCPUUsage2 compute cpu usage by computing data at regular interval
func GetCPUUsage()int{
	// Loop not run, launch it
	if cpuStatsRun == false {
		cpuStatsRun = true
		go runCPUUsageStats()
	}
	return currentUsage
}

// Run in infinite loop and calc cpu usage for a processus by delta during a period (1 second)
func runCPUUsageStats(){
	nbCPU := float64(runtime.NumCPU())
	params := fmt.Sprintf("(Get-process -Id %d).CPU",os.Getpid())
	for {
		cmd := exec.Command("powershell", params)
		data, _ := cmd.Output()
		current,_ := strconv.ParseFloat(strings.Replace(string(data),"\r\n","",-1),32)
		if previous == 0 {
			previous = current
		}
		currentUsage = int(((current - previous)*float64(100))/(waitTime*nbCPU) )
		previous = current
		time.Sleep(time.Duration(waitTime )*time.Second)
	}
}
