package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func qrCodeHandler(w http.ResponseWriter, r *http.Request) {
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

	// In production, generate actual QR code using a library like github.com/skip2/go-qrcode
	// For now, return QR code data URL or API endpoint
	qrData := map[string]interface{}{
		"short_code": shortCode,
		"short_url":  fmt.Sprintf("http://localhost:8080/%s", shortCode),
		"qr_code_url": fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", 
			fmt.Sprintf("http://localhost:8080/%s", shortCode)),
		"original_url": mapping.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(qrData)
}



