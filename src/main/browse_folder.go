package main
import (
    "music"
    "arguments"
)

func main(){
    args := arguments.ParseArgs()

    switch args["task"]{
        case "browse":
        dico := music.LoadDictionnary(args["workingFolder"])
        dico.Browse(args["browse"])
        case "index":
        music.IndexArtists(args["workingFolder"])
    }


}