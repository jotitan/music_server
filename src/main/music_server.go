package main

import (
	"arguments"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"logger"
	"math/rand"
	"music"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

/* Launch a server to treat resize image */

type SSEWriter struct {
	w io.Writer
	f http.Flusher
}

//MusicServer manage all request of music server. Delegate to many objects
type MusicServer struct {
	// Folder where indexes are stored
	folder string
	// Folder where webresources are stored
	webfolder string
	// main music folder to update
	musicFolder string
	addressMask [4]int
	// Used to read musics
	library *music.MusicLibrary
	// manage index access
	indexManager *music.SearchIndex
	// Used to manage device host
	devices *music.Devices
	// Manage favorites
	favorites *music.FavoritesManager
}

func (sse SSEWriter) Write(message string) {
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}

func (ms *MusicServer) root(response http.ResponseWriter, request *http.Request) {
	ms.devices.Reset()
	if url := request.RequestURI; url == "/" {
		// Reinit at each reload page
		http.ServeFile(response, request, filepath.Join(ms.webfolder, "music.html"))
	} else {
		http.ServeFile(response, request, filepath.Join(ms.webfolder, url[1:]))
	}
}

// Use to find node with very short timeout
func (ms MusicServer) status(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Up"))
}

func (ms MusicServer) checkRequester(request *http.Request) bool {
	addr := request.RemoteAddr[:strings.LastIndex(request.RemoteAddr, ":")]
	if "[::1]" != addr {
		// [::1] means localhost. Otherwise, compare to mask
		for i, val := range strings.Split(addr, ".") {
			if intval, e := strconv.ParseInt(val, 10, 32); e != nil {
				logger.GetLogger().Error("User attempt to update data from outside", request.Host, request.RemoteAddr)
				return false
			} else {
				if int(intval)&ms.addressMask[i] != int(intval) {
					logger.GetLogger().Error("User attempt to update data from outside", request.Host, request.RemoteAddr)
					return false
				}
			}
		}
	}
	return true
}

// Create index used by search
func (ms *MusicServer) index(response http.ResponseWriter, request *http.Request) {
	// Always check addressMask. If no define, mask is 0.0.0.0 and nothing is accepted (except localhost)
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		textIndexer := music.IndexArtists(ms.folder)
		ms.indexManager.UpdateIndexer(textIndexer)
	}
}

// Redindex all data but keep all index in memories to increase treatment
func (ms *MusicServer) fullReindex(response http.ResponseWriter, request *http.Request) {
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		dico := music.LoadDictionnary(ms.folder)
		output := music.NewOutputDictionnary(ms.folder)
		textIndexer := dico.FullReindex(ms.musicFolder, output)
		ms.indexManager.UpdateIndexer(textIndexer)
		// Reload library
		ms.library = music.NewMusicLibrary(ms.folder)
	}
}

func (ms MusicServer) createSSEHeader(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "text/event-stream")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Connection", "keep-alive")
	response.Header().Set("Access-Control-Allow-Origin", "*")
}

// Return stats by server side event
func (ms MusicServer) statsAsSSE(response http.ResponseWriter, request *http.Request) {
	ms.createSSEHeader(response)
	ms.sendStats(response)
}

func (ms *MusicServer) getAllArtists(response http.ResponseWriter, request *http.Request) {
	genre := request.FormValue("genre")
	logger.GetLogger().Info("Get all artists", genre)
	// if genre exist, filter artist list
	filterArtist := ms.indexManager.SearchArtistByGenre(genre)

	// Response with name and url
	artists := music.LoadArtistIndex(ms.folder).FindAll()
	artistsData := make([]map[string]string, 0, len(artists))
	for artist, id := range artists {
		// test if artist id is in the filtered genre list
		if _, exist := filterArtist[id]; exist || (len(filterArtist) == 0 && genre == "") {
			artistsData = append(artistsData, map[string]string{"name": artist, "url": fmt.Sprintf("id=%d", id)})
		}
	}
	sort.Sort(music.SortByArtist(artistsData))
	bdata, _ := json.Marshal(artistsData)
	response.Write(bdata)
}

//Return all favorites as musics
func (ms MusicServer) getFavorites(response http.ResponseWriter, request *http.Request) {
	ms.getMusics(response, request, ms.favorites.GetFavorites(), false, []string{"artist"})
}

func (ms MusicServer) setFavorite(response http.ResponseWriter, request *http.Request) {
	if id, err := strconv.ParseInt(request.FormValue("id"), 10, 32); err == nil {
		favorite := request.FormValue("value") == "true"
		ms.favorites.Set(int(id), favorite)
		logger.GetLogger().Info("Update favorite", id, request.FormValue("value"))
		response.Write([]byte(fmt.Sprintf("{\"value\":%t}", favorite)))
	} else {
		response.Write([]byte("{\"error\":true}"))
	}
}

