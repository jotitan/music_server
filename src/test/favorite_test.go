package test

import (
	"fmt"
	"music"
	"testing"
)

func TestFavorite(t *testing.T) {
	fv := music.NewFavoritesManager("", 100)

	fv.Set(12, true)
	fv.Set(17, true)
	fv.Set(25, true)
	fv.Set(87, true)
	fv.Set(112, true)
	fv.Set(145, true)
	testFavorite(fv, 11, false, t)
	testFavorite(fv, 12, true, t)
	testFavorite(fv, 13, false, t)
	testFavorite(fv, 14, false, t)
	testFavorite(fv, 17, true, t)
	testFavorite(fv, 18, false, t)
	testFavorite(fv, 23, false, t)
	testFavorite(fv, 25, true, t)
	testFavorite(fv, 27, false, t)
	testFavorite(fv, 81, false, t)
	testFavorite(fv, 86, false, t)
	testFavorite(fv, 87, true, t)
	testFavorite(fv, 100, false, t)
	testFavorite(fv, 112, true, t)
	testFavorite(fv, 113, false, t)
	testFavorite(fv, 115, false, t)
	testFavorite(fv, 140, false, t)
	testFavorite(fv, 145, true, t)
	testFavorite(fv, 144, false, t)
	testFavorite(fv, 146, false, t)
}

func testFavorite(fv *music.FavoritesManager, value int, result bool, t *testing.T) {
	if r := fv.IsFavorite(value); r != result {
		t.Error(fmt.Sprintf("Must be %t but found %t", result, r))
	}
}
