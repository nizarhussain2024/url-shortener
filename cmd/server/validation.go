package main

import (
	"fmt"
	"net/url"
	"strings"
)

func validateURLFormat(urlStr string) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

func validateCustomCode(code string) error {
	if len(code) < 3 {
		return fmt.Errorf("Custom code must be at least 3 characters")
	}

	if len(code) > 20 {
		return fmt.Errorf("Custom code must be at most 20 characters")
	}

	// Only allow alphanumeric and hyphens
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("Custom code can only contain letters, numbers, and hyphens")
		}
	}

	return nil
}

func sanitizeURL(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	// Add http:// if no scheme
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}
	return urlStr
}

