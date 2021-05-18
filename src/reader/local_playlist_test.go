package reader

import (
	"fmt"
	"testing"
)

func TestAddPlaylist(t *testing.T){
	pl := NewPlaylist()
	pl.Add(1,"path1")
	pl.Add(2,"path2")

	if len(pl.musics) != 2 {
		t.Error(fmt.Sprintf("Must find 2 elements but find %d",len(pl.musics)))
	}
}

func TestRemovePlaylist(t *testing.T){
	pl := NewPlaylist()
	pl.Add(1,"path1")
	pl.Add(2,"path2")
	pl.Add(3,"path3")
	pl.Add(4,"path4")

	pl.Remove(3)
	if len(pl.musics) != 3 {
		t.Error(fmt.Sprintf("Must find 3 elements but find %d",len(pl.musics)))
	}
	if pl.musics[1].idMusic != 2 {
		t.Error(fmt.Sprintf("2 element must have id 1 but found %d",pl.musics[1].idMusic))
	}
	if pl.musics[2].idMusic != 4 {
		t.Error(fmt.Sprintf("3 element must have id 4 but found %d",pl.musics[2].idMusic))
	}
	pl.Remove(0)
	pl.Remove(0)
	if len(pl.musics) != 1 {
		t.Error(fmt.Sprintf("Must find 1 elements but find %d",len(pl.musics)))
	}
	pl.Remove(0)
	if len(pl.musics) != 0 {
		t.Error(fmt.Sprintf("Must find no elments but find %d",len(pl.musics)))
	}
}
