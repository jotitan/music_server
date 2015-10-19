package hardwareutil

import (
	"os/exec"
	"fmt"
	"os"
	"bytes"
	"strings"
	"strconv"
	"syscall"
	"regexp"
	"math"
)


const(
	// Minimum memory (256Mo)
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

// GetTemperature return the temperature
func GetTemperature()float32{
	cmd := exec.Command("/opt/vc/bin/vcgencmd","measure_temp")
	if data,err := cmd.Output() ; err == nil {
		if reg,e := regexp.Compile("[0-9]{1,2}\\.[0-9]{0,2}") ; e == nil {

			strTemp := reg.FindString(string(data))
			if temperature,e := strconv.ParseFloat(strTemp,32) ; e == nil {
				return float32(math.Ceil(temperature*10)/10)
			}
		}
	}
	return 0
}
