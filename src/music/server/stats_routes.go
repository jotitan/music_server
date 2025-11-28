package server

import (
	"net/http"
	"time"
)

func (ms MusicServer) createSSEHeader(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*")
}

// StatsAsSSE return stats by server side event
func (ms MusicServer) StatsAsSSE(response http.ResponseWriter, request *http.Request) {
	ms.createSSEHeader(response)
	ms.sendStats(response)
}

func (ms MusicServer) sendStats(r http.ResponseWriter) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	stop := false
	go func() {
		<-r.(http.CloseNotifier).CloseNotify()
		stop = true
	}()

	for {
		r.Write([]byte("data: " + "hello" + "\n\n"))
		if stop == true {
			break
		}
		r.(http.Flusher).Flush()
		time.Sleep(1 * time.Second)
	}
}
