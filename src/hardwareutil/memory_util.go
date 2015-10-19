package hardwareutil

import "runtime"

// GetCurrentMemory return current load memory of procces
func GetCurrentMemory()int{
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	return int(mem.Alloc / 1024)
}
