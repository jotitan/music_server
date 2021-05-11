package reader

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var playlist *Playlist
var reader *musicReader
var player *PlayerPlaylist

func InitLocalPlayer(deviceName, localUrl, urlMusicService string){
	playlist = NewPlaylist()
	reader = NewMusicReader()
	player = NewPlayerPlaylist(playlist, reader, deviceName, localUrl, urlMusicService)
	player.connectToServer()
	player.runHeartBeat()
}

func Play(_ http.ResponseWriter, r * http.Request){
	if index, err := strconv.ParseInt(r.FormValue("index"),10,32) ; err == nil {
		player.Play(int(index))
	}else{
		player.Play(0)
	}
}

func Health(w http.ResponseWriter, _ * http.Request){
	w.Write([]byte("UP"))
}

func Pause(_ http.ResponseWriter, _ * http.Request){
	reader.Pause()
}

func Next(_ http.ResponseWriter, _ * http.Request){
	player.Next()
}

func Previous(_ http.ResponseWriter, _ * http.Request){
	player.Previous()
}

func Add(_ http.ResponseWriter, r * http.Request){
	path := r.FormValue("path")
	if idMusic, err := strconv.ParseInt(r.FormValue("id"),10,32) ; err == nil {
		playlist.Add(int(idMusic), path)
	}
}

func Remove(w http.ResponseWriter, r * http.Request){
	if index, err := strconv.ParseInt(r.FormValue("index"),10,32) ; err == nil {
		playlist.Remove(int(index))
	}
}

func Clean(w http.ResponseWriter, r * http.Request){
	playlist.Clean()
}

func State(w http.ResponseWriter, r * http.Request) {
	ids := make([]int,len(playlist.musics))
	for i,music := range playlist.musics {
		ids[i] = music.idMusic
	}
	state := struct{
		Ids []int `json:"ids"`
		Current int `json:"current"`
		Playing bool `json:"playing"`
		Volume int `json:"volume"`
		//radio and position
	}{
		ids,
		playlist.currentMusic,
		player.IsPause(),
		0,
	}
	data,_ := json.Marshal(state)
	w.Write(data)
}

func List(w http.ResponseWriter, r * http.Request){
	list := make([]string,len(playlist.musics))
	for i,value := range playlist.musics {
		list[i] = value.path
	}
	data,_ := json.Marshal(list)
	w.Write(data)
}
