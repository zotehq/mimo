package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthCheck struct {
	Status     string `json:"status"`
	ResponseMs int64  `json:"response_ms"`
}

func HealthStats(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	hc := HealthCheck{Status: "OK"}
	jsonData, err := json.Marshal(hc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	hc.ResponseMs = time.Since(start).Milliseconds()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
