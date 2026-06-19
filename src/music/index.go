package music

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
	"math"
	"sort"
	"strconv"
)

// Coordinates all index for searching

type SearchIndex struct {
	folder string
	// Manage index by album
	albumManager *AlbumManager
	// Full text search
	musicIndexer TextIndexer
	// Manage index by genre
	genreReader   *GenreReader
	artistManager ArtistManager
}

// NewSearchIndex create a manager with everything to request indexes
func NewSearchIndex(folder string) *SearchIndex {
	s := SearchIndex{folder: folder}
	s.musicIndexer = LoadTextIndexer(folder, TextIndexerFilename)
	s.albumManager = NewAlbumManager(folder, true)
	s.genreReader = NewGenreReader(folder)
	s.artistManager = LoadArtistIndex(folder)
	return &s
}

func (si *SearchIndex) UpdateAlbumManager(albumManager *AlbumManager) {
	si.albumManager = albumManager
}

// UpdateIndexer update text indexer and reload genre
func (si *SearchIndex) UpdateIndexer(textIndexer TextIndexer) {
	si.musicIndexer = textIndexer
	si.artistManager = LoadArtistIndex(si.folder)
	si.genreReader = NewGenreReader(si.folder)
}

// ListAllAlbums list all albums in database
func (si *SearchIndex) ListAllAlbums(genre string) []map[string]string {
	filterAlbums := map[int]struct{}{}
	if genre != "" {
		filterAlbums = si.genreReader.GetAlbums(genre)
	}
	albums := si.albumManager.LoadAllAlbums()

	//logger.GetLogger().Info(albums, genre)
	albumsData := make([]map[string]string, 0, len(albums))
	for album, id := range albums {
		// test if album id is in the filtered genre list
		if _, exist := filterAlbums[id]; exist || len(filterAlbums) == 0 {
			albumsData = append(albumsData, map[string]string{"name": album, "id": fmt.Sprintf("%d", id), "url": fmt.Sprintf("idAlbum=%d", id)})
		}
	}
	sort.Sort(SortByArtist(albumsData))
	return albumsData
}

// ListFullAlbumById load musics of a specific album
func (si *SearchIndex) ListFullAlbumById(albumID int) []int {
	logger.GetLogger().Info("Get all musics of album", albumID)
	return si.albumManager.GetMusicsAll(albumID)
}

// ListAlbumById load musics of a specific album
func (si *SearchIndex) ListAlbumById(albumID int) []int {
	logger.GetLogger().Info("Get all musics of album", albumID)
	return si.albumManager.GetMusics(albumID)
}

// ListAlbumByArtist load an album by id
func (si *SearchIndex) ListAlbumByArtist(artistID int) []map[string]string {
	logger.GetLogger().Info("Get all albums of artist", artistID)
	albums := NewAlbumByArtist().GetAlbums(si.folder, artistID)
	albumsData := make([]map[string]string, 0, len(albums))
	for _, album := range albums {
		albumsData = append(albumsData, map[string]string{"name": album.Name, "url": fmt.Sprintf("idAlbum=%d", album.Id)})
	}
	sort.Sort(SortByArtist(albumsData))
	return albumsData
}

// SearchText search musics from prefix text (search in title and artist)
func (si *SearchIndex) SearchText(text string, strSize string) []int32 {
	size := float64(10)
	if intSize, e := strconv.ParseInt(strSize, 10, 32); e == nil {
		size = float64(intSize)
	}

	musics := si.musicIndexer.Search(text)
	musics32 := make([]int32, len(musics))
	for i, m := range musics {
		musics32[i] = int32(m)
	}
	logger.GetLogger().Info("Search", text, len(musics))
	return musics32[:int(math.Min(size, float64(len(musics))))]
}

// SearchArtistByGenre return all artists who have musics of this genre
func (si *SearchIndex) SearchArtistByGenre(genre string) map[int]struct{} {
	if genre == "" {
		return make(map[int]struct{})
	}
	return si.genreReader.GetArtist(genre)
}

func (si *SearchIndex) SearchArtistsByTerm(term string) map[int]string {
	foundArtists := si.artistManager.textIndexer.Search(term)
	artists := make(map[int]string, len(foundArtists))

	for _, id := range foundArtists {
		if name, exists := si.artistManager.artistsById[id]; exists {
			artists[id] = name
		}
	}
	return artists
}

func (si *SearchIndex) SearchAlbumsByTerm(term string) []map[string]string {
	foundAlbumsId := si.albumManager.textIndexer.Search(term)

	albumsData := make([]map[string]string, 0)
	for _, albumID := range foundAlbumsId {
		if name, exists := si.albumManager.mbaa.reverseNames[albumID]; exists {
			albumsData = append(albumsData, map[string]string{"name": name, "id": fmt.Sprintf("%d", albumID), "url": fmt.Sprintf("idAlbum=%d", albumID)})
		}
	}
	return albumsData
}

// ListGenres return all genres
func (si *SearchIndex) ListGenres() []string {
	return si.genreReader.GetGenres()
}
