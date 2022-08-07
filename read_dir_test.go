package main

import (
	"os"
	"path/filepath"
	"testing"
)

var testDirectories = []string{
	"testdir/folder_underscore",
	"testdir/folder-dash",
	"testdir/folder space",
}

var testFiles = map[string]string{
	"empty_file":              "",
	"file_with_extension.txt": "Abababa",
	"file_with_rune":          `1000011110101000010011000111101011011011100001100011010110011111000011001100100100010011101000111001110100010101100111000001011100000010111111011010100011111000`,
}

func setupTestDirectories() {
	for _, d := range testDirectories {
		os.MkdirAll(d, os.ModePerm)
		for fname, fdata := range testFiles {
			dst := filepath.Join(d, fname)
			if err := os.WriteFile(dst, []byte(fdata), 0666); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func init() {
	setupTestDirectories()
}

func deleteTestDirectories() {
	if err := os.RemoveAll("testdir"); err != nil {
		log.Fatal(err)
	}
}

func TestGetDirectoryInfo(t *testing.T) {

	for _, testDir := range testDirectories {
		testDir = filepath.Clean(testDir)

		t.Run(testDir, func(t *testing.T) {
			cacheDir, err := GetDirectoryInfo(testDir)
			if err != nil {
				log.Error(err)
			}

			if cacheDir.path != testDir {
				t.Errorf("Expected %q, got %q", testDir, cacheDir.path)
			}

			log.Infof("%+v", cacheDir)
		})
	}

	deleteTestDirectories()
}
