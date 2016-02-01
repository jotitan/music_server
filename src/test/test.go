package main

import (
	"os/exec"
	"bytes"
	"bufio"
	"logger"
	"strings"
	"regexp"
)

func main(){
	//run()

	//args := arguments.ParseArgs()
	//dico := music.LoadDictionnary(args["workingFolder"])
	//music.IndexArtists(args["workingFolder"])
	//dico.Browse(args["browse"])

    //aa := music.LoadAllAlbums(args["workingFolder"])
	//logger.GetLogger().Info(aa)

	tot := "v1"
	//tot2 := "v1 & v2, v3/v4 /v5/v6&v7 , v8"
	reg,_ := regexp.Compile("&|,|/")
	vals := reg.Split(tot,-1)
	results := make([]string,len(vals))
	for i,v := range vals {
		results[i] = strings.Trim(v," ")
	}
	logger.GetLogger().Info(results)
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