func (ms MusicServer) getMusics(response http.ResponseWriter, request *http.Request, musicsIds []int, sortByTrack bool, fields []string) {
	// Get genre, if exist, filter music with
	genre := strings.ToLower(request.FormValue("genre"))
	musics := make([]map[string]interface{}, 0, len(musicsIds))
	for _, musicID := range musicsIds {
		m := ms.library.GetMusicInfo(int32(musicID))
		if genre == "" || strings.ToLower(m["genre"]) == genre {
			delete(m, "path") // Cause no need to return
			infos := make(map[string]string)
			infos["favorite"] = fmt.Sprintf("%t", ms.favorites.IsFavorite(musicID))
			infos["track"] = "#" + m["track"]
			infos["time"] = m["length"]
			// Recopy wanted fields
			for _, field := range fields {
				infos[field] = m[field]
			}
			musics = append(musics, map[string]interface{}{"name": m["title"], "id": fmt.Sprintf("%d", musicID), "infos": infos})
		}
	}
	if sortByTrack {
		sort.Sort(music.SortByAlbum(musics))
	}
	data, _ := json.Marshal(musics)
	response.Write(data)
}

func (ms *MusicServer) listByArtist(response http.ResponseWriter, request *http.Request) {
	if id := request.FormValue("id"); id == "" {
		ms.getAllArtists(response, request)
	} else {
		logger.GetLogger().Info("Load music of artist", id)
		artistID, _ := strconv.ParseInt(id, 10, 32)
		musicsIds := music.LoadArtistMusicIndex(ms.folder).MusicsByArtist[int(artistID)]
		ms.getMusics(response, request, musicsIds, false, []string{})
	}
}

func (ms *MusicServer) listByOnlyAlbums(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		albumID, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
		musicsIds := ms.indexManager.ListFullAlbumById(int(albumID))
		ms.getMusics(response, request, musicsIds, true, []string{})
	default:
		albumsData := ms.indexManager.ListAllAlbums(request.FormValue("genre"))
		data, _ := json.Marshal(albumsData)
		response.Write(data)
	}
}

func (ms MusicServer) listByAlbum(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		artistID, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
		albumsData := ms.indexManager.ListAlbumByArtist(int(artistID))
		bdata, _ := json.Marshal(albumsData)
		response.Write(bdata)
	case request.FormValue("idAlbum") != "":
		albumID, _ := strconv.ParseInt(request.FormValue("idAlbum"), 10, 32)
		musicsIDs := ms.indexManager.ListAlbumById(int(albumID))
		ms.getMusics(response, request, musicsIDs, true, []string{})

	default:
		ms.getAllArtists(response, request)
	}
}

func writeCrossAccessHeader(response http.ResponseWriter) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
}

func (ms *MusicServer) listGenres(response http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(ms.indexManager.ListGenres())
	response.Write(data)
}

// Load a resource like a cover
func (ms MusicServer) get(response http.ResponseWriter, request *http.Request) {
	url := request.FormValue("src")
	if f, e := os.Open(url); e == nil {
		defer f.Close()
		io.Copy(response, f)
	}
}

// Return info about music
func (ms MusicServer) musicInfo(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
	logger.GetLogger().Info("Load music info with id", id)
	isfavorite := ms.favorites.IsFavorite(int(id))
	musicInfoData := ms.library.GetMusicInfoAsJSON(int32(id), isfavorite)
	writeCrossAccessHeader(response)
	response.Write(musicInfoData)
}

// Return info about many musics
func (ms MusicServer) musicsInfo(response http.ResponseWriter, request *http.Request) {
	var ids []int32
	json.Unmarshal([]byte(request.FormValue("ids")), &ids)
	logger.GetLogger().Info("Load musics", len(ids))

	ms.musicsResponse(ids, response)
}

// Get informations from ids of music
func (ms MusicServer) musicsResponse(ids []int32, response http.ResponseWriter) {
	musics := ms.library.GetMusicsInfo(ids)
	musicsExport := make([]map[string]string, 0, len(musics))
	for _, musicInfo := range musics {
		if musicInfo != nil && len(musicInfo) > 0 {
			delete(musicInfo, "path")
			musicInfo["src"] = fmt.Sprintf("music?id=%s", musicInfo["id"])
			musicsExport = append(musicsExport, musicInfo)
		}
	}
	bdata, _ := json.Marshal(musicsExport)
	response.Write(bdata)
}

func (ms MusicServer) musicsInfoInline(response http.ResponseWriter, request *http.Request) {
	strIds := strings.Split(request.FormValue("ids"), ",")
	ids := make([]int32, len(strIds))
	for i, strID := range strIds {
		if id, err := strconv.ParseInt(strID, 10, 32); err == nil {
			ids[i] = int32(id)
		} else {
			ids[i] = 0
		}
	}
	ms.musicsResponse(ids, response)
}

//search musics by free text
func (ms *MusicServer) search(response http.ResponseWriter, request *http.Request) {
	musics := ms.indexManager.SearchText(request.FormValue("term"), request.FormValue("size"))
	ms.musicsResponse(musics, response)
}

func (ms MusicServer) nbMusics(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(fmt.Sprintf("%d", ms.library.GetNbMusics())))
}

