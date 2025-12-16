package main

import (
	"sync"
	"time"
)

type URLExpiration struct {
	Code      string
	ExpiresAt time.Time
}

var expirationStore = struct {
	mu         sync.RWMutex
	expirations map[string]time.Time
}{
	expirations: make(map[string]time.Time),
}

func setExpiration(code string, duration time.Duration) {
	expirationStore.mu.Lock()
	defer expirationStore.mu.Unlock()
	expirationStore.expirations[code] = time.Now().Add(duration)
}

func isExpired(code string) bool {
	expirationStore.mu.RLock()
	defer expirationStore.mu.RUnlock()

	expiresAt, exists := expirationStore.expirations[code]
	if !exists {
		return false
	}
	return time.Now().After(expiresAt)
}

func cleanupExpiredURLs() {
	expirationStore.mu.Lock()
	defer expirationStore.mu.Unlock()

	now := time.Now()
	for code, expiresAt := range expirationStore.expirations {
		if now.After(expiresAt) {
			delete(expirationStore.expirations, code)
			store.mu.Lock()
			delete(store.urls, code)
			store.mu.Unlock()
		}
	}
}

func startCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cleanupExpiredURLs()
		}
	}()
}

