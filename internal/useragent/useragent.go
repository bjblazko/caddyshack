// Package useragent derives client type, browser and OS names from
// User-Agent strings.
package useragent

import (
	"cmp"
	"strings"
)

// Kind is the type of client that sent a request.
type Kind string

// Client kinds, see Classify.
const (
	KindBrowser Kind = "Browser"
	KindMonitor Kind = "Monitor" // uptime and health checks
	KindBot     Kind = "Bot"     // crawlers, link previews, feed readers
	KindScript  Kind = "Script"  // command-line tools, HTTP libraries, scanners
)

// rule maps any of its tokens found in a User-Agent to a name.
type rule struct {
	name   string
	tokens []string
}

// kindRules are checked in order; the first match wins. Monitors come first
// because several identify as "bot" (UptimeRobot, Pingdom).
var kindRules = []rule{
	{string(KindMonitor), []string{
		"uptime-kuma", "uptimerobot", "pingdom", "statuscake", "better uptime",
		"betteruptime", "betterstack", "better stack", "healthchecks", "site24x7",
		"updown.io", "hetrixtools", "freshping", "gatus", "blackbox-exporter",
		"nodeping", "checkly", "upptime", "uptime.com", "cron-job.org",
	}},
	{string(KindBot), []string{
		"bot", "spider", "crawl", "slurp", "facebookexternalhit", "meta-externalagent",
		"whatsapp", "mastodon", "pleroma", "akkoma", "misskey", "feed", "rss",
		"newsblur", "inoreader", "netnewswire", "headlesschrome", "lighthouse",
		"preview", "ia_archiver", "semrush", "ahrefs", "bytespider", "chatgpt-user",
		"claude-web", "anthropic-ai", "perplexity", "ccbot", "yandex", "qwant",
		"googleother", "google-extended", "dataforseo", "embedly", "iframely",
	}},
	{string(KindScript), []string{
		"curl/", "wget/", "python-requests", "python-urllib", "python-httpx",
		"aiohttp", "go-http-client", "java/", "okhttp", "apache-httpclient",
		"libwww-perl", "guzzlehttp", "node-fetch", "axios/", "undici", "ruby",
		"powershell", "httpie", "postmanruntime", "scrapy", "reqwest", "dart:io",
		"zgrab", "nuclei", "masscan", "nmap", "sqlmap", "nikto", "wpscan",
		"gobuster", "dirbuster", "ffuf", "censys", "expanse",
	}},
}

// osRules are checked in order, case-sensitively: iOS before macOS (iPhone
// UAs contain "Mac OS X"), Android before Linux.
var osRules = []rule{
	{"iOS", []string{"iPhone", "iPad"}},
	{"Windows", []string{"Windows"}},
	{"macOS", []string{"Macintosh", "Mac OS X"}},
	{"Android", []string{"Android"}},
	{"ChromeOS", []string{"CrOS"}},
	{"Linux", []string{"Linux"}},
}

// Classify returns the kind of client that sent a request with this
// User-Agent. An empty User-Agent is a script; anything unrecognised is a
// browser.
func Classify(ua string) Kind {
	if strings.TrimSpace(ua) == "" {
		return KindScript
	}
	return Kind(match(kindRules, strings.ToLower(ua), string(KindBrowser)))
}

// Parse extracts the client kind and the browser and OS names from a
// User-Agent string. For clients that are not browsers the browser name is
// their kind (Monitor, Bot, Script).
func Parse(ua string) (kind Kind, browser, os string) {
	os = match(osRules, ua, "Other")
	if kind = Classify(ua); kind != KindBrowser {
		return kind, string(kind), os
	}
	return kind, browserName(ua), os
}

// FromNames derives the client kind from browser and OS names that an
// anonymizer kept in place of the User-Agent. Non-browser names are mapped to
// their kind ("curl" → Script), as Parse would name them.
func FromNames(browser, os string) (kind Kind, browserName, osName string) {
	switch strings.ToLower(browser) {
	case "monitor":
		kind = KindMonitor
	case "bot":
		kind = KindBot
	case "script", "curl", "wget":
		kind = KindScript
	default:
		kind = KindBrowser
	}
	if kind != KindBrowser {
		browser = string(kind)
	}
	return kind, browser, cmp.Or(os, "Other")
}

// browserName checks browsers built on Chrome before Chrome, and Chrome
// before Safari (Chrome UAs contain "Safari/").
func browserName(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/"):
		return "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera"):
		return "Opera"
	case strings.Contains(ua, "Vivaldi/"):
		return "Vivaldi"
	case strings.Contains(ua, "Brave"):
		return "Brave"
	case strings.Contains(ua, "Chrome/") && strings.Contains(ua, "Safari/"):
		return "Chrome"
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome/") && strings.Contains(ua, "Version/"):
		return "Safari"
	case strings.Contains(ua, "Firefox/"):
		return "Firefox"
	}
	return "Other"
}

// match returns the name of the first rule with a token contained in s.
func match(rules []rule, s, fallback string) string {
	for _, r := range rules {
		for _, token := range r.tokens {
			if strings.Contains(s, token) {
				return r.name
			}
		}
	}
	return fallback
}
