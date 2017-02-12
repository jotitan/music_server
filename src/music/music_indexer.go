package music

import (
	"logger"
)

func IndexArtists(folder string)TextIndexer{
    // Recreate albums index at each time (very quick)
    artists := LoadArtists(folder)
    logger.GetLogger().Info("Launch index with",len(artists),"artists")
    dico := LoadDictionnary(folder)
    musicsByArtist := LoadArtistMusicIndex(folder)

    am := NewAlbumManager(folder)
	// Contains all album name and id
    genreIndexer := NewGenreIndexer()
    // Index album by genre (consider only one genre by album)

	albumsByGenre := make(map[string]map[int]struct{})

    for n, artistId := range artists {
        // For each artist, compile all genre of musics
        musicsIds := musicsByArtist.Get(artistId)
        // Load all tracks and group by album
        albums  := make(map[string][]int)
        genres := make(map[string]struct{})
        logger.GetLogger().Info("=>",n,artistId,dico.GetMusicsFromIds(musicsIds))
        for i,music := range dico.GetMusicsFromIds(musicsIds)  {
            musicId := musicsIds[i]
			// Index genre by artist
			// TODO transform empty as undefined
			if music["genre"] != "" {
				genres[music["genre"]] = struct{}{}
			}
            if ids,ok := albums[music["album"]] ; ok {
                albums[music["album"]] = append(ids,musicId)
            }else{
                albums[music["album"]] = []int{musicId}
            }
            albumId := am.AddMusic(music["album"],musicId)
			logger.GetLogger().Info("Info album",music["album"],albumId)
			am.IndexText(musicId,music["title"],music["artist"])
			// Index genre by album
			if music["genre"] != "" {
				if listAlbums, exist := albumsByGenre[music["genre"]]; !exist {
					albumsByGenre[music["genre"]] = map[int]struct{}{albumId:struct{}{}}
				}else{
					listAlbums[albumId] = struct{}{}
					//albumsByGenre[music["genre"]] = append(listAlbums,albumId)
				}
			}
		}
		genreIndexer.AddManyGenresForArtist(genres,artistId)
        am.AddAlbumsByArtist(artistId,albums)
    }
	// Save albums by genre
	for genre,albums := range albumsByGenre {
	 	for idAlbum := range albums {
			genreIndexer.AddAlbum(genre,idAlbum)
		}
	}
    am.textIndexer.Build()
    am.Save()
	genreIndexer.Save(folder)
    logger.GetLogger().Info("End index")
    return am.textIndexer
}
