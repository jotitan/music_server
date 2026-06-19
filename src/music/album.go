package music

import (
	"encoding/gob"
	"errors"
	"github.com/jotitan/music_server/logger"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Give methods to manage album

type AlbumsList struct {
	albums []string
}

// ElementsIndex One file with link between a father element id and son element id (map)
type ElementsIndex struct {
}

// ElementsByFather store all album for artist
type ElementsByFather map[int][]int

func (ebf *ElementsByFather) Add(fatherID, sonId int) {
	if albums, ok := (*ebf)[fatherID]; ok {
		(*ebf)[fatherID] = append(albums, sonId)
	} else {
		(*ebf)[fatherID] = []int{sonId}
	}
}

func (ebf *ElementsByFather) Adds(fatherId int, sonsId []int) {
	if list, present := (*ebf)[fatherId]; present {
		(*ebf)[fatherId] = append(list, sonsId...)
	} else {
		(*ebf)[fatherId] = sonsId
	}
}

func (ebf *ElementsByFather) Save(folder string) {
	path := filepath.Join(folder, "artist_music.index")
	f, _ := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer f.Close()
	enc := gob.NewEncoder(f)
	logger.LogE(enc.Encode(ebf))
}

type Album struct {
	Id   int
	Name string
}

func NewAlbum(id int, name string) Album {
	return Album{id, name}
}

type AlbumByArtist struct {
	idxByArtist        map[int][]Album
	header             []int64
	currentArtist      int
	previousPosition   int64
	previousDataLength int64
	max                int
}

func NewAlbumByArtist() *AlbumByArtist {
	return &AlbumByArtist{idxByArtist: make(map[int][]Album)}
}

func (aba *AlbumByArtist) AddAlbum(idArtist int, album Album) {
	if albums, ok := aba.idxByArtist[idArtist]; !ok {
		(*aba).idxByArtist[idArtist] = []Album{album}
	} else {
		(*aba).idxByArtist[idArtist] = append(albums, album)
	}
}

func (aba *AlbumByArtist) Save(folder string) {
	path := filepath.Join(folder, "album_by_artist.index")
	f, _ := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	// Get max artist id
	maxId := 0
	for id := range aba.idxByArtist {
		if id > maxId {
			maxId = id
		}
	}
	// Prepare header (nb elements and size artist
	aba.header = make([]int64, maxId)
	aba.previousPosition = int64(4 + 8*maxId)
	aba.max = maxId
	logger.LogE(f.Write(getInt32AsByte(int32(maxId))))
	logger.LogE(f.Write(make([]byte, 8*maxId)))

	// Copy data
	logger.LogE(io.Copy(f, aba))
	logger.LogE(f.WriteAt(getInts64AsByte(aba.header), 4))

	logger.LogE(f.Close())

}

// Save position of data in header. header | len album (2) | id (4) | len name album (1) | album name
func (aba *AlbumByArtist) Read(p []byte) (int, error) {
	// Length of written bytes
	dataLength := 0
	for {
		if aba.currentArtist > aba.max {
			return dataLength, io.EOF
		}
		// Evaluate block size
		artist, ok := aba.idxByArtist[aba.currentArtist]
		if ok {
			// Artist id start at one
			// write first header, only if not done yet
			if aba.header[aba.currentArtist-1] == 0 {
				aba.header[aba.currentArtist-1] = aba.previousPosition + aba.previousDataLength
				aba.previousPosition = aba.header[aba.currentArtist-1]
			}

			// Check enough place
			estimateSize := 2
			for _, album := range artist {
				estimateSize += 5 + len(album.Name)
			}

			aba.previousDataLength = int64(estimateSize)
			if dataLength+estimateSize > len(p) {
				return dataLength, nil
			}
			writeBytes(p, getInt16AsByte(int16(len(artist))), dataLength)
			dataLength += 2
			for _, album := range artist {
				writeBytes(p, getInt32AsByte(int32(album.Id)), dataLength)
				p[dataLength+4] = byte(len(album.Name))
				writeBytes(p, []byte(album.Name), dataLength+5)
				dataLength += 5 + len(album.Name)
			}
		}
		aba.currentArtist++
	}
}

func (aba *AlbumByArtist) GetAlbums(folder string, artistId int) []Album {
	path := filepath.Join(folder, "album_by_artist.index")
	f, _ := os.Open(path)
	defer f.Close()

	// Read artist position
	// Check number of element
	nbArtists := int(getInt32FromFile(f, 0))
	if artistId > nbArtists {
		return []Album{}
	}
	posInHeader := int64(4 + (artistId-1)*8)

	posInFile := getInt64FromFile(f, posInHeader)
	if posInFile == 0 {
		return []Album{}
	}
	nbAlbums := int(getInt16FromFile(f, posInFile))
	posInFile += 2
	albums := make([]Album, nbAlbums)
	for i := 0; i < nbAlbums; i++ {
		id := getInt32FromFile(f, posInFile)
		lengthName := getInt8FromFile(f, posInFile+4)
		nameTab := make([]byte, lengthName)
		logger.LogE(f.ReadAt(nameTab, posInFile+5))
		albums[i] = NewAlbum(int(id), string(nameTab))
		posInFile += int64(5 + lengthName)
	}
	return albums
}

// AlbumManager manage albums (musics for each album, index...)
type AlbumManager struct {
	// Store musics by album for one artist (not forcebly all music of album). Position is id of album for an artist
	musicsByAlbum [][]int
	// Store albums by artist
	aba *AlbumByArtist
	// Music by album alone
	mbaa *AlbumsIndex
	// Working folder
	folder string
	// Text indexer for album names
	textIndexer TextIndexer
}

func NewAlbumManager(folder string, loadIndex bool) *AlbumManager {
	am := AlbumManager{}
	am.folder = folder
	am.musicsByAlbum = make([][]int, 0)
	am.aba = NewAlbumByArtist()
	am.mbaa = NewAlbumsIndex()
	am.textIndexer = NewTextIndexer()
	if loadIndex {
		am.init()
	}
	return &am
}

func (am *AlbumManager) AddAlbumsByArtist(artistID int, albums map[string][]int) {
	for album, musicsIds := range albums {
		albumID := len(am.musicsByAlbum) + 1
		am.musicsByAlbum = append(am.musicsByAlbum, musicsIds)
		am.aba.AddAlbum(artistID, NewAlbum(albumID, album))
	}
}

// AddMusic return id of the album
func (am *AlbumManager) AddMusic(album string, idMusic int, title string) (int, error) {
	return am.mbaa.Add(album, idMusic, strings.ToLower(title))
}

func (am *AlbumManager) Save() {
	logger.GetLogger().Info("Save musics by album by artist")
	NewMusicAlbumSaver(am.musicsByAlbum).Save(filepath.Join(am.folder, "album_music.index"))

	logger.GetLogger().Info("Save all musics of albums")
	NewMusicAlbumSaver(am.mbaa.index).Save(filepath.Join(am.folder, "all_albums_music.index"))
	(&(IndexSaver{am.mbaa.toSave, 0})).Save(filepath.Join(am.folder, "albums.dico"), true)

	am.aba.Save(am.folder)
	am.textIndexer.Build()
	am.textIndexer.SaveInFile(am.folder, albumTextIndexerFilename)
}

func (am *AlbumManager) getMusicsFrom(filename string, albumId int) []int {
	path := filepath.Join(am.folder, filename)
	f, _ := os.Open(path)
	defer f.Close()

	// Check number of elements
	nbAlbums := int(getInt32FromFile(f, 0))
	if albumId > nbAlbums {
		return []int{}
	}
	// Album id start at 1
	posInHeader := int64((albumId-1)*8 + 4)
	posInFile := getInt64FromFile(f, posInHeader)
	nbMusics := int32(getInt16FromFile(f, posInFile))
	//logger.GetLogger().Info("Load musics of album", albumId, ", pos :", posInFile, ", musics :", nbMusics)
	musicsTab := make([]byte, nbMusics*4)
	logger.LogE(f.ReadAt(musicsTab, posInFile+2))
	return getBytesAsInts32Int(musicsTab)
}

func (am *AlbumManager) GetMusics(albumId int) []int {
	return am.getMusicsFrom("album_music.index", albumId)
}

func (am *AlbumManager) GetMusicsAll(albumId int) []int {
	return am.getMusicsFrom("all_albums_music.index", albumId)
}

// LoadAllAlbums load all album names with id for each
func (am *AlbumManager) LoadAllAlbums() map[string]int {
	return am.mbaa.names
}

// Store some data in cache
func (am *AlbumManager) init() {
	am.mbaa.names = IndexReader{}.Load(filepath.Join(am.folder, "albums.dico"))
	for name, id := range am.mbaa.names {
		am.mbaa.reverseNames[id] = name
	}
	am.textIndexer = LoadTextIndexer(am.folder, albumTextIndexerFilename)
}

type musicByAlbumSaver struct {
	data    [][]int
	current int
	// Store all positions
	header []int64
	// used to define next position data in header
	currentAlbumSize int
}

func NewMusicAlbumSaver(albums [][]int) musicByAlbumSaver {
	return musicByAlbumSaver{data: albums}
}

func (mas musicByAlbumSaver) Save(path string) {
	f, _ := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	// Reserve header size (nb elements * 8 + 4)
	logger.LogE(f.Write(getInt32AsByte(int32(len(mas.data)))))
	logger.LogE(f.Write(make([]byte, len(mas.data)*8)))
	mas.header = make([]int64, len(mas.data))
	logger.LogE(io.Copy(f, &mas))
	// Rewrite header at the beginning
	logger.LogE(f.WriteAt(getInts64AsByte(mas.header), 4))

	logger.LogE(f.Close())
}

func (mas *musicByAlbumSaver) Read(p []byte) (int, error) {
	lengthData := 0
	for {
		if mas.current >= len(mas.data) {
			return lengthData, io.EOF
		}
		// Check if enough place to length
		// Check enougth place to write data nb element (2o)
		if len(p) < lengthData+2 {
			return lengthData, nil
		}
		album := mas.data[mas.current]

		// Write in header in not already done
		if mas.header[mas.current] == 0 {
			writeBytes(p, getInt16AsByte(int16(len(album))), lengthData)

			lengthData += 2
			// first position is just after the header
			dataPosition := int64(4 + 8*len(mas.data))
			if mas.current > 0 {
				// When new turn, album size can be change. Impossible to get correct position. Save in file
				// Take last position and add last length data
				dataPosition = mas.header[mas.current-1] + int64(mas.currentAlbumSize*4+2)
			}
			//logger.GetLogger().Info("Write album header", mas.current, "length", len(album), "with pos in file", dataPosition)
			mas.header[mas.current] = dataPosition
			mas.currentAlbumSize = len(album)
		}

		// Write in header only if header is empty (cause partial write could append)
		// Check enough place to write musics. If not, check number of music which can be written
		nbWritable := (len(p) - lengthData) / 4
		if len(album) > nbWritable {
			logger.GetLogger().Info("Write partial", nbWritable, mas.current)
			// Partial write, just some musics
			data := getInts32AsByte(album[:nbWritable])
			writeBytes(p, data, lengthData)
			mas.data[mas.current] = album[nbWritable:]
			lengthData += len(data)
			return lengthData, nil
		}
		// write all music
		data := getInts32AsByte(album)
		writeBytes(p, data, lengthData)
		mas.current++
		lengthData += len(data)
	}
	//return lengthData, nil
}

// AlbumsIndex store all albums, no matter artist, only based on name
type AlbumsIndex struct {
	// Names albums with id
	names        map[string]int
	reverseNames map[int]string
	toSave       []string
	index        [][]int
	exists       []map[int]struct{}
	existsName   []map[string]struct{}
}

func NewAlbumsIndex() *AlbumsIndex {
	return &AlbumsIndex{make(map[string]int), make(map[int]string), []string{}, make([][]int, 0), make([]map[int]struct{}, 0), make([]map[string]struct{}, 0)}
}

// Add a music into album index, if album no already exist, create it
// @return : the id of the album
func (ai *AlbumsIndex) Add(album string, idMusic int, title string) (int, error) {
	lowerAlbum := strings.ToLower(album)
	idAlbum, ok := ai.names[lowerAlbum]
	if !ok {
		idAlbum = len(ai.names) + 1
		ai.names[lowerAlbum] = idAlbum
		ai.reverseNames[idAlbum] = album
		ai.toSave = append(ai.toSave, album)
		ai.index = append(ai.index, []int{idMusic})
		ai.exists = append(ai.exists, map[int]struct{}{idMusic: {}})
		ai.existsName = append(ai.existsName, map[string]struct{}{title: {}})
		return idAlbum, nil
	}
	// Check if id music already indexed
	if _, ok := ai.exists[idAlbum-1][idMusic]; !ok {
		// Check if name is not already indexed
		if _, ok := ai.existsName[idAlbum-1][title]; !ok {
			// Position in index list is id - 1
			ai.index[idAlbum-1] = append(ai.index[idAlbum-1], idMusic)
			ai.exists[idAlbum-1][idMusic] = struct{}{}
			ai.existsName[idAlbum-1][title] = struct{}{}
			return idAlbum, nil
		}
		return idAlbum, errors.New("music name is already indexed")
	}
	return idAlbum, errors.New("music id is already indexed")
}
