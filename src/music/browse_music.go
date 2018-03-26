package music

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"logger"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mjibson/id3"
)

/* Give methods to browse musics in a specific directory */

const (
	limitMusicFile = 1000
)

type MusicSaver interface {
	AddToSave(music Music)
	FinishEnd()
	LoadExistingMusicsInfo() map[string]map[string]string
}

func NewOldMusicSaver(folder string) OldMusicSaver {
	return OldMusicSaver{folder: folder}
}

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

func (oms OldMusicSaver) LoadExistingMusicsInfo() map[string]map[string]string {
	musicsMap := make(map[string]map[string]string)
	for fileId := 0; ; fileId++ {
		path := filepath.Join(oms.folder, fmt.Sprintf("music_%d.dico", fileId))
		logger.GetLogger().Info("Full RI", path)
		if f, err := os.Open(path); err == nil {
			data, _ := ioutil.ReadAll(f)
			total := int(getInt64FromBytes(data[0:8]))
			for j := 0; j < total; j++ {
				id := fileId*limitMusicFile + j + 1
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

// Save the file. Header with fix size (fixe number of element, 10000 for example). Header link position of element.
// nb element (8) | pos 1 (8) | pos 2 (8) | ... | dataLength 1 (4) | data 1 (n) | ...
/*func (oms * OldMusicSaver)Save(){
	// Save in the file. Create data in buffer instead of write everytime
	fileId,_ := findLastFile("","")
	// Get next file
	if oms.changeFolder {
		fileId++
		oms.changeFolder = false
	}
	path := filepath.Join("folder",fmt.Sprintf("music_%d.dico",fileId))
	logger.GetLogger().Info("Save in file",path)
	f,err := os.OpenFile(path,os.O_CREATE|os.O_RDWR|os.O_EXCL,os.ModePerm)
	// If exist, just append result at the end
	*oms.header = make([]int64,0,len(oms.musics))
	headerPos := int64(8)
	totalElements := int64(0)
	if err != nil {
		f,_ = os.OpenFile(path,os.O_RDWR,os.ModePerm)
		defer f.Close()
		info,_ := f.Stat()
		*oms.header = append(*oms.header,info.Size())
		// Get total elements
		totalElements = getInt64FromFile(f,0)

		if totalElements == limitMusicFile {
			oms.changeFolder = true
			oms.Save()
			return
		}

		f.Seek(0,2)	// Back to the end
		// Position in header depend on number element
		headerPos += totalElements*8
	}else{
		defer f.Close()
		// Create header at begin
		//f.Write(getInt64AsByte(int64(len(md.musics))))
		f.Write(make([]byte,limitMusicFile*8))
		//md.header = append(md.header,8*(1+limitMusicFile))
	}
	//md.currentRead = 0
	// Use a reader over md. Write header at the end
	//io.Copy(f,md)
	// Write total elements
	//f.WriteAt(getInt64AsByte(totalElements + int64(len(md.musics))),0)
	// Write header
	//f.WriteAt(getInts64AsByte(md.header[:len(md.header)-1]),headerPos)
}  */

// Read used in copy to save data in file
/*func (oms * OldMusicSaver)Read(tab []byte)(int,error){
	// Read md musics, evaluate if enough place in tab (int32 for length + len data
	nbWrite := 0
	for{
		// Check if all data at been read
		if md.currentRead >= len(md.musics){
			return nbWrite,io.EOF
		}
		data := md.musics[md.currentRead].toJSON()
		if size := 8 + len(data) ; nbWrite + size < len(tab) {
			writeBytes(tab,getInt64AsByte(int64(len(data))),nbWrite)
			writeBytes(tab,data,nbWrite+8)
			nbWrite+=size
			md.currentRead++
			// Save position in temp header
			// First case, init
			md.header = append(md.header,md.header[len(md.header)-1]+int64(size))
		}else{
			break
		}
	}

	return nbWrite,nil
}  */

// Find id file with biggest id
/*func findLastFile(folder,pattern string)(int64,error){
	r,_ := regexp.Compile(pattern)
	max := int64(-1)
	filesFolder,_ := os.Open(folder)
	defer filesFolder.Close()
	names,_ := filesFolder.Readdirnames(-1)
	for _,name := range names {
		if result := r.FindStringSubmatch(name) ; len(result) >1 {
			if id,err := strconv.ParseInt(result[1],10,32) ; err == nil && id > max {
				max = id
			}
		}
	}
	if max == -1 {
		return 0,errors.New("No dictionnary yet")
	}
	return max,nil
}  */

// Download existing index to index new musics (key is music path)
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

	return md.Browse2(folderName, musics, musicSaver)
}

// Browse a folder to get all data
/*func (md * MusicDictionnary)Browse(folderName string, musicSaver MusicSaver)TextIndexer{
	logger.GetLogger().Info("Begin index")
	dictionnary := LoadDictionnary(md.indexFolder)
	return dictionnary.Browse2(folderName,musicSaver)
} */

func (md *MusicDictionnary) Browse2(folderName string, musics map[string]map[string]string, musicSaver MusicSaver) TextIndexer {
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

type MusicDictionnary struct {
	//previousSize int
	// Used to store header to write. List of position
	//header []int64
	// Nb of element when opening file
	//currentRead int
	// Next id for file
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

/*func (md MusicDictionnary)currentSize()int{
	return md.previousSize + len(md.musics)
} */

// return err
/*func (md MusicDictionnary)findLastFile()(int64,error){
	return findLastFile(md.indexFolder,"music_([0-9]+).dico")
} */

func NewMusic(data id3.File, idMusic int64, path, cover string) Music {
	return Music{file: data, id: idMusic, path: path, cover: cover}
}

// Algo to keep same music id when reinindex :
// 1) Save id in structure when loading
// 2) Detect useless id for creating sequence. After using all missing, use normal sequence starting at number of element
// 2) When indexing, if id music belong to range, insert in good place

// Add music in dictionnary. If file limit is reach, save the file
/*func (md * MusicDictionnary)Add(musicInfo Music){
	if md.currentSize() >= limitMusicFile {
		//md.Save()
		//md.changeFolder = true
		// Save, use new file
		md.musics = make([]Music,0,limitMusicFile)
		md.previousSize = 0
	}
	md.musics = append(md.musics, musicInfo)
	// split artist when & or / or , is present
	for _,artist := range  splitArtists(musicInfo.file.Artist) {
		idArtist := md.artistIndex.Add(artist)
		md.artistMusicIndex.Add(idArtist,int(musicInfo.id))
	}
} */

func splitArtists(artistsList string) []string {
	reg, _ := regexp.Compile("&|,|/|;")
	vals := reg.Split(artistsList, -1)
	artists := make([]string, len(vals))
	for i, v := range vals {
		artists[i] = strings.Trim(v, " ")
	}
	return artists
}

// Get many musics by id
//@Deprecated
/*func (md MusicDictionnary)GetMusicsFromIds(ids []int)[]map[string]string{
	musicResults := make([]map[string]string,0,len(ids))
	// Group ids by file id
	groupsIds := make(map[int][]int)
	for _,id := range ids {
		fileId := (id-1) / limitMusicFile
		if group,ok := groupsIds[fileId] ; ok {
			groupsIds[fileId] = append(group,id)
		}else{
			groupsIds[fileId] = []int{id}
		}
	}
	for fileId,musicsId := range groupsIds {
		path := filepath.Join(md.indexFolder,fmt.Sprintf("music_%d.dico",fileId))
		if f,err := os.Open(path) ; err == nil {
			defer f.Close()
			// Load all musics
			for _,id := range musicsId {
				pos := int64(id - fileId*limitMusicFile)*8
				posInFile := getInt64FromFile(f,pos)
				lengthData := getInt64FromFile(f,posInFile)
				data := make([]byte,lengthData)
				f.ReadAt(data,posInFile+8)

				var results map[string]string
				json.Unmarshal(data,&results)
				results["id"] = fmt.Sprintf("%d",id)
				musicResults = append(musicResults,results)
			}
		}
	}
	return musicResults
}  */

// GetMusicFromId return the music to an id
/*func (md MusicDictionnary)GetMusicFromId(id int)map[string]string{
	// Id begin at 1
	fileId := (id-1) / limitMusicFile

	path := filepath.Join(md.indexFolder,fmt.Sprintf("music_%d.dico",fileId))
	if f,err := os.Open(path) ; err == nil {
		defer f.Close()
		pos := int64(id - fileId*limitMusicFile)*8
		posInFile := getInt64FromFile(f,pos)
		lengthData := getInt64FromFile(f,posInFile)
		data := make([]byte,lengthData)
		f.ReadAt(data,posInFile+8)

		var results map[string]string
		json.Unmarshal(data,&results)
		return results
	}
	return nil
}  */

/*func NewDictionnary(workingDirectory string)MusicDictionnary {
	return MusicDictionnary{indexFolder:workingDirectory}
} */

/*func GetNbMusics(workingDirectory string)int64{
	md := LoadDictionnary(workingDirectory)
	return md.nextId-1
} */

// LoadDictionnary load the dictionnary which store music info by id
func LoadDictionnary(workingDirectory string) MusicDictionnary {
	md := MusicDictionnary{indexFolder: workingDirectory}

	// Load music info
	/*fileId,notExist := md.findLastFile()
	if notExist == nil{
		// Load the last file and get current element inside
		path := filepath.Join(md.indexFolder,fmt.Sprintf("music_%d.dico",fileId))
		f,_ := os.Open(path)
		defer f.Close()
		tabNb := make([]byte,8)
		f.ReadAt(tabNb,0)
		md.previousSize = int(getInt64FromFile(f,0))
		md.nextId = fileId*limitMusicFile + int64(md.previousSize+1)
	}else{
		md.previousSize = 0
		md.nextId = 1
	} */

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
	music := id3.Read(r)
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

func (md MusicDictionnary) getTimeMusic(filename string) string {
	f, _ := os.Open(filename)
	fmt.Sprintf("%v", f.Fd())

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
