#!/usr/bin/env python3
"""Fetch content from cn.investing.com using curl_cffi to bypass Cloudflare."""

import argparse
import sys
import urllib.parse

from curl_cffi import requests as cf_requests


def fetch_page(url, content_type="html"):
    """Fetch a page from cn.investing.com using Safari impersonation."""
    headers = {
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    }
    if content_type == "json":
        headers["Accept"] = "application/json, text/plain, */*"

    try:
        r = cf_requests.get(
            url,
            impersonate="safari17_0",
            headers=headers,
            timeout=30,
            allow_redirects=True,
        )
    except Exception as e:
        sys.stderr.write(f"ERR: Failed to fetch {url}: {e}\n")
        sys.exit(1)

    if r.status_code != 200:
        sys.stderr.write(
            f"ERR: HTTP {r.status_code} for {url}: {r.text[:200]}\n"
        )
        sys.exit(1)

    if r.text.strip() == "403" or len(r.text) < 100:
        sys.stderr.write(f"ERR: Suspicious response from {url}\n")
        sys.exit(1)

    sys.stdout.write(r.text)


def fetch_search(q):
    """Call the Search API endpoint."""
    url = f"https://api.investing.com/api/search?q={urllib.parse.quote(q)}"
    try:
        r = cf_requests.get(
            url,
            impersonate="safari17_0",
            timeout=15,
            allow_redirects=True,
        )
    except Exception as e:
        sys.stderr.write(f"ERR: Search failed: {e}\n")
        sys.exit(1)

    if r.status_code != 200:
        sys.stderr.write(
            f"ERR: Search API returned {r.status_code}: {r.text[:200]}\n"
        )
        sys.exit(1)

    sys.stdout.write(r.text)


def main():
    parser = argparse.ArgumentParser(description="Fetch content from investing.com")
    subparsers = parser.add_subparsers(dest="command", required=True)

    page_parser = subparsers.add_parser("page", help="Fetch a page")
    page_parser.add_argument("url", help="URL to fetch")
    page_parser.add_argument("--type", choices=["html", "json"], default="html")

    search_parser = subparsers.add_parser("search", help="Call Search API")
    search_parser.add_argument("q", help="Search query")

    args = parser.parse_args()

    if args.command == "page":
        fetch_page(args.url, args.type)
    elif args.command == "search":
        fetch_search(args.q)


if __name__ == "__main__":
    main()
