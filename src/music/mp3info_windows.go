package music
import "path/filepath"


// GetMp3InfoPath : return path of mp3info. mp3info.exe must be in folder
func GetMp3InfoPath(folder string)string{
    return filepath.Join(folder,"mp3info.exe")
}
