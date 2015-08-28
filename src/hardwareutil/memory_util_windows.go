package hardwareutil

import (
	"strconv"
	"os/exec"
	"strings"
)


/* Give tools to detect max memory */

const(
	// Minimum memory (2Go)
	defaultMemory = uint64(256*1024*1024)
)

// GetAvailableMemory return available memory on the system. Return half memory.
func GetAvailableMemory()uint64{

	// Use powershell instead of cmd on dows
	cmd := exec.Command("PowerShell","(Get-WMIObject Win32_PhysicalMemory | Measure-Object Capacity -Sum).Sum")
	data,_:= cmd.Output()
	if mem,err := strconv.ParseInt(strings.Replace(string(data),"\r\n","",-1),10,0) ; err == nil {
		return uint64(mem/2)
	}
	return defaultMemory
}

