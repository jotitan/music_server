package main

import (
	"github.com/jotitan/music_server/logger"
	"github.com/jotitan/music_server/music"
	"os"
)

func main() {
	inputFolder := os.Args[1]
	outputFolder := os.Args[2]
	logger.GetLogger().Info("Convert from " + inputFolder + " to " + outputFolder)
	music.CreateNewDictionnary(inputFolder, outputFolder)

}
