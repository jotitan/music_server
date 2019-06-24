package server

import (
	"fmt"
	"io"
	"logger"
	"net/http"
	"os"
	"strconv"
)

// Return music content
func (ms *MusicServer) Readmusic(response http.ResponseWriter, request *http.Request) {
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
