package music
import (
    "errors"
    "net/http"
    "encoding/xml"
    "net/url"
    "time"
    "io/ioutil"
    "logger"
    "regexp"
)

type RootResponse struct {
    XMLName xml.Name `xml:"metadata"`
    Releases []Release `xml:"release-list>release"`
}

type Release struct {
    //XMLName xml.Name `xml:"release"`
    //Field string `xml:"field"`
    Id string `xml:"id,attr"`
    Group ReleaseGroup `xml:"release-group"`
}

type ReleaseGroup struct {
    Id string `xml:"id,attr"`
}

func (rl RootResponse)GetUrl()(string,error){
    if len(rl.Releases) == 0 {
        return "",errors.New("Impossible to get id")
    }
    // read 10 firsts
    for i:= 0 ; i < 10 && i < len(rl.Releases) ; i++{
        r := rl.Releases[i]
        if r.Group.Id != "" {
            // Check if exist
            urlImage := "http://coverartarchive.org/release-group/" + r.Group.Id + "/front-250"
            resp,_ := http.Get(urlImage)
            if resp != nil && resp.Body!=nil {
                defer resp.Body.Close()
            }
            if resp.StatusCode == 200 {
                return urlImage,nil
            }
        }
        if r.Id != "" {
            urlImage := "http://coverartarchive.org/release/" + r.Id + "/front-250"
            resp,_ := http.Get(urlImage)
            if resp != nil && resp.Body!=nil {
                defer resp.Body.Close()
            }
            if resp.StatusCode == 200 {
                return urlImage,nil
            }
        }
    }
    return "",errors.New("No id")
}

func extractInfo(data []byte)(string,error){
    var r RootResponse
    xml.Unmarshal(data,&r)
    return r.GetUrl()
}

// key if composed of artist-album
var coverCache = make(map[string]string)

// use to request music brainz one time max by second
var lastGet = 0

var waitMBTime = 2000

func checkLastGet(){
    defer func(){
        lastGet = time.Now().Nanosecond()
    }()
    if lastGet == 0 {
        return
    }
    if rest := waitMBTime * 1000000 - time.Now().Nanosecond() - lastGet ; rest > 0 {
        time.Sleep(time.Nanosecond * time.Duration(rest))
    }
}

var musicBrainzUrl = "http://musicbrainz.org/ws/2/release/?"
var totalMBRequest = 0
var retryMBRequest = 0

// Get cover musicbrainz id and check if resource exist
// Rules : check :
// 1 : artist + album, then artist + first real word (not the, a...) of album
// 2 : artist + song title (as release) for single case
// 3 : artist only
func GetCover(artist,album,title string)string{
    // Take first artist (before ( , and &)
    formatArtist,_ := regexp.Compile("[a-zA-Z0-9 ]*")
    if values := formatArtist.FindAllString(artist,1) ; len(values) == 1 {
        artist = values[0]
    }
    key := artist + "-" + album
    if url,ok:= coverCache[key] ; ok {
        return url
    }
    params := "artist:\"" + artist + "\"";
    if album != "" {
        params+=" AND release:\"" + album + "\""
    } else{
        if title !="" {
            params+=" AND release:\"" + title + "\""
        }
    }
    params = url.Values{"query":[]string{params}}.Encode()
    checkLastGet()
    totalMBRequest++
    logger.GetLogger().Info("req",params,totalMBRequest,retryMBRequest)
    if resp,e := http.Get(musicBrainzUrl + params); e == nil {
        defer resp.Body.Close()
        // Quota exceed, relaunch after time
        if resp.StatusCode == 503 {
            retryMBRequest++
            logger.GetLogger().Error("Limit exceed, retry",totalMBRequest,retryMBRequest,params)
            d,_ := ioutil.ReadAll(resp.Body)
            logger.GetLogger().Error(string(d))
            time.Sleep(time.Duration(waitMBTime) * time.Millisecond)
            return GetCover(artist,album,title)
        }
        // Check if release field are present, if not, limit
        d,_ := ioutil.ReadAll(resp.Body)
        if cover,err := extractInfo(d) ; err == nil {
            coverCache[key] = cover
            return cover
        } else{
            // case 3, only artist(title is already used as album
            if title == ""{
                return GetCover(artist, "", "")
            }
            countWords,_ := regexp.Compile("[a-zA-Z0-9]+")
            // Case 2 (single release)
            if len(countWords.FindAllString(album,-1)) ==1 {
                return GetCover(artist, title, "")
            }
            // No cover
            if album == "" && title == "" {
                coverCache[key] = ""
                return ""
            }
            // Case 1 bis : try to relaunch with first real word of album (more than 3 carac, separator are different to [a-Z0-9]
            sa,_ := regexp.Compile("[a-zA-Z0-9]{4,}")
            if values := sa.FindAllString(album,1) ; len(values) == 1 {
                return GetCover(artist,values[0],title)
            }
            // Case 3, only artist
            return GetCover(artist,"","")

        }
    }
    return ""
}
