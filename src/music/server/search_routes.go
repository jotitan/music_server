package server

import (
	"encoding/json"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"github.com/jotitan/music_server/music"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (ms *MusicServer) getAllArtists(response http.ResponseWriter, request *http.Request) {
	begin := time.Now()
	genre := request.FormValue("genre")
	logger.GetLogger().Info("Get all artists", genre)
	// if genre exist, filter artist list
	filterArtist := ms.indexManager.SearchArtistByGenre(genre)

	// Response with name and url
	artists := music.LoadArtistIndex(ms.folder).FindAll()
	artistsData := make([]map[string]string, 0, len(artists))
	for artist, id := range artists {
		// test if artist id is in the filtered genre list
		if _, exist := filterArtist[id]; exist || (len(filterArtist) == 0 && genre == "") {
			artistsData = append(artistsData, map[string]string{"name": artist, "url": fmt.Sprintf("id=%d", id)})
		}
	}
	sort.Sort(music.SortByArtist(artistsData))
	bdata, _ := json.Marshal(artistsData)
	response.Write(bdata)
	logger.GetLogger().Info("Get all artists", genre,"in",time.Now().Sub(begin))
}

func (ms *MusicServer) getMusics(response http.ResponseWriter, request *http.Request, musicsIds []int, sortByTrack bool, fields []string) {
	// Get genre, if exist, filter music with
	genre := strings.ToLower(request.FormValue("genre"))
	musics := make([]map[string]interface{}, 0, len(musicsIds))
	for _, musicID := range musicsIds {
		m := ms.library.GetMusicInfo(int32(musicID))
		if genre == "" || strings.ToLower(m["genre"]) == genre {
			delete(m, "path") // Cause no need to return
			infos := make(map[string]string)
			infos["favorite"] = fmt.Sprintf("%t", ms.favorites.IsFavorite(musicID))
			infos["track"] = "#" + m["track"]
			infos["time"] = m["length"]
			// Recopy wanted fields
			for _, field := range fields {
				infos[field] = m[field]
			}
			musics = append(musics, map[string]interface{}{"name": m["title"], "id": fmt.Sprintf("%d", musicID), "infos": infos})
		}
	}
	if sortByTrack {
		sort.Sort(music.SortByAlbum(musics))
	}
	data, _ := json.Marshal(musics)
	response.Write(data)
}

func (ms *MusicServer) ListByArtist(response http.ResponseWriter, request *http.Request) {
	begin := time.Now()
	if id := request.FormValue("id"); id == "" {
		ms.getAllArtists(response, request)
	} else {
		artistID, _ := strconv.ParseInt(id, 10, 32)
		musicsIds := music.LoadArtistMusicIndex(ms.folder).MusicsByArtist[int(artistID)]
		ms.getMusics(response, request, musicsIds, false, []string{})
		logger.GetLogger().Info("Load music of artist", id,"in",time.Now().Sub(begin))
	}

}

func (ms *MusicServer) ListByOnlyAlbums(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		albumID, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
		musicsIds := ms.indexManager.ListFullAlbumById(int(albumID))
		ms.getMusics(response, request, musicsIds, true, []string{})
	default:
		albumsData := ms.indexManager.ListAllAlbums(request.FormValue("genre"))
		data, _ := json.Marshal(albumsData)
		response.Write(data)
	}
}

func (ms *MusicServer) ListByAlbum(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		artistID, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
		albumsData := ms.indexManager.ListAlbumByArtist(int(artistID))
		bdata, _ := json.Marshal(albumsData)
		response.Write(bdata)
	case request.FormValue("idAlbum") != "":
		albumID, _ := strconv.ParseInt(request.FormValue("idAlbum"), 10, 32)
		musicsIDs := ms.indexManager.ListAlbumById(int(albumID))
		ms.getMusics(response, request, musicsIDs, true, []string{})

	default:
		ms.getAllArtists(response, request)
	}
}

// Return info about music
func (ms *MusicServer) MusicInfo(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.ParseInt(request.FormValue("id"), 10, 32)
	logger.GetLogger().Info("Load music info with id", id)
	isfavorite := ms.favorites.IsFavorite(int(id))
	musicInfoData := ms.library.GetMusicInfoAsJSON(int32(id), isfavorite)
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Write(musicInfoData)
}

// Return info about many musics
func (ms *MusicServer) MusicsInfo(response http.ResponseWriter, request *http.Request) {
	var ids []int32
	json.Unmarshal([]byte(request.FormValue("ids")), &ids)
	//logger.GetLogger().Info("Load musics", len(ids))
	if request.FormValue("short") != "" {
		ms.getMusics(response, request, int32asInt(ids), false, []string{"artist"})
	}else {
		ms.musicsResponse(ids, response)
	}
}


//search musics by free text
func (ms *MusicServer) Search(response http.ResponseWriter, request *http.Request) {
	musics := ms.indexManager.SearchText(request.FormValue("term"), request.FormValue("size"))
	ms.musicsResponse(musics, response)
}

// Get informations from ids of music
func (ms *MusicServer) musicsResponse(ids []int32, response http.ResponseWriter) {
	musics := ms.library.GetMusicsInfo(ids)
	musicsExport := make([]map[string]string, 0, len(musics))
	for _, musicInfo := range musics {
		if musicInfo != nil && len(musicInfo) > 0 {
			delete(musicInfo, "path")
			musicInfo["src"] = fmt.Sprintf("music?id=%s", musicInfo["id"])
			musicsExport = append(musicsExport, musicInfo)
		}
	}
	bdata, _ := json.Marshal(musicsExport)
	response.Write(bdata)
}

// ???
func (ms *MusicServer) MusicsInfoInline(response http.ResponseWriter, request *http.Request) {
	strIds := strings.Split(request.FormValue("ids"), ",")
	ids := make([]int32, len(strIds))
	for i, strID := range strIds {
		if id, err := strconv.ParseInt(strID, 10, 32); err == nil {
			ids[i] = int32(id)
		} else {
			ids[i] = 0
		}
	}
	ms.musicsResponse(ids, response)
}


func (ms *MusicServer) NbMusics(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(fmt.Sprintf("%d", ms.library.GetNbMusics())))
}

func int32asInt(ids []int32)[]int{
	idsInt := make([]int,len(ids))
	for i,id := range ids {
		idsInt[i] = int(id)
	}
	return idsInt
}

func (ms *MusicServer) ListGenres(response http.ResponseWriter, request *http.Request) {
	data, _ := json.Marshal(ms.indexManager.ListGenres())
	response.Write(data)
}
