package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"
)

type HealthCheck struct {
	Status     string `json:"status"`
	ResponseMs int64  `json:"response_ms"`
}

func main() {
	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/proxy/image", imageHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	start := time.Now()

	hc := HealthCheck{Status: "OK"}

	jsonData, err := json.Marshal(hc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	hc.ResponseMs = time.Since(start).Milliseconds()
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	start := time.Now()

	remoteURL, err := url.Parse(queryPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheKey := strings.Replace(remoteURL.String(), "/", "_", -1)
	cachePath := filepath.Join(os.TempDir(), cacheKey+".webp")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		resp, err := http.Get(remoteURL.String())
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

		webpData, err := webp.EncodeRGBA(img, 80)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(cachePath, webpData, 0644)
		if err != nil {
			log.Printf("Failed to write cache file: %s", err.Error())
		}
	}

	webpData, err := os.ReadFile(cachePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cacheStatus := "MISS"

	if err == nil {
		cacheStatus = "HIT"
	}

	w.Header().Set("Cache-Status", cacheStatus)
	w.Header().Set("Response-Time", fmt.Sprint(time.Since(start).Milliseconds()))
	w.Header().Set("Content-Type", "image/webp")
	w.Write(webpData)
}
