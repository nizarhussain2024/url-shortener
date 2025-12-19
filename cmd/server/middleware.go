package main

import (
	"log"
	"net/http"
	"time"
)

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next(w, r)
		log.Printf("Completed in %v", time.Since(start))
	}
}

func validateURL(url string) bool {
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}




