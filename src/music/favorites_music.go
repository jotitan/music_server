package music
import (
    "os"
    "path/filepath"
)

// Manage favorites music by writing a 1 at position music index. If music has 123 as id, the flag at position 123 can be set

type FavoritesManager struct {
    file *os.File
}

func NewFavoritesManager(folder string)FavoritesManager{
    fm := FavoritesManager{}
    // If not exist, create with music number
    if f,err := os.OpenFile(filepath.Join(folder,"favorites"),os.O_RDWR,os.ModePerm) ; err != nil {
        if f,err := os.OpenFile(filepath.Join(folder,"favorites"),os.O_RDWR,os.ModePerm) ; err == nil {
            // Init file
            fm.file = f
        }
    }else{
        fm.file = f
    }
    return fm
}

func (fm * FavoritesManager)Set(idMusic int, favorite bool){

}
