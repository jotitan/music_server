package music
import (
    "encoding/gob"
    "os"
    "path/filepath"
    "io"
    "strings"
)

// Give methods to manage album


type AlbumsList struct{
    albums []string
}

// One file with link between a father element id and son element id (map)
type ElementsIndex struct{

}

// AlbumByArtist store all album for artist
type ElementsByFather map[int][]int

func (ebf * ElementsByFather)Add(fatherId,sonId int){
    if albums,ok := (*ebf)[fatherId] ; ok {
        (*ebf)[fatherId] = append(albums,sonId)
    }else{
        (*ebf)[fatherId] = []int{sonId}
    }
}

func (ebf * ElementsByFather)Adds(fatherId int,sonsId []int){
    if list,present := (*ebf)[fatherId] ; present {
        (*ebf)[fatherId] = append(list,sonsId...)
    }else{
        (*ebf)[fatherId] = sonsId
    }
}

func (ebf ElementsByFather)Save(folder string){
    path := filepath.Join(folder,"artist_music.index")
    f,_ := os.OpenFile(path,os.O_TRUNC|os.O_CREATE|os.O_RDWR,os.ModePerm)
    defer f.Close()
    enc := gob.NewEncoder(f)
    enc.Encode(ebf)
}

func LoadElementsByFather(folder, filename string)ElementsByFather{
    path := filepath.Join(folder,filename + ".index")
    ebf := ElementsByFather{}
    if f,err := os.Open(path);err == nil {
        dec := gob.NewDecoder(f)
        dec.Decode(&ebf)
        f.Close()
    }else{
        ebf = ElementsByFather(make(map[int][]int))
    }
    return ebf
}

type Album struct{
    Id int
    Name string
}

func NewAlbum(id int, name string)Album{
    return Album{id,name}
}

type AlbumByArtist struct{
    idxByArtist map[int][]Album
    header []int64
    currentArtist int
    previousPosition int64
    previousDataLength int64
    max int
}

func NewAlbumByArtist()*AlbumByArtist{
    return &AlbumByArtist{idxByArtist:make(map[int][]Album)}
}

func (aba * AlbumByArtist)AddAlbum(idArtist int,album Album){
    if albums,ok := aba.idxByArtist[idArtist] ; !ok{
        (*aba).idxByArtist[idArtist] = []Album{album}
    }else{
        (*aba).idxByArtist[idArtist] = append(albums,album)
    }
}

func (aba AlbumByArtist)Save(folder string){
    path := filepath.Join(folder,"album_by_artist.index")
    f,_ := os.OpenFile(path,os.O_CREATE|os.O_TRUNC|os.O_RDWR,os.ModePerm)
    // Get max artist id
    max := 0
    for id := range aba.idxByArtist {
        if id > max {
            max = id
        }
    }
    // Prepare header (nb elements and size artist
    aba.header = make([]int64,max)
    aba.previousPosition = int64(4 + 8*max)
    aba.max = max
    f.Write(getInt32AsByte(int32(max)))
    f.Write(make([]byte,8*max))

    // Copy data
    io.Copy(f,&aba)
    f.WriteAt(getInts64AsByte(aba.header),4)

    f.Close()

}

// Save position of data in header. header | len album (2) | id (4) | len name album (1) | album name
func (aba * AlbumByArtist)Read(p []byte)(int,error){
    dataLength := 0

    for {
        if aba.currentArtist > aba.max {
            return dataLength,io.EOF
        }
        // Evaluate block size
        artist,ok := aba.idxByArtist[aba.currentArtist]
        if ok {
            // Artist id start at one
            // write first header
            aba.header[aba.currentArtist-1] = aba.previousPosition + aba.previousDataLength
            aba.previousPosition = aba.header[aba.currentArtist-1]

            // Check enough place
            estimateSize := 2
            for _,album := range artist {
                estimateSize+=5 + len(album.Name)
            }
            aba.previousDataLength=int64(estimateSize)
            if dataLength + estimateSize > len(p) {
                return dataLength,nil
            }
            writeBytes(p,getInt16AsByte(int16(len(artist))),dataLength)
            dataLength+=2
            for _,album := range artist {
                writeBytes(p,getInt32AsByte(int32(album.Id)),dataLength)
                p[dataLength+4] = byte(len(album.Name))
                writeBytes(p,[]byte(album.Name),dataLength+5)
                dataLength+=5+len(album.Name)

            }
        }
        aba.currentArtist++
    }
}

func (mba AlbumByArtist)GetAlbums(folder string,artistId int)[]Album{
    path := filepath.Join(folder,"album_by_artist.index")
    f,_ := os.Open(path)
    defer f.Close()

    // Read artist position
    // Check number of element
    nbArtists := int(getInt32FromFile(f,0))

    if artistId > nbArtists  {
        return []Album{}
    }
    posInHeader := int64(4 + (artistId -1)*8)
    posInFile := getInt64FromFile(f,posInHeader)
    if posInFile == 0 {
        return []Album{}
    }
    nbAlbums := int(getInt16FromFile(f,posInFile))

    posInFile+=2
    albums := make([]Album,nbAlbums)
    for i := 0 ; i < nbAlbums ; i++ {
        id := getInt32FromFile(f,posInFile)
        lengthName := getInt8FromFile(f,posInFile+4)
        nameTab := make([]byte,lengthName)
        f.ReadAt(nameTab,posInFile+5)
        albums[i] = NewAlbum(int(id),string(nameTab))
        posInFile+=int64(5+lengthName)
    }
    return albums
}

type AlbumManager struct{
    musicsByAlbum [][]int
    // Store albums by artist
    aba * AlbumByArtist
    // Music by album alone
    mbaa * AlbumsIndex
    // Working folder
    folder string
}

