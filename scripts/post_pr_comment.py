#!/usr/bin/env python3
import argparse
import json
import os
import sys
from urllib.request import Request, urlopen


API = "https://api.github.com"


def read_token():
    token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
    if not token and os.path.exists(".github_token"):
        token = open(".github_token", "r", encoding="utf-8").read().strip()
    if not token:
        raise RuntimeError("GitHub token not found in env or .github_token")
    return token


def post_comment(owner: str, repo: str, pr: int, body: str, token: str):
    url = f"{API}/repos/{owner}/{repo}/issues/{pr}/comments"
    payload = json.dumps({"body": body}).encode("utf-8")
    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github+json",
        "User-Agent": "tea-commenter/1.0",
        "Content-Type": "application/json",
    }
    req = Request(url, data=payload, headers=headers, method="POST")
    with urlopen(req) as resp:
        return json.load(resp)


def main():
    ap = argparse.ArgumentParser(description="Post a comment to a PR")
    ap.add_argument("--owner", required=True)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--pr", type=int, required=True)
    ap.add_argument("--body", help="Comment body text")
    ap.add_argument("--body-file", help="Path to file with comment body")
    ap.add_argument("--prefix", help="Optional prefix heading")
    args = ap.parse_args()

    if not args.body and not args.body_file:
        print("Error: --body or --body-file is required", file=sys.stderr)
        return 2
    body_text = args.body or open(args.body_file, "r", encoding="utf-8").read()
    if args.prefix:
        body_text = f"{args.prefix}\n\n" + body_text

    token = read_token()
    res = post_comment(args.owner, args.repo, args.pr, body_text, token)
    print(res.get("html_url", "commented"))
    return 0


if __name__ == "__main__":
    sys.exit(main() or 0)
