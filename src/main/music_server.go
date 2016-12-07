package main
import (
    "net/http"
	"runtime"
	"arguments"
	"logger"
	"fmt"
	"net"
	"strings"
	"time"
	"io"
	"path/filepath"
	"os"
	"encoding/json"
	"music"
	"strconv"
	"sort"
	"crypto/md5"
	"math/rand"
	"encoding/hex"
	"math"
)

/* Launch a server to treat resize image */

type SSEWriter struct{
	w io.Writer
	f http.Flusher
}

type MusicServer struct{
	folder string
	webfolder string
	// main music folder to update
	musicFolder string
	addressMask [4]int
	dico music.MusicDictionnary
	albumManager *music.AlbumManager
	textIndexer music.TextIndexer
}

func (sse SSEWriter)Write(message string){
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}

func (ms MusicServer)root(response http.ResponseWriter, request *http.Request){
	if url := request.RequestURI ; url == "/"{
		http.ServeFile(response,request,filepath.Join(ms.webfolder,"music.html"))
	}else{
		http.ServeFile(response,request,filepath.Join(ms.webfolder,url[1:]))
	}
}

// Use to find node with very short timeout
func (ms MusicServer)status(response http.ResponseWriter, request *http.Request){
	response.Write([]byte("Up"))
}

func (ms MusicServer)checkRequester(request *http.Request)bool{
	addr := request.RemoteAddr[:strings.LastIndex(request.RemoteAddr,":")]
	if "[::1]" != addr {
		// [::1] means localhost. Otherwise, compare to mask
		for i,val := range strings.Split(addr,".") {
			if intval,e := strconv.ParseInt(val,10,32) ; e!= nil {
				logger.GetLogger().Error("User attempt to update data from outside",request.Host,request.RemoteAddr)
				return false
			}else{
				if int(intval) & ms.addressMask[i] != int(intval){
					logger.GetLogger().Error("User attempt to update data from outside",request.Host,request.RemoteAddr)
					return false
				}
			}
		}
	}
	return true
}

// Create index used by search
func (ms * MusicServer)index(response http.ResponseWriter, request *http.Request){
	// Always check addressMask. If no define, mask is 0.0.0.0 and nothing is accepted (except localhost)
	if !ms.checkRequester(request){
		return
	}
	if ms.musicFolder!="" {
		ms.textIndexer = music.IndexArtists(ms.folder)
	}
}

// update local folder if exist
func (ms * MusicServer)update(response http.ResponseWriter, request *http.Request){
	// Always check addressMask. If no define, mask is 0.0.0.0 and nothing is accepted (except localhost)
	if !ms.checkRequester(request){
		return
	}
	if ms.musicFolder!="" {
		dico := music.LoadDictionnary(ms.folder)
		ms.textIndexer = dico.Browse(ms.musicFolder)
	}
}

// Redindex all data but keep all index in memories to increase treatment
func (ms MusicServer)fullReindex(response http.ResponseWriter, request *http.Request){
	if !ms.checkRequester(request){
		return
	}
	if ms.musicFolder!="" {
		dico := music.LoadDictionnary(ms.folder)
		dico.FullReindex(ms.musicFolder)
	}
}

func (ms MusicServer)createSSEHeader(response http.ResponseWriter){
	response.Header().Set("Content-Type","text/event-stream")
	response.Header().Set("Cache-Control","no-cache")
	response.Header().Set("Connection","keep-alive")
	response.Header().Set("Access-Control-Allow-Origin","*")
}

// Return stats by server side event
func (ms MusicServer)statsAsSSE(response http.ResponseWriter, request *http.Request){
	ms.createSSEHeader(response)
	ms.sendStats(response)
}

type sortByArtist []map[string]string
type sortByAlbum []map[string]interface{}

func (a sortByArtist)Len() int{return len(a)}
func (a sortByArtist)Less(i, j int) bool{return strings.ToLower(a[i]["name"]) < strings.ToLower(a[j]["name"])}
func (a sortByArtist)Swap(i, j int) {a[i],a[j] = a[j],a[i]}

