package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"logger"
	"math/rand"
	"music"
	"net/http"
	"strconv"
	"time"
)

func (ms MusicServer) GetShares(response http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(music.GetSharesInfo())
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Write(data)
}

func (ms MusicServer) KillShare(response http.ResponseWriter, request *http.Request) {
	if ss := getShare(request, "id"); ss != nil {
		ss.ForwardEvent(sessionID(response, request), "close", "")
	}
}

func getShare(request *http.Request, idName string) *music.SharedSession {
	if id, err := strconv.ParseInt(request.FormValue(idName), 10, 32); err == nil {
		return music.GetShareConnection(int(id))
	}
	return nil
}

func (ms MusicServer) Share(response http.ResponseWriter, request *http.Request) {
	// If id is present, connect as clone
	if ss := getShare(request, "id"); ss != nil {
		// Create new SessionID at each connection
		ss.ConnectToShare(response, request.FormValue("device"), sessionIDWithOpts(response, request,true))
	} else {
		music.CreateShareConnection(response, request.FormValue("device"), sessionID(response, request))
	}
}

func (ms MusicServer) ShareUpdate(response http.ResponseWriter, request *http.Request) {
	if ss := getShare(request, "id"); ss != nil {
		response.Header().Set("Access-Control-Allow-Origin", "*")
		ss.ForwardEvent(sessionID(response, request), request.FormValue("event"), request.FormValue("data"))
	}
}


func (me MusicServer)ComputeLatency(response http.ResponseWriter, request *http.Request){
	// Get original time (original_time), two differents times (local_receive & local_push) and add current
	originalTime := parseInt(request,"original_time")
	localReceive := parseInt(request,"local_receive")
	localPush := parseInt(request,"local_push")
	idShare := parseInt(request,"id")

	latency := music.ComputeLatency(originalTime,time.Now().UnixNano(),localReceive,localPush)
	if average,done := music.UpdateLatency(latency,int(idShare)) ; done {
		ss := getShare(request,"id")
		ss.SetLatency(average)
		fmt.Println("Average done : ",getSessionID(request),idShare,average)
	}
}

func parseInt(request * http.Request, name string)int64{
	if value := request.FormValue(name) ; value != "" {
		if numValue, err := strconv.ParseInt(value,10,64) ; err == nil {
			return numValue
		}
	}
	return 0
}

func getSessionID(request *http.Request) string {
	for _, c := range request.Cookies() {
		if c.Name == "jsessionid" {
			return c.Value
		}
	}
	return ""
}

func sessionID(response http.ResponseWriter, request *http.Request) string {
	return sessionIDWithOpts(response,request,false)
}

// If forceNew == true, create a new sessionID
func sessionIDWithOpts(response http.ResponseWriter, request *http.Request, forceNew bool) string {
	if id := getSessionID(request); !forceNew && id != "" {
		return id
	}

	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d-%d", time.Now().Nanosecond(), rand.Int())))
	hash := h.Sum(nil)
	hexValue := hex.EncodeToString(hash)
	logger.GetLogger().Info("Set cookie session", hexValue)
	http.SetCookie(response, &http.Cookie{Name: "jsessionid", Value: hexValue})
	return hexValue
}

