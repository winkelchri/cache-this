package main

import (
	"bufio"
	"io"
	"io/fs"
	"math"
	"time"

	"os"
	"sort"

	"github.com/karrick/godirwalk"
	"github.com/sirupsen/logrus"
	"github.com/winkelchri/cache-this/logging"
)

var log *logrus.Logger
var bufSize = 4 * 1024

func init() {
	log = logging.NewLogger("main")
	log.SetLevel(logrus.DebugLevel)
}

type File struct {
	path string
	info fs.FileInfo
}

// var files []fs.FileInfo
var files []File

func handleWalk(osPathname string, de *godirwalk.Dirent) error {
	stats, err := os.Stat(osPathname)

	if err != nil {
		log.Println(err)
		return godirwalk.SkipThis
	}

	files = append(files, File{path: osPathname, info: stats})
	return nil
}

func SortFileSizeAscend(files []File) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].info.Size() < files[j].info.Size()
	})
}

func SortFileSizeDescend(files []File) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].info.Size() > files[j].info.Size()
	})
}

func SizeInMB(size int64) float64 {
	return float64(size) / math.Pow(1024, 2)
}

func readFile(file File) error {
	log.Infof("Reading file '%s': %.2f MB", file.path, SizeInMB(file.info.Size()))

	f, err := os.Open(file.path)

	if err != nil {
		return err
	}

	defer f.Close()
	r := bufio.NewReader(f)

	nr := int64(0)
	// buf := make([]byte, 0, 4*1024)
	buf := make([]byte, 0, bufSize)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]

		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return err
		}

		// Do something with buf
		// fmt.Print(".")
		// fmt.Printf("%v", buf)
		nr += int64(len(buf))

		if err != nil && err != io.EOF {
			return err
		}
	}
	// fmt.Println()

	return nil
}

func main() {
	// TODO
	// 1. 	Setup read which is using a smaller buffer to not
	// 		fill up all memory. Inspiration: https://stackoverflow.com/questions/54028660/how-to-read-huge-files-with-samll-ram-in-golang

	// 2.	Create small GUI application using bubbletea to show
	// 		1.	Read-speed of the current file
	// 		2.	Overall speed and elapsed time
	// 		3.	Option to stop the process

	// 3. 	Option to only read the biggest files?
	d := os.Getenv("CACHE_THIS_DIR")
	if d == "" {
		d = "."
	}

	log.Infof("Getting files of directory: '%s'", d)
	err := godirwalk.Walk(d, &godirwalk.Options{
		Callback: handleWalk,
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Files total: '%d'", len(files))

	SortFileSizeDescend(files)
	var totalSize int64 = 0
	for _, file := range files {
		totalSize += file.info.Size()
	}
	// fs.FileInfo.Size() returns file size in bytes
	log.Infof("Size total: %.2f MB", float64(totalSize)/math.Pow(1024, 2))

	for _, file := range files {
		fileSize := float64(file.info.Size()) / math.Pow(1024, 2)
		start := time.Now()
		readFile(file)
		dur := time.Since(start)
		speed := fileSize / dur.Seconds()
		log.Infof("File size: %.2f MB; read speed: %.2f MB/s", fileSize, speed)
	}
}
