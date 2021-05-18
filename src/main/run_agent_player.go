package main

import (
	"fmt"
	"github.com/jotitan/music_server/reader"
	"os"
)

func main() {
	if len(os.Args) != 5 {
		panic("need to specify <music_server> <device_name> <local_url> <port>")
	}
	// Register to music server
	urlMusicService := os.Args[1]	// http://localhost:9004
	deviceName := os.Args[2]		// local_service
	port := os.Args[4]				// 9007
	localUrl := fmt.Sprintf("%s:%s",os.Args[3],port)	//http://localhost:9007
	reader.InitLocalPlayer(deviceName, localUrl, urlMusicService)
	reader.RunLocalServer(port)
}
