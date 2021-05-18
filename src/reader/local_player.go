package reader

import (
	"errors"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/jotitan/music_server/logger"
	"io"
	"net/http"
	"os"
	"time"
)

type musicReader struct {
	control       *beep.Ctrl
	running       bool
	volumeManager *effects.Volume
	currentVolume float64
}

func NewMusicReader()*musicReader {
	return &musicReader{running: false}
}

func (mr * musicReader)Play(path string)error{
	logger.GetLogger().Info("Play",path)
	f,err:= os.Open(path)
	if err != nil {
		return err
	}
	return mr.readStream(f)
}

func (mr * musicReader)StopRadio(){
	speaker.Close()
	mr.running = false
}

func (mr * musicReader)PlayRadio(urlRadio string)error{
	logger.GetLogger().Info("Read radio",urlRadio)
	resp,err := http.Get(urlRadio)
	if err != nil {
		return err
	}
	return mr.readStream(resp.Body)
}

func (mr * musicReader)readStream(stream io.ReadCloser)error{
	defer stream.Close()
	streamer, format,err := mp3.Decode(stream)
	if err != nil {
		return err
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	mr.control = &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false}
	mr.volumeManager = &effects.Volume{Base:2,Volume: mr.currentVolume,Silent: false,Streamer: mr.control}
	mr.running = true

	// Detect end and read next music (chanel to playlist)
	end := make(chan bool)
	speaker.Play(beep.Seq(mr.volumeManager, beep.Callback(func() {
		end <- true
	})))
	<- end
	// For radio, impossible, must continue if stream stop
	return nil
}

func (mr * musicReader)Pause()error{
	logger.GetLogger().Info("Pause")
	if !mr.running {
		return errors.New("no music running")
	}
	speaker.Lock()
	mr.control.Paused = !mr.control.Paused
	speaker.Unlock()
	return nil
}

func (mr *musicReader) updateVolume(step float64) {
	speaker.Lock()
	logger.GetLogger().Info("Update volume",step,mr.volumeManager.Volume)
	mr.currentVolume+=step
	mr.volumeManager.Volume=mr.currentVolume
	speaker.Unlock()
}
