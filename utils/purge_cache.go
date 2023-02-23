package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

func PurgeCache() {
	cacheDir := os.TempDir()
	expiredTime := time.Now().Add(-time.Minute * 30)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("Failed to read cache directory: %s", err.Error())
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			log.Printf("Failed to get file info for %s: %s", file.Name(), err.Error())
			continue
		}
		if info.ModTime().Before(expiredTime) {
			err := os.Remove(filepath.Join(cacheDir, file.Name()))
			if err != nil {
				log.Printf("Failed to delete file %s: %s", file.Name(), err.Error())
			} else {
				log.Printf("Deleted expired cache file %s", file.Name())
			}
		}
	}
}
