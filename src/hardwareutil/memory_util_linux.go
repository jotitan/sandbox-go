package hardwareutil

import (
	"strconv"
	"os/exec"
	"strings"
	"bytes"
	//"syscall"
	"syscall"
)


/* Give tools to detect max memory */

const(
	// Minimum memory (2Go)
	defaultMemory = uint64(256*1024*1024)
)

// GetAvailableMemory return available memory on the system. Return half memory.
func GetAvailableMemory()uint64{
	// Check memory limit (ulimit -v call with Getrlimit)
	var r syscall.Rlimit

	if err := syscall.Getrlimit(syscall.RLIMIT_AS,&r) ; err == nil {
		return r.Cur / 2
	}
	// Check /proc/meminfo
	cmdGrep := exec.Command("grep", "MemTotal", "/proc/meminfo")
	cmdAwk := exec.Command("gawk", "{print $2}")
	cmdAwk.Stdin, _ = cmdGrep.StdoutPipe()
	buf := bytes.NewBuffer(make([]byte, 0))
	cmdAwk.Stdout = buf
	// Run asynchrone and wait Grep output
	cmdAwk.Start()
	cmdGrep.Run()
	cmdAwk.Wait()
	if mem, err := strconv.ParseInt(strings.Trim(string(buf.Bytes()), "\n"), 10, 0) ; err == nil {
		// Mem get in Kilo bytes, convert to bytes
		return uint64(mem / 2) * 1024
	}
	return defaultMemory
}
