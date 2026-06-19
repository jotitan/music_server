package music

import (
	"encoding/binary"
	"encoding/gob"
	"github.com/jotitan/music_server/logger"
	"io"
	"os"
	"path/filepath"
)

type IndexSaver struct {
	values  []string
	current int
}

// Save only new artists
func (is *IndexSaver) Save(pathfile string, trunc bool) {
	path := filepath.Join(pathfile)
	// TRUNC or NOT
	var f *os.File
	var err error
	if trunc {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		logger.LogE(f.Write(getInt32AsByte(int32(len(is.values)))))
	} else {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
		if err == nil {
			// New, write size
			logger.LogE(f.Write(getInt32AsByte(int32(len(is.values)))))
		} else {
			f, _ = os.OpenFile(path, os.O_RDWR, os.ModePerm)
			logger.LogE(f.WriteAt(getInt32AsByte(int32(len(is.values))), 0))
			logger.LogE(f.Seek(0, 2))
		}
	}
	is.current = 0
	logger.LogE(io.Copy(f, is))
	logger.LogE(f.Close())
}

// Read data from artist index
func (is *IndexSaver) Read(p []byte) (int, error) {
	pos := 0
	for {
		if is.current >= len(is.values) {
			return pos, io.EOF
		}
		value := is.values[is.current]
		if pos+2+len(value) > len(p) {
			return pos, nil
		}
		writeBytes(p, getInt16AsByte(int16(len(value))), pos)
		writeBytes(p, []byte(value), pos+2)
		pos += 2 + len(value)
		is.current++
	}
}

// ArtistManager store all artists (avoid double)
type ArtistManager struct {
	// Used to define if an artist exist (id of artist)
	artists     map[string]int
	artistsById map[int]string
	// Used in write
	tempBuffer []byte
	currentId  int
	// new artists
	artistsToSave []string
	currentSave   int
	textIndexer   TextIndexer
}

type IndexReader struct {
	data       map[string]int
	tempBuffer []byte
	currentId  int
}

// Load Get artist index to search...
// first id start at 1
func (ir IndexReader) Load(path string) map[string]int {
	f, err := os.Open(path)
	if err == nil {
		logger.LogE(io.Copy(&ir, f))
		logger.LogE(f.Close())
		return ir.data
	}
	return map[string]int{}
}

// Write get data in p and write in object
// nb artist (4) | lenght name (2) | data name...
// The id of element is position when reading (start at 1)
func (ir *IndexReader) Write(p []byte) (int, error) {
	pos := 0
	if ir.data == nil || len(ir.data) == 0 {
		// Load number, first 4 bytes
		ir.data = make(map[string]int, int(binary.LittleEndian.Uint32(p[0:4])))
		ir.currentId = 1
		pos = 4
	}
	pSize := len(p)
	if ir.tempBuffer != nil && len(ir.tempBuffer) > 0 {
		p = append(ir.tempBuffer, p...)
		ir.tempBuffer = nil
	}
	for {
		if pos+2 > len(p) {
			// Save rest in buffer
			ir.tempBuffer = p[pos:]
			return pSize, nil
		}
		artistSize := int(binary.LittleEndian.Uint16(p[pos : pos+2]))
		if pos+2+artistSize > len(p) {
			ir.tempBuffer = p[pos:]
			return pSize, nil
		}
		ir.data[string(p[pos+2:pos+2+artistSize])] = ir.currentId
		ir.currentId++
		pos += 2 + artistSize
	}
}

func LoadArtists(folder string) map[string]int {
	return IndexReader{}.Load(filepath.Join(folder, "artist.dico"))
}

// LoadArtistIndex Get artist index to search...
func LoadArtistIndex(folder string) ArtistManager {
	ai := ArtistManager{artists: make(map[string]int), artistsToSave: make([]string, 0), textIndexer: NewTextIndexer()}
	ai.artists = LoadArtists(folder)
	ai.artistsById = make(map[int]string, len(ai.artists))
	for name, id := range ai.artists {
		ai.textIndexer.IndexText(id, name)
		ai.artistsById[id] = name
	}
	ai.textIndexer.Build()
	ai.currentId = len(ai.artists) + 1
	return ai
}

