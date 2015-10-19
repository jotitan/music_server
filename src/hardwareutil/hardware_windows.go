package hardwareutil

import (
	"runtime"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"strconv"
)


var currentUsage = 0
var previous = 0.0
var cpuStatsRun = false

const (
	waitTime = 3
	// Minimum memory (256Mo)
	defaultMemory = uint64(256*1024*1024)
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

// GetTemperature return the temperature
func GetTemperature()float32{
	return 0
}
