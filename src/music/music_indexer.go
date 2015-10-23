package music

func IndexArtists(folder string){
    // Recreate albums index at each time (very quick)
    artists := LoadArtistIndex(folder)
    dico := LoadDictionnary(folder)
    musicsByArtist := LoadArtistMusicIndex(folder)

    am := NewAlbumManager(folder)

    // Index album by genre (consider only one genre by album)

    for _, artistId := range artists.FindAll() {
        musicsIds := musicsByArtist.Get(artistId)
        // Load all tracks and group by album
        albums  := make(map[string][]int)
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
}
