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

func (ms *MusicServer) findAlbumsFromTerm(response http.ResponseWriter, request *http.Request) {
	term := request.FormValue("term")
	albums := ms.indexManager.SearchAlbumsByTerm(term)
	data, _ := json.Marshal(albums)
	logger.LogE(response.Write(data))
}

func (ms *MusicServer) findArtistsFromTerm(response http.ResponseWriter, request *http.Request) {
	term := request.FormValue("artist")
	artists := ms.indexManager.SearchArtistsByTerm(term)

	artistsData := make([]map[string]string, 0, len(artists))
	for id, artist := range artists {
		artistsData = append(artistsData, map[string]string{"name": artist, "id": fmt.Sprintf("%d", id)})
	}
	sort.Sort(music.SortByArtist(artistsData))
	data, _ := json.Marshal(artistsData)
	logger.LogE(response.Write(data))
}

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
	logger.LogE(response.Write(bdata))
	logger.GetLogger().Info("Get all artists", genre, "in", time.Now().Sub(begin))
}

func (ms *MusicServer) getMusics(response http.ResponseWriter, request *http.Request, musicsIds []int, sortByTrack bool, fields []string) {
	// Get genre, if exists, filter music with
	genre := strings.ToLower(request.FormValue("genre"))
	musics := make([]map[string]interface{}, 0, len(musicsIds))
	fields = append(fields, []string{"album", "artist"}...)
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
				if field == "src" {
					infos["src"] = fmt.Sprintf("music?id=%d", musicID)
				} else {
					infos[field] = m[field]
				}
			}
			musics = append(musics, map[string]interface{}{"name": m["title"], "id": fmt.Sprintf("%d", musicID), "infos": infos})
		}
	}
	if sortByTrack {
		sort.Sort(music.SortByAlbum(musics))
	}
	data, _ := json.Marshal(musics)
	logger.LogE(response.Write(data))
}

func (ms *MusicServer) ListArtists(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("term") != "":
		ms.findAlbumsFromTerm(response, request)
	case request.FormValue("artist") != "":
		ms.findArtistsFromTerm(response, request)
	default:
		ms.getAllArtists(response, request)
	}
}

func (ms *MusicServer) ListMusicsByArtist(response http.ResponseWriter, request *http.Request) {
	begin := time.Now()

	if id := request.FormValue("id"); id == "" {
		ms.getAllArtists(response, request)
	} else {
		artistID, _ := strconv.Atoi(id)
		musicsIds := music.LoadArtistMusicIndex(ms.folder).MusicsByArtist[artistID]
		fields := make([]string, 0)
		if request.FormValue("detail") == "true" {
			fields = append(fields, "src")
		}
		ms.getMusics(response, request, musicsIds, false, fields)
		logger.GetLogger().Info("Load music of artist", id, "in", time.Now().Sub(begin))
	}

}

func (ms *MusicServer) ListByOnlyAlbums(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		albumID, _ := strconv.Atoi(request.FormValue("id"))
		musicsIds := ms.indexManager.ListFullAlbumById(albumID)
		ms.getMusics(response, request, musicsIds, true, []string{})
	default:
		albumsData := ms.indexManager.ListAllAlbums(request.FormValue("genre"))
		data, _ := json.Marshal(albumsData)
		logger.LogE(response.Write(data))
	}
}

func (ms *MusicServer) ListMusicsByAlbum(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("id") != "":
		artistID, _ := strconv.Atoi(request.FormValue("id"))
		albumsData := ms.indexManager.ListAlbumByArtist(artistID)
		bdata, _ := json.Marshal(albumsData)
		logger.LogE(response.Write(bdata))
	case request.FormValue("idAlbum") != "":
		albumID, _ := strconv.Atoi(request.FormValue("idAlbum"))
		musicsIDs := ms.indexManager.ListAlbumById(albumID)
		ms.getMusics(response, request, musicsIDs, true, []string{})
	default:
		albumsData := ms.indexManager.ListAllAlbums(request.FormValue("genre"))
		data, _ := json.Marshal(albumsData)
		logger.LogE(response.Write(data))
	}
}

func (ms *MusicServer) ListAlbums(response http.ResponseWriter, request *http.Request) {
	switch {
	// return albums of artist
	case request.FormValue("term") != "":
		ms.findAlbumsFromTerm(response, request)
	default:
		albumsData := ms.indexManager.ListAllAlbums(request.FormValue("genre"))
		data, _ := json.Marshal(albumsData)
		logger.LogE(response.Write(data))
	}
}

// MusicInfo Return info about music
func (ms *MusicServer) MusicInfo(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(request.FormValue("id"))
	logger.GetLogger().Info("Load music info with id", id)
	isFavorite := ms.favorites.IsFavorite(id)
	musicInfoData := ms.library.GetMusicInfoAsJSON(int32(id), isFavorite)
	response.Header().Set("Access-Control-Allow-Origin", "*")
	logger.LogE(response.Write(musicInfoData))
}

func (ms *MusicServer) PathMusic(response http.ResponseWriter, request *http.Request) {
	id, _ := strconv.Atoi(request.FormValue("id"))
	logger.GetLogger().Info("Load path info with id", id)
	response.Header().Set("Access-Control-Allow-Origin", "*")
	m := ms.library.GetMusicInfo(int32(id))

	logger.LogE(response.Write([]byte(m["path"])))
}

// MusicsInfo Return info about many musics
func (ms *MusicServer) MusicsInfo(response http.ResponseWriter, request *http.Request) {
	var ids []int32
	logger.LogE(json.Unmarshal([]byte(request.FormValue("ids")), &ids))
	//logger.GetLogger().Info("Load musics", len(ids))
	if request.FormValue("short") != "" {
		ms.getMusics(response, request, int32asInt(ids), false, []string{"artist"})
	} else {
		ms.musicsResponse(ids, response)
	}
}

// Search musics by free text
func (ms *MusicServer) Search(response http.ResponseWriter, request *http.Request) {
	musics := ms.indexManager.SearchText(request.FormValue("term"), request.FormValue("size"))
	ms.musicsResponse(musics, response)
}

// musicsResponse Get details from ids of music
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
	logger.LogE(response.Write(bdata))
}

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

func (ms *MusicServer) NbMusics(response http.ResponseWriter, _ *http.Request) {
	logger.LogE(response.Write([]byte(fmt.Sprintf("%d", ms.library.GetNbMusics()))))
}

func int32asInt(ids []int32) []int {
	idsInt := make([]int, len(ids))
	for i, id := range ids {
		idsInt[i] = int(id)
	}
	return idsInt
}

func (ms *MusicServer) ListGenres(response http.ResponseWriter, _ *http.Request) {
	data, _ := json.Marshal(ms.indexManager.ListGenres())
	logger.LogE(response.Write(data))
}
