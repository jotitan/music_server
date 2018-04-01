package main

import (
	"logger"
	"music"
	"os"
)

func main() {
	inputFolder := os.Args[1]
	outputFolder := os.Args[2]
	logger.GetLogger().Info("Convert from " + inputFolder + " to " + outputFolder)
	music.CreateNewDictionnary(inputFolder, outputFolder)

}
