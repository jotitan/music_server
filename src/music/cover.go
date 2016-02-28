package music
import (
    "errors"
    "net/http"
    "encoding/xml"
    "net/url"
    "time"
    "io/ioutil"
    "strings"
    "logger"
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
            if resp.StatusCode == 200 {
                return urlImage,nil
            }
        }
        if r.Id != "" {
            urlImage := "http://coverartarchive.org/release/" + r.Id + "/front-250"
            resp,_ := http.Get(urlImage)
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

func checkLastGet(){
    defer func(){
        lastGet = time.Now().Nanosecond()
    }()
    if lastGet == 0 {
        return
    }
    if rest := 110000000 - time.Now().Nanosecond() - lastGet ; rest > 0 {
        time.Sleep(time.Nanosecond * time.Duration(rest))
    }
}

// Get cover musicbrainz id and check if resource exist
func GetCover(artist,album string)string{
    key := artist + "-" + album
    if url,ok:= coverCache[key] ; ok {
        return url
    }
    params := "artist:\"" + artist + "\"";
    if album != "" {
        params+=" AND release:\"" + album + "\""
    }
    params = url.Values{"query":[]string{params}}.Encode()
    checkLastGet()
    if resp,e := http.Get("http://musicbrainz.org/ws/2/release/?" + params); e == nil {
        defer resp.Body.Close()
        // Quota exceed, relaunch after 1000ms
        if resp.StatusCode == 503 {
            logger.GetLogger().Error("Limit exceed, retry",params)
            time.Sleep(time.Second)
            return GetCover(artist,album)
        }
        // Check if release field are present, if not, limit
        d,_ := ioutil.ReadAll(resp.Body)
        if cover,err := extractInfo(d) ; err == nil {
            coverCache[key] = cover
            return cover
        } else{
            if album == "" {
                // No more solution
                coverCache[key] = ""
                return ""
            }
            // try to relaunch with first word of album
            if strings.Contains(album," ") {
                album = album[:strings.Index(album," ")]
                return GetCover(artist,album)
            }else{
                return GetCover(artist,"")
            }
        }
    }
    return ""
}
