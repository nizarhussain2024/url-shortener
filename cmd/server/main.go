package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type URLMapping struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int       `json:"click_count"`
}

type URLStore struct {
	mu       sync.RWMutex
	urls     map[string]*URLMapping
	counter  int64
}

var store = &URLStore{
	urls:    make(map[string]*URLMapping),
	counter: 0,
}

func main() {
	http.HandleFunc("/api/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/api/stats/", statsHandler)
	http.HandleFunc("/health", healthHandler)

	fmt.Println("URL Shortener running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "url-shortener",
	})
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	store.counter++
	shortCode := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", store.counter)))[:8]
	
	mapping := &URLMapping{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CreatedAt:   time.Now(),
		ClickCount:  0,
	}
	store.urls[shortCode] = mapping
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"short_url":    fmt.Sprintf("http://localhost:8080/%s", shortCode),
		"original_url": req.URL,
		"short_code":   shortCode,
	})
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "URL Shortener API",
			"endpoints": "POST /api/shorten, GET /{shortCode}, GET /api/stats/{shortCode}",
		})
		return
	}

	shortCode := r.URL.Path[1:]
	
	store.mu.Lock()
	mapping, exists := store.urls[shortCode]
	if exists {
		mapping.ClickCount++
	}
	store.mu.Unlock()

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mapping.OriginalURL, http.StatusFound)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[len("/api/stats/"):]
	if shortCode == "" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	mapping, exists := store.urls[shortCode]
	store.mu.RUnlock()

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}
