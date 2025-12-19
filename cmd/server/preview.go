package main

import (
	"encoding/json"
	"net/http"
)

type LinkPreview struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
}

func linkPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Query().Get("code")
	if shortCode == "" {
		http.Error(w, "Code parameter is required", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	mapping, exists := store.urls[shortCode]
	store.mu.RUnlock()

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	// In production, fetch actual metadata from the URL
	preview := LinkPreview{
		Title:       "Link Preview",
		Description: "Preview for " + mapping.OriginalURL,
		URL:         mapping.OriginalURL,
		Image:       "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}




