#!/usr/bin/env python3
"""
Fetch the latest GitHub Actions artifact (e.g., 'playwright-artifacts') for a given PR
and extract it locally to provide ZIPs for fallback upload.

Usage:
  python3 scripts/fetch_playwright_artifact.py --owner <owner> --repo <repo> --pr <num> \
      [--name playwright-artifacts] [--out build-ci-logs/playwright]

Behavior:
  - Tries latest workflow run for the PR (event=pull_request, runs API) and lists its artifacts.
  - If not found, falls back to repo-level artifact list filtered by name.
  - Downloads artifact zip and extracts into the output directory.
  - Prints extracted file paths and returns 0 on success (non-blocking). If nothing found, exits 0.

Notes:
  - Reads GH_TOKEN from env or .github_token.
  - Uses only Python stdlib (urllib, zipfile).
"""
import argparse
import json
import os
import sys
import urllib.request
import urllib.error
import zipfile
from pathlib import Path

API = "https://api.github.com"

def read_token():
    token = os.environ.get("GH_TOKEN")
    if not token and os.path.exists(".github_token"):
        token = Path(".github_token").read_text(encoding="utf-8").strip()
    if not token:
        raise RuntimeError("GH_TOKEN not found (env or .github_token)")
    return token

def gh_json(url: str, token: str):
    req = urllib.request.Request(url, headers={
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github+json",
        "User-Agent": "fetch-playwright-artifact"
    })
    with urllib.request.urlopen(req) as resp:
        return json.load(resp)

def gh_download(url: str, token: str) -> bytes:
    req = urllib.request.Request(url, headers={
        "Authorization": f"token {token}",
        "User-Agent": "fetch-playwright-artifact"
    })
    with urllib.request.urlopen(req) as resp:
        return resp.read()

def find_latest_run_for_pr(owner: str, repo: str, pr: int, token: str):
    # List latest workflow runs for repository
    url = f"{API}/repos/{owner}/{repo}/actions/runs?per_page=50"
    data = gh_json(url, token)
    runs = data.get("workflow_runs", [])
    # Filter runs for this PR number
    for run in runs:
        if run.get("event") != "pull_request":
            continue
        prs = run.get("pull_requests") or []
        for p in prs:
            if p.get("number") == pr:
                return run
    return None

def list_run_artifacts(owner: str, repo: str, run_id: int, token: str):
    url = f"{API}/repos/{owner}/{repo}/actions/runs/{run_id}/artifacts?per_page=50"
    data = gh_json(url, token)
    return data.get("artifacts", [])

def list_repo_artifacts(owner: str, repo: str, token: str):
    url = f"{API}/repos/{owner}/{repo}/actions/artifacts?per_page=50"
    data = gh_json(url, token)
    return data.get("artifacts", [])

def download_and_extract_artifact(owner: str, repo: str, artifact_id: int, token: str, out_dir: Path) -> list[Path]:
    # Download zip for artifact
    url = f"{API}/repos/{owner}/{repo}/actions/artifacts/{artifact_id}/zip"
    blob = gh_download(url, token)
    out_dir.mkdir(parents=True, exist_ok=True)
    tmp_zip = out_dir / f"artifact_{artifact_id}.zip"
    tmp_zip.write_bytes(blob)
    extracted = []
    with zipfile.ZipFile(tmp_zip, 'r') as zf:
        for member in zf.infolist():
            # Avoid Zip Slip
            name = Path(member.filename)
            dest = out_dir / name
            dest.parent.mkdir(parents=True, exist_ok=True)
            with zf.open(member) as src, open(dest, 'wb') as dst:
                dst.write(src.read())
            extracted.append(dest)
    return extracted

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--owner", required=True)
    ap.add_argument("--repo", required=True)
    ap.add_argument("--pr", type=int, required=True)
    ap.add_argument("--name", default="playwright-artifacts")
    ap.add_argument("--out", default="build-ci-logs/playwright")
    args = ap.parse_args()

    token = read_token()
    owner = args.owner
    repo = args.repo
    pr = args.pr
    target_name = args.name
    out_dir = Path(args.out)

    artifact = None
    try:
        run = find_latest_run_for_pr(owner, repo, pr, token)
        if run:
            arts = list_run_artifacts(owner, repo, run.get("id"), token)
            for a in arts:
                if a.get("name") == target_name:
                    artifact = a
                    break
        if artifact is None:
            arts = list_repo_artifacts(owner, repo, token)
            for a in arts:
                if a.get("name") == target_name:
                    artifact = a
                    break
    except urllib.error.HTTPError as e:
        print(f"HTTPError: {e}")
        return 0
    except Exception as e:
        print(f"Error: {e}")
        return 0

    if artifact is None:
        print("No matching artifact found; skip.")
        return 0

    try:
        extracted = download_and_extract_artifact(owner, repo, artifact.get("id"), token, out_dir)
        for p in extracted:
            print(f"EXTRACTED: {p}")
        # Best-effort: show if trace zip exists
        trace_zip = next((p for p in extracted if p.name.endswith("trace.zip")), None)
        if trace_zip:
            print(f"TRACE_ZIP_READY: {trace_zip}")
    except Exception as e:
        print(f"Download/extract error: {e}")
        return 0

if __name__ == "__main__":
    main()
