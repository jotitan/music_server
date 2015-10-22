package music

func IndexArtists(folder string){
    // Recreate albums index at each time (very quick)
    artists := LoadArtistIndex(folder)
    dico := LoadDictionnary(folder)
    artistsIdx := LoadArtistMusicIndex(folder)
    // Index music by album and artist
    mba := MusicByAlbum{}
    // Index album by artist
    aba := NewAlbumByArtist()
    // Index by album (based on name), no matter artist (case on multiple artist)
    idxAlbums := NewAlbumsIndex()

    // Index album by genre (consider only one genre by album)

    for _, id := range  artists.FindAll() {
        musicsIds := artistsIdx.MusicsByArtist[id]
        // Load all tracks and group by album
        albums  := make(map[string][]int)
        for i,music := range dico.GetMusicsFromIds(musicsIds)  {
            musicId := musicsIds[i]
            if ids,ok := albums[music["album"]] ; ok {
                albums[music["album"]] = append(ids,musicId)
            }else{
                albums[music["album"]] = []int{musicId}
            }
            idxAlbums.Add(music["album"],musicId)
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
    idxAlbums.Save(folder)

}
