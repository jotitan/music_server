package music

import (
	"fmt"
	"math"
	"time"
)

/**
Give methods to compute latency between endpoints
1) Method to compute latency between server and client
2) Mecanism to synchronize shift between server and client (server tiome must be used at reference
 */


// Compute locally latency before set
var localLatency = make(map[int][]int64)


func SendLatencyRequest(device  * Device, id int){
 	device.send("check-latency",fmt.Sprintf("{\"time\":%d,\"id\":%d}",time.Now().UnixNano(),id))
 }

 // Two firsts are in nano, two lasts are in ms
 func ComputeLatency(originalTime, endTIme, localReceive, localPush int64)int64{
	clientLatency := localPush - localReceive
	return (endTIme - originalTime - clientLatency) / 1000000
 }


// Return average latency after all values received and true to specify latency is well computed, false otherwise
func UpdateLatency(latency int64, id int)(float64,bool){
	localLatency[id] = append(localLatency[id],latency)
	if len(localLatency[id]) == 10{
		// All latency received, compute average
		// Get max and min
		max := maxInSeries(localLatency[id])
		min := minInSeries(localLatency[id])
		average :=  float64(sumSeries(localLatency[id]) - max - min)/8
		delete(localLatency,id)
		return average, true
	}
	return 0,false
}

func sumSeries(list []int64)int64{
	sum := int64(0)
	for _,value := range list {
		sum+=value
	}
	return sum
}

func maxInSeries(list []int64)int64{
	max := int64(math.MinInt64)
	for _,value := range list {
		if max < value {
			max = value
		}
	}
	return max
}

func minInSeries(list []int64)int64{
	min := int64(math.MaxInt64)
	for _,value := range list {
		if min > value {
			min = value
		}
	}
	return min
}


// Id is used to know which connection is on measure
func computeLatency(device * Device, id int ){
	localLatency[id] = make([]int64,0,10)
	for i := 0 ; i < 10 ; i++ {
		SendLatencyRequest(device,id)
		time.Sleep(200*time.Millisecond)
	}
}