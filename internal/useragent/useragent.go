package useragent

import "strings"

// Parse extracts browser and OS names from a User-Agent string.
// Detection order: more specific checks first (iOS before macOS,
// since iPhone/iPad UAs also contain "Mac OS X").
func Parse(ua string) (browser, os string) {
	browser = "Other"
	os = "Other"

	// OS detection — specific before generic
	switch {
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		os = "iOS"
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		os = "macOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "CrOS"):
		os = "ChromeOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	}

	// Browser detection (order matters)
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "curl/"):
		browser = "curl"
	case strings.Contains(lower, "bot") || strings.Contains(lower, "spider") || strings.Contains(lower, "crawl"):
		browser = "Bot"
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera"):
		browser = "Opera"
	case strings.Contains(ua, "Vivaldi/"):
		browser = "Vivaldi"
	case strings.Contains(ua, "Brave"):
		browser = "Brave"
	case strings.Contains(ua, "Chrome/") && strings.Contains(ua, "Safari/"):
		browser = "Chrome"
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome/"):
		browser = "Safari"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	}

	return browser, os
}
