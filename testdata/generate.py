#!/usr/bin/env python3
"""Generate a realistic Caddy JSONL test log file modeled on example.com.

Uses sample page URIs, asset paths, User-Agent strings, and IP ranges
from various countries to produce a plausible 14-day traffic sample.
"""

import json
import random
import math
from datetime import datetime, timedelta, timezone

random.seed(42)  # reproducible

OUTPUT = "testdata/sample-example.jsonl"
DAYS = 14
HOST = "example.com"
# Approximate daily visits for a small personal site
BASE_DAILY_REQUESTS = 80

# ── Sample pages ────────────────────────────────────────────────────

PAGES = [
    ("/", 30),                                  # homepage, highest weight
    ("/about.html", 10),
    ("/products/", 12),
    ("/products/product-a.html", 15),
    ("/products/product-b.html", 12),
    ("/blog/", 8),
    ("/blog/2026-03-welcome.html", 6),
]

# Assets loaded per page view (not all every time)
ASSETS = [
    "/css/style.css",
    "/css/material-symbols-outlined.ttf",
    "/js/site.js",
    "/img/logo-256.png",
    "/img/logo-alpha-256.png",
    "/img/logo-product-b.png",
    "/img/logo-product-b.svg",
    "/img/logo-product-a-96.png",
    "/img/logo-product-a.svg",
    "/img/screenshot-1-overview.jpg",
    "/img/screenshot-2-fullscreen-and-exif.jpg",
    "/img/screenshot-3-filemanager.jpg",
    "/img/screenshot-4-wastebin.jpg",
    "/img/screenshot-5-addgeolocation.jpg",
    "/img/product-b-bio.jpg",
    "/img/product-b-nodes.jpg",
    "/img/product-b-tree.jpg",
    "/favicon.ico",
]

# Pages that scanners/404s hit
SCAN_PATHS = [
    "/wp-login.php", "/wp-admin/", "/.env", "/xmlrpc.php",
    "/admin/", "/phpmyadmin/", "/.git/config", "/api/v1/",
    "/../../etc/passwd", "/actuator/health", "/.well-known/security.txt",
]

# ── User-Agents ─────────────────────────────────────────────────────

UA_POOL = [
    # Chrome on macOS (most common)
    ("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36", 25),
    # Chrome on Windows
    ("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36", 20),
    # Chrome on Linux
    ("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36", 8),
    # Firefox on Windows
    ("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0", 10),
    # Firefox on Linux
    ("Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0", 5),
    # Safari on macOS
    ("Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15", 8),
    # Safari on iPhone
    ("Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1", 7),
    # Safari on iPad
    ("Mozilla/5.0 (iPad; CPU OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1", 3),
    # Edge on Windows
    ("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0", 5),
    # Chrome on Android
    ("Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.6778.81 Mobile Safari/537.36", 6),
    # Vivaldi
    ("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Vivaldi/7.0", 2),
    # Opera
    ("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 OPR/117.0.0.0", 2),
    # Googlebot
    ("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", 4),
    # Bingbot
    ("Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", 2),
    # curl
    ("curl/8.7.1", 2),
    # Uptime monitor
    ("Mozilla/5.0 (compatible; UptimeRobot/2.0; http://www.uptimerobot.com/)", 1),
    # Brave (identifies as Chrome but with Brave marker in some builds)
    ("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Brave Chrome/131.0.0.0 Safari/537.36", 1),
]

# ── IP pools by country (using realistic public ranges) ─────────────

