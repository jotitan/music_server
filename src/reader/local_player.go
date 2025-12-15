package reader

import (
	"errors"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/jotitan/music_server/logger"
	"io"
	"net/http"
	"os"
	"time"
)

type musicReader struct {
	control            *beep.Ctrl
	running            bool
	volumeManager      *effects.Volume
	currentVolume      float64
	isInitialize       bool
	initResampler      beep.SampleRate
	playingMusicDetail *musicDetail
}

type musicDetail struct {
	streamer beep.StreamSeekCloser
	rate     beep.SampleRate
}

func (md *musicDetail) Length() int64 {
	if md == nil {
		return 0
	}
	return md.rate.D(md.streamer.Len()).Milliseconds()
}

func (md *musicDetail) Pos() int64 {
	if md == nil {
		return 0
	}
	return md.rate.D(md.streamer.Position()).Milliseconds()
}

func NewMusicReader() *musicReader {
	return &musicReader{running: false}
}

func (mr *musicReader) Play(path string) error {
	logger.GetLogger().Info("Play", path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return mr.readStream(f)
}

func (mr *musicReader) StopRadio() {
	speaker.Close()
	mr.running = false
}

func (mr *musicReader) PlayRadio(urlRadio string) error {
	logger.GetLogger().Info("Read radio", urlRadio)
	resp, err := http.Get(urlRadio)
	if err != nil {
		return err
	}
	return mr.readStream(resp.Body)
}

func (mr *musicReader) ForceClose() {
	speaker.Clear()
}

func (mr *musicReader) initPlayer(format beep.Format) error {
	if mr.isInitialize {
		return nil
	}
	mr.isInitialize = true
	mr.initResampler = format.SampleRate
	return speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
}

func (mr *musicReader) readStream(stream io.ReadCloser) error {
	defer stream.Close()
	streamer, format, err := mp3.Decode(stream)
	if err != nil {
		return err
	}
	defer streamer.Close()
	err = mr.initPlayer(format)
	if err != nil {
		return err
	}
	resampler := beep.Resample(4, format.SampleRate, mr.initResampler, beep.Loop(1, streamer))
	//resampler := beep.Resample(4, mr.initResampler, format.SampleRate, beep.Loop(1, streamer))

	mr.control = &beep.Ctrl{Streamer: resampler, Paused: false}
	mr.playingMusicDetail = &musicDetail{streamer: streamer, rate: format.SampleRate}
	mr.volumeManager = &effects.Volume{Base: 2, Volume: mr.currentVolume, Silent: false, Streamer: mr.control}
	mr.running = true

	// Detect end and read next music (chanel to playlist)
	end := make(chan bool)
	speaker.Play(beep.Seq(mr.volumeManager, beep.Callback(func() {
		end <- true
	})))
	<-end
	// For radio, impossible, must continue if stream stop
	return nil
}

func (mr *musicReader) Pause() error {
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
	logger.GetLogger().Info("Update volume", step, mr.volumeManager.Volume)
	mr.currentVolume += step
	mr.volumeManager.Volume = mr.currentVolume
	speaker.Unlock()
}
