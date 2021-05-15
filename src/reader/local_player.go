package reader

import (
	"errors"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/jotitan/music_server/logger"
	"os"
	"time"
)

type musicReader struct{
	control *beep.Ctrl
	running bool
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
	streamer, format,err := mp3.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	mr.control = &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	mr.running = true

	// Detect end and read next music (chanel to playlist)
	end := make(chan bool)
	speaker.Play(beep.Seq(mr.control, beep.Callback(func() {
		end <- true
	})))
	logger.GetLogger().Info("Wait end music")
	<- end
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
