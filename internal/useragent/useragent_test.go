package useragent

import "testing"

const (
	chromeMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	safariIOS = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	firefox   = "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0"
	edgeWin   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0"
)

func TestClassify(t *testing.T) {
	cases := map[string]Kind{
		chromeMac:             KindBrowser,
		safariIOS:             KindBrowser,
		firefox:               KindBrowser,
		"Uptime-Kuma/1.23.11": KindMonitor,
		"Mozilla/5.0+(compatible; UptimeRobot/2.0; http://www.uptimerobot.com/)":                                 KindMonitor,
		"Pingdom.com_bot_version_1.4_(http://www.pingdom.com/)":                                                  KindMonitor,
		"Better Stack Better Uptime Bot Mozilla/5.0":                                                             KindMonitor,
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)":                               KindBot,
		"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)": KindBot,
		"Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)":                                     KindBot,
		"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)":                              KindBot,
		"WhatsApp/2.23.20.0": KindBot,
		"Mastodon/4.2.0 (http.rb/5.1.1; +https://social.example/)": KindBot,
		"Feedly/1.0 (+http://www.feedly.com/fetcher.html)":         KindBot,
		"curl/8.4.0":             KindScript,
		"Wget/1.21.4":            KindScript,
		"python-requests/2.31.0": KindScript,
		"Go-http-client/1.1":     KindScript,
		"Mozilla/5.0 zgrab/0.x":  KindScript,
		"Mozilla/5.0 (compatible; Nmap Scripting Engine; https://nmap.org/book/nse.html)": KindScript,
		"": KindScript,
	}
	for ua, want := range cases {
		if got := Classify(ua); got != want {
			t.Errorf("Classify(%q) = %s, want %s", ua, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	cases := []struct{ ua, browser, os string }{
		{chromeMac, "Chrome", "macOS"},
		{safariIOS, "Safari", "iOS"},
		{firefox, "Firefox", "Linux"},
		{edgeWin, "Edge", "Windows"},
		{"Uptime-Kuma/1.23.11", "Monitor", "Other"},
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", "Bot", "Other"},
		{"curl/8.4.0", "Script", "Other"},
		{"Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Mobile Safari/537.36", "Chrome", "Android"},
	}
	for _, c := range cases {
		if _, b, o := Parse(c.ua); b != c.browser || o != c.os {
			t.Errorf("Parse(%q) = %s/%s, want %s/%s", c.ua, b, o, c.browser, c.os)
		}
	}
}

func TestFromNames(t *testing.T) {
	cases := []struct {
		browser, os         string
		kind                Kind
		wantBrowser, wantOS string
	}{
		{"Firefox", "Linux", KindBrowser, "Firefox", "Linux"},
		{"Bot", "Other", KindBot, "Bot", "Other"},
		{"curl", "", KindScript, "Script", "Other"},
		{"Monitor", "Other", KindMonitor, "Monitor", "Other"},
		{"Other", "Windows", KindBrowser, "Other", "Windows"},
	}
	for _, c := range cases {
		k, b, o := FromNames(c.browser, c.os)
		if k != c.kind || b != c.wantBrowser || o != c.wantOS {
			t.Errorf("FromNames(%q, %q) = %s/%s/%s", c.browser, c.os, k, b, o)
		}
	}
}