// Add the artist in index. Return id
func (ai *ArtistManager) Add(artist string) int {
	// Check if exist
	if id, exist := ai.artists[artist]; exist {
		return id
	}
	id := ai.currentId
	ai.artists[artist] = id
	ai.textIndexer.IndexText(id, artist)
	ai.artistsToSave = append(ai.artistsToSave, artist)
	logger.GetLogger().Info("Add artist", artist, " :", ai.currentId)
	ai.currentId++
	return id
}

// FindAll return all artists with id
func (ai ArtistManager) FindAll() map[string]int {
	return ai.artists
}

// Save only new artists
func (ai *ArtistManager) Save(folder string, trunc bool) {
	is := IndexSaver{ai.artistsToSave, 0}
	logger.GetLogger().Info("Save artists", len(ai.artistsToSave))
	is.Save(filepath.Join(folder, "artist.dico"), trunc)
}

// Write get data in p and write in object
// nb artist (4) | lenght name (2) | data name...
func (ai *ArtistManager) Write(p []byte) (int, error) {
	pos := 0
	if ai.artists == nil || len(ai.artists) == 0 {
		// Load number, first 4 bytes
		ai.artists = make(map[string]int, int(binary.LittleEndian.Uint32(p[0:4])))
		ai.currentId = 1
		pos = 4
	}
	pSize := len(p)
	if ai.tempBuffer != nil && len(ai.tempBuffer) > 0 {
		p = append(ai.tempBuffer, p...)
		ai.tempBuffer = nil
	}
	for {
		if pos+2 > len(p) {
			// Save rest in buffer
			ai.tempBuffer = p[pos:]
			return pSize, nil
		}
		artistSize := int(binary.LittleEndian.Uint16(p[pos : pos+2]))
		if pos+2+artistSize > len(p) {
			ai.tempBuffer = p[pos:]
			return pSize, nil
		}
		ai.artists[string(p[pos+2:pos+2+artistSize])] = ai.currentId
		ai.currentId++
		pos += 2 + artistSize
	}
}

// ArtistMusicIndex is an index music by artist. Use id artist and id music.
// Save with temporary method with gob decode / encode
// TODO change with ElementsByFather
type ArtistMusicIndex struct {
	// map with id artist of key and list of music
	MusicsByArtist        map[int][]int
	checkDuplicateArtists map[int]map[int]struct{}
}

// Get all musics of a specific artist (id)
func (ami *ArtistMusicIndex) Get(artistID int) []int32 {
	musics := ami.MusicsByArtist[artistID]
	musicsIds := make([]int32, len(musics))
	for i, id := range musics {
		musicsIds[i] = int32(id)
	}
	return musicsIds
}

// Add a music to an artist
func (ami *ArtistMusicIndex) Add(artist, music int) {
	if musics, present := ami.MusicsByArtist[artist]; present {
		// Check if number already exists
		if _, exist := ami.checkDuplicateArtists[artist][music]; !exist {
			ami.MusicsByArtist[artist] = append(musics, music)
			ami.checkDuplicateArtists[artist][music] = struct{}{}
		}
	} else {
		ami.MusicsByArtist[artist] = []int{music}
		ami.checkDuplicateArtists[artist] = map[int]struct{}{music: {}}
	}
}

// Save the musics for each artist (id)
func (ami ArtistMusicIndex) Save(folder string) {
	path := filepath.Join(folder, "artist_music.index")
	f, _ := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer f.Close()
	enc := gob.NewEncoder(f)
	logger.LogE(enc.Encode(ami))
}

func LoadArtistMusicIndex(folder string) ArtistMusicIndex {
	path := filepath.Join(folder, "artist_music.index")
	ami := ArtistMusicIndex{}
	if f, err := os.Open(path); err == nil {
		dec := gob.NewDecoder(f)
		logger.LogE(dec.Decode(&ami))
		logger.LogE(f.Close())
	} else {
		ami.MusicsByArtist = make(map[int][]int)
	}
	ami.checkDuplicateArtists = make(map[int]map[int]struct{})
	return ami
}
