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
	cache map[string]io.Reader
	mu    sync.RWMutex
}

func (c *imageCache) Get(key string) (io.Reader, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	reader, ok := c.cache[key]
	return reader, ok
}

func (c *imageCache) Set(key string, reader io.Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = reader
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
		if reader, ok := cache.Get(cacheKey); ok {
			cacheStatus = "HIT"
			w.Header().Set("Cache-Status", cacheStatus)
			w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
			w.Header().Set("Content-Type", "image/webp")
			_, err := io.Copy(w, reader)
			if err != nil {
				log.Printf("Failed to write response: %s", err.Error())
			}
			return
		}
	}

	webpFile, err := os.Open(cachePath)
	if err == nil {
		cacheStatus = "HIT"
		if cache != nil {
			cache.Set(cacheKey, webpFile)
		}
		w.Header().Set("Cache-Status", cacheStatus)
		w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
		w.Header().Set("Content-Type", "image/webp")
		_, err := io.Copy(w, webpFile)
		if err != nil {
			log.Printf("Failed to write response: %s", err.Error())
		}
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

	webpFile, err = os.Create(cachePath)
	if err != nil {
		log.Printf("Failed to create cache file: %s", err.Error())
	} else {
		defer webpFile.Close()

		err = webp.Encode(webpFile, img, nil)
		if err != nil {
			log.Printf("Failed to encode image to WebP: %s", err.Error())
		}

		if cache != nil {
			_, err = webpFile.Seek(0, 0)
			if err != nil {
				log.Printf("Failed to seek cache file: %s", err.Error())
			} else {
				cache.Set(cacheKey, webpFile)
			}
		}
	}

	w.Header().Set("Cache-Status", cacheStatus)
	w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
	w.Header().Set("Content-Type", "image/webp")
	_, err = io.Copy(w, bytes.NewReader(imageData))
	if err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}
