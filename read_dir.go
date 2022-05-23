package main

import (
	"bufio"
	"io"
	"io/fs"
	"math"
	"os"
	"sort"

	"github.com/karrick/godirwalk"
)

var bufSize = 4 * 1024

type CacheDir struct {
	path     string
	numFiles int64
	sizeDir  int64
	files    []File
}

type File struct {
	path string
	info fs.FileInfo
}

// SortFileSizeAscend sorts the files by size ascending
func (c *CacheDir) SortFileSizeAscend() {
	files := c.files
	sort.Slice(files, func(i, j int) bool {
		return files[i].info.Size() < files[j].info.Size()
	})
}

// SortFileSizeAscend sorts the files by size descending
func (c *CacheDir) SortFileSizeDescend() {
	files := c.files
	sort.Slice(files, func(i, j int) bool {
		return files[i].info.Size() > files[j].info.Size()
	})
}

// SizeInMB converts the given size int64 into a megabyte value. Assuming byte value.
func SizeInMB(size int64) float64 {
	return float64(size) / math.Pow(1024, 2)
}

// Read reads the given file considering the buffer size
func (file File) Read() error {
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

// GetDirectoryInfo walks the given directory and returns a CacheDir object
func GetDirectoryInfo(path string) (CacheDir, error) {
	var c = CacheDir{path: path}

	log.Infof("Getting files of directory: '%s'", path)
	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			stats, err := os.Stat(osPathname)

			if err != nil {
				log.Errorln(err)
				return godirwalk.SkipThis
			}

			c.files = append(c.files, File{path: osPathname, info: stats})
			if !stats.Mode().IsDir() {
				c.numFiles++
			}
			c.sizeDir += stats.Size()

			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

	if err != nil {
		return CacheDir{}, err
	}

	c.SortFileSizeDescend()
	return c, nil
}
