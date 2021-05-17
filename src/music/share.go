package music

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"io/ioutil"
	"net/http"
	"strconv"
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
	// Used to monitor heartbeat of real service
	heartbeartChannel chan struct{}
}

func (ss *SharedSession)NotifyHeartbeat(sessionID string){
	if strings.EqualFold(sessionID,ss.original.sessionID) && ss.heartbeartChannel != nil {
		ss.heartbeartChannel <- struct{}{}
	}
}

func (ss *SharedSession)startHeartbeatChecker(sharedSessions map[int]*SharedSession){
	if ss.original.isBrowser {
		return
	}
	ss.heartbeartChannel = make(chan struct{})
	go func(){
		for{
			select {
			case <-ss.heartbeartChannel:
			case <- time.NewTimer(3*time.Minute).C:
				// To late, stop shareSession
				delete(sharedSessions,ss.id)
				logger.GetLogger().Info("Remove service session, timeout",ss.id)
				return
			}
		}
	}()
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
				newEvent, newData, success := ss.original.send(event, data)
				if success && !strings.EqualFold("",newEvent) {
					logger.GetLogger().Info("Send service new event", newEvent)
					// Send event to clone
					for _, clone := range ss.clones {
						clone.send(newEvent, newData)
					}
				}
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
	isBrowser bool
	// only if isBrowser false
	url          string
	getMusicsInfo func([]int32) []map[string]string
}

func (d Device) send(event string, data string) (newEvent, message string,success bool) {
	if d.isBrowser {
		return "", "",d.sendBrowser(event, data)
	}else{
		return d.sendService(event, data)
	}
}

func jsonToParams(data string)string{
	mapData := make(map[string]string)
	json.Unmarshal([]byte(data),&mapData)
	params := make([]string,0,len(mapData))
	for key,value := range mapData {
		params = append(params,fmt.Sprintf("%s=%s",key,value))
	}
	return strings.Join(params,"&")
}

func extractFieldFromJson(data, field string)interface{}{
	dataAsMap := make(map[string]interface{})
	json.Unmarshal([]byte(data),&dataAsMap)
	return dataAsMap[field]
}

// Send event thru api
func (d Device) sendService(event string, data string) (newEvent, message string, success bool) {
	// Work without data for pause, next, previous
	urlToCall := ""
	switch event {
	case "playMusic":
		// Extract field position from json and send as index
		urlToCall = fmt.Sprintf("%s/music/play?index=%d",d.url,int(extractFieldFromJson(data,"position").(float64)))
	case "play","pause","next","previous":
		urlToCall = fmt.Sprintf("%s/music/%s?%s",d.url,event,jsonToParams(data))
	case "add":
		return d.postMusicsToServer(data)
		// Get path to inject
		/*if id, err := strconv.ParseInt(data, 10,32) ; err == nil {
			musicInfo := d.getMusicsInfo([]int32{int32(id)})
			// Encode path
			urlToCall = fmt.Sprintf("%s/playlist/%s?id=%s&path=%s", d.url, event, data, url.PathEscape(musicInfo[0]["path"]))
		}*/
	case "askPlaylist":
		urlToCall = fmt.Sprintf("%s/playlist/state",d.url)
	case "volumeUp","volumeDown":
		urlToCall = fmt.Sprintf("%s/control/%s",d.url,event)
	case "remove","list","clean":
		urlToCall = fmt.Sprintf("%s/playlist/%s?%s",d.url,event,jsonToParams(data))
	}

	resp,err := http.Get(urlToCall)
	return manageServiceResponse(event,resp,err)
}

func manageServiceResponse(event string,resp *http.Response,err error ) (string,string, bool){
	if err == nil && resp.StatusCode == 200 {
		switch event {
		case "askPlaylist":
			if data,err := ioutil.ReadAll(resp.Body) ; err == nil {
				return "playlist",string(data),true
			}
		}
		return "", "", true

	}
	return "", "",false
}

func (d Device)postMusicsToServer(data string)(string,string,bool){
	// First unsplit
	musics := stringArrayToIntArray(data)
	// Load musics and create request
	musicsInfo := d.getMusicsInfo(musics)
	request := make([]map[string]string,len(musicsInfo))
	for pos, musicInfo := range musicsInfo {
		request[pos] = map[string]string{
			"id":   musicInfo["id"],
			"path": musicInfo["path"],
		}
	}
	dataRequest,_ := json.Marshal(request)
	postUrl := fmt.Sprintf("%s/playlist/add",d.url)

	resp,err := http.Post(postUrl,"application/json",bytes.NewBuffer(dataRequest))
	return manageServiceResponse("add",resp,err)
}

func stringArrayToIntArray(data string)[]int32{
	ids := strings.Split(data,",")
	musics := make([]int32,0,len(ids))
	for _,strId := range ids {
		if id,err := strconv.ParseInt(strId, 10, 32) ; err == nil {
			musics = append(musics, int32(id))
		}
	}
	return musics
}

func (d Device)isUp()bool{
	if d.isBrowser {
		return true
	}
	resp,err := http.Get(fmt.Sprintf("%s/health",d.url))
	return err == nil && resp.StatusCode == 200
}

func (d Device) sendBrowser(event string, data string) (success bool) {
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
	// Check if original still exist
	if !ss.original.isUp() {
		logger.GetLogger().Error("Impossible to connect to share")
		return
	}
	var device *Device
	logger.GetLogger().Info("Connect clone", ss.id, "with sessionID",sessionID)
	// Check if sessionID exist
	CreateSSEHeader(response)
	if v, dev := ss.isClone(sessionID); !v {
		//check device exist
		device = &Device{name: deviceName, sessionID: sessionID, response: response, connected: true, isBrowser: true}
		ss.clones = append(ss.clones, device)
	} else {
		dev.response = response
		device = dev
	}
	device.send("id", fmt.Sprintf("%d", ss.id))
	newEvent, data, success := ss.original.send("askPlaylist", "")
	if success && !strings.EqualFold(newEvent,"") {
		device.send(newEvent, data)
	}
	// Block until connexion is up
	checkConnection(device)
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
	//logger.GetLogger().Info("End device", d.sessionID)
}

func CreateShareConnectionService(deviceName, url, sessionID string, getMusicsInfo func([]int32)[]map[string]string)int{
	device := &Device{name: deviceName, sessionID: sessionID, connected: true, isBrowser: false, url : url, getMusicsInfo:getMusicsInfo}
	ss := &SharedSession{id: generateShareCode(), connected: true, original: device}
	sharedSessions[ss.id] = ss
	logger.GetLogger().Info("Create share service", ss.id)
	ss.startHeartbeatChecker(sharedSessions)
	return ss.id
}

//CreateShareConnection create an original connexion
func CreateShareConnection(response http.ResponseWriter, deviceName, sessionID string) {
	CreateSSEHeader(response)
	// Generate unique code to receive order
	device := &Device{name: deviceName, response: response, sessionID: sessionID, connected: true, isBrowser: true}
	ss := &SharedSession{id: generateShareCode(), connected: true, original: device}
	sharedSessions[ss.id] = ss
	logger.GetLogger().Info("Create share", ss.id)
	ss.original.send("id", fmt.Sprintf("%d", ss.id))
	//computeLatency(ss.original,ss.id)
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
