package music

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"logger"
	"os"
	"path/filepath"
	"sort"
)

const (
	limitMusicsInFile = 1000
)

// Save original dictionnary in new structure : one file header with for all id, file id and position in file. Some files with informatiosn inside
// For reading, load header file in memory and keep until next scan

// Structure to read musics
type MusicLibrary struct {
	folder string
	// For each music, store the position in int64 : 4 first bytes for the fileId and 4 next for position in file
	musicPositions map[int32]int64
}

func NewMusicLibrary(folder string) *MusicLibrary {
	ml := &MusicLibrary{folder: folder}
	// Preload index in memory
	ml.loadMusicIndex()
	return ml
}

//loadMusicIndex load all index (idMusic=>position)
func (ml *MusicLibrary) loadMusicIndex() {
	if f, err := os.Open(filepath.Join(ml.folder, "dico_index.dico")); err == nil {
		defer f.Close()
		nbMusics := getInt32FromFile(f, int64(0))
		data := make([]byte, 12*nbMusics)
		f.ReadAt(data, 4)
		ml.musicPositions = make(map[int32]int64, nbMusics)
		for i := int32(0); i < nbMusics; i++ {
			musicID := int32(binary.LittleEndian.Uint32(data[i*12 : i*12+4]))
			position := int64(binary.LittleEndian.Uint64(data[i*12+4 : i*12+12]))
			ml.musicPositions[musicID] = position
		}
		logger.GetLogger().Info("Load", nbMusics, "musics from dictionnary")
	} else {
		logger.GetLogger().Error("Impossible to load index data", err.Error())
	}
}

func (ml MusicLibrary) GetPosition(id int32) int64 {
	return ml.musicPositions[id]
}

func (ml MusicLibrary) GetNbMusics() int {
	return len(ml.musicPositions)
}

//GetMusicInfoAsJSON
func (ml MusicLibrary) GetMusicInfoAsJSON(id int32, isfavorite bool) []byte {
	musicInfo := ml.GetMusicInfo(id)
	delete(musicInfo, "path")
	musicInfo["id"] = fmt.Sprintf("%d", id)
	musicInfo["src"] = fmt.Sprintf("music?id=%d", id)
	if isfavorite {
		musicInfo["favorite"] = "true"
	}
	bdata, _ := json.Marshal(musicInfo)

	return bdata
}

//GetMusicInfo return music info from id
func (ml MusicLibrary) GetMusicInfo(id int32) map[string]string {
	// Find position in header
	if pointer, ok := ml.musicPositions[id]; ok {
		fileId := int32(pointer >> 32)
		position := int64(int32(pointer))
		filename := filepath.Join(ml.folder, fmt.Sprintf("dico_music_%d.dico", fileId))
		if dataFile, err := os.Open(filename); err == nil {
			// Read size data
			size := getInt64FromFile(dataFile, position)
			musicInfo := make(map[string]string)
			json.Unmarshal(getBytesFromFile(dataFile, position+8, size), &musicInfo)
			return musicInfo
		}
	}
	return nil
}

//GetMusicsInfo return musics information from ids
func (ml MusicLibrary) GetMusicsInfo(ids []int32) []map[string]string {
	// Find positions in header and group by fileId
	positionsByFileId := make(map[int32][]int)
	for _, id := range ids {
		if pointer, ok := ml.musicPositions[id]; ok {
			fileId := int32(pointer >> 32)
			position := int(int32(pointer))
			if positions, exist := positionsByFileId[fileId]; exist {
				positionsByFileId[fileId] = append(positions, position)
			} else {
				positionsByFileId[fileId] = []int{position}
			}
		}
	}
	musicsInfo := make([]map[string]string, 0, len(ids))
	// For each file, search every positions inside
	for fileId, positions := range positionsByFileId {
		filename := filepath.Join(ml.folder, fmt.Sprintf("dico_music_%d.dico", fileId))
		if dataFile, err := os.Open(filename); err == nil {
			defer dataFile.Close()
			// Sort to increase performance by sequential access
			sort.Ints(positions)
			for _, position := range positions {
				musicsInfo = append(musicsInfo, readMusicFromFile(dataFile, int64(position)))
			}
		}
	}
	return musicsInfo
}

func readMusicFromFile(dataFile *os.File, position int64) map[string]string {
	// Read data size
	size := getInt64FromFile(dataFile, position)
	musicInfo := make(map[string]string)
	json.Unmarshal(getBytesFromFile(dataFile, position+8, size), &musicInfo)
	return musicInfo
}

// Used to save dictionnary in files
type OutputDictionnary struct {
	folder string
	// position (file and byte position) for each id
	headerIds map[int32]int64
	// List of ids, pointer to modify without pointer on object
	ids *[]int
	// Actual number of element to flush in file
	musicToSave *[]Music

	// Used for saving
	currentMusicRead      *int
	currentPositionInFile *int
	// Id of the file where musics are saved
	currentFileId *int32
}

