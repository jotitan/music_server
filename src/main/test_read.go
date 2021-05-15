package main

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/jotitan/music_server/logger"
	"os"
	"time"
)

func main() {
	// /home/osmc/test.mp3
	err := read(os.Args[1])
	logger.GetLogger().Info("UC",err)
}

func read(path string)error{
	logger.GetLogger().Info("Play")
	f,err:= os.Open(path)
	if err != nil {
		return err
	}
	streamer, format,err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	control := &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false}

	// Detect end and read next music (chanel to playlist)
	end := make(chan bool)
	logger.GetLogger().Info("Run music")
	speaker.Play(beep.Seq(control, beep.Callback(func() {
		logger.GetLogger().Info("STOP END")
		end <- true
	})))
	logger.GetLogger().Info("Wait end")
	<- end
	logger.GetLogger().Info("End reach")
	return nil
}
