package music

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/jotitan/music_server/logger"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ascherkus/go-id3/src/id3"
)

/* Give methods to browse musics in a specific directory */

const (
	limitMusicFile = 1000
)

// MusicSaver represent the contract to be able to save musics
type MusicSaver interface {
	AddToSave(music Music)
	FinishEnd()
	LoadExistingMusicsInfo() map[string]map[string]string
}

// NewOldMusicSaver create an old implementation
func NewOldMusicSaver(folder string) OldMusicSaver {
	return OldMusicSaver{folder: folder}
}

// OldMusicSaver is the old implementation of index
type OldMusicSaver struct {
	// If file is full, change directory
	changeFolder bool
	folder       string
	header       *[]int64
}

func foundMusicInCache(musics map[string]map[string]string, filename string) (*id3.File, string, int64) {
	if jsonInfo, ok := musics[filename]; ok {
		if info, oldcover, id := fromJSON(jsonInfo); strings.HasSuffix(oldcover, ".mp3") {
			cover := GetCover(info.Artist, info.Album, info.Name, filename)
			logger.GetLogger().Info("Bad cover", oldcover, ", find", cover)
			coverCache[info.Artist+"-"+info.Album] = cover
			return info, cover, id
		} else {
			return info, oldcover, id
		}
	}
	return nil, "", 0
}

// LoadExistingMusicsInfo load already existing music in index
func (oms OldMusicSaver) LoadExistingMusicsInfo() map[string]map[string]string {
	musicsMap := make(map[string]map[string]string)
	for fileID := 0; ; fileID++ {
		path := filepath.Join(oms.folder, fmt.Sprintf("music_%d.dico", fileID))
		logger.GetLogger().Info("Full RI", path)
		if f, err := os.Open(path); err == nil {
			data, _ := ioutil.ReadAll(f)
			total := int(getInt64FromBytes(data[0:8]))
			for j := 0; j < total; j++ {
				id := fileID*limitMusicFile + j + 1
				pos := getInt64FromBytes(data[(j+1)*8 : (j+2)*8])
				lengthInfo := getInt64FromBytes(data[pos : pos+8])
				musicInfo := data[pos+8 : pos+8+lengthInfo]
				var results map[string]string
				json.Unmarshal(musicInfo, &results)
				results["id"] = fmt.Sprintf("%d", id)
				musicsMap[results["path"]] = results
			}
			f.Close()
		} else {
			break
		}
	}
	return musicsMap
}

// FullReindex load all musics in index, browse and find new ones
func (md *MusicDictionnary) FullReindex(folderName string, musicSaver MusicSaver) TextIndexer {
	logger.GetLogger().Info("Launch FullReindex")

	// Load in map (by path) music info
	musics := musicSaver.LoadExistingMusicsInfo()
	// Define nextId as highiest id
	max := int64(0)
	for _, m := range musics {
		if id, err := strconv.ParseInt(m["id"], 10, 32); err == nil && max < id {
			max = id
		}
	}
	md.nextId = max + 1
	logger.GetLogger().Info("Set next Id at ", md.nextId)
	logger.GetLogger().Info("Load", len(musics), "elements")

	// Remove index, dico, index (will be recomputed) and map (used for simple update)
	if dir, err := os.Open(md.indexFolder); err == nil {
		if names, err := dir.Readdirnames(-1); err == nil {
			for _, n := range names {
				if strings.HasSuffix(n, ".index") || strings.HasSuffix(n, ".dico") || strings.HasSuffix(n, ".map") {
					os.Remove(filepath.Join(md.indexFolder, n))
				}
			}
		}
	}
	// Reinit artist index
	md.artistIndex = LoadArtistIndex(md.indexFolder)
	md.artistMusicIndex = LoadArtistMusicIndex(md.indexFolder)

	return md.Browse(folderName, musics, musicSaver)
}

// Browse all musics in root folder, detect unindexed musics and add into library
func (md *MusicDictionnary) Browse(folderName string, musics map[string]map[string]string, musicSaver MusicSaver) TextIndexer {
	// Useless for full reindex (read everything and search in path)
	md.loadIndexedInodes()

	md.browseFolder(folderName, musics, musicSaver)
	musicSaver.FinishEnd()
	md.saveExistingMusic()
	md.artistIndex.Save(md.indexFolder, false)
	md.artistMusicIndex.Save(md.indexFolder)

	return IndexArtists(md.indexFolder)
}

// Load existing music (by inode) in index. Only used for incremental updates
func (md *MusicDictionnary) loadIndexedInodes() {
	md.musicInIndex = make(map[int]struct{})
	// Load list of int32
	if f, err := os.Open(filepath.Join(md.indexFolder, "existing_music.map")); err == nil {
		// Load in memory and parse int32 (or load by block if to heavy)
		data, _ := ioutil.ReadAll(f)
		md.musicInIndex = make(map[int]struct{}, len(data)/4)
		for i := 0; i < len(data)/4; i++ {
			md.musicInIndex[int(binary.LittleEndian.Uint32(data[i*4:(i+1)*4]))] = struct{}{}
		}
		f.Close()
	} else {
		md.musicInIndex = make(map[int]struct{})
	}
}

