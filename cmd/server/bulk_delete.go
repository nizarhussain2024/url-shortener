package main

import (
	"encoding/json"
	"net/http"
)

type BulkDeleteRequest struct {
	Codes []string `json:"codes"`
}

func bulkDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	deleted := []string{}
	notFound := []string{}

	for _, code := range req.Codes {
		if _, exists := store.urls[code]; exists {
			delete(store.urls, code)
			deleted = append(deleted, code)
		} else {
			notFound = append(notFound, code)
		}
	}
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted":   deleted,
		"not_found": notFound,
		"total":     len(req.Codes),
	})
}


