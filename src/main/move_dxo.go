package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main(){
	folder := os.Args[1]
	outputFolder := os.Args[3]
	pattern := strings.ToLower(os.Args[2])

	fmt.Println("Work on",folder)
	if root,err := os.Open(folder) ; err == nil {
		defer root.Close()
		// List all folders
		folders,_ := root.Readdir(-1)
		for _,dir := range folders{
			folderName := filepath.Join(root.Name(),dir.Name())
			// Test hd
			// Test if pattern file exists (like DxO)
			if folder,err := os.Open(folderName) ; err == nil {
				if files,err := folder.Readdirnames(-1) ; err == nil {
					patternFiles := make([]string,0,len(files))
					for _,f := range files{
						if strings.Contains(strings.ToLower(f),pattern) {
							patternFiles = append(patternFiles,f)
						}
					}
					if len(patternFiles) > 0 {
						fmt.Println("Create HD folder and move inside",folderName,"=>",len(patternFiles))
						dirName := dir.Name()
						if dirName[0] == '_' {
							dirName = dirName[1:]
						}
						moveFiles(patternFiles,folderName,filepath.Join(outputFolder,dirName,"hd"))
					}
				}
			}
		}
	}
}

func moveFiles(from []string,rootFrom,to string){
	if os.MkdirAll(to,os.ModePerm) == nil {
		success := 0
		// Move files
		for _,toMove := range from {
			fmt.Println("Move",filepath.Join(rootFrom,toMove),"=>",filepath.Join(to,toMove))
			if move(filepath.Join(rootFrom,toMove),filepath.Join(to,toMove)) {
				success++
			}
		}
		fmt.Println("Move",success,"/",len(from))
	}
}

func move(fromName, toName string)bool{
	if from,eF := os.Open(fromName) ; eF== nil {
		defer func(){
			from.Close()
			os.Remove(fromName)
		}()
		if to,eT := os.OpenFile(toName, os.O_TRUNC|os.O_RDWR|os.O_CREATE,os.ModePerm) ; eT == nil {
			defer to.Close()
			if size,err := io.Copy(to,from) ; size > 0 && err == nil {
				return true
			}else{
				fmt.Println("TO",err)
			}
		}else{
			fmt.Println("TO",eT)
		}
	}else{
		fmt.Println("FROM",eF)
	}
	return false
}
