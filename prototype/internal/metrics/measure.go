package metrics

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type DirInfo struct {
	totalSize Byte
	maxSize   Byte
	maxPath   string
	minSize   Byte
	minPath   string
}

func (d DirInfo) Size() Byte {
	return d.totalSize
}

func (d DirInfo) Max() Byte {
	return d.maxSize
}

func MeasureDir(dirPath string) (DirInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return DirInfo{}, fmt.Errorf("error reading directory: %w", err)
	}

	if len(entries) == 0 {
		return DirInfo{}, fmt.Errorf("no entries on %s to measure\n", dirPath)
	}

	firstInfo, err := entries[0].Info()
	if err != nil {
		return DirInfo{}, fmt.Errorf("error reading info for entry '%s'", entries[0].Name())
	}

	var dirInfo = DirInfo{}
	dirInfo.minSize = Byte(firstInfo.Size())
	dirInfo.minPath = filepath.Join(dirPath, entries[0].Name())
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Fatalf("error reading info for entry '%s'", entry.Name())
		}
		var currSize int64
		var currMax int64
		var currMaxPath string
		var currMin int64
		var currMinPath string
		entryPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			entryInfo, err := MeasureDir(entryPath)
			if err != nil {
				return DirInfo{}, fmt.Errorf("error measuring sub-directory '%s': %w", entryPath, err)
			}
			currSize = int64(entryInfo.totalSize)
			currMax = int64(entryInfo.maxSize)
			currMaxPath = entryInfo.maxPath
			currMin = int64(entryInfo.minSize)
			currMinPath = entryInfo.minPath
		} else {
			currSize = info.Size()
			currMax = info.Size()
			currMaxPath = entryPath
			currMin = info.Size()
			currMinPath = entryPath
		}

		dirInfo.totalSize += Byte(currSize)

		if currMax > int64(dirInfo.maxSize) {
			dirInfo.maxSize = Byte(currMax)
			dirInfo.maxPath = currMaxPath
		}
		if currMin < int64(dirInfo.minSize) {
			dirInfo.minSize = Byte(currMin)
			dirInfo.minPath = currMinPath
		}
	}

	return dirInfo, nil

}
