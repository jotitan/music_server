package reader

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"net/http"
)

func RunLocalServer(port string){
	server := http.ServeMux{}
	server.HandleFunc("/health",Health)
	server.HandleFunc("/music/play",Play)
	server.HandleFunc("/music/pause",Pause)
	server.HandleFunc("/music/next",Next)
	server.HandleFunc("/music/previous",Previous)
	server.HandleFunc("/control/volumeUp",VolumeUp)
	server.HandleFunc("/control/volumeDown",VolumeDown)
	server.HandleFunc("/playlist/add",Add)
	server.HandleFunc("/radio/play",Radio)
	server.HandleFunc("/radio/stop",StopRadio)
	server.HandleFunc("/playlist/remove",Remove)
	server.HandleFunc("/playlist/clean",Clean)
	server.HandleFunc("/playlist/list",List)
	server.HandleFunc("/playlist/state",State)
	logger.GetLogger().Info("Start server on",port)
	http.ListenAndServe(fmt.Sprintf(":%s",port),&server)
}
