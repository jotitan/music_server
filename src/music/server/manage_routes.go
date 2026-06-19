package server

import (
	"github.com/jotitan/music_server/music"
	"net/http"
)

// Index Create index used by search
func (ms *MusicServer) Index(_ http.ResponseWriter, request *http.Request) {
	// Always check addressMask. If no define, mask is 0.0.0.0 and nothing is accepted (except localhost)
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		textIndexer, am := music.IndexArtists(ms.folder)
		ms.indexManager.UpdateIndexer(textIndexer)
		ms.indexManager.UpdateAlbumManager(am)
	}
}

// FullReindex all data but keep all index in memories to increase treatment
func (ms *MusicServer) FullReindex(_ http.ResponseWriter, request *http.Request) {
	if !ms.checkRequester(request) {
		return
	}
	if ms.musicFolder != "" {
		dico := music.LoadDictionary(ms.folder)
		output := music.NewOutputDictionary(ms.folder)
		textIndexer := dico.FullReindex(ms.musicFolder, output)
		ms.indexManager.UpdateIndexer(textIndexer)
		// Reload library
		ms.library = music.NewMusicLibrary(ms.folder)
	}
}