IP_POOLS = [
    # (country weight, ip_prefix_generator)
    # Germany — heaviest
    ("DE", 35, lambda: f"{random.choice([84,85,87,91,93,95,130,134,141,178,188,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # United States
    ("US", 15, lambda: f"{random.choice([3,8,12,15,17,18,23,24,32,35,38,40,44,45,47,50,52,54,55,56,57,63,64,65,66,67,68,69,70,71,72,73,74,75,76,96,97,98,99,100,104,108])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Austria
    ("AT", 8, lambda: f"{random.choice([77,78,80,81,83,86,88,89,131,143,146,185,193,194,195])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Switzerland
    ("CH", 5, lambda: f"{random.choice([31,46,77,80,81,82,83,84,85,86,87,130,131,141,146,178,185,188,193,194,195,212,213,217])}.{random.randint(100,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # France
    ("FR", 5, lambda: f"{random.choice([2,5,31,37,62,77,78,80,81,82,83,84,86,88,89,90,92,109])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Netherlands
    ("NL", 4, lambda: f"{random.choice([2,5,31,37,62,77,80,81,82,83,84,85,86,87,89,130,131,141,143,145,146,149,154,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # United Kingdom
    ("GB", 5, lambda: f"{random.choice([2,5,25,31,37,51,62,77,78,80,81,82,83,86,87,88,89,90,91,92,109,130,131,141,146,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Japan
    ("JP", 3, lambda: f"{random.choice([1,14,27,36,42,49,58,59,60,61,101,106,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,133,150,153,157,158,160,163,175,180,182,183,202,203,210,211,219,220,221])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Brazil
    ("BR", 3, lambda: f"{random.choice([131,138,139,143,146,152,161,168,170,177,179,186,187,189,191,200,201])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # India
    ("IN", 3, lambda: f"{random.choice([1,14,27,36,42,43,49,59,61,101,103,106,110,111,112,114,115,116,117,119,120,121,122,124,125])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Poland
    ("PL", 3, lambda: f"{random.choice([5,31,37,46,62,77,78,80,81,83,85,86,87,89,91,93,95,141,153,156,159,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Sweden
    ("SE", 2, lambda: f"{random.choice([2,5,31,46,62,77,78,80,81,83,84,85,86,87,89,90,91,130,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Canada
    ("CA", 2, lambda: f"{random.choice([24,64,65,66,67,68,69,70,72,74,96,97,99,142,192,198,199,204,205,206,207,208,209])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Australia
    ("AU", 2, lambda: f"{random.choice([1,14,27,36,42,43,49,58,59,60,61,101,103,106,110,112,114,115,116,117,119,120,121,122,124,125,139,143,144,175,180,182,183,202,203,210,211])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Russia
    ("RU", 2, lambda: f"{random.choice([2,5,31,37,46,62,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,109,141,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Spain
    ("ES", 2, lambda: f"{random.choice([2,5,31,37,62,77,78,80,81,82,83,84,85,86,87,88,89,90,109,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
    # Italy
    ("IT", 2, lambda: f"{random.choice([2,5,31,37,46,62,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,109,151,176,178,185,188,193,194,195,212,213,217])}.{random.randint(1,254)}.{random.randint(1,254)}.{random.randint(1,254)}"),
]

# Also some IPv6 visitors (~5%)
IPV6_PREFIXES = [
    "2a01:4f8:",   # Hetzner DE
    "2a02:810:",   # Telekom DE
    "2a02:8109:",  # Vodafone DE
    "2001:1a81:",  # NetCologne DE
    "2607:f8b0:",  # Google US
    "2a00:1450:",  # Google EU
]

# ── TLS configs ─────────────────────────────────────────────────────

TLS_CONFIGS = [
    # TLS 1.3 / h2 (majority)
    {"resumed": False, "version": 772, "cipher_suite": 4865, "proto": "h2", "server_name": HOST, "ech": False},
    {"resumed": True,  "version": 772, "cipher_suite": 4865, "proto": "h2", "server_name": HOST, "ech": False},
    {"resumed": False, "version": 772, "cipher_suite": 4867, "proto": "h2", "server_name": HOST, "ech": False},
    # TLS 1.2 / http/1.1 (bots, curl, older clients)
    {"resumed": False, "version": 771, "cipher_suite": 49199, "proto": "http/1.1", "server_name": HOST, "ech": False},
]

# ── Helpers ─────────────────────────────────────────────────────────

def weighted_choice(items_with_weights):
    items, weights = zip(*items_with_weights)
    return random.choices(items, weights=weights, k=1)[0]

def pick_ip():
    """Pick a random IP from the country-weighted pool."""
    if random.random() < 0.05:
        # IPv6
        prefix = random.choice(IPV6_PREFIXES)
        return f"{prefix}{random.randint(0,0xffff):x}:{random.randint(0,0xffff):x}::1"
    countries, weights, generators = zip(*IP_POOLS)
    idx = random.choices(range(len(IP_POOLS)), weights=weights, k=1)[0]
    return IP_POOLS[idx][2]()

def pick_ua():
    return weighted_choice(UA_POOL)

def pick_tls(ua):
    """Bots and curl get TLS 1.2/http1.1, others mostly TLS 1.3/h2."""
    if "bot" in ua.lower() or "crawl" in ua.lower() or "curl" in ua.lower():
        return TLS_CONFIGS[3]
    if random.random() < 0.08:
        return TLS_CONFIGS[3]  # occasional TLS 1.2
    return random.choice(TLS_CONFIGS[:3])

def make_entry(ts, ip, ua, method, uri, status, size, duration):
    proto = "HTTP/2.0" if "h2" in pick_tls(ua).get("proto", "") else "HTTP/1.1"
    tls = pick_tls(ua)
    if tls["proto"] == "http/1.1":
        proto = "HTTP/1.1"
    else:
        proto = "HTTP/2.0"

    headers = {"User-Agent": [ua], "Accept": ["text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"]}
    if "css" in uri or "js/" in uri:
        headers["Accept"] = ["text/css,*/*;q=0.1"] if ".css" in uri else ["*/*"]
        headers["Accept-Encoding"] = ["gzip, deflate, br"]
    if "img/" in uri or ".ico" in uri or ".jpg" in uri or ".png" in uri:
        headers["Accept"] = ["image/avif,image/webp,image/apng,image/*,*/*;q=0.8"]
    if random.random() < 0.4:
        headers["Accept-Language"] = [random.choice([
            "de-DE,de;q=0.9,en;q=0.5",
            "en-US,en;q=0.9,de;q=0.8",
            "en-US,en;q=0.9",
            "de-DE,de;q=0.9",
            "fr-FR,fr;q=0.9,en;q=0.5",
        ])]
    if random.random() < 0.3 and status == 200:
        headers["Referer"] = [f"https://{HOST}/"]

    resp_headers = {"Server": ["Caddy"]}
    if status == 200 and size > 0:
        if uri.endswith((".html", "/")) or "/" == uri:
            resp_headers["Content-Type"] = ["text/html; charset=utf-8"]
        elif uri.endswith(".css"):
            resp_headers["Content-Type"] = ["text/css; charset=utf-8"]
            resp_headers["Content-Encoding"] = ["gzip"]
            resp_headers["Vary"] = ["Accept-Encoding"]
        elif uri.endswith(".js"):
            resp_headers["Content-Type"] = ["application/javascript; charset=utf-8"]
            resp_headers["Content-Encoding"] = ["gzip"]
            resp_headers["Vary"] = ["Accept-Encoding"]
        elif uri.endswith((".png", ".jpg", ".svg", ".ico")):
            ext_map = {".png": "image/png", ".jpg": "image/jpeg", ".svg": "image/svg+xml", ".ico": "image/x-icon"}
            for ext, ct in ext_map.items():
                if uri.endswith(ext):
                    resp_headers["Content-Type"] = [ct]
                    break
        elif uri.endswith(".ttf"):
            resp_headers["Content-Type"] = ["font/ttf"]

    return {
        "level": "info",
        "ts": ts,
        "logger": "http.log.access.log0",
        "msg": "handled request",
        "request": {
            "remote_ip": ip,
            "remote_port": str(random.randint(1024, 65535)),
            "client_ip": ip,
            "proto": proto,
            "method": method,
            "host": HOST,
            "uri": uri,
            "headers": headers,
            "tls": tls,
        },
        "bytes_read": 0,
        "user_id": "",
        "duration": duration,
        "size": size,
        "status": status,
        "resp_headers": resp_headers,
    }

# ── Main generation ─────────────────────────────────────────────────

def generate():
    entries = []
    now = datetime(2026, 3, 24, 12, 0, 0, tzinfo=timezone.utc)
    start = now - timedelta(days=DAYS)

    # Generate a pool of "returning visitors" (IPs that appear multiple times)
    returning_ips = [(pick_ip(), pick_ua()) for _ in range(30)]

    for day_offset in range(DAYS):
        day_start = start + timedelta(days=day_offset)

        # Vary daily volume: weekdays more, weekends less, some randomness
        weekday = day_start.weekday()
        if weekday < 5:
            daily_count = int(BASE_DAILY_REQUESTS * random.uniform(0.8, 1.3))
        else:
            daily_count = int(BASE_DAILY_REQUESTS * random.uniform(0.4, 0.7))

        for _ in range(daily_count):
            # Time within the day — weighted toward business hours (UTC)
            hour = int(random.triangular(6, 22, 14))
            minute = random.randint(0, 59)
            second = random.randint(0, 59)
            micro = random.randint(0, 999999)
            ts_dt = day_start.replace(hour=hour, minute=minute, second=second, microsecond=micro)
            ts = ts_dt.timestamp()

            # 30% chance of being a returning visitor
            if random.random() < 0.3 and returning_ips:
                ip, ua = random.choice(returning_ips)
            else:
                ip = pick_ip()
                ua = pick_ua()

            # Decide what kind of request this is
            roll = random.random()

            if roll < 0.03:
                # Scanner / vulnerability probe
                uri = random.choice(SCAN_PATHS)
                entries.append(make_entry(ts, ip, "Mozilla/5.0", "GET", uri, 404, 0, random.uniform(0.0001, 0.001)))

            elif roll < 0.08:
                # Bot crawling real pages
                page_uri = weighted_choice(PAGES)
                entries.append(make_entry(ts, ip, ua, "GET", page_uri,
                    200, random.randint(2000, 12000), random.uniform(0.001, 0.008)))

            elif roll < 0.12:
                # HEAD request (monitoring/curl)
                entries.append(make_entry(ts, ip, "curl/8.7.1", "HEAD", "/",
                    200, 0, random.uniform(0.0001, 0.0005)))

            else:
                # Normal page view + associated asset requests
                page_uri = weighted_choice(PAGES)
                page_size = random.randint(3000, 15000)
                page_duration = random.uniform(0.002, 0.015)

                # Occasional 304 Not Modified
                if random.random() < 0.1:
                    entries.append(make_entry(ts, ip, ua, "GET", page_uri, 304, 0, random.uniform(0.0003, 0.001)))
                else:
                    entries.append(make_entry(ts, ip, ua, "GET", page_uri, 200, page_size, page_duration))

                # 2-6 asset requests follow within ~100ms
                num_assets = random.randint(2, 6)
                selected_assets = random.sample(ASSETS, min(num_assets, len(ASSETS)))
                for asset in selected_assets:
                    asset_ts = ts + random.uniform(0.05, 0.3)
                    asset_size = random.randint(200, 120000) if ".jpg" in asset else random.randint(200, 30000)
                    # Occasional 304
                    if random.random() < 0.15:
                        entries.append(make_entry(asset_ts, ip, ua, "GET", asset, 304, 0, random.uniform(0.0002, 0.0008)))
                    else:
                        entries.append(make_entry(asset_ts, ip, ua, "GET", asset, 200, asset_size, random.uniform(0.0003, 0.003)))

    # Sort by timestamp
    entries.sort(key=lambda e: e["ts"])

    # Write JSONL
    with open(OUTPUT, "w") as f:
        for entry in entries:
            f.write(json.dumps(entry, separators=(",", ":")) + "\n")

    print(f"Generated {len(entries)} log entries over {DAYS} days → {OUTPUT}")
    print(f"Date range: {start.strftime('%Y-%m-%d')} to {(now - timedelta(days=1)).strftime('%Y-%m-%d')}")

if __name__ == "__main__":
    generate()
