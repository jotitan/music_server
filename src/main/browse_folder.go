package main
import (
    "music"
    "arguments"
    "fmt"
)

func main(){
    args := arguments.ParseArgs()

    switch args["task"]{
        case "browse":
        dico := music.LoadDictionnary(args["workingFolder"])
        dico.Browse(args["browse"])
        case "index":
        music.IndexArtists(args["workingFolder"])
        default:
        fmt.Println("No task define. Available tasks (-task) : browse (-workingFolder and -browse required), index (-workingFolder required)")
    }


}