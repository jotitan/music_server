package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"github.com/jotitan/music_server/music"
	"math/rand"
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
	if id, err := strconv.Atoi(request.FormValue(idName)); err == nil {
		return music.GetShareConnection(id)
	}
	return nil
}

// Heartbeat is used to monitor connected service player. After specific timeout, cut connection
func (ms MusicServer) Heartbeat(response http.ResponseWriter, request *http.Request) {
	sessionID := sessionID(response, request)
	shareId, err := strconv.Atoi(request.FormValue("id"))

	if err != nil {
		logger.GetLogger().Error("Impossible to manage heartbeat, bad share id")
		return
	}
	// Search shared, if not exist or ids doesn't match, log error
	sc := music.GetShareConnection(shareId)
	if sc == nil {
		logger.GetLogger().Error("Impossible to manage heartbeat of", shareId)
		return
	}
	sc.NotifyHeartbeat(sessionID)
}

func (ms MusicServer) ShareService(response http.ResponseWriter, request *http.Request) {
	// Return id of shared
	id := music.CreateShareConnectionService(request.FormValue("device"), request.FormValue("url"), sessionID(response, request), ms.library.GetMusicsInfo)
	response.Write([]byte(fmt.Sprintf("%d", id)))
}

func (ms MusicServer) Share(response http.ResponseWriter, request *http.Request) {
	// If id is present, connect as clone
	if ss := getShare(request, "id"); ss != nil {
		// Create new SessionID at each connection ?
		ss.ConnectToShare(response, request.FormValue("device"), sessionIDWithOpts(response, request, false))
	} else {
		music.CreateShareConnection(response, request.FormValue("device"), sessionID(response, request))
	}
}

func (ms MusicServer) SendRequest(w http.ResponseWriter, r *http.Request) {
	if ss := getShare(r, "id"); ss != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		data, err := ss.SendRequest(sessionID(w, r), r.FormValue("event"), r.FormValue("data"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(data))
	} else {
		http.Error(w, "bad session id", http.StatusNotFound)
	}
}

func (ms MusicServer) ShareUpdate(response http.ResponseWriter, request *http.Request) {
	if ss := getShare(request, "id"); ss != nil {
		response.Header().Set("Access-Control-Allow-Origin", "*")
		ss.ForwardEvent(sessionID(response, request), request.FormValue("event"), request.FormValue("data"))
	}
}

func (ms MusicServer) ComputeLatency(response http.ResponseWriter, request *http.Request) {
	// Get original time (original_time), two differents times (local_receive & local_push) and add current
	originalTime := parseInt(request, "original_time")
	localReceive := parseInt(request, "local_receive")
	localPush := parseInt(request, "local_push")
	idShare := parseInt(request, "id")

	latency := music.ComputeLatency(originalTime, time.Now().UnixNano(), localReceive, localPush)
	if average, done := music.UpdateLatency(latency, int(idShare)); done {
		ss := getShare(request, "id")
		ss.SetLatency(average)
		fmt.Println("Average done : ", getSessionID(request), idShare, average)
	}
}

func parseInt(request *http.Request, name string) int64 {
	if value := request.FormValue(name); value != "" {
		if numValue, err := strconv.ParseInt(value, 10, 64); err == nil {
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
	return sessionIDWithOpts(response, request, false)
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
