package server

import (
	"github.com/jotitan/music_server/music"
	"net/http"
)

// Create index used by search
func (ms *MusicServer) Index(response http.ResponseWriter, request *http.Request) {
	// Always check addressMask. If no define, mask is 0.0.0.0 and nothing is accepted (except localhost)
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		textIndexer := music.IndexArtists(ms.folder)
		ms.indexManager.UpdateIndexer(textIndexer)
	}
}

// Redindex all data but keep all index in memories to increase treatment
func (ms *MusicServer) FullReindex(response http.ResponseWriter, request *http.Request) {
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		dico := music.LoadDictionnary(ms.folder)
		output := music.NewOutputDictionnary(ms.folder)
		textIndexer := dico.FullReindex(ms.musicFolder, output)
		ms.indexManager.UpdateIndexer(textIndexer)
		// Reload library
		ms.library = music.NewMusicLibrary(ms.folder)
	}
}
