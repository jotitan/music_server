package music

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
)

/** Index artist, album by genre */
/* final file : list of genre and pointer to each genre with 2 list : artist ids and album ids */

type GenreData struct {
	artists []int
	albums  []int
}

func (gd *GenreData) addArtist(artistId int) {
	gd.artists = append(gd.artists, artistId)
}

func (gd *GenreData) addAlbum(albumId int) {
	gd.albums = append(gd.albums, albumId)
}

type GenreIndexer struct {
	genres     map[string]*GenreData
	genreNames []string
	nameLength int
}

func NewGenreIndexer() GenreIndexer {
	return GenreIndexer{make(map[string]*GenreData), make([]string, 0), 0}
}

func (gi *GenreIndexer) AddArtist(genre string, artistId int) {
	genreData := gi.getGenreData(genre)
	genreData.addArtist(artistId)
}

//AddManyGenresForArtist index many genres for an artist
func (gi *GenreIndexer) AddManyGenresForArtist(genres map[string]struct{}, artistID int) {
	for genre := range genres {
		genreData := gi.getGenreData(genre)
		genreData.addArtist(artistID)
	}
}

func (gi *GenreIndexer) AddAlbum(genre string, albumID int) {
	genreData := gi.getGenreData(genre)
	genreData.addAlbum(albumID)
}

func (gi *GenreIndexer) getGenreData(genre string) *GenreData {
	genreData, ok := gi.genres[genre]
	if !ok {
		genreData = &GenreData{make([]int, 0), make([]int, 0)}
		gi.genres[genre] = genreData
		gi.genreNames = append(gi.genreNames, genre)
		gi.nameLength += len(genre)
	}
	return genreData
}

// Save data into a file
func (gi GenreIndexer) Save(folder string) {
	filename := filepath.Join(folder, "genres.index")
	// Build header : nb genre (4) | header length (4) | length genre 1 (1) | genre 1 (str) | pos genre 1 (4)
	// data structure : nb album (4) | list album (n x 4) | nb artist (4) | list artist (n x 4)
	lengthHeader := 4 + 4 + gi.nameLength + 5*len(gi.genreNames)
	// represent a cursor on where file are
	cursor := lengthHeader
	header := bytes.NewBuffer(make([]byte, 0, lengthHeader))
	header.Write(getIntAsByte(len(gi.genreNames)))
	header.Write(getIntAsByte(lengthHeader - 4))
	for _, genre := range gi.genreNames {
		genreData, _ := gi.genres[genre]
		header.Write([]byte{byte(len(genre))})
		header.Write([]byte(genre))
		header.Write(getIntAsByte(cursor))
		cursor += 8 + 4*(len(genreData.albums)+len(genreData.artists))
	}
	out, _ := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	out.Write(header.Bytes())

	for _, genre := range gi.genreNames {
		genreData, _ := gi.genres[genre]
		data := bytes.NewBuffer(make([]byte, 0, 8+4*(len(genreData.albums)+len(genreData.artists))))
		sort.Ints(genreData.albums)
		sort.Ints(genreData.artists)
		data.Write(getIntAsByte(len(genreData.albums)))
		data.Write(getInts32AsByte(genreData.albums))
		data.Write(getIntAsByte(len(genreData.artists)))
		data.Write(getInts32AsByte(genreData.artists))
		out.Write(data.Bytes())
	}
	out.Close()
}

type GenreReader struct {
	// for each genre, position in file of data
	genres     map[string]int64
	genreNames []string
	filename   string
}

func NewGenreReader(folder string) *GenreReader {
	filename := filepath.Join(folder, "genres.index")
	in, _ := os.Open(filename)
	nbGenre := getNextInt32FromFile(in)
	lengthHeader := getNextInt32FromFile(in)
	header := make([]byte, lengthHeader)
	in.Read(header) 
	genres := make(map[string]int64, nbGenre)
	genreNames := make([]string, nbGenre)
	pos := 0
	for i := 0; i < int(nbGenre); i++ {
		lengthName := int(header[pos])
		name := string(header[pos+1 : pos+1+lengthName])
		pos += 1 + lengthName
		genres[name] = int64(getInt32FromBytes(header[pos : pos+4]))
		genreNames[i] = name
		pos += 4
	}
	in.Close()
	sort.Strings(genreNames)
	return &GenreReader{genres, genreNames, filename}
}

func (gr GenreReader) GetGenres() []string {
	return gr.genreNames
}

// album first in block
func (gr GenreReader) GetAlbum(genre string) map[int]struct{} {
	if pos, exist := gr.genres[genre]; exist {
		in, _ := os.Open(gr.filename)
		nbAlbums := int(getInt32FromFile(in, pos))
		dataAlbums := make([]byte, 4*nbAlbums)
		in.ReadAt(dataAlbums, pos+4)
		results := make(map[int]struct{}, nbAlbums)
		for i := 0; i < nbAlbums; i++ {
			results[int(getInt32FromBytes(dataAlbums[i*4:(i+1)*4]))] = struct{}{}
		}
		return results
	}
	return map[int]struct{}{}
}

// artist in second position
// Return a map, easier to filter
func (gr GenreReader) GetArtist(genre string) map[int]struct{} {
	if pos, exist := gr.genres[genre]; exist {
		in, _ := os.Open(gr.filename)
		nbAlbums := int(getInt32FromFile(in, pos))
		posArtists := pos + (1+int64(nbAlbums))*4
		nbArtists := int(getInt32FromFile(in, posArtists))
		dataArtists := make([]byte, 4*nbArtists)
		in.ReadAt(dataArtists, posArtists+4)
		results := make(map[int]struct{}, nbArtists)
		for i := 0; i < nbArtists; i++ {
			results[int(getInt32FromBytes(dataArtists[i*4:(i+1)*4]))] = struct{}{}
		}
		in.Close()
		return results
	}
	return map[int]struct{}{}
}
