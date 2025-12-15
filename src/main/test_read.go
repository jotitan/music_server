package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/jotitan/music_server/logger"
	"os"
	"time"
)

func test() {
	fmt.Println(pa(1))
	fmt.Println(pa(2))
	fmt.Println(pa(0))
}

func pa(v int) *int {
	defer func() {
		if err := recover(); err != nil {
			logger.GetLogger().Error("Error when", err)
		}
	}()
	return throwPanic(v)
}

func throwPanic(v int) *int {
	if v == 0 {
		panic("Fuck of")
	}
	val := 1
	return &val
}

func main() {
	test()
	return
	// /home/osmc/test.mp3
	err := read(os.Args[1])
	logger.GetLogger().Info("UC", err)
}

func read(path string) error {
	logger.GetLogger().Info("Play")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(f)
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
	<-end
	logger.GetLogger().Info("End reach")
	return nil
}
