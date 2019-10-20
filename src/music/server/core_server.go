package server

import (
	"io"
	"io/ioutil"
	"logger"
	"music"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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


func (ms *MusicServer) Root(response http.ResponseWriter, request *http.Request) {
	proxyRedirect := request.Header.Get("proxy-redirect")
	ms.devices.Reset()
	if url := request.RequestURI; url == "/" || strings.Index(url, "/?") == 0 {
		// Reinit at each reload page
		ms.serveFile(response, request, filepath.Join(ms.webfolder, "music.html"),proxyRedirect)
		//http.ServeFile(response, request, filepath.Join(ms.webfolder, "music.html"))
	} else {
		if url == "/remote" {
			ms.serveFile(response, request, filepath.Join(ms.webfolder, "html/remote_control_full.html"),proxyRedirect)
			//http.ServeFile(response, request, filepath.Join(ms.webfolder, "html/remote_control_full.html"))
		}else {
			ms.serveFile(response, request, filepath.Join(ms.webfolder, url[1:]),proxyRedirect)
			//http.ServeFile(response, request, filepath.Join(ms.webfolder, url[1:]))
		}
	}
}

func (ms * MusicServer)serveFile(response http.ResponseWriter, request * http.Request, file, proxyRedirect string){
	if strings.HasSuffix(file,".html") && proxyRedirect != ""{
		// Rewrite all urls
		data,_ := ioutil.ReadFile(file)
		formatData := strings.Replace(strings.Replace(string(data),"src=\"","src=\"/" + proxyRedirect,-1),"href=\"","href=\"/" + proxyRedirect,-1)
		formatData = strings.Replace(formatData,"var basename=\"/\";","var basename=\"/" + proxyRedirect + "\";",-1)
		response.Write([]byte(formatData))
	}else{
		http.ServeFile(response, request, file)
	}

}

// Use to find node with very short timeout
func (ms MusicServer) Status(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Up"))
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

// Modify volumn of music on different server by calling a distant service on 9098
func (ms *MusicServer) Volume(response http.ResponseWriter, request *http.Request) {
	host := request.Host[:strings.Index(request.Host, ":")]
	ms.devices.SetVolume(request.FormValue("volume") == "down", host)
}


func (ms MusicServer) Create(port string, indexFolder, musicFolder, addressMask, webfolder string, routes func(ms * MusicServer)*http.ServeMux ) {
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

	mux := routes(&ms)

	logger.GetLogger().Info("Runner start on :", localIP, port)
	e := http.ListenAndServe(":"+port, mux)

	logger.GetLogger().Error("Runner ko", e)
}

// Load a resource like a cover
func (ms *MusicServer) Get(response http.ResponseWriter, request *http.Request) {
	url := request.FormValue("src")
	if f, e := os.Open(url); e == nil {
		defer f.Close()
		io.Copy(response, f)
	}
}

func (ms MusicServer) checkRequester(request *http.Request) bool {
	// Trouble with IPV6
	return true
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
