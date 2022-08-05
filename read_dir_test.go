package main

import (
	"os"
	"testing"
)

var testDirectories = []string{
	"testdir/folder_underscore",
	"testdir/folder-dash",
	"testdir/folder space",
}

func setupTestDirectories() {
	for _, d := range testDirectories {
		os.MkdirAll(d, os.ModePerm)
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