func NewAlbumManager(folder string)*AlbumManager{
    am := AlbumManager{}
    am.folder = folder
    am.musicsByAlbum = make([][]int,0)
    am.aba = NewAlbumByArtist()
    am.mbaa = NewAlbumsIndex()
    return &am
}

func (am * AlbumManager)AddAlbumsByArtist(artistId int,albums map[string][]int) {
    for album,musicsIds := range albums {
        idAlbum := len(am.musicsByAlbum)+1
        am.musicsByAlbum = append(am.musicsByAlbum,musicsIds)
        am.aba.AddAlbum(artistId,NewAlbum(idAlbum,album))
    }
}

func (am * AlbumManager)AddMusic(album string,idMusic int){
    am.mbaa.Add(album,idMusic)
}

func (am * AlbumManager) Save(){
    NewMusicAlbumSaver(am.musicsByAlbum).Save(filepath.Join(am.folder,"album_music.index"))

	NewMusicAlbumSaver(am.mbaa.index).Save(filepath.Join(am.folder,"all_albums_music.index"))
	(&(IndexSaver{am.mbaa.toSave,0})).Save(filepath.Join(am.folder,"albums.dico"),true)

    am.aba.Save(am.folder)
}

func (am * AlbumManager)getMusicsFrom(filename string,albumId int)[]int{
	path := filepath.Join(am.folder,filename)
	f,_ := os.Open(path)
	defer f.Close()

	// Check number of elements
	nbAlbums := int(getInt32FromFile(f,0))
	if albumId > nbAlbums {
		return []int{}
	}
	// Album id start at 1
	posInHeader := int64((albumId-1)*8+4)
	posInFile :=  getInt64FromFile(f,posInHeader)
	nbMusics := int32(getInt16FromFile(f,posInFile))

	musicsTab := make([]byte,nbMusics*4)
	f.ReadAt(musicsTab,posInFile+2)
	return getBytesAsInts32Int(musicsTab)
}

func (am * AlbumManager)GetMusics(albumId int)[]int {
	return am.getMusicsFrom("album_music.index",albumId)
}

func (am * AlbumManager)GetMusicsAll(albumId int)[]int {
	return am.getMusicsFrom("all_albums_music.index",albumId)
}

// LoadArtistIndex Get artist index to search...
func (am * AlbumManager)LoadAllAlbums()map[string]int{
	return IndexReader{}.Load(filepath.Join(am.folder,"albums.dico"))
}

type musicByAlbumSaver struct{
    data [][]int
    current int
    // Store all positions
    header []int64
    // used to define next position data in header
    currentAlbumSize int
}

func NewMusicAlbumSaver(albums [][]int)musicByAlbumSaver{
    return musicByAlbumSaver{data:albums}
}

func (mas musicByAlbumSaver)Save(path string){
    f,_ := os.OpenFile(path,os.O_CREATE|os.O_TRUNC|os.O_RDWR,os.ModePerm)
    // Reserve header size (nb elements * 8 + 4)
    f.Write(getInt32AsByte(int32(len(mas.data))))
    f.Write(make([]byte,len(mas.data)*8))
    mas.header = make([]int64,len(mas.data))
    io.Copy(f,&mas)
    // Rewrite header at the beginning
    f.WriteAt(getInts64AsByte(mas.header),4)

    f.Close()
}

func (mas * musicByAlbumSaver)Read(p []byte)(int,error){
    lengthData := 0
    for {
        if mas.current >= len(mas.data){
            return lengthData,io.EOF
        }
        // Check if enough place to length
        // Check enougth place to write data nb element (2o)
        if len(p) < lengthData + 2 {
            return lengthData,nil
        }
        album := mas.data[mas.current]

        if mas.header[mas.current] == 0{

            // Only if not write already
            writeBytes(p,getInt16AsByte(int16(len(album))),lengthData)
            lengthData+=2
            if mas.current == 0 {
                // first position is just after the header
                mas.header[mas.current] = int64(4 + 8*len(mas.data))
            }else{
                // When new turn, album size can be change. Impossible to get correct position. Save in file
                // Take last position and add last length data
                mas.header[mas.current] = mas.header[mas.current-1] + int64(mas.currentAlbumSize*4 + 2)
            }
            mas.currentAlbumSize = len(album)
        }

        // Write in header only if header is empty (cause partial write could append)
        // Check enough place to write musics. If not, check number of music which can be written
        nbWritable := (len(p) - lengthData)/4
        if len(album)>nbWritable {
            // Partial write, just some musics
            data := getInts32AsByte(album[:nbWritable])
            writeBytes(p,data,lengthData)
            mas.data[mas.current] = album[nbWritable:]
            lengthData+=len(data)
            //logger.GetLogger().Fatal(nbWritable,len(p),lengthData,len(data),mba.currentWriteId)
            return lengthData,nil
        }else{
            // write all music
            data := getInts32AsByte(album)
            writeBytes(p,data,lengthData)
            mas.current++
            lengthData+=len(data)
        }
    }
    return lengthData,nil
}

// AlbumsIndex store all albums, no matter artist, only based on name
type AlbumsIndex struct{
    // Names albums with id
	names map[string]int
	toSave []string
    index [][]int
}

func NewAlbumsIndex()*AlbumsIndex{
    return &AlbumsIndex{make(map[string]int),[]string{},make([][]int,0)}
}

func (ai * AlbumsIndex)Add(album string,idMusic int){
    lowerAlbum := strings.ToLower(album)
	if id,ok := ai.names[lowerAlbum];!ok {
		ai.names[lowerAlbum] = len(ai.names)
		ai.toSave = append(ai.toSave,album)
		ai.index = append(ai.index,[]int{idMusic})
	}else{
		ai.index[id] = append(ai.index[id],idMusic)
	}
}
