package music

import (
    "logger"
    "strings"
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

    for _, artistId := range artists {
        // For each artist, compile all genre of musics
        musicsIds := musicsByArtist.Get(artistId)
        // Load all tracks and group by album
        albums := make(map[string][]int)
        genres := make(map[string]struct {})
        //logger.GetLogger().Info("=>",n,artistId,dico.GetMusicsFromIds(musicsIds))
        for i, music := range dico.GetMusicsFromIds(musicsIds) {
            musicId := musicsIds[i]
            // Index genre by artist
            // TODO transform empty as undefined
            genre := ""
            if music["genre"] != "" {
                // Reformat genre
                genre = strings.Replace(strings.Replace(music["genre"], "/", " ", -1), "-", " ", -1)
                genre = strings.ToUpper(genre[0:1]) + strings.ToLower(genre[1:])
                genres[genre] = struct {}{}
            }

            // If returned error, id is already indexed
            if albumId, err := am.AddMusic(music["album"], musicId,music["title"]); err == nil {
                if ids, ok := albums[music["album"]]; ok {
                    albums[music["album"]] = append(ids, musicId)
                }else {
                    albums[music["album"]] = []int{musicId}
                }
                //logger.GetLogger().Info("Info album",music["album"],albumId)
                am.IndexText(musicId, music["title"], music["artist"])
                // Index genre by album
                if genre != "" {
                    if listAlbums, exist := albumsByGenre[genre]; !exist {
                        albumsByGenre[genre] = map[int]struct {}{albumId:struct {}{}}
                    }else {
                        listAlbums[albumId] = struct {}{}
                        //albumsByGenre[music["genre"]] = append(listAlbums,albumId)
                    }
                }
            } else{
                //logger.GetLogger().Info(music["title"],":",err.Error())
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
