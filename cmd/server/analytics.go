package main

import (
	"sync"
	"time"
)

type Analytics struct {
	mu           sync.RWMutex
	TotalURLs    int64
	TotalClicks  int64
	TopURLs      map[string]int64
	DailyClicks  map[string]int64
}

var analytics = &Analytics{
	TopURLs:     make(map[string]int64),
	DailyClicks: make(map[string]int64),
}

func (a *Analytics) RecordClick(shortCode string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.TotalClicks++
	a.TopURLs[shortCode]++
	today := time.Now().Format("2006-01-02")
	a.DailyClicks[today]++
}

func (a *Analytics) RecordURL() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.TotalURLs++
}

func (a *Analytics) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return map[string]interface{}{
		"total_urls":   a.TotalURLs,
		"total_clicks": a.TotalClicks,
		"top_urls":     a.TopURLs,
		"daily_clicks": a.DailyClicks,
	}
}
