package music
import (
    "os"
    "logger"
    "path/filepath"
    "github.com/mjibson/id3"
    "encoding/json"
    "io"
	"strings"
	"regexp"
	"strconv"
	"fmt"
	"errors"
	"os/exec"
	"time"
)

/* Give methods to browse musics in a specific directory */


const (
	limitMusicFile = 1000
)

// Browse a folder to get all data
func (md * MusicDictionnary)Browse(folderName string){
    dictionnary := LoadDictionnary(md.indexFolder)

	dictionnary.browseFolder(folderName)
	dictionnary.Save()
	dictionnary.artistIndex.Save(md.indexFolder)
	dictionnary.artistMusicIndex.Save(md.indexFolder)

	IndexArtists(md.indexFolder)
}

func (md * MusicDictionnary)browseFolder(folderName string){
	if folder,err := os.Open(folderName) ; err == nil {
		defer folder.Close()
		// List all files
		files,_ := folder.Readdir(-1)
		for _,file := range files {
			path := filepath.Join(folderName,file.Name())
			if file.IsDir() {
				logger.GetLogger().Info("Parse",path)
				md.browseFolder(path)
			} else{
				// TODO : check in bloomfilter if exist. If true, final check in dictionnary
				if strings.HasSuffix(file.Name(),".mp3") {
					if info := md.extractInfo(path) ; info != nil {
						logger.GetLogger().Info("Index",info)
						md.Add(path, *info)
					}else{
						logger.GetLogger().Error("Impossible to add",path)
					}
				}
			}
		}
	} else{
		logger.GetLogger().Error(err,folderName)
	}
}

type MusicDictionnary struct{
    previousSize int
    // Data in json format
    musics []Music
    // Used to store header to write. List of position
    header []int64
    // Nb of element when opening file
    currentRead int
	// If file is full, change directory
	changeFolder bool
	// Next id for file
	nextId int64
    // Directory where indexes are
    indexFolder string
	// Artist index
	artistIndex ArtistIndex
    artistMusicIndex ArtistMusicIndex
}

func (md MusicDictionnary)currentSize()int{
    return md.previousSize + len(md.musics)
}

type Music struct{
	file id3.File
    path string
	id int64
}

func (m Music)toJSON()[]byte{
	data := make(map[string]string,7)
	data["title"] = m.file.Name
	data["artist"] = m.file.Artist
	data["album"] = m.file.Album
	data["length"] = m.file.Length
	data["year"] = m.file.Year
	data["genre"] = m.file.Genre
	data["track"] = m.file.Track
    data["path"] = m.path
	jsonData,_ := json.Marshal(data)
	return jsonData
}

// Find id file with biggest id
func findLastFile(folder,pattern string)(int64,error){
    r,_ := regexp.Compile(pattern)
    max := int64(-1)
    filesFolder,_ := os.Open(folder)
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
}

// return err
func (md MusicDictionnary)findLastFile()(int64,error){
	return findLastFile(md.indexFolder,"music_([0-9]+).dico")
}

// Save the file. Header with fix size (fixe number of element, 10000 for example). Header link position of element.
// nb element (8) | pos 1 (8) | pos 2 (8) | ... | dataLength 1 (4) | data 1 (n) | ...
func (md * MusicDictionnary)Save(){
    // Save in the file. Create data in buffer instead of write everytime
    fileId,_ := md.findLastFile()
	// Get next file
	if md.changeFolder {
		fileId++
		md.changeFolder = false
	}
	path := filepath.Join(md.indexFolder,fmt.Sprintf("music_%d.dico",fileId))
	logger.GetLogger().Info("Save in file",path)
    f,err := os.OpenFile(path,os.O_CREATE|os.O_RDWR|os.O_EXCL,os.ModePerm)
    // If exist, just append result at the end
    md.header = make([]int64,0,len(md.musics))
    headerPos := int64(8)
	totalElements := int64(0)
	if err != nil {
        f,_ = os.OpenFile(path,os.O_RDWR,os.ModePerm)
		defer f.Close()
        info,_ := f.Stat()
        md.header = append(md.header,info.Size())
		// Get total elements
		totalElements = getInt64FromFile(f,0)

		if totalElements == limitMusicFile {
			md.changeFolder = true
			md.Save()
			return
		}

		f.Seek(0,2)	// Back to the end
		// Position in header depend on number element
		headerPos += totalElements*8
    }else{
		defer f.Close()
        // Create header at begin
        f.Write(getInt64AsByte(int64(len(md.musics))))
        f.Write(make([]byte,limitMusicFile*8))
        md.header = append(md.header,8*(1+limitMusicFile))
    }
    md.currentRead = 0
    // Use a reader over md. Write header at the end
    io.Copy(f,md)
	// Write total elements
	f.WriteAt(getInt64AsByte(totalElements + int64(len(md.musics))),0)
    // Write header
    f.WriteAt(getInts64AsByte(md.header[:len(md.header)-1]),headerPos)
}

