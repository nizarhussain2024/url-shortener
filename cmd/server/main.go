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
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ClickCount  int       `json:"click_count"`
	CustomCode  string    `json:"custom_code,omitempty"`
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
	http.HandleFunc("/api/shorten", loggingMiddleware(rateLimitMiddleware(shortenHandler)))
	http.HandleFunc("/api/shorten/bulk", loggingMiddleware(bulkShortenHandler))
	http.HandleFunc("/api/analytics", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics.GetStats())
	}))
	http.HandleFunc("/api/qrcode", loggingMiddleware(qrCodeHandler))
	http.HandleFunc("/api/preview", loggingMiddleware(linkPreviewHandler))
	http.HandleFunc("/api/bulk-delete", loggingMiddleware(bulkDeleteHandler))
	http.HandleFunc("/", loggingMiddleware(redirectHandler))
	http.HandleFunc("/api/stats/", loggingMiddleware(statsHandler))
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
		URL        string `json:"url"`
		CustomCode string `json:"custom_code,omitempty"`
		ExpiresIn  int    `json:"expires_in_days,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if !validateURL(req.URL) {
		http.Error(w, "Invalid URL format. Must start with http:// or https://", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	var shortCode string
	if req.CustomCode != "" {
		if _, exists := store.urls[req.CustomCode]; exists {
			store.mu.Unlock()
			http.Error(w, "Custom code already exists", http.StatusConflict)
			return
		}
		shortCode = req.CustomCode
	} else {
		store.counter++
		shortCode = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", store.counter)))[:8]
	}
	
	mapping := &URLMapping{
		ShortCode:   shortCode,
		OriginalURL: req.URL,
		CreatedAt:   time.Now(),
		ClickCount:  0,
		CustomCode:  req.CustomCode,
	}
	
	if req.ExpiresIn > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.ExpiresIn)
		mapping.ExpiresAt = &expiresAt
	}
	
	store.urls[shortCode] = mapping
	store.mu.Unlock()
	
	analytics.RecordCreation()

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
		if mapping.ExpiresAt != nil && time.Now().After(*mapping.ExpiresAt) {
			delete(store.urls, shortCode)
			store.mu.Unlock()
			http.Error(w, "Short URL has expired", http.StatusGone)
			return
		}
		mapping.ClickCount++
		analytics.RecordClick(shortCode)
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
