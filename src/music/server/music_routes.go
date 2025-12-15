package server

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"io"
	"net/http"
	"os"
	"strconv"
)

func (ms *MusicServer) RecreateIndex(response http.ResponseWriter, request *http.Request) {
	ms.library.RecreateIndex()
}

// Return music content
func (ms *MusicServer) Readmusic(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(request.FormValue("id"))
	logger.GetLogger().Info("Get music id", id)
	musicInfo := ms.library.GetMusicInfo(int32(id))

	m, _ := os.Open(musicInfo["path"])
	info, _ := m.Stat()
	logger.GetLogger().Info("load", musicInfo["path"])
	response.Header().Set("Content-type", "audio/mpeg")
	response.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	io.Copy(response, m)
}
