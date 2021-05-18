package server

import (
	"github.com/jotitan/music_server/music"
	"net/http"
)

/* Manage synchronous playing of same music */
/* A player join an other and play the same music with time correction on slave */


func SynchroShare(w http.ResponseWriter, r * http.Request){

//	from := getShare(r,"id")
//	to := getShare(r,"idMaster")


	// Create SSE to idShareFrom
	music.CreateSSEHeader(w)
	// Force master to communicate
}

