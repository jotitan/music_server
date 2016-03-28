package music
import "logger"

func IndexArtists(folder string){
    // Recreate albums index at each time (very quick)
    artists := LoadArtists(folder)
    logger.GetLogger().Info("Launch index with",len(artists),"artists")
    dico := LoadDictionnary(folder)
    musicsByArtist := LoadArtistMusicIndex(folder)

    am := NewAlbumManager(folder)

    // Index album by genre (consider only one genre by album)

    for n, artistId := range artists {
        musicsIds := musicsByArtist.Get(artistId)
        // Load all tracks and group by album
        albums  := make(map[string][]int)
        logger.GetLogger().Info("=>",n,artistId,dico.GetMusicsFromIds(musicsIds))
        for i,music := range dico.GetMusicsFromIds(musicsIds)  {
            musicId := musicsIds[i]
            if ids,ok := albums[music["album"]] ; ok {
                albums[music["album"]] = append(ids,musicId)
            }else{
                albums[music["album"]] = []int{musicId}
            }
            am.AddMusic(music["album"],musicId)
        }
        am.AddAlbumsByArtist(artistId,albums)
    }
    am.Save()
    logger.GetLogger().Info("End index")
}
