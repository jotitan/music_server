package main
import (
	"music"
	"arguments"
	"logger"
)

func main(){
	//run()

	args := arguments.ParseArgs()
	//dico := music.LoadDictionnary(args["workingFolder"])
	music.IndexArtists(args["workingFolder"])
	//dico.Browse(args["browse"])

    aa := music.LoadAllAlbums(args["workingFolder"])
	logger.GetLogger().Info(aa)


}

