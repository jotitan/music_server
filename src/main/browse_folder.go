package main
import (
    "music"
    "arguments"
)

func main(){
    args := arguments.ParseArgs()

    dico := music.LoadDictionnary(args["workingFolder"])
    dico.Browse(args["browse"])
}