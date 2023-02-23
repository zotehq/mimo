package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chai2010/webp"
)

func ServeWebPImage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(queryPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash := sha256.Sum256(imageData)
	cacheKey := hex.EncodeToString(hash[:])
	cachePath := filepath.Join(os.TempDir(), cacheKey+".webp")

	var cacheStatus string

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		img, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		webpData, err := webp.EncodeRGBA(img, 80)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(cachePath, webpData, 0644)
		if err != nil {
			log.Printf("Failed to write cache file: %s", err.Error())
		}

		cacheStatus = "MISS"
	} else {
		cacheStatus = "HIT"
	}

	webpData, err := os.ReadFile(cachePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Status", cacheStatus)
	w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
	w.Header().Set("Content-Type", "image/webp")
	w.Write(webpData)
}
