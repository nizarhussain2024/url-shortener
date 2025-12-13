package main

import (
	"encoding/json"
	"net/http"
)

type BulkShortenRequest struct {
	URLs []struct {
		URL        string `json:"url"`
		CustomCode string `json:"custom_code,omitempty"`
	} `json:"urls"`
}

func bulkShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BulkShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := make([]map[string]interface{}, 0, len(req.URLs))

	for _, urlReq := range req.URLs {
		if !validateURL(urlReq.URL) {
			results = append(results, map[string]interface{}{
				"url":   urlReq.URL,
				"error": "Invalid URL format",
			})
			continue
		}

		store.mu.Lock()
		var shortCode string
		if urlReq.CustomCode != "" {
			if _, exists := store.urls[urlReq.CustomCode]; exists {
				store.mu.Unlock()
				results = append(results, map[string]interface{}{
					"url":   urlReq.URL,
					"error": "Custom code already exists",
				})
				continue
			}
			shortCode = urlReq.CustomCode
		} else {
			store.counter++
			shortCode = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", store.counter)))[:8]
		}

		mapping := &URLMapping{
			ShortCode:   shortCode,
			OriginalURL: urlReq.URL,
			CreatedAt:   time.Now(),
			ClickCount:  0,
			CustomCode:  urlReq.CustomCode,
		}
		store.urls[shortCode] = mapping
		analytics.RecordURL()
		store.mu.Unlock()

		results = append(results, map[string]interface{}{
			"url":        urlReq.URL,
			"short_code": shortCode,
			"short_url":  fmt.Sprintf("http://localhost:8080/%s", shortCode),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"total":   len(results),
	})
}


