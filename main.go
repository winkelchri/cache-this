package main

import (
	"github.com/sirupsen/logrus"
	"github.com/winkelchri/cache-this/logging"
)

var log *logrus.Logger

func init() {
	log = logging.NewLogger("main")
	log.SetLevel(logrus.DebugLevel)
}

// func main() {
// 	// TODO
// 	// 1. 	Setup read which is using a smaller buffer to not
// 	// 		fill up all memory. Inspiration: https://stackoverflow.com/questions/54028660/how-to-read-huge-files-with-samll-ram-in-golang

// 	// 2.	Create small GUI application using bubbletea to show
// 	// 		1.	Read-speed of the current file
// 	// 		2.	Overall speed and elapsed time
// 	// 		3.	Option to stop the process

// 	// 3. 	Option to only read the biggest files?
// 	c, err := GetDirectoryInfo("")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Infof("Files total: %d", c.numFiles)

// 	// fs.FileInfo.Size() returns file size in bytes
// 	log.Infof("Size total: %.2f MB", float64(c.sizeDir)/math.Pow(1024, 2))

// 	for _, file := range c.files {
// 		fileSize := float64(file.info.Size()) / math.Pow(1024, 2)
// 		start := time.Now()
// 		readFile(file)
// 		dur := time.Since(start)
// 		speed := fileSize / dur.Seconds()
// 		log.Infof("File size: %.2f MB; read speed: %.2f MB/s", fileSize, speed)
// 	}
// }

func main() {
	StartUI()
}
