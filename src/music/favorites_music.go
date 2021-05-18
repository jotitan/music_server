package music

import (
	"github.com/jotitan/music_server/logger"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"time"
)

//FavoritesManager manage favorites music by writing a 1 at position music index. If music has 123 as id, the flag at position 123 can be set
type FavoritesManager struct {
	// list of all musics id (8 by bytes). The specific bit is 1 when music is favorite
	musics []byte
	// if true, data must be saved
	toSave bool
}

//NewFavoritesManager create a new favorite manager to load and
func NewFavoritesManager(folder string, nbMusics int) *FavoritesManager {
	fm := FavoritesManager{}
	// If not exist, create with music number
	if f, err := os.OpenFile(filepath.Join(folder, "favorites"), os.O_RDWR, os.ModePerm); err != nil {
		fm.musics = make([]byte, int(math.Ceil(float64(nbMusics)/8)))
	} else {
		defer f.Close()
		// Load favorites from file
		if fm.musics, err = ioutil.ReadAll(f); err != nil {
			fm.musics = make([]byte, int(math.Ceil(float64(nbMusics)/8)))
		}
	}
	go fm.save(folder)
	return &fm
}

// Run save every 5 minutes if necessary
func (fm *FavoritesManager) save(folder string) {
	if fm.toSave {
		if f, err := os.OpenFile(filepath.Join(folder, "favorites"), os.O_RDWR|os.O_CREATE|os.O_CREATE, os.ModePerm); err == nil {
			defer f.Close()
			f.Write(fm.musics)
			logger.GetLogger().Info("Save favorites in file")
		} else {
			logger.GetLogger().Error("Impossible to save error", err.Error())
		}
		fm.toSave = false
	}
	time.Sleep(time.Minute * 5)
	fm.save(folder)
}

//GetFavorites return all favorites
func (fm FavoritesManager) GetFavorites() []int {
	favorites := make([]int, 0, fm.cap())
	for pos, block := range fm.musics {
		for i := 0; i < 8; i++ {
			if block&(1<<uint(i)) != 0 {
				favorites = append(favorites, pos*8+i)
			}
		}
	}
	return favorites
}

//IsFavorite check if a music is a favorite
func (fm FavoritesManager) IsFavorite(musicID int) bool {
	if musicID >= fm.cap() {
		return false
	}
	return fm.musics[musicID/8]&(1<<uint(musicID%8)) != 0
}

//Set a music as favorite, or not
func (fm *FavoritesManager) Set(musicID int, favorite bool) {
	if musicID >= fm.cap() {
		fm.resize(musicID)
	}
	b := uint(fm.musics[musicID/8])
	if favorite {
		// Set a bit 1 at good position with OR operation
		b |= uint(1 << uint((musicID % 8)))
	} else {
		// put only a zero a position with AND operation
		b &= ^(1 << uint(musicID%8))
	}
	fm.musics[musicID/8] = byte(b)
	fm.toSave = true
}

func (fm FavoritesManager) cap() int {
	return len(fm.musics) * 8
}

// resize music array if id can go inside.Compute new length and append empty bytes
func (fm *FavoritesManager) resize(targetID int) {
	extendCapacity := int(math.Ceil(float64(targetID+1-fm.cap()) / 8))
	fm.musics = append(fm.musics, make([]byte, extendCapacity)...)
}