func (a sortByAlbum)Len() int{return len(a)}
func (a sortByAlbum)Less(i, j int) bool{
	infos1 := a[i]["infos"].(map[string]string)
	infos2 := a[j]["infos"].(map[string]string)
	t1,_ := strconv.ParseInt(infos1["track"][1:],10,32)
	t2,_ := strconv.ParseInt(infos2["track"][1:],10,32)
	return t1 < t2
}
func (a sortByAlbum)Swap(i, j int) {a[i],a[j] = a[j],a[i]}

func (ms MusicServer)getAllArtists(response http.ResponseWriter){
	logger.GetLogger().Info("Get all artists")
	// Response with nampe and url
	artists := music.LoadArtistIndex(ms.folder).FindAll()
	artistsData := make([]map[string]string,0,len(artists))
	for artist,id := range artists{
		artistsData = append(artistsData,map[string]string{"name":artist,"url":fmt.Sprintf("id=%d",id)})
	}
	sort.Sort(sortByArtist(artistsData))
	bdata,_ := json.Marshal(artistsData)
	response.Write(bdata)
}

func (ms MusicServer)getMusics(response http.ResponseWriter,musicsIds []int,sortByTrack bool){
	musics := make([]map[string]interface{},len(musicsIds))
	for i,musicId := range musicsIds{
		m := ms.dico.GetMusicFromId(musicId)
		delete(m,"path")	// Cause no need to return
		infos := make(map[string]string)
		infos["track"] = "#" + m["track"]
		infos["time"] = m["length"]
		musics[i] = map[string]interface{}{"name":m["title"],"id":fmt.Sprintf("%d",musicId),"infos":infos}
	}
	if sortByTrack {
	   sort.Sort(sortByAlbum(musics))
	}
	data,_:= json.Marshal(musics)
	response.Write(data)
}

func (ms MusicServer)listByArtist(response http.ResponseWriter, request *http.Request){
	if id := request.FormValue("id") ; id == "" {
		ms.getAllArtists(response)
	}else{
		logger.GetLogger().Info("Load music of artist",id)
		artistId,_ := strconv.ParseInt(id,10,32)
		musicsIds := music.LoadArtistMusicIndex(ms.folder).MusicsByArtist[int(artistId)]
		ms.getMusics(response,musicsIds,false)
	}
}

func (ms MusicServer)listByOnlyAlbums(response http.ResponseWriter, request *http.Request){
	switch{
		// return albums of artist
		case request.FormValue("id") != "" :
		logger.GetLogger().Info("Get musics of album")
		idAlbum,_ := strconv.ParseInt(request.FormValue("id"),10,32)
		    musics := ms.albumManager.GetMusicsAll(int(idAlbum))
			ms.getMusics(response,musics,true)
	default :
		albums := ms.albumManager.LoadAllAlbums()
		albumsData := make([]map[string]string,0,len(albums))
		for album,id := range albums{
			albumsData = append(albumsData,map[string]string{"name":album,"url":fmt.Sprintf("id=%d",id)})
		}
		sort.Sort(sortByArtist(albumsData))
		data,_ := json.Marshal(albumsData)
		response.Write(data)
	}

}

func (ms MusicServer)listByAlbum(response http.ResponseWriter, request *http.Request){
	 switch{
		 // return albums of artist
	  case request.FormValue("id") != "" :
	  logger.GetLogger().Info("Get all albums")
	  idArtist,_:= strconv.ParseInt(request.FormValue("id"),10,32)
		albums := music.NewAlbumByArtist().GetAlbums(ms.folder,int(idArtist))
	  	albumsData := make([]map[string]string,0,len(albums))
		for _,album := range albums{
			albumsData = append(albumsData,map[string]string{"name":album.Name,"url":fmt.Sprintf("idAlbum=%d",album.Id)})
		}
		  sort.Sort(sortByArtist(albumsData))
		  bdata,_ := json.Marshal(albumsData)
		  response.Write(bdata)
	  case request.FormValue("idAlbum") != "" :
	  	idAlbum,_ := strconv.ParseInt(request.FormValue("idAlbum"),10,32)
	  	  musics := ms.albumManager.GetMusics(int(idAlbum))
	  	ms.getMusics(response,musics,true)

	  default : ms.getAllArtists(response)
	}

}

func writeCrossAccessHeader(response http.ResponseWriter){
	response.Header().Set("Access-Control-Allow-Origin","*")
}

