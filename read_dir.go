package main

import (
	"bufio"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
)

var bufSize = 16 * 1024

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

	// FIXME: 	Currently issue with os specific paths
	// 			Reading seems to be broken due to filepath object?

	// Convert path to os specific path notation
	path = filepath.Clean(path)

	var c = CacheDir{path: path}

	log.Infof("Getting files of directory: '%s'", path)

	err := filepath.Walk(path, func(osPathname string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Errorln(err)
			// TODO: Possible cause of error during runtime. Investigate how to skip files which cannot be readed.
		}

		c.files = append(c.files, File{path: osPathname, info: info})
		if !info.IsDir() {
			c.numFiles++
		}
		c.sizeDir += info.Size()

		return nil
	})

	// Windows returns EOF in case of empty folders
	// if err != nil && err.Error() != "EOF" {
	if err != nil {
		return CacheDir{}, err
	}

	c.SortFileSizeDescend()
	return c, nil
}
