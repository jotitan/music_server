package music

import (
	"github.com/jotitan/music_server/logger"
	"strings"
)

//IndexArtists index all artists of musics with musics for each
func IndexArtists(folder string) TextIndexer {
	library := NewMusicLibrary(folder)
	// Recreate albums index at each time (very quick)
	artists := LoadArtists(folder)
	musicsByArtist := LoadArtistMusicIndex(folder)
	logger.GetLogger().Info("Launch index with", len(artists), "artists")
	// Contains all album name and id
	am := NewAlbumManager(folder)

	genreIndexer := NewGenreIndexer()
	// Index album by genre (consider only one genre by album)

	albumsByGenre := make(map[string]map[int]struct{})

	for _, artistID := range artists {
		// For each artist, compile all genre of musics
		musicsIds := musicsByArtist.Get(artistID)
		// Load all tracks and group by album
		albums := make(map[string][]int)
		genres := make(map[string]struct{})
		for i, music := range library.GetMusicsInfo(musicsIds) {
			musicID := musicsIds[i]
			// Index genre by artist
			genre := "Undefined"
			if music["genre"] != "" {
				// Reformat genre
				genre = strings.Replace(strings.Replace(music["genre"], "/", " ", -1), "-", " ", -1)
				genre = strings.ToUpper(genre[0:1]) + strings.ToLower(genre[1:])
				genres[genre] = struct{}{}
			}

			// If returned error, id is already indexed
			albumID, err := am.AddMusic(music["album"], int(musicID), music["title"])
			if ids, ok := albums[music["album"]]; ok {
				albums[music["album"]] = append(ids, int(musicID))
			} else {
				albums[music["album"]] = []int{int(musicID)}
			}
			if err == nil {
				am.IndexText(int(musicID), music["title"], music["artist"])
				// Index genre by album
				if genre != "" {
					if listAlbums, exist := albumsByGenre[genre]; !exist {
						albumsByGenre[genre] = map[int]struct{}{albumID: struct{}{}}
					} else {
						listAlbums[albumID] = struct{}{}
					}
				}
			}
		}
		genreIndexer.AddManyGenresForArtist(genres, artistID)
		am.AddAlbumsByArtist(artistID, albums)
	}
	// Save albums by genre
	for genre, albums := range albumsByGenre {
		for idAlbum := range albums {
			genreIndexer.AddAlbum(genre, idAlbum)
		}
	}
	am.textIndexer.Build()
	am.Save()
	genreIndexer.Save(folder)
	logger.GetLogger().Info("End index")
	return am.textIndexer
}
