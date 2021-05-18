package server

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"net/http"
	"strconv"

)

//Return all favorites as musics
func (ms MusicServer) GetFavorites(response http.ResponseWriter, request *http.Request) {
	ms.getMusics(response, request, ms.favorites.GetFavorites(), false, []string{"artist"})
}

func (ms MusicServer) SetFavorite(response http.ResponseWriter, request *http.Request) {
	if id, err := strconv.ParseInt(request.FormValue("id"), 10, 32); err == nil {
		favorite := request.FormValue("value") == "true"
		ms.favorites.Set(int(id), favorite)
		logger.GetLogger().Info("Update favorite", id, request.FormValue("value"))
		response.Write([]byte(fmt.Sprintf("{\"value\":%t}", favorite)))
	} else {
		response.Write([]byte("{\"error\":true}"))
	}
}

