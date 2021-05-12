package reader

import (
	"errors"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"net/http"
	"time"
)

type Playlist struct{
	musics []musicLink
	currentMusic int
}

type musicLink struct {
	idMusic int
	path string
}

func NewPlaylist()*Playlist {
	return &Playlist{make([]musicLink,0),-1}
}

func (pl * Playlist)HasNext()bool {
	return pl.currentMusic+1 < len(pl.musics)
}

func (pl * Playlist)HasPrevious()bool {
	return pl.currentMusic != 0
}

func (pl * Playlist)SetCurrent(index int)(musicLink, error){
	if index > len(pl.musics) || index < 0{
		return musicLink{}, errors.New("music out of range")
	}
	pl.currentMusic = index
	return pl.musics[index], nil
}

func (pl * Playlist)Add(idMusic int, path string){
	logger.GetLogger().Info("Add music")
	pl.musics = append(pl.musics,musicLink{idMusic ,path})
}

func (pl * Playlist)Remove(indexMusic int){
	logger.GetLogger().Info("Remove music")
	if indexMusic == 0 {
		pl.musics = pl.musics[1:]
	}else {
		pl.musics = append(pl.musics[0:indexMusic-1], pl.musics[indexMusic:]...)
	}
}

func (pl * Playlist)Clean() {
	logger.GetLogger().Info("Clean playlist")
	pl.musics = make([]musicLink,0)
}

type PlayerPlaylist struct {
	playlist        *Playlist
	player          *musicReader
	deviceName      string
	sessionID       string
	shareID         string
	urlMusicService string
	localUrl        string
}

func NewPlayerPlaylist(playlist * Playlist, player * musicReader, deviceName, localUrl, urlMusicService string)*PlayerPlaylist{
	return &PlayerPlaylist{playlist: playlist, player: player, localUrl: localUrl, deviceName: deviceName, urlMusicService:urlMusicService}
}

func (pp *PlayerPlaylist)connectToServer(){
	pp.sessionID, pp.shareID = RegisterService(pp.urlMusicService,pp.deviceName,pp.localUrl)
}

func (pp *PlayerPlaylist)runHeartBeat(){
	// Send hearbeat request to musicserver
	go func(){
		for {
			urlToCall := fmt.Sprintf("%s/heartbeat?id=%s",pp.urlMusicService,pp.shareID)
			req,_ := http.NewRequest("GET",urlToCall,nil)
			req.Header.Add("Cookie",fmt.Sprintf("jsessionid=%s",pp.sessionID))
			resp,err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != 200 {
				// Trouble, try to reconnect after 10 seconds
				time.Sleep(time.Second*10)
				pp.connectToServer()
			}
			time.Sleep(time.Minute)
		}
	}()
}

func (pp *PlayerPlaylist) notifyCurrent(){
	// Specify which music is played with position and id
	position := pp.playlist.currentMusic
	id := pp.playlist.musics[position].idMusic
	urlToCall := fmt.Sprintf("%s/shareUpdate?id=%s&event=playMusic&data={\"position\":%d,\"id\":%d}",pp.urlMusicService, pp.shareID,position,id)
	// Add jsessionid in cookie
	req,_ := http.NewRequest("GET",urlToCall,nil)
	req.Header.Add("Cookie",fmt.Sprintf("jsessionid=%s",pp.sessionID))
	http.DefaultClient.Do(req)
}

func (pp * PlayerPlaylist)Next()bool{
	if playlist.HasNext() {
		logger.GetLogger().Info("Next music")
		music,_ := pp.playlist.SetCurrent(pp.playlist.currentMusic+1)
		pp.playWithEndDetection(music.path)
		return true
	}
	return false
}

func (pp * PlayerPlaylist)Previous(){
	if playlist.HasPrevious() {
		logger.GetLogger().Info("Previous music")
		music,_ := pp.playlist.SetCurrent(pp.playlist.currentMusic-1)
		pp.playWithEndDetection(music.path)
	}
}

func (pp * PlayerPlaylist)IsPause()bool {
	return pp.player.control != nil && pp.player.control.Paused
}

func (pp * PlayerPlaylist)Pause(){
	logger.GetLogger().Info("Pause")
	pp.player.Pause()
}

func (pp * PlayerPlaylist)playWithEndDetection(path string){
	go func(){
		pp.notifyCurrent()
		if err := pp.player.Play(path) ; err != nil {
			logger.GetLogger().Error("Impossible to read music",err)
		}
		if pp.Next() {
			pp.notifyCurrent()
		}
	}()
}

func (pp * PlayerPlaylist)Play(index int){
	if music, err := playlist.SetCurrent(index) ; err == nil {
		logger.GetLogger().Info("Play music",music.idMusic)
		pp.playWithEndDetection(music.path)
	}
}
