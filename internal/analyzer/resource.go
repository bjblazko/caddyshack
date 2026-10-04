package analyzer

import (
	"strings"
)

var assetPrefixes = []string{"/css/", "/js/", "/img/", "/fonts/", "/api"}
var assetExtensions = []string{".css", ".js", ".png", ".jpg", ".svg", ".ttf", ".woff", ".woff2", ".ico"}

func isAsset(uri string) bool {
	for _, p := range assetPrefixes {
		if strings.HasPrefix(uri, p) {
			return true
		}
	}
	for _, e := range assetExtensions {
		if strings.HasSuffix(uri, e) {
			return true
		}
	}
	return false
}

// uriPath strips the query string from a URI for extension matching.
func uriPath(uri string) string {
	if i := strings.IndexByte(uri, '?'); i >= 0 {
		return uri[:i]
	}
	return uri
}

var staticExtensions = []string{".css", ".js", ".map", ".woff", ".woff2", ".ttf", ".eot", ".otf"}
var staticPrefixes = []string{"/css/", "/js/", "/fonts/"}
var staticFiles = []string{"robots.txt", "sitemap.xml"}

func isStaticResource(uri string) bool {
	p := uriPath(uri)
	for _, px := range staticPrefixes {
		if strings.HasPrefix(p, px) {
			return true
		}
	}
	for _, e := range staticExtensions {
		if strings.HasSuffix(p, e) {
			return true
		}
	}
	for _, f := range staticFiles {
		if strings.HasSuffix(p, "/"+f) || p == "/"+f {
			return true
		}
	}
	return false
}

var imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico", ".bmp", ".avif"}
var imagePrefixes = []string{"/img/", "/images/"}

func isImageResource(uri string) bool {
	p := uriPath(uri)
	for _, px := range imagePrefixes {
		if strings.HasPrefix(p, px) {
			return true
		}
	}
	for _, e := range imageExtensions {
		if strings.HasSuffix(p, e) {
			return true
		}
	}
	return false
}