// Create dictionnary from old version
func CreateNewDictionnary(fromFolderOldVersion, toFolderNewVersion string) {
	oldDico := NewOldMusicSaver(fromFolderOldVersion)
	musics := oldDico.LoadExistingMusicsInfo()
	output := NewOutputDictionnary(toFolderNewVersion)

	// Browse all path
	for path := range musics {
		m, cover, id := foundMusicInCache(musics, path)
		output.AddToSave(NewMusic(*m, id, path, cover))
	}
	output.FinishEnd()

	// Copy list artist
	ai := LoadArtistIndex(fromFolderOldVersion)
	ai.artistsToSave = make([]string, len(ai.artists))
	for name, id := range ai.artists {
		ai.artistsToSave[id-1] = name
	}
	ai.Save(toFolderNewVersion, true)

	ami := LoadArtistMusicIndex(fromFolderOldVersion)
	ami.Save(toFolderNewVersion)
}

// Load all existing musics in memory. Used to avoid full reparsing
func (od OutputDictionnary) LoadExistingMusicsInfo() map[string]map[string]string {
	musicsMap := make(map[string]map[string]string)
	for fileId := 0; ; fileId++ {
		path := filepath.Join(od.folder, fmt.Sprintf("dico_music_%d.dico", fileId))
		logger.GetLogger().Info("Read existing musics", path)
		if f, err := os.Open(path); err == nil {
			data, _ := ioutil.ReadAll(f)
			pos := int64(0)
			for {
				if pos >= int64(len(data)) {
					break
				}
				lengthMusic := getInt64FromBytes(data[pos : pos+8])
				dataMusic := data[pos+8 : pos+8+lengthMusic]
				var results map[string]string
				json.Unmarshal(dataMusic, &results)
				musicsMap[results["path"]] = results
				pos += 8 + lengthMusic
			}
			f.Close()
		} else {
			break
		}
	}
	return musicsMap
}

func NewOutputDictionnary(folder string) *OutputDictionnary {
	musicsList := make([]Music, 0, limitMusicsInFile)
	ids := make([]int, 0)
	currentFileId := int32(0)
	currentMusicRead := 0
	currentPositionInFile := 0
	return &OutputDictionnary{
		folder:                folder,
		headerIds:             make(map[int32]int64),
		ids:                   &ids,
		musicToSave:           &musicsList,
		currentFileId:         &currentFileId,
		currentMusicRead:      &currentMusicRead,
		currentPositionInFile: &currentPositionInFile}
}

func (od OutputDictionnary) AddToSave(music Music) {
	if len(*od.musicToSave) >= limitMusicsInFile {
		od.save()
	}
	*(od.musicToSave) = append(*(od.musicToSave), music)
}

func (od OutputDictionnary) FinishEnd() {
	od.save()
	od.saveHeader()
	logger.GetLogger().Info("End saving dictionnary")
}

func (od OutputDictionnary) saveHeader() {
	logger.GetLogger().Info("Save header :", len(*od.ids), "elements")
	sort.IntsAreSorted(*od.ids)
	data := make([]byte, 12*len(*od.ids)+4)
	writeBytes(data, getInt32AsByte(int32(len(*od.ids))), 0)
	for pos, id := range *od.ids {
		writeBytes(data, getInt32AsByte(int32(id)), 4+pos*12)
		writeBytes(data, getInt64AsByte(od.headerIds[int32(id)]), 4+pos*12+4)
	}
	// Save in file
	if f, err := os.OpenFile(filepath.Join(od.folder, "dico_index.dico"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err == nil {
		defer f.Close()
		f.Write(data)
	} else {
		logger.GetLogger().Error("Impossible to asve index dico", err.Error())
	}
}

func (od OutputDictionnary) save() {
	filename := filepath.Join(od.folder, fmt.Sprintf("dico_music_%d.dico", *od.currentFileId))
	logger.GetLogger().Info("Save in file", filename, ":", len(*od.musicToSave), "elements")
	if f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err == nil {
		defer f.Close()
		io.Copy(f, &od)
		*od.currentFileId++
		*od.currentMusicRead = 0
		*od.currentPositionInFile = 0
	}
	*od.musicToSave = make([]Music, 0, limitMusicsInFile)
}

func (od *OutputDictionnary) Read(tab []byte) (int, error) {
	// Read md musics, evaluate if enough place in tab (int32 for length + len data
	nbWrite := 0
	for {
		// Check if all data at been read
		if *od.currentMusicRead >= len(*od.musicToSave) {
			return nbWrite, io.EOF
		}
		l := *(od.musicToSave)
		music := l[*od.currentMusicRead]
		data := music.toJSON()
		if size := 8 + len(data); nbWrite+size < len(tab) {
			writeBytes(tab, getInt64AsByte(int64(len(data))), nbWrite)
			writeBytes(tab, data, nbWrite+8)
			nbWrite += size
			*od.currentMusicRead++
			position := int64(*od.currentFileId)<<32 + int64(*od.currentPositionInFile)
			*od.currentPositionInFile += size
			od.headerIds[int32(music.id)] = position
			*od.ids = append(*od.ids, int(music.id))
		} else {
			break
		}
	}
	return nbWrite, nil
}