// Return info about music
func (ms MusicServer)musicInfo(response http.ResponseWriter, request *http.Request){
	id,_ := strconv.ParseInt(request.FormValue("id"),10,32)
	logger.GetLogger().Info("Load music info with id",id)
	musicInfo := ms.dico.GetMusicFromId(int(id))
	delete(musicInfo,"path")
	musicInfo["id"] = fmt.Sprintf("%d",id)
	musicInfo["src"] = fmt.Sprintf("music?id=%d",id)
	bdata,_ := json.Marshal(musicInfo)
	writeCrossAccessHeader(response)
	response.Write(bdata)
}

// Return info about many musics
func (ms MusicServer)musicsInfo(response http.ResponseWriter, request *http.Request){
	var ids []int
	json.Unmarshal([]byte(request.FormValue("ids")),&ids)
	logger.GetLogger().Info("Load musics",len(ids))

	ms.musicsResponse(ids,response)
}

func (ms MusicServer)musicsResponse(ids []int,response http.ResponseWriter){
	musics := ms.dico.GetMusicsFromIds(ids)
	for _,musicInfo := range musics{
		delete(musicInfo,"path")
		musicInfo["src"] = fmt.Sprintf("music?id=%s",musicInfo["id"])
	}
	bdata,_ := json.Marshal(musics)
	response.Write(bdata)
}

func (ms MusicServer)browse(response http.ResponseWriter, request *http.Request){
	folder := request.FormValue("folder")
	ms.dico.Browse(folder)
}

//search musics by free text
func (ms * MusicServer)search(response http.ResponseWriter, request *http.Request){
	text := request.FormValue("term")
	size := float64(10)
	if s := request.FormValue("size") ; s != "" {
		if intSize,e := strconv.ParseInt(s,10,32) ; e == nil {
			size = float64(intSize)
		}
	}
	musics := ms.textIndexer.Search(text)
	logger.GetLogger().Info("Search",text,len(musics))
	ms.musicsResponse(musics[:int(math.Min(size,float64(len(musics))))],response)
}

func (ms MusicServer)nbMusics(response http.ResponseWriter, request *http.Request){
	response.Write([]byte(fmt.Sprintf("%d",music.GetNbMusics(ms.folder))))
}

// Return music content
func (ms MusicServer)readmusic(response http.ResponseWriter, request *http.Request){
	id,_ := strconv.ParseInt(request.FormValue("id"),10,32)
	logger.GetLogger().Info("Get music id",id)
	musicInfo := ms.dico.GetMusicFromId(int(id))

	m,_ := os.Open(musicInfo["path"])
	info,_ := m.Stat()
	logger.GetLogger().Info("load",musicInfo["path"])
	response.Header().Set("Content-type","audio/mpeg")
	response.Header().Set("Content-Length",fmt.Sprintf("%d",info.Size()))
	io.Copy(response,m)
}

func getSessionID(request *http.Request)string {
	for _, c := range request.Cookies() {
		if c.Name == "jsessionid" {
			return c.Value
		}
	}
	return ""
}

func sessionID(response http.ResponseWriter,request *http.Request)string{
	if id := getSessionID(request) ; id != ""{
		return id
	}
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d-%d",time.Now().Nanosecond(),rand.Int())))
	hash := h.Sum(nil)
	hexValue := hex.EncodeToString(hash)
	logger.GetLogger().Info("Set cookie session",hexValue)
	http.SetCookie(response,&http.Cookie{Name:"jsessionid",Value:hexValue})
	return hexValue
}

func (ms MusicServer)getShares(response http.ResponseWriter, request *http.Request){
	data,_ := json.Marshal(music.GetSharesInfo())
	writeCrossAccessHeader(response)
	response.Write(data)
}

func (ms MusicServer)killShare(response http.ResponseWriter, request *http.Request){
	if ss := getShare(request,"id") ; ss!=nil{
		ss.ForwardEvent(sessionID(response,request),"close","")
	}
}

func getShare(request *http.Request, idName string)*music.SharedSession {
	if id,err := strconv.ParseInt(request.FormValue(idName),10,32) ; err == nil {
		return music.GetShareConnection(int(id))
	}
	return nil
}

