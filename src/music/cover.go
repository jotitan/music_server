package music

import (
	"encoding/xml"
	"errors"
	"github.com/jotitan/music_server/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RootResponse struct {
	XMLName xml.Name    `xml:"metadata"`
	Info    InfoRelease `xml:"release-list"`
	//Releases []Release `xml:"release-list>release"`
}

type InfoRelease struct {
	Count    string    `xml:"count,attr"`
	Releases []Release `xml:"release"`
}

type Release struct {
	//XMLName xml.Name `xml:"release"`
	//Field string `xml:"field"`
	Id    string       `xml:"id,attr"`
	Group ReleaseGroup `xml:"release-group"`
}

type ReleaseGroup struct {
	Id string `xml:"id,attr"`
}

func (rl RootResponse) GetCount() int {
	//logger.GetLogger().Info("=>",rl.Info)
	if val, err := strconv.Atoi(rl.Info.Count); err == nil {
		return val
	}
	return 0
}

func (rl RootResponse) GetUrl() (string, error) {
	if len(rl.Info.Releases) == 0 {
		return "", errors.New("Impossible to get id")
	}
	// read 10 firsts
	for i := 0; i < 10 && i < len(rl.Info.Releases); i++ {
		r := rl.Info.Releases[i]
		if r.Group.Id != "" {
			// Check if exist
			urlImage := "http://coverartarchive.org/release-group/" + r.Group.Id + "/front-250"
			resp, _ := http.Get(urlImage)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			if resp != nil && resp.StatusCode == 200 {
				return urlImage, nil
			}
		}
		if r.Id != "" {
			urlImage := "http://coverartarchive.org/release/" + r.Id + "/front-250"
			resp, _ := http.Get(urlImage)
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			if resp != nil && resp.StatusCode == 200 {
				return urlImage, nil
			}
		}
	}
	return "", errors.New("No id")
}

// @param threashold : max number of results
func extractInfo(data []byte, threashold int) (string, error) {
	var r RootResponse
	xml.Unmarshal(data, &r)
	if threashold != 0 && r.GetCount() > threashold {
		return "", errors.New("Too much results")
	}
	return r.GetUrl()
}

// key if composed of artist-album
var coverCache = make(map[string]string)

// use to request music brainz one time max by second
var lastGet = 0

var waitMBTime = 2000

func checkLastGet() {
	defer func() {
		lastGet = time.Now().Nanosecond()
	}()
	if lastGet == 0 {
		return
	}
	if rest := waitMBTime*1000000 - time.Now().Nanosecond() - lastGet; rest > 0 {
		time.Sleep(time.Nanosecond * time.Duration(rest))
	}
}

var musicBrainzUrl = "http://musicbrainz.org/ws/2/release/?"
var totalMBRequest = 0
var retryMBRequest = 0

func doSearch(params string, threashold int) string {
	fParams := url.Values{"query": []string{params}}.Encode()
	checkLastGet()
	totalMBRequest++
	//logger.GetLogger().Info("req",params,totalMBRequest,retryMBRequest)
	if resp, e := http.Get(musicBrainzUrl + fParams); e == nil {
		defer resp.Body.Close()
		// Quota exceed, relaunch after time
		if resp.StatusCode == 503 {
			retryMBRequest++
			logger.GetLogger().Error("Limit exceed, retry", totalMBRequest, retryMBRequest, params)
			time.Sleep(time.Duration(waitMBTime) * time.Millisecond)
			return doSearch(params, threashold)
		}
		// Check if release field are present, if not, limit
		d, _ := ioutil.ReadAll(resp.Body)
		if cover, err := extractInfo(d, threashold); err == nil {
			return cover
		}
	}
	return ""
}

// Get cover musicbrainz id and check if resource exist
// Rules : check :
// 0 : check if cover.jpg exist in music folder
// 1 : artist + album
// 2 : artist + song title (as release) for single case
// 3 : artist and half album
// 4 : only album with threashold
// 3 : artist only
func GetCover(artist, album, title, filename string) string {
	if artist == "" && album == "" {
		return ""
	}

	// Check if cover.jpg exist in folder, if true, return url like file:/
	if cover, e := os.Open(filepath.Join(filepath.Dir(filename), "cover.jpg")); e == nil {
		// A jpg cover is found, return it
		cover.Close()
		return "/get?src=" + filepath.Join(filepath.Dir(filename), "cover.jpg")
	}
	// Take first artist (before ( , and &)
	formatArtist, _ := regexp.Compile("[a-zA-Z0-9\\. ]*")
	if values := formatArtist.FindAllString(artist, 1); len(values) == 1 {
		artist = values[0]
	}
	album = strings.Replace(strings.ToLower(album), " ost", "", -1)
	if p := strings.Index(album, "cd"); p != -1 {
		album = album[:p]
	}
	key := artist + "-" + album
	if url, ok := coverCache[key]; ok {
		return url
	}

	// 1
	if cover := doSearch("artist:\""+artist+"\" AND release:\""+album+"\"", 0); cover != "" {
		return cover
	}
	// 2
	if cover := doSearch("artist:\""+artist+"\" AND release:\""+title+"\"", 0); cover != "" {
		return cover
	}
	// 3 remove end useless of album : stop at first carac != aZ09., stop before CD and OST
	patternAlbum, _ := regexp.Compile("[a-zA-Z0-9 \\.]*")
	if fAlbum := patternAlbum.FindAllString(album, 1); len(fAlbum) == 1 && len(fAlbum[0]) != len(album) {
		if cover := doSearch("artist:\""+artist+"\" AND release:\""+fAlbum[0]+"\"", 0); cover != "" {
			return cover
		}
	}
	// 4
	if cover := doSearch("release:\""+album+"\"", 50); cover != "" {
		return cover
	}
	// 5
	if cover := doSearch("artist:\""+artist+"\"", 0); cover != "" {
		return cover
	}
	return ""
}
