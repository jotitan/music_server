package main

import (
	"os/exec"
	"bytes"
	"bufio"
	"strings"
	"net/http"
	"io/ioutil"
	"net/url"
	"time"
	"encoding/xml"
	"regexp"
	"music"
	"logger"
	"github.com/mjibson/id3"
	"os"
	"fmt"
	"path/filepath"
)



func main(){


	fmt.Println(filepath.Dir("c:/toto/blabla"))

	return

	fmt.Println(music.Intersect([]int{1,3,5,6},[]int{2,3,4,5,6,7}))
	fmt.Println(music.Intersect([]int{1,3,5,6},[]int{2,4,7,9}))
	fmt.Println(music.Intersect([]int{1,2,3,5,6,7},[]int{4,7,9}))
	fmt.Println(music.Intersect([]int{4,7,9},[]int{1,2,3,5,6,7}))

	ti := music.NewTextIndexer()
	ti.Add("aaa",1)
	ti.Add("aaa",22)
	ti.Add("aaa",222)
	ti.Add("cc",45)
	ti.Add("ccc",3)
	ti.Add("ccc",33)
	ti.Add("ccc",333)
	ti.Add("bbb",2)
	ti.Add("bbb",22)
	ti.Add("bbb",222)
	ti.Add("ddd",4)
	ti.Add("ggg",7)
	ti.Add("fff",6)
	ti.Add("lll",12)

	ti.Build()

	fmt.Println(ti.Search("bbb"))
	fmt.Println(ti.Search("d"))
	fmt.Println(ti.Search("g"))
	fmt.Println(ti.Search("t"))
	fmt.Println(ti.Search("l"))
	fmt.Println(ti.Search("a"))
	fmt.Println(ti.Search("c"))
	fmt.Println(ti.Search("de"))
	fmt.Println(ti.Search("aab"))
	fmt.Println(ti.Search("bbb a"))




	return


	pathMu := "D:\\test\\audio\\02-II. Dies Irae - Dies Irae.mp3"
	//pathMu := "D:\\test\\audio\\01-I. Requiem.mp3"
	ff,_ := os.Open(pathMu)
	defer ff.Close()
	info := id3.Read(ff)
	logger.GetLogger().Info(info)
	return

	album := "totoand vvc cd 1"
	if p:= strings.Index(album,"cd") ; p!=-1 {
		logger.GetLogger().Info(album[:p] + "d")
	}
	return

	uuu := "http://musicbrainz.org/ws/2/release/?query=artist%3A%22John+Barry%22+AND+release%3A%22you%20only%20live%20twice%22"
	if resp,e := http.Get(uuu); e == nil {
		// Check if release field are present, if not, limit
		d,_ := ioutil.ReadAll(resp.Body)
		var m music.RootResponse
		e := xml.Unmarshal(d,&m)

		logger.GetLogger().Info(e,m)
	}

	return

	searchStr := "Thomas Newman and helmut P.  (ok bob)"
	rrr,_ := regexp.Compile("[a-zA-Z0-9\\.]+")
	logger.GetLogger().Info(len(rrr.FindAllString(searchStr,-1)))
	return

	//artist := "Queen"
	artist := "Alicia%20Keys"
	album = "Unplugged"
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

