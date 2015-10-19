package music

func IndexArtists(folder string){
    // Recreate albums index at each time (very quick)
    artists := LoadArtistIndex(folder)
    dico := LoadDictionnary(folder)
    artistsIdx := LoadArtistMusicIndex(folder)
    mba := MusicByAlbum{}
    aba := NewAlbumByArtist()

    for _, id := range  artists.FindAll() {
        musicsIds := artistsIdx.MusicsByArtist[id]
        // Load all tracks and group by album
        albums  := make(map[string][]int)
        for i,music := range dico.GetMusicsFromIds(musicsIds)  {
            if ids,ok := albums[music["album"]] ; ok {
                albums[music["album"]] = append(ids,musicsIds[i])
            }else{
                albums[music["album"]] = []int{musicsIds[i]}
            }
        }
        // Save all albums
        for album,musicsIds := range albums {
            idAlbum := mba.Adds(musicsIds)
            // Add idAlbum in album artist index
            aba.AddAlbum(id,NewAlbum(idAlbum,album))
        }

    }
    mba.Save(folder)
    aba.Save(folder)

}
