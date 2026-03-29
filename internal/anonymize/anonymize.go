package anonymize

import "strings"

// IP anonymizes an IP address for GDPR compliance.
// IPv4: zeros the last octet. IPv6: keeps first 3 groups.
func IP(ip string) string {
	if strings.Contains(ip, ":") {
		return ipv6(ip)
	}
	return ipv4(ip)
}

func ipv4(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		parts[3] = "0"
		return strings.Join(parts, ".")
	}
	return ip
}

func ipv6(ip string) string {
	parts := strings.Split(ip, ":")
	if len(parts) >= 4 {
		return strings.Join(parts[:3], ":") + "::"
	}
	return ip
}