func (ms MusicServer)share(response http.ResponseWriter, request *http.Request){
	// If id is present, connect as clone
	if ss := getShare(request,"id") ; ss!=nil{
		ss.ConnectToShare(response, request.FormValue("device"),sessionID(response,request))
	}else {
		music.CreateShareConnection(response, request.FormValue("device"),sessionID(response, request))
	}
}

func (ms MusicServer)shareUpdate(response http.ResponseWriter, request *http.Request){
	if ss := getShare(request,"id") ; ss!=nil{
		writeCrossAccessHeader(response)
		ss.ForwardEvent(sessionID(response,request),request.FormValue("event"),request.FormValue("data"))
	}
}

func (ms MusicServer)sendStats(r http.ResponseWriter){
	defer func(){
		if err := recover() ; err != nil {}
	}()
	stop := false
	go func(){
		<-r.(http.CloseNotifier).CloseNotify()
		stop=true
	}()

	for {
		r.Write([]byte("data: " + "hello" + "\n\n"))
		if stop == true{
			break
		}
		r.(http.Flusher).Flush()
		time.Sleep(1 * time.Second)
	}
}

func (ms MusicServer)findExposedURL()string{
	adr,_ := net.InterfaceAddrs()
	for _,a := range adr {
		if a.String() != "0.0.0.0" && !strings.Contains(a.String(),"127.0.0.1"){
			if idx := strings.Index(a.String(),"/"); idx != -1 {
				return a.String()[:idx]
			}
			return a.String()
		}
	}
	return "localhost"
}

func (ms MusicServer)create(port string,indexFolder,musicFolder,addressMask,webfolder string){
	ms.folder = indexFolder
	ms.textIndexer = music.LoadTextIndexer(ms.folder)
	ms.albumManager = music.NewAlbumManager(ms.folder)
	ms.webfolder = "resources/"
	if musicFolder != "" {
		ms.musicFolder = musicFolder
		if addressMask!=""{
			for i,val := range strings.Split(addressMask,".") {
				if intVal,e := strconv.ParseInt(val,10,32);e == nil {
					ms.addressMask[i] = int(intVal)
				}
			}
		}
	}
	if webfolder != ""{
		ms.webfolder = webfolder
	}
	ms.dico = music.LoadDictionnary(ms.folder)
	if port == ""{
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := ms.findExposedURL()

	mux := ms.createRoutes()
	logger.GetLogger().Info("Runner ok on :",localIP,port)
    http.ListenAndServe(":" + port,mux)

	logger.GetLogger().Error("Runner ko")
}

func (ms *MusicServer)createRoutes()*http.ServeMux{
	mux := http.NewServeMux()

	mux.HandleFunc("/status",ms.status)
	mux.HandleFunc("/statsAsSSE",ms.statsAsSSE)

	mux.HandleFunc("/music",ms.readmusic)
	mux.HandleFunc("/nbMusics",ms.nbMusics)
	mux.HandleFunc("/musicInfo",ms.musicInfo)
	mux.HandleFunc("/musicsInfo",ms.musicsInfo)
	mux.HandleFunc("/listByArtist",ms.listByArtist)
	mux.HandleFunc("/listByAlbum",ms.listByAlbum)
	mux.HandleFunc("/listByOnlyAlbums",ms.listByOnlyAlbums)
	mux.HandleFunc("/browse",ms.browse)
	mux.HandleFunc("/search",ms.search)

	mux.HandleFunc("/update",ms.update)
	mux.HandleFunc("/index",ms.index)
	mux.HandleFunc("/fullReindex",ms.fullReindex)

	mux.HandleFunc("/share",ms.share)
	mux.HandleFunc("/killshare",ms.killShare)
	mux.HandleFunc("/shares",ms.getShares)
	mux.HandleFunc("/shareUpdate",ms.shareUpdate)
	mux.HandleFunc("/",ms.root)
	return mux
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]

	if logFolder, ok := args["log"] ; ok {
		logger.InitLogger(filepath.Join(logFolder, "music_"+port+".log"), true)
	}

	ms := MusicServer{}
	ms.create(port,args["folder"],args["musicFolder"],args["addressMask"],args["webfolder"])
}