// Save inodes of readed files into file
func (md *MusicDictionnary) saveExistingMusic() {
	if f, err := os.OpenFile(filepath.Join(md.indexFolder, "existing_music.map"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err == nil {
		list := make([]int, 0, len(md.musicInIndex))
		for inode := range md.musicInIndex {
			list = append(list, inode)
		}
		f.Write(getInts32AsByte(list))
		f.Close()
	}
}

func readInodes(folder string) map[string]int {
	data, _ := exec.Command("ls", folder, "-i1").Output()
	inodes := make(map[string]int)
	r := bufio.NewReader(bytes.NewBuffer(data))
	for {
		if line, _, error := r.ReadLine(); error == nil {
			value := strings.Trim(string(line), " ")
			pos := strings.Index(value, " ")
			if inode, err := strconv.ParseInt(value[:pos], 10, 32); err == nil {
				inodes[value[pos+1:]] = int(inode)
			}
		} else {
			break
		}
	}
	return inodes
}

func (md *MusicDictionnary) browseFolder(folderName string, musics map[string]map[string]string, musicSaver MusicSaver) {
	if folder, err := os.Open(folderName); err == nil {
		defer folder.Close()
		// List all files
		files, _ := folder.Readdir(-1)
		inodes := readInodes(folderName)
		for _, file := range files {
			inode := inodes[file.Name()]
			path := filepath.Join(folderName, file.Name())
			if file.IsDir() {
				if strings.HasPrefix(file.Name(), "_") {
					logger.GetLogger().Info("Escape folder starting width _", file.Name())
				} else {
					logger.GetLogger().Info("Parse", path)
					md.browseFolder(path, musics, musicSaver)
				}
			} else {
				logger.GetLogger().Info("Treat", file.Name())
				if strings.HasSuffix(file.Name(), ".mp3") {
					// Used in update to only import new ones
					if _, exist := md.musicInIndex[inode]; inode == 0 || !exist {
						if info, cover, id := md.extractInfo(path, musics); info != nil {
							m := NewMusic(*info, id, path, cover)
							musicSaver.AddToSave(m)
							// Save artist in index
							for _, artist := range splitArtists(m.file.Artist) {
								idArtist := md.artistIndex.Add(artist)
								md.artistMusicIndex.Add(idArtist, int(m.id))
							}

							logger.GetLogger().Info("Index", m.id, info.Artist, info.Name, info.Album, cover)
							md.musicInIndex[inode] = struct{}{}
						} else {
							logger.GetLogger().Error("Impossible to add", path)
						}
					} else {
						//logger.GetLogger().Info(path, "already in index")
					}
				}
			}
		}
	} else {
		logger.GetLogger().Error(err, folderName)
	}
}

// MusicDictionnary manage music search, index search and music browsing
type MusicDictionnary struct {
	// Next id of music
	nextId int64
	// Directory where indexes are
	indexFolder string
	// Artist index
	artistIndex      ArtistIndex
	artistMusicIndex ArtistMusicIndex
	// Map which contains inode of indexed music
	musicInIndex map[int]struct{}
	// Store the dictionnary of music
	dictionnary *OutputDictionnary
}

func splitArtists(artistsList string) []string {
	reg, _ := regexp.Compile("&|,|/|;")
	vals := reg.Split(artistsList, -1)
	artists := make([]string, len(vals))
	for i, v := range vals {
		artists[i] = strings.Trim(v, " ")
	}
	return artists
}

// LoadDictionnary load the dictionnary which store music info by id
func LoadDictionnary(workingDirectory string) MusicDictionnary {
	md := MusicDictionnary{indexFolder: workingDirectory}

	// Load artist index (list of artist, list of music by artist)
	md.artistIndex = LoadArtistIndex(workingDirectory)
	md.artistMusicIndex = LoadArtistMusicIndex(workingDirectory)
	return md
}

// extractInfo get id3tag info
func (md *MusicDictionnary) extractInfo(filename string, musics map[string]map[string]string) (*id3.File, string, int64) {
	// check in temp cache
	file, cover, id := foundMusicInCache(musics, filename)
	if file != nil {
		return file, cover, id
	}
	r, _ := os.Open(filename)
	defer r.Close()
	music := readMusic(r)
	if music == nil {
		music = &id3.File{}
	}
	if music.Name == "" {
		music.Name = filepath.Base(filename)
	}
	if music.Artist != "" {
		cover = GetCover(music.Artist, music.Album, music.Name, filename)
		coverCache[music.Artist+"-"+music.Album] = cover
	}
	if music.Album == "" {
		music.Album = "Unknown"
	}
	if music.Artist == "" {
		music.Artist = "Unknown"
	}

	// Too long where file is distant, copy in local
	music.Length = md.getTimeMusic(filename)
	id = md.nextId
	md.nextId++
	return music, cover, id
}

func readMusic(file *os.File) *id3.File {
	defer func() {
		if err := recover(); err != nil {
			logger.GetLogger().Error("Impossible to read", err)
		}
	}()
	return id3.Read(file)

}

func (md MusicDictionnary) getTimeMusic(filename string) string {
	f, _ := os.Open(filename)

	tmpName := fmt.Sprintf("%d", time.Now().Nanosecond())
	tmpPath := filepath.Join(os.TempDir(), tmpName)

	ftmp, _ := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	io.Copy(ftmp, f)
	f.Close()
	ftmp.Close()

	defer os.Remove(tmpPath)

	mp3InfoPath := GetMp3InfoPath(md.indexFolder)
	cmd := exec.Command(mp3InfoPath, "-p", "%S", tmpPath)

	if result, error := cmd.Output(); error == nil {
		return string(result)
	}
	return ""
}