// Modify volumn of music on different server by calling a distant service on 9098
func (ms *MusicServer) volume(response http.ResponseWriter, request *http.Request) {
	host := request.Host[:strings.Index(request.Host, ":")]
	ms.devices.SetVolume(request.FormValue("volume") == "down", host)
}

// Return music content
func (ms MusicServer) readmusic(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
	logger.GetLogger().Info("Get music id", id)
	musicInfo := ms.library.GetMusicInfo(int32(id))

	m, _ := os.Open(musicInfo["path"])
	info, _ := m.Stat()
	logger.GetLogger().Info("load", musicInfo["path"])
	response.Header().Set("Content-type", "audio/mpeg")
	response.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	io.Copy(response, m)
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
	if id := getSessionID(request); id != "" {
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

func (ms MusicServer) getShares(response http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(music.GetSharesInfo())
	writeCrossAccessHeader(response)
	response.Write(data)
}

func (ms MusicServer) killShare(response http.ResponseWriter, request *http.Request) {
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

func (ms MusicServer) share(response http.ResponseWriter, request *http.Request) {
	// If id is present, connect as clone
	if ss := getShare(request, "id"); ss != nil {
		ss.ConnectToShare(response, request.FormValue("device"), sessionID(response, request))
	} else {
		music.CreateShareConnection(response, request.FormValue("device"), sessionID(response, request))
	}
}

func (ms MusicServer) shareUpdate(response http.ResponseWriter, request *http.Request) {
	if ss := getShare(request, "id"); ss != nil {
		writeCrossAccessHeader(response)
		ss.ForwardEvent(sessionID(response, request), request.FormValue("event"), request.FormValue("data"))
	}
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

func (ms MusicServer) findExposedURL() string {
	adr, _ := net.InterfaceAddrs()
	for _, a := range adr {
		if a.String() != "0.0.0.0" && !strings.Contains(a.String(), "127.0.0.1") {
			if idx := strings.Index(a.String(), "/"); idx != -1 {
				return a.String()[:idx]
			}
			return a.String()
		}
	}
	return "localhost"
}

func (ms MusicServer) create(port string, indexFolder, musicFolder, addressMask, webfolder string) {
	ms.folder = indexFolder
	ms.indexManager = music.NewSearchIndex(ms.folder)
	ms.library = music.NewMusicLibrary(ms.folder)
	ms.devices = music.NewDevices()
	ms.webfolder = "resources/"
	ms.favorites = music.NewFavoritesManager(ms.folder, 100)
	if musicFolder != "" {
		ms.musicFolder = musicFolder
		if addressMask != "" {
			for i, val := range strings.Split(addressMask, ".") {
				if intVal, e := strconv.ParseInt(val, 10, 32); e == nil {
					ms.addressMask[i] = int(intVal)
				}
			}
		}
	}
	if webfolder != "" {
		ms.webfolder = webfolder
	}

	if port == "" {
		logger.GetLogger().Fatal("Impossible to run node, port is not defined")
	}
	localIP := ms.findExposedURL()

	mux := ms.createRoutes()
	logger.GetLogger().Info("Runner ok on :", localIP, port)
	e := http.ListenAndServe(":"+port, mux)

	logger.GetLogger().Error("Runner ko", e)
}

func (ms *MusicServer) createRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", ms.status)
	mux.HandleFunc("/statsAsSSE", ms.statsAsSSE)

	mux.HandleFunc("/music", ms.readmusic)
	mux.HandleFunc("/nbMusics", ms.nbMusics)

	// Manage search
	mux.HandleFunc("/musicInfo", ms.musicInfo)
	mux.HandleFunc("/get", ms.get)
	mux.HandleFunc("/musicsInfo", ms.musicsInfo)
	mux.HandleFunc("/musicsInfoInline", ms.musicsInfoInline)
	mux.HandleFunc("/listByArtist", ms.listByArtist)
	mux.HandleFunc("/listByAlbum", ms.listByAlbum)
	mux.HandleFunc("/listByOnlyAlbums", ms.listByOnlyAlbums)
	mux.HandleFunc("/search", ms.search)

	// Manage musics
	mux.HandleFunc("/genres", ms.listGenres)
	mux.HandleFunc("/index", ms.index)
	mux.HandleFunc("/fullReindex", ms.fullReindex)

	// Manage favorites
	mux.HandleFunc("/setFavorite", ms.setFavorite)
	mux.HandleFunc("/getFavorites", ms.getFavorites)

	// Manage share device
	mux.HandleFunc("/share", ms.share)
	mux.HandleFunc("/killshare", ms.killShare)
	mux.HandleFunc("/shares", ms.getShares)
	mux.HandleFunc("/shareUpdate", ms.shareUpdate)
	mux.HandleFunc("/volume", ms.volume)

	// Serve files
	mux.HandleFunc("/", ms.root)
	return mux
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	port := args["port"]

	if logFolder, ok := args["log"]; ok {
		logger.InitLogger(filepath.Join(logFolder, "music_"+port+".log"), true)
	}

	ms := MusicServer{}
	ms.create(port, args["folder"], args["musicFolder"], args["addressMask"], args["webfolder"])
}
