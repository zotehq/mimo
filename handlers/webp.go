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
	"github.com/gin-gonic/gin"
)

type imageCache struct {
	cache      map[string]*bytes.Reader
	order      []string
	maxEntries int
	mu         sync.Mutex
}

func (c *imageCache) Get(key string) (*bytes.Reader, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	reader, ok := c.cache[key]
	if ok {
		for i, k := range c.order {
			if k == key {
				c.order = append(c.order[:i], c.order[i+1:]...)
				c.order = append(c.order, key)
				break
			}
		}
	}
	return reader, ok
}

func (c *imageCache) Set(key string, reader *bytes.Reader) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.maxEntries {
		lruKey := c.order[0]
		delete(c.cache, lruKey)
		c.order = c.order[1:]
	}

	c.cache[key] = reader
	c.order = append(c.order, key)
}

var cache = &imageCache{
	cache:      make(map[string]*bytes.Reader),
	maxEntries: 1000,
}

const cacheDir = "/tmp"

type ResettableBuffer struct {
	*bytes.Buffer
}

func (b *ResettableBuffer) Reset() {
	b.Buffer.Reset()
}

func ServeWebP(c *gin.Context) {
	start := time.Now()

	queryURL := c.Request.URL.Query().Get("url")
	if queryURL == "" {
		c.String(http.StatusBadRequest, "Missing 'url' query parameter")
		return
	}

	hash := sha256.Sum256([]byte(queryURL))
	cacheKey := hex.EncodeToString(hash[:])

	if reader, ok := cache.Get(cacheKey); ok {
		c.Header("Cache-Status", "HIT")
		c.Header("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
		c.Header("Content-Type", "image/webp")

		reader.Seek(0, io.SeekStart)
		_, err := io.Copy(c.Writer, reader)
		if err != nil {
			log.Printf("Failed to write response: %s", err.Error())
		}
		return
	}

	diskCachePath := filepath.Join(cacheDir, cacheKey+".webp")
	webpFile, err := os.Open(diskCachePath)
	if err == nil {
		defer webpFile.Close()

		webpData, err := io.ReadAll(webpFile)
		if err != nil {
			log.Printf("Failed to read disk cache file: %s", err.Error())
		} else {
			webpReader := bytes.NewReader(webpData)
			cache.Set(cacheKey, webpReader)

			c.Header("Cache-Status", "HIT")
			c.Header("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
			c.Header("Content-Type", "image/webp")
			_, err = io.Copy(c.Writer, webpReader)
			if err != nil {
				log.Printf("Failed to write response: %s", err.Error())
			}
			return
		}
	}

	resp, err := http.Get(queryURL)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	webpBuffer := &ResettableBuffer{bytes.NewBuffer(nil)}
	err = webp.Encode(webpBuffer, img, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to encode image to WebP")
		return
	}

	webpReaderForCache := bytes.NewReader(webpBuffer.Bytes())
	cache.Set(cacheKey, webpReaderForCache)

	diskCacheFile, err := os.Create(diskCachePath)
	if err != nil {
		log.Printf("Failed to create disk cache file: %s", err.Error())
	} else {
		defer diskCacheFile.Close()

		_, err = io.Copy(diskCacheFile, webpReaderForCache)
		if err != nil {
			log.Printf("Failed to write to disk cache file: %s", err.Error())
		}
	}

	c.Header("Cache-Status", "MISS")
	c.Header("Content-Type", "image/webp")

	responseTime := time.Since(start).Milliseconds()
	c.Header("X-Response-Time", fmt.Sprintf("%d ms", responseTime))

	combinedReader := io.MultiReader(webpReaderForCache, &ResettableBuffer{bytes.NewBuffer(webpBuffer.Bytes())})
	_, err = io.Copy(c.Writer, combinedReader)
	if err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}
