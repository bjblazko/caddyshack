package analyzer

import (
	"sort"

	"github.com/bjblazko/caddyshack/internal/geoip"
)

func topN(m map[string]int, n int) []NameCount {
	result := make([]NameCount, 0, len(m))
	for k, v := range m {
		result = append(result, NameCount{Name: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Name < result[j].Name
	})
	if len(result) > n {
		result = result[:n]
	}
	return result
}

func sortedIntNameCounts(m map[int]int) []NameCount {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	result := make([]NameCount, len(keys))
	for i, k := range keys {
		result[i] = NameCount{Name: statusString(k), Count: m[k]}
	}
	return result
}

func statusString(code int) string {
	s := ""
	s += string(rune('0' + code/100))
	s += string(rune('0' + (code/10)%10))
	s += string(rune('0' + code%10))
	return s
}

func sortedDays(m map[string]int) []DayCount {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]DayCount, len(keys))
	for i, k := range keys {
		result[i] = DayCount{Date: k, Count: m[k]}
	}
	return result
}

func topVisitors(m map[string]int, n int) []VisitorInfo {
	type ipCount struct {
		ip    string
		count int
	}
	items := make([]ipCount, 0, len(m))
	for ip, count := range m {
		items = append(items, ipCount{ip, count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count != items[j].count {
			return items[i].count > items[j].count
		}
		return items[i].ip < items[j].ip
	})
	if len(items) > n {
		items = items[:n]
	}
	result := make([]VisitorInfo, len(items))
	for i, item := range items {
		code := geoip.Lookup(item.ip)
		result[i] = VisitorInfo{
			IP:          item.ip,
			Count:       item.count,
			Country:     code,
			CountryName: geoip.CountryName(code),
		}
	}
	return result
}

func topCountries(m map[string]int, n int) []CountryCount {
	type cc struct {
		code  string
		count int
	}
	items := make([]cc, 0, len(m))
	for code, count := range m {
		items = append(items, cc{code, count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count != items[j].count {
			return items[i].count > items[j].count
		}
		return items[i].code < items[j].code
	})
	if len(items) > n {
		items = items[:n]
	}
	result := make([]CountryCount, len(items))
	for i, item := range items {
		result[i] = CountryCount{
			Code:  item.code,
			Name:  geoip.CountryName(item.code),
			Count: item.count,
		}
	}
	return result
}
