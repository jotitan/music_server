package main

import (
	"arguments"
	"encoding/json"
	"io"
	"logger"
	"music/server"
	"net/http"
	"path/filepath"
	"runtime"
)

/* Launch a server to treat resize image */

type SSEWriter struct {
	w io.Writer
	f http.Flusher
}

func (sse SSEWriter) Write(message string) {
	sse.w.Write([]byte("data: " + message + "\n\n"))
	sse.f.Flush()
}

/**
@basename : use when apache redirection is front of app. Basename is the /basename. Default is /
 */
func createRoutes(ms * server.MusicServer) *http.ServeMux {
	mux := http.NewServeMux()

	registerRoute(mux,"/status", "",ms.Status)
	registerRoute(mux,"/statsAsSSE", "", ms.StatsAsSSE)

	registerRoute(mux,"/music", "", ms.Readmusic)
	registerRoute(mux,"/nbMusics", "", ms.NbMusics)


	registerRoute(mux,"/get", "", ms.Get)

	// Manage search
	registerRoute(mux,"/genres", "", ms.ListGenres)
	registerRoute(mux,"/musicInfo", "", ms.MusicInfo)
	registerRoute(mux,"/musicsInfo", "", ms.MusicsInfo)
	registerRoute(mux,"/musicsInfoInline", "", ms.MusicsInfoInline)
	registerRoute(mux,"/listByArtist", "", ms.ListByArtist)
	registerRoute(mux,"/listByAlbum", "", ms.ListByAlbum)
	registerRoute(mux,"/listByOnlyAlbums", "", ms.ListByOnlyAlbums)
	registerRoute(mux,"/search", "", ms.Search)

	// Manage musics
	registerRoute(mux,"/index", "", ms.Index)
	registerRoute(mux,"/fullReindex", "", ms.FullReindex)

	// Manage favorites
	registerRoute(mux,"/setFavorite", "", ms.SetFavorite)
	registerRoute(mux,"/getFavorites", "", ms.GetFavorites)

	// Manage share device
	registerRoute(mux,"/share", "", ms.Share)
	registerRoute(mux,"/killshare", "", ms.KillShare)
	registerRoute(mux,"/shares", "", ms.GetShares)
	registerRoute(mux,"/shareUpdate", "", ms.ShareUpdate)
	registerRoute(mux,"/volume", "", ms.Volume)
	registerRoute(mux,"/latency", "", ms.ComputeLatency)

	// Serve files
	registerRoute(mux,"/help", "", Help)
	registerRoute(mux,"/", "", ms.Root)
	return mux
}

var routesDefinitions = make(map[string]string)

func registerRoute(mux * http.ServeMux,pattern, doc string, handler func(w http.ResponseWriter,r * http.Request)) {
	mux.HandleFunc(pattern,handler)
	routesDefinitions[pattern] = doc
}

func Help(w http.ResponseWriter,r * http.Request){
	if data,err := json.Marshal(routesDefinitions) ; err == nil {
		w.Write(data)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := arguments.ParseArgs()
	if !checkArguments(args) {
		return
	}
	port := args.GetString("port")

	if args.Has("log") {
		logger.InitLogger(filepath.Join(args.GetString("log"), "music_"+port+".log"), true)
	}
	logger.GetLogger().Info("Starting music server....")
	ms := server.MusicServer{}
	ms.Create(
		port,
		args.GetString("folder"),
		args.GetString("musicFolder"),
		args.GetString("addressMask"),
		args.GetString("webfolder"),
		createRoutes)
}

func checkArguments(args arguments.Arguments) bool {
	nbErrors := 0
	if !args.Has("folder") {
		logger.GetLogger().Info("Specify -folder : contains index")
		nbErrors++
	}
	if !args.Has("musicFolder") {
		logger.GetLogger().Info("Specify -musicFolder : contains musics to read")
		nbErrors++
	}
	if !args.Has("port") {
		logger.GetLogger().Info("Specify -port : port on which server runs")
		nbErrors++
	}
	return nbErrors == 0
}
