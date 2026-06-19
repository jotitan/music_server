package reader

import (
	"encoding/json"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"io"
	"net/http"
	"strconv"
)

var playlist *Playlist
var reader *musicReader
var player *PlayerPlaylist

func InitLocalPlayer(deviceName, localUrl, urlMusicService string) {
	playlist = NewPlaylist()
	reader = NewMusicReader()
	player = NewPlayerPlaylist(playlist, reader, deviceName, localUrl, urlMusicService)
	player.connectToServer()
	player.runHeartBeat()
}

func Play(_ http.ResponseWriter, r *http.Request) {
	if index, err := strconv.Atoi(r.FormValue("index")); err == nil {
		player.ForceClose()
		player.Play(index)
	} else {
		// Call Pause or Play
		player.PauseOrPlayFirst()
	}
}

func VolumeUp(_ http.ResponseWriter, _ *http.Request) {
	player.UpdateVolume(0.5)
}
func VolumeDown(_ http.ResponseWriter, _ *http.Request) {
	player.UpdateVolume(-0.5)
}

func Health(w http.ResponseWriter, _ *http.Request) {
	logger.LogE(w.Write([]byte("UP")))
}

func Pause(_ http.ResponseWriter, _ *http.Request) {
	logger.LogE(reader.Pause())
}

func Next(_ http.ResponseWriter, _ *http.Request) {
	player.Next()
}

func Previous(_ http.ResponseWriter, _ *http.Request) {
	player.Previous()
}

func StopRadio(_ http.ResponseWriter, _ *http.Request) {
	player.StopRadio()
}

func Radio(_ http.ResponseWriter, r *http.Request) {
	radio := r.FormValue("data")
	player.Radio(radio)
}

func ForceClose(_ http.ResponseWriter, _ *http.Request) {
	player.ForceClose()
}

func Add(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)
	musics := make([]map[string]string, 0)
	if err := json.Unmarshal(data, &musics); err == nil {
		for _, music := range musics {
			if idMusic, err := strconv.Atoi(music["id"]); err == nil {
				playlist.Add(idMusic, music["path"])
			}
		}
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, err.Error(), 404)
	}
}

func Remove(_ http.ResponseWriter, r *http.Request) {
	if index, err := strconv.Atoi(r.FormValue("index")); err == nil {
		playlist.Remove(index)
	}
}

func Clean(_ http.ResponseWriter, _ *http.Request) {
	playlist.Clean()
}

func State(w http.ResponseWriter, _ *http.Request) {
	ids := make([]int, len(playlist.musics))
	for i, music := range playlist.musics {
		ids[i] = music.idMusic
	}
	state := struct {
		Ids      []int `json:"ids"`
		Current  int   `json:"current"`
		Playing  bool  `json:"playing"`
		Volume   int   `json:"volume"`
		Length   int64 `json:"length"`
		Position int64 `json:"position"`
		//radio and position
	}{
		ids,
		playlist.currentMusic,
		player.IsPlaying(),
		0,
		reader.playingMusicDetail.Length(),
		reader.playingMusicDetail.Pos(),
	}
	data, _ := json.Marshal(state)
	logger.LogE(w.Write(data))
}

func Current(w http.ResponseWriter, _ *http.Request) {
	logger.LogE(w.Write([]byte(fmt.Sprintf("{\"current\":%d}", playlist.currentMusic))))
}

func Reconnect(_ http.ResponseWriter, _ *http.Request) {
	player.connectToServer()
}

func List(w http.ResponseWriter, _ *http.Request) {
	list := make([]string, len(playlist.musics))
	for i, value := range playlist.musics {
		list[i] = value.path
	}
	data, _ := json.Marshal(list)
	logger.LogE(w.Write(data))
}
