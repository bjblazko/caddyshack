package geoip

import (
	"encoding/binary"
	"encoding/csv"
	"io"
	"log"
	"net"
	"os"
	"sort"
	"strings"
)

var (
	starts    []uint32
	ends      []uint32
	countries []string
	loaded    bool
)

// CountryNames maps ISO 3166-1 alpha-2 codes to display names.
var CountryNames = map[string]string{
	"AD": "Andorra", "AE": "UAE", "AF": "Afghanistan", "AL": "Albania",
	"AM": "Armenia", "AR": "Argentina", "AT": "Austria", "AU": "Australia",
	"AZ": "Azerbaijan", "BA": "Bosnia", "BD": "Bangladesh", "BE": "Belgium",
	"BG": "Bulgaria", "BR": "Brazil", "BY": "Belarus", "CA": "Canada",
	"CH": "Switzerland", "CL": "Chile", "CN": "China", "CO": "Colombia",
	"CZ": "Czechia", "DE": "Germany", "DK": "Denmark", "DZ": "Algeria",
	"EC": "Ecuador", "EE": "Estonia", "EG": "Egypt", "ES": "Spain",
	"FI": "Finland", "FR": "France", "GB": "United Kingdom", "GE": "Georgia",
	"GH": "Ghana", "GR": "Greece", "HK": "Hong Kong", "HR": "Croatia",
	"HU": "Hungary", "ID": "Indonesia", "IE": "Ireland", "IL": "Israel",
	"IN": "India", "IQ": "Iraq", "IR": "Iran", "IS": "Iceland",
	"IT": "Italy", "JP": "Japan", "KE": "Kenya", "KR": "South Korea",
	"KZ": "Kazakhstan", "LT": "Lithuania", "LU": "Luxembourg", "LV": "Latvia",
	"MA": "Morocco", "MD": "Moldova", "ME": "Montenegro", "MK": "N. Macedonia",
	"MX": "Mexico", "MY": "Malaysia", "NG": "Nigeria", "NL": "Netherlands",
	"NO": "Norway", "NZ": "New Zealand", "PE": "Peru", "PH": "Philippines",
	"PK": "Pakistan", "PL": "Poland", "PT": "Portugal", "RO": "Romania",
	"RS": "Serbia", "RU": "Russia", "SA": "Saudi Arabia", "SE": "Sweden",
	"SG": "Singapore", "SI": "Slovenia", "SK": "Slovakia", "TH": "Thailand",
	"TN": "Tunisia", "TR": "Türkiye", "TW": "Taiwan", "UA": "Ukraine",
	"US": "United States", "UY": "Uruguay", "UZ": "Uzbekistan", "VN": "Vietnam",
	"ZA": "South Africa", "ZZ": "Unknown",
}

// Load reads a DB-IP country CSV into memory for binary search lookups.
// Only IPv4 ranges are loaded. If the file doesn't exist, lookups return "??".
func Load(path string) {
	if loaded {
		return
	}
	loaded = true

	f, err := os.Open(path)
	if err != nil {
		log.Printf("GeoIP database not found at %s — country lookups disabled", path)
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 3 {
			continue
		}
		startIP, endIP, country := record[0], record[1], record[2]
		if strings.Contains(startIP, ":") {
			continue // skip IPv6
		}
		s := ipToUint32(startIP)
		e := ipToUint32(endIP)
		if s == 0 && e == 0 {
			continue
		}
		starts = append(starts, s)
		ends = append(ends, e)
		countries = append(countries, country)
	}

	log.Printf("GeoIP loaded: %d IPv4 ranges", len(starts))
}

// Lookup returns the ISO country code for an IPv4 address.
// Returns "??" for IPv6 addresses or if no database is loaded.
func Lookup(ip string) string {
	if strings.Contains(ip, ":") {
		return "??"
	}
	if len(starts) == 0 {
		return "??"
	}
	ipInt := ipToUint32(ip)
	if ipInt == 0 {
		return "??"
	}

	idx := sort.Search(len(starts), func(i int) bool {
		return starts[i] > ipInt
	}) - 1

	if idx >= 0 && ipInt <= ends[idx] {
		return countries[idx]
	}
	return "??"
}

// CountryName returns the display name for an ISO country code.
func CountryName(code string) string {
	if name, ok := CountryNames[code]; ok {
		return name
	}
	return code
}

func ipToUint32(ip string) uint32 {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return 0
	}
	v4 := parsed.To4()
	if v4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(v4)
}