// Read used in copy to save data in file
func (md * MusicDictionnary)Read(tab []byte)(int,error){
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
}

// Add music in dictionnary. If file limit is reach, save the file
func (md * MusicDictionnary)Add(path string,music id3.File){
    if md.currentSize() >= limitMusicFile {
        md.Save()
		md.changeFolder = true
        // Save, use new file
        md.musics = make([]Music,0,limitMusicFile)
		md.previousSize = 0
    }
    idMusic := md.nextId
    md.nextId++
    md.musics = append(md.musics, Music{file:music,id:idMusic,path:path})
	idArtist := md.artistIndex.Add(music.Artist)
	md.artistMusicIndex.Add(idArtist,int(idMusic))
}

// Get many musics by id
func (md MusicDictionnary)GetMusicsFromIds(ids []int)[]map[string]string{
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
				musicResults = append(musicResults,results)
			}
		}
	}
	return musicResults
}

// GetMusicFromId return the music to an id
func (md MusicDictionnary)GetMusicFromId(id int)map[string]string{
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
}

func NewDictionnary(workingDirectory string)MusicDictionnary {
    return MusicDictionnary{changeFolder:false, indexFolder:workingDirectory}
}

// LoadDictionnary load the dictionnary which store music info by id
func LoadDictionnary(workingDirectory string)MusicDictionnary{
    md := MusicDictionnary{changeFolder:false,indexFolder:workingDirectory}

	// Load music info
    fileId,notExist := md.findLastFile()
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
        md.musics = make([]Music,0,limitMusicFile)
		md.nextId = 1
    }

    // Load artist index (list of artist, list of music by artist)
	md.artistIndex = LoadArtistIndex(workingDirectory)
	md.artistMusicIndex = LoadArtistMusicIndex(workingDirectory)
    return md
}

// extractInfo get id3tag info
func (md MusicDictionnary)extractInfo(filename string)*id3.File{
  	r,_ := os.Open(filename)
	defer r.Close()
	music := id3.Read(r)
	if music == nil {
		music = &id3.File{}
	}
	if music.Name == "" {
		music.Name = filepath.Base(filename)
	}
	if music.Album == "" {
		music.Album = "Unknown"
	}
	if music.Artist == "" {
		music.Artist = "Unknown"
	}
	// Too long where file is distant, copy in local
	music.Length = md.getTimeMusic(filename)

	return music
}


func (md MusicDictionnary)getTimeMusic(filename string) string{
	f,_ := os.Open(filename)
	fmt.Sprintf("%v",f.Fd())

	tmpName := fmt.Sprintf("%d",time.Now().Nanosecond())
	tmpPath := filepath.Join(os.TempDir(),tmpName)

	ftmp,_ := os.OpenFile(tmpPath,os.O_CREATE|os.O_RDWR,os.ModeTemporary)
	io.Copy(ftmp,f)
	f.Close()
	ftmp.Close()

	defer os.Remove(tmpPath)

	mp3InfoPath := filepath.Join(md.indexFolder,"mp3info.exe")
	cmd := exec.Command(mp3InfoPath,"-p","%S",tmpPath)
	if result,error := cmd.Output() ; error == nil {
		return string(result)
	}
	return ""
}
