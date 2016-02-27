package main

import (
	"os/exec"
	"bytes"
	"bufio"
	"logger"
	"strings"
	"net/http"
	"io/ioutil"
	"net/url"
	"time"
	"encoding/xml"
	"music"
)



func main(){

	urlMB := "http://musicbrainz.org/ws/2/release/?query=artist%3A%22Minnie+Riperton%22"
	if resp,e := http.Get(urlMB) ; e == nil {
		if resp.StatusCode == 200 {
			var rx music.RootResponse
			data,_ := ioutil.ReadAll(resp.Body)
			xml.Unmarshal(data,&rx)
			logger.GetLogger().Info(rx)
			logger.GetLogger().Info(rx.GetUrl())
		}
	}

	return

	//artist := "Queen"
	artist := "Alicia%20Keys"
	album := "Unplugged"
	params := "artist:\"" + artist + "\"";
	if album != "" {
		params+=" AND release:\"" + album + "\""
	}

	u := url.Values{"query":[]string{params}}

	params = u.Encode()

	url := "http://musicbrainz.org/ws/2/release/?" + params
	if resp,e := http.Get(url); e == nil {
		// Quota exceed, relaunch after 1000ms
		if resp.StatusCode == 503 {
			time.Sleep(time.Second)
		}
		// Check if release field are present, if not, limit
		d,_ := ioutil.ReadAll(resp.Body)
		var m map[string]string
		e := xml.Unmarshal(d,&m)

		logger.GetLogger().Info(e,m)
	}


	return
	//run()

	//args := arguments.ParseArgs()
	//dico := music.LoadDictionnary(args["workingFolder"])
	//music.IndexArtists(args["workingFolder"])
	//dico.Browse(args["browse"])

    //aa := music.LoadAllAlbums(args["workingFolder"])
	//logger.GetLogger().Info(aa)

	t := []int{2,4,6,8,10}

	pos := 4
	logger.GetLogger().Info(append(t[:pos],t[pos+1:]...))
	return

	path := "C:\\DataBE\\Bernardo1"

	data,_ := exec.Command("ls",path,"-i1").Output()

	r := bufio.NewReader(bytes.NewBuffer(data))
	for {
     if line, _, error := r.ReadLine() ; error == nil {
		 info := strings.Split(string(line)," ")
		 logger.GetLogger().Info(info[1],"=>",info[0])
	 }else{
		 break
	 }

	}
}

