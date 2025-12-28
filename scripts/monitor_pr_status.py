#!/usr/bin/env python3
import argparse
import json
import os
import sys
import time
from datetime import datetime, timezone
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError


def http_get(url: str, token: str | None = None) -> dict:
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "tea-monitor/1.0",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = Request(url, headers=headers)
    try:
        with urlopen(req, timeout=20) as resp:
            data = resp.read().decode("utf-8")
            return json.loads(data)
    except HTTPError as e:
        return {"error": f"HTTPError {e.code}", "detail": str(e)}
    except URLError as e:
        return {"error": "URLError", "detail": str(e)}


def ts():
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def log_line(fp, line: str):
    fp.write(line + "\n")
    fp.flush()


def get_pr_head_sha(owner: str, repo: str, pr: int, token: str | None) -> str | None:
    url = f"https://api.github.com/repos/{owner}/{repo}/pulls/{pr}"
    data = http_get(url, token)
    if isinstance(data, dict) and data.get("head") and data["head"].get("sha"):
        return data["head"]["sha"]
    return None


def get_commit_status(owner: str, repo: str, sha: str, token: str | None) -> dict:
    # Combined status API
    url = f"https://api.github.com/repos/{owner}/{repo}/commits/{sha}/status"
    return http_get(url, token)


def summarize_status(status_json: dict) -> tuple[str, list[tuple[str, str]]]:
    state = status_json.get("state", "unknown")
    contexts = []
    for s in status_json.get("statuses", []) or []:
        contexts.append((s.get("context", "unknown"), s.get("state", "unknown")))
    return state, contexts


def main():
    parser = argparse.ArgumentParser(description="Monitor PR commit statuses and log failures")
    parser.add_argument("--owner", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--pr", type=int, required=True)
    parser.add_argument("--interval", type=int, default=30, help="Polling interval seconds")
    parser.add_argument("--log", default="build-ci-logs/monitor_pr65.log")
    parser.add_argument("--once", action="store_true", help="Run only once and exit")
    args = parser.parse_args()

    token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")

    # Ensure log dir
    log_dir = os.path.dirname(args.log) or "."
    os.makedirs(log_dir, exist_ok=True)

    # Open log file append
    with open(args.log, "a", encoding="utf-8") as fp:
        log_line(fp, f"[{ts()}] monitor start owner={args.owner} repo={args.repo} pr={args.pr} interval={args.interval}s")

        last_state = None
        last_sha = None
        while True:
            sha = get_pr_head_sha(args.owner, args.repo, args.pr, token)
            if not sha:
                msg = f"[{ts()}] ERROR: unable to retrieve PR head sha"
                print(msg)
                log_line(fp, msg)
                if args.once:
                    return 1
                time.sleep(args.interval)
                continue

            if sha != last_sha:
                log_line(fp, f"[{ts()}] head sha={sha}")
                last_sha = sha

            st_json = get_commit_status(args.owner, args.repo, sha, token)
            if st_json.get("error"):
                msg = f"[{ts()}] ERROR: status api failed: {st_json.get('error')} {st_json.get('detail')}"
                print(msg)
                log_line(fp, msg)
                if args.once:
                    return 1
                time.sleep(args.interval)
                continue

            state, contexts = summarize_status(st_json)
            ctx_str = ", ".join([f"{c}:{s}" for c, s in contexts])
            line = f"[{ts()}] combined_state={state} contexts=[{ctx_str}]"
            log_line(fp, line)

            # Print alerts for failures
            if state in ("failure", "error") or any(s in ("failure", "error") for _, s in contexts):
                alert = f"[{ts()}] ALERT: failure detected combined_state={state} contexts=[{ctx_str}]"
                print(alert)
                log_line(fp, alert)

            if args.once:
                return 0

            time.sleep(args.interval)


if __name__ == "__main__":
    sys.exit(main() or 0)
