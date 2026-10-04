package analyzer

import (
	"net"
	"net/url"
	"strings"
)

// canonicalHost returns the form under which requests to the same site are
// grouped: lower case, without the default ports :80 and :443 that some
// clients include in the Host header. Other ports stay part of the host.
func canonicalHost(host string) string {
	host = strings.ToLower(host)
	name, port, err := net.SplitHostPort(host)
	if err != nil || (port != "80" && port != "443") {
		return host // no port, or a non-default one
	}
	if strings.Contains(name, ":") {
		return "[" + name + "]" // IPv6 literal
	}
	return name
}

// referrerHost returns the canonical host of a Referer URL, or "" if the value
// is not a URL with a host.
func referrerHost(referer string) string {
	u, err := url.Parse(referer)
	if err != nil {
		return ""
	}
	return canonicalHost(u.Host)
}
