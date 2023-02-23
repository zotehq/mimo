package handlers

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
)

func ServeWebPImage(w http.ResponseWriter, r *http.Request) {
	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	start := time.Now()

	cacheKey := strings.ReplaceAll(queryPath, "/", "_")
	cachePath := filepath.Join(os.TempDir(), cacheKey+".webp")

	if _, err := os.Stat(cachePath); err == nil {
		webpData, err := os.ReadFile(cachePath)
		if err != nil {
			log.Printf("Failed to read cache file: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Status", "HIT")
		w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
		w.Header().Set("Content-Type", "image/webp")
		w.Write(webpData)
		return
	}

	resp, err := http.Get(queryPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch remote image", resp.StatusCode)
		return
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Cache-Status", "MISS")
	w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
	w.Header().Set("Content-Type", "image/webp")
	w.Write(webpData)
}
