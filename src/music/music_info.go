package music

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ascherkus/go-id3/src/id3"
)

//Music represent a music with all informations
type Music struct {
	file  id3.File
	path  string
	cover string
	id    int64
}

//NewMusic create a new music from id3tag, musicId, path and cover
func NewMusic(data id3.File, idMusic int64, path, cover string) Music {
	return Music{file: data, id: idMusic, path: path, cover: cover}
}

func fromJSON(data map[string]string) (*id3.File, string, int64) {
	file := id3.File{}
	file.Name = data["title"]
	file.Artist = data["artist"]
	file.Album = data["album"]
	file.Length = data["length"]
	file.Year = data["year"]
	file.Genre = data["genre"]
	file.Track = data["track"]
	id := int64(0)
	if iID, err := strconv.ParseInt(data["id"], 10, 32); err == nil {
		id = iID
	}
	return &file, data["cover"], id
}

func (m Music) toJSON() []byte {
	data := make(map[string]string, 9)
	data["id"] = fmt.Sprintf("%d", m.id)
	data["title"] = m.file.Name
	data["artist"] = m.file.Artist
	data["album"] = m.file.Album
	data["length"] = m.file.Length
	data["year"] = m.file.Year
	data["genre"] = m.file.Genre
	data["track"] = m.file.Track
	data["path"] = m.path
	data["cover"] = m.cover
	jsonData, _ := json.Marshal(data)
	return jsonData
}

//SortByArtist is used to sort musics by artist
type SortByArtist []map[string]string

func (a SortByArtist) Len() int { return len(a) }
func (a SortByArtist) Less(i, j int) bool {
	return strings.ToLower(a[i]["name"]) < strings.ToLower(a[j]["name"])
}
func (a SortByArtist) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

//SortByAlbum is used to sort musics by album
type SortByAlbum []map[string]interface{}

func (a SortByAlbum) Len() int { return len(a) }
func (a SortByAlbum) Less(i, j int) bool {
	infos1 := a[i]["infos"].(map[string]string)
	infos2 := a[j]["infos"].(map[string]string)
	return getTrack(infos1["track"]) < getTrack(infos2["track"])
}
func (a SortByAlbum) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func getTrack(track string) int {
	if track[0] == '#' {
		track = track[1:]
	}
	if pos := strings.Index(track, "/"); pos != -1 {
		track = track[:pos]
	}
	if n, err := strconv.ParseInt(track, 10, 32); err == nil {
		return int(n)
	}
	return 0
}
