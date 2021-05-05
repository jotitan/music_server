package music

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"net/http"
	"strings"
	"time"
)

// SharedSession represent a relation between an interface (original) and copy (clones)
type SharedSession struct {
	// Id of the share, different from sessionID
	id        int
	original  *Device
	clones    []*Device
	response  http.ResponseWriter
	connected bool
	latency float64
}

func (ss SharedSession) isOriginal(sessionID string) bool {
	return strings.EqualFold(ss.original.sessionID, sessionID)
}

func (ss SharedSession) isClone(sessionID string) (bool, *Device) {
	for _, clone := range ss.clones {
		if strings.EqualFold(clone.sessionID, sessionID) {
			return true, clone
		}
	}
	return false, nil
}

func (ss * SharedSession)SetLatency(latency float64){
	ss.latency = latency
}

func (ss SharedSession)GetLatency()float64{
	return ss.latency
}

//ForwardEvent from a remote control. Possible event : add music, remove music, play music, play, pause, next, previous
// Detect the sender from his session ID
func (ss SharedSession) ForwardEvent(sessionID string, event, data string) {
	logger.GetLogger().Info("Receive event", event, "from", sessionID,"(",len(ss.clones),")")
	// Detect sender
	if ss.isOriginal(sessionID) {
		// Send to all clone
		for _, clone := range ss.clones {
			clone.send(event, data)
		}
		if strings.EqualFold(event, "close") {
			// Remove share
			removeSharedSession(ss.id)
		}
	} else {
		if v, dev := ss.isClone(sessionID); v {
			if strings.EqualFold(event, "close") {
				dev.connected = false
				ss.removeClone(sessionID)

			} else {
				ss.original.send(event, data)
			}
		}
	}
}

//Device represent a shared session
type Device struct {
	name      string
	response  http.ResponseWriter
	sessionID string
	connected bool
}

func (d Device) send(event string, data string) (success bool) {
	defer func() {
		if e := recover(); e != nil {
			success = false
		}
	}()
	logger.GetLogger().Info("SEND", event, data, d.sessionID)
	if event != "" {
		d.response.Write([]byte(fmt.Sprintf("event: %s\n", event)))
	}
	d.response.Write([]byte("data: " + data + "\n\n"))
	d.response.(http.Flusher).Flush()
	success = true
	return
}

var sharedSessions = make(map[int]*SharedSession)

// Create standard header for SSE
func CreateSSEHeader(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*")
}

func removeSharedSession(id int) {
	logger.GetLogger().Info("Remove share", id)
	delete(sharedSessions, id)
}

//GetShareConnection return a shared connection from ID
func GetShareConnection(id int) *SharedSession {
	if ss, exist := sharedSessions[id]; exist {
		return ss
	}
	return nil
}

func (ss *SharedSession) removeClone(sessionID string) {
	for i, d := range ss.clones {
		if strings.EqualFold(d.sessionID, sessionID) {
			ss.clones = append(ss.clones[:i], ss.clones[i+1:]...)
			logger.GetLogger().Info("Remove clone", sessionID, ", ", len(ss.clones), "left")
			return
		}
	}
}

// create new connection at each time, no connection recup
func (ss *SharedSession) ConnectToShare(response http.ResponseWriter, deviceName, sessionID string) {
	var device *Device
	logger.GetLogger().Info("Connect clone", ss.id)
	// Check if sessionID exist
	CreateSSEHeader(response)
	if v, dev := ss.isClone(sessionID); !v {
		//check device exist
		device = &Device{name: deviceName, sessionID: sessionID, response: response, connected: true}
		ss.clones = append(ss.clones, device)
	} else {
		dev.response = response
		device = dev
	}
	device.send("id", fmt.Sprintf("%d", ss.id))
	ss.original.send("askPlaylist", "")
	checkConnection(device)
	// remove clone cause connection is lost and checkconnection ended
	ss.removeClone(sessionID)
}

func checkConnection(d *Device) {
	disconnect := false
	go func() {
		defer func() {
			if err := recover(); err != nil {
				disconnect = true
			}
		}()
		<-d.response.(http.CloseNotifier).CloseNotify()
		disconnect = true
	}()
	for {
		if !d.connected || disconnect {
			break
		}
		time.Sleep(5 * time.Second)
	}
	logger.GetLogger().Info("End device", d.sessionID)
	// Check disconnexion
}

//CreateShareConnection create an original connexion
func CreateShareConnection(response http.ResponseWriter, deviceName, sessionID string) {
	CreateSSEHeader(response)
	// Generate unique code to receive order
	device := &Device{name: deviceName, response: response, sessionID: sessionID, connected: true}
	ss := &SharedSession{id: generateShareCode(), connected: true, original: device}
	sharedSessions[ss.id] = ss
	logger.GetLogger().Info("Create share", ss.id)
	ss.original.send("id", fmt.Sprintf("%d", ss.id))
	computeLatency(ss.original,ss.id)
	checkConnection(device)
	removeSharedSession(ss.id)
}


type ShareInfo struct {
	Name string
	Id   int
}

func GetSharesInfo() []ShareInfo {
	shares := make([]ShareInfo, 0, len(sharedSessions))
	for id, ss := range sharedSessions {
		shares = append(shares, ShareInfo{Name: ss.original.name, Id: id})
	}
	return shares
}

// Generate unique code
func generateShareCode() int {
	return time.Now().Nanosecond()
}
