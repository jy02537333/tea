#!/usr/bin/env python3
"""
Check health of evidence links in a PR body.

Usage:
  python3 scripts/check_evidence_links.py --owner <owner> --repo <repo> --pr <num>

Notes:
  - Reads GH_TOKEN from env or .github_token file.
  - Parses four evidence links from grouped bullets (supports optional checkboxes) or table.
  - Performs HEAD requests (timeout 10s) and prints a compact report.
  - Exits 0 always (non-blocking). Add --strict to exit non-zero on any failure.
"""
import argparse
import json
import os
import re
import sys
import urllib.request
import urllib.error

API = "https://api.github.com"


def read_token():
    token = os.environ.get("GH_TOKEN")
    if not token and os.path.exists(".github_token"):
        token = open(".github_token", "r", encoding="utf-8").read().strip()
    if not token:
        raise RuntimeError("GH_TOKEN not found (env or .github_token)")
    return token


def fetch_issue_body(owner, repo, pr, token):
    url = f"{API}/repos/{owner}/{repo}/issues/{pr}"
    req = urllib.request.Request(
        url,
        headers={
            "Authorization": f"token {token}",
            "Accept": "application/vnd.github+json",
            "User-Agent": "evidence-link-checker",
        },
    )
    with urllib.request.urlopen(req) as resp:
        data = json.load(resp)
    return data.get("body") or ""


def extract_urls(body: str):
    cb = r"(?:\[[ xX]\]\s*)?"  # optional checkbox
    patterns_bullets = [
        (rf"-\s*{cb}申请前截图（页面）[:：]\s*(\S+)", "申请前截图（页面）"),
        (rf"-\s*{cb}审核弹窗截图[:：]\s*(\S+)", "审核弹窗截图"),
        (
            rf"-\s*{cb}Playwright\s*Trace\s*压缩包[:：]\s*(\S+)",
            "Playwright Trace 压缩包",
        ),
        (rf"-\s*{cb}Trace\s*Report\s*HTML[:：]\s*(\S+)", "Trace Report HTML"),
    ]
    urls = {}
    for pat, key in patterns_bullets:
        m = re.search(pat, body)
        if m:
            urls[key] = m.group(1)
    if len(urls) >= 4:
        return urls
    # Try table
    lines = body.splitlines()
    for i in range(len(lines) - 1):
        if re.match(r"^\|\s*项目\s*\|\s*链接\s*\|\s*$", lines[i]) and re.match(
            r"^\|\s*-+\s*\|\s*-+\s*\|\s*$", lines[i + 1]
        ):
            j = i + 2
            while j < len(lines) and lines[j].strip().startswith("|"):
                m = re.match(r"^\|\s*(.*?)\s*\|\s*(\S+)\s*\|\s*$", lines[j])
                if m:
                    label = m.group(1).strip()
                    link = m.group(2).strip()
                    if label in (
                        "申请前截图（页面）",
                        "审核弹窗截图",
                        "Playwright Trace 压缩包",
                        "Trace Report HTML",
                    ):
                        urls[label] = link
                j += 1
            break
    return urls


def head_ok(url: str, timeout: float = 10.0) -> tuple[bool, int]:
    try:
        req = urllib.request.Request(
            url, method="HEAD", headers={"User-Agent": "evidence-link-checker"}
        )
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            code = getattr(resp, "status", 200)
            return (200 <= code < 400), code
    except urllib.error.HTTPError as e:
        return False, e.code
    except Exception:
        return False, 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--owner", required=True)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--pr", type=int, required=True)
    ap.add_argument("--strict", action="store_true")
    args = ap.parse_args()

    token = read_token()
    body = fetch_issue_body(args.owner, args.repo, args.pr, token)
    urls = extract_urls(body)
    labels = [
        "申请前截图（页面）",
        "审核弹窗截图",
        "Playwright Trace 压缩包",
        "Trace Report HTML",
    ]
    failures = 0
    for label in labels:
        url = urls.get(label, "")
        ok, code = head_ok(url) if url else (False, 0)
        status = "OK" if ok else "FAIL"
        print(f"- {label}: {status} ({code}) {url}")
        if not ok:
            failures += 1
    if failures and args.strict:
        sys.exit(1)


if __name__ == "__main__":
    main()
