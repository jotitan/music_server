package music

import (
	"io"
	"os"
	"path/filepath"
	"encoding/binary"
	"encoding/gob"
	"logger"
)


type IndexSaver struct{
	values []string
	current int
}

// Save only new artists
func (is * IndexSaver)Save(pathfile string,trunc bool){
	path := filepath.Join(pathfile)
	// TRUNC or NOT
	var f *os.File
	var err error
	if trunc {
		f,err = os.OpenFile(path,os.O_CREATE|os.O_TRUNC|os.O_RDWR,os.ModePerm)
		f.Write(getInt32AsByte(int32(len(is.values))))
	}else {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
		if err == nil {
			// New, write size
			f.Write(getInt32AsByte(int32(len(is.values))))
		}else {
			f, _ = os.OpenFile(path, os.O_RDWR, os.ModePerm)
			f.WriteAt(getInt32AsByte(int32(len(is.values))), 0)
			f.Seek(0, 2)
		}
	}
	is.current = 0
	io.Copy(f,is)
	f.Close()
}

// Read data from artist index
func (is * IndexSaver)Read(p []byte)(int,error){
	pos := 0
	for {
		if is.current >= len(is.values){
			return pos,io.EOF
		}
		value := is.values[is.current]
		if pos + 2 + len(value) > len(p){
			return pos,nil
		}
		writeBytes(p,getInt16AsByte(int16(len(value))),pos)
		writeBytes(p,[]byte(value),pos+2)
		pos+=2+len(value)
		is.current++
	}
}

// ArtistIndex store all artists (avoid double)
type ArtistIndex struct{
	// Used to define if an artist exist (id of artist)
	artists map[string]int
	// Used in write
	tempBuffer []byte
	currentId int
	// new artists
	artistsToSave []string
	currentSave int
}

// LoadArtistIndex Get artist index to search...
func LoadArtistIndex(folder string)ArtistIndex{
	path := filepath.Join(folder,"artist.dico")
	f,err := os.Open(path)
	ai := ArtistIndex{artists:make(map[string]int),currentId:1,artistsToSave:make([]string,0)}
	if err == nil {
		io.Copy(&ai,f)
		f.Close()
	}
	return ai
}

// Add the artist in index. Return id
func (ai * ArtistIndex)Add(artist string)int{
	// Check if exist
	if id,exist := ai.artists[artist] ; exist {
		return id
	}
	id := ai.currentId
	ai.artists[artist] = id
	ai.artistsToSave = append(ai.artistsToSave,artist)
	ai.currentId++
	return id
}

// FindAll return all artists with id
func (ai ArtistIndex)FindAll()map[string]int {
	return ai.artists
}

// Save only new artists
func (ai * ArtistIndex)Save(folder string){
	IndexSaver{ai.artistsToSave,0}.Save(filepath.Join(folder,"artist.dico"),false)
}

// Write get data in p and write in object
// nb artist (4) | lenght name (2) | data name...
func (ai * ArtistIndex)Write(p []byte)(int,error){
	pos := 0
	if ai.artists == nil || len(ai.artists) == 0{
		// Load number, first 4 bytes
		ai.artists = make(map[string]int,int(binary.LittleEndian.Uint32(p[0:4])))
		ai.currentId = 1
		pos=4
	}
	pSize := len(p)
	if ai.tempBuffer != nil && len(ai.tempBuffer) > 0{
		p = append(ai.tempBuffer,p...)
		ai.tempBuffer = nil
	}
	for {
		if pos + 2 > len(p) {
			// Save rest in buffer
			ai.tempBuffer = p[pos:]
			return pSize,nil
		}
		artistSize := int(binary.LittleEndian.Uint16(p[pos:pos+2]))
		if pos + 2 + artistSize > len(p)   {
			ai.tempBuffer = p[pos:]
			return pSize,nil
		}
		ai.artists[string(p[pos+2:pos+2+artistSize])] = ai.currentId
		ai.currentId++
		pos+=2+artistSize
	}
	return pSize,nil
}

// ArtistMusicIndex is an index music by artist. Use id artist and id music.
// Save with temporary method with gob decode / encode
// TODO change with ElementsByFather
type ArtistMusicIndex struct{
	// map with id artist of key and list of music
	MusicsByArtist map[int][]int
}

func (ami * ArtistMusicIndex)Add(artist,music int){
	if musics,present := ami.MusicsByArtist[artist] ; present {
		ami.MusicsByArtist[artist] = append(musics,music)
	}else{
		ami.MusicsByArtist[artist] = []int{music}
	}
}

func (ami * ArtistMusicIndex)Adds(artist int,listMusics []int){
	if musics,present := ami.MusicsByArtist[artist] ; present {
		ami.MusicsByArtist[artist] = append(musics,listMusics...)
	}else{
		ami.MusicsByArtist[artist] = listMusics
	}
}

func (ami ArtistMusicIndex)Save(folder string){
	path := filepath.Join(folder,"artist_music.index")
	f,_ := os.OpenFile(path,os.O_TRUNC|os.O_CREATE|os.O_RDWR,os.ModePerm)
	defer f.Close()
	enc := gob.NewEncoder(f)
	enc.Encode(ami)
}

func LoadArtistMusicIndex(folder string)ArtistMusicIndex{
	path := filepath.Join(folder,"artist_music.index")
	ami := ArtistMusicIndex{}
	if f,err := os.Open(path);err == nil {
		dec := gob.NewDecoder(f)
		dec.Decode(&ami)
		f.Close()
	}else{
		ami.MusicsByArtist = make(map[int][]int)
	}
	return ami
}

// Artist store a list of album : idArtist, list album (id,name)
