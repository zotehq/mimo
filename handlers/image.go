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
	"sync"
	"time"

	"github.com/chai2010/webp"
)

type imageCache struct {
	cache map[string][]byte
	mu    sync.RWMutex
}

func (c *imageCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.cache[key]
	return data, ok
}

func (c *imageCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = data
}

func ServeWebPImage(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(queryPath))
	cacheKey := hex.EncodeToString(hash[:])
	cachePath := filepath.Join(os.TempDir(), cacheKey+".webp")

	cacheStatus := "MISS"
	cache, ok := r.Context().Value("cache").(*imageCache)
	if ok {
		if data, ok := cache.Get(cacheKey); ok {
			cacheStatus = "HIT"
			w.Header().Set("Cache-Status", cacheStatus)
			w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
			w.Header().Set("Content-Type", "image/webp")
			w.Write(data)
			return
		}
	}

	webpData, err := os.ReadFile(cachePath)
	if err == nil {
		cacheStatus = "HIT"
		if cache != nil {
			cache.Set(cacheKey, webpData)
		}
		w.Header().Set("Cache-Status", cacheStatus)
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

	webpData, err = webp.EncodeRGBA(img, 80)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile(cachePath, webpData, 0644)
	if err != nil {
		log.Printf("Failed to write cache file: %s", err.Error())
	}

	if cache != nil {
		cache.Set(cacheKey, webpData)
	}

	w.Header().Set("Cache-Status", cacheStatus)
	w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
	w.Header().Set("Content-Type", "image/webp")
	w.Write(webpData)
}
