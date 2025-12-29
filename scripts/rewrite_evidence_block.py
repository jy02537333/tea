#!/usr/bin/env python3
"""
Rewrite a GitHub PR evidence block to grouped format with checkboxes.

Usage:
  python3 scripts/rewrite_evidence_block.py --owner jy02537333 --repo tea --pr 65 \
      [--checkbox checked|unchecked] [--urls-file build-ci-logs/e2e_urls.json]

Notes:
- GH_TOKEN is read from environment or .github_token file.
- Prefer URLs from --urls-file; fallback to parsing existing PR body (table or bullets).
- Replaces an existing table or bullet block of four items, or inserts under a header containing "验证证据".
- Writes intermediate outputs under build-ci-logs and prints updated_at.
"""
import argparse
import json
import os
import re
import sys
import urllib.request
import difflib

API = "https://api.github.com"
LOG_DIR = "build-ci-logs"


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
            "User-Agent": "evidence-rewriter",
        },
    )
    with urllib.request.urlopen(req) as resp:
        data = json.load(resp)
    os.makedirs(LOG_DIR, exist_ok=True)
    with open(
        os.path.join(LOG_DIR, f"pr{pr}_detail_for_update.json"), "w", encoding="utf-8"
    ) as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    return data.get("body") or ""


def extract_urls_from_body(body: str):
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


def build_group_block(urls: dict, checkbox: str):
    ck = "[x]" if checkbox == "checked" else "[ ]"
    lines = []
    lines.append("截图")
    lines.append(f"- {ck} 申请前截图（页面）： {urls.get('申请前截图（页面）','')}")
    lines.append(f"- {ck} 审核弹窗截图： {urls.get('审核弹窗截图','')}")
    lines.append("")
    lines.append("Trace")
    lines.append(
        f"- {ck} Playwright Trace 压缩包： {urls.get('Playwright Trace 压缩包','')}"
    )
    lines.append(f"- {ck} Trace Report HTML： {urls.get('Trace Report HTML','')}")
    return "\n".join(lines)


def replace_block(body: str, group_block: str):
    lines = body.splitlines()
    # Table replacement
    table_start = None
    table_end = None
    for i in range(len(lines) - 1):
        if re.match(r"^\|\s*项目\s*\|\s*链接\s*\|\s*$", lines[i]) and re.match(
            r"^\|\s*-+\s*\|\s*-+\s*\|\s*$", lines[i + 1]
        ):
            table_start = i
            j = i + 2
            while j < len(lines) and lines[j].strip().startswith("|"):
                j += 1
            table_end = j
            break
    if table_start is not None and table_end is not None:
        new_lines = lines[:table_start] + group_block.splitlines() + lines[table_end:]
        return "\n".join(new_lines)
    # Bullet replacement
    cb = r"(?:\[[ xX]\]\s*)?"
    labels_patterns = [
        rf"-\s*{cb}申请前截图（页面）[:：]\s*\S+",
        rf"-\s*{cb}审核弹窗截图[:：]\s*\S+",
        rf"-\s*{cb}Playwright\s*Trace\s*压缩包[:：]\s*\S+",
        rf"-\s*{cb}Trace\s*Report\s*HTML[:：]\s*\S+",
    ]
    indices = []
    for i, line in enumerate(lines):
        for pat in labels_patterns:
            if re.search(pat, line):
                indices.append(i)
                break
    if indices:
        # 计算替换范围：自上而下尽可能包含“截图/Trace”标题、相关项目符号与相邻空行
        start = min(indices)
        end = max(indices)
        bullet_union = re.compile(r"(?:" + r")|(?:".join(labels_patterns) + r")")
        # 向上扩展包含上一行标题或空行（最多各一次）
        up_steps = 0
        while start - 1 >= 0 and up_steps < 2:
            prev = lines[start - 1].strip()
            if prev in ("截图", "Trace", ""):
                start -= 1
                up_steps += 1
            else:
                break
        # 向下扩展，直到遇到非标题且非项目符号且非空行
        while end + 1 < len(lines):
            nxt = lines[end + 1]
            if nxt.strip() in ("截图", "Trace", "") or bullet_union.search(nxt):
                end += 1
            else:
                break
        new_lines = []
        new_lines.extend(lines[:start])
        new_lines.extend(group_block.splitlines())
        new_lines.extend(lines[end + 1 :])
        return "\n".join(new_lines)
    # Header insertion
    hdr = re.search(r"(?m)^.*验证证据.*$", body)
    if hdr:
        pos = body.find(hdr.group(0))
        before = body[:pos]
        after = body[pos:]
        after_lines = after.splitlines()
        if after_lines:
            after_lines.insert(1, group_block)
            return before + "\n".join(after_lines)
        else:
            return group_block + "\n\n" + body
    return group_block + "\n\n" + body


def patch_issue_body(owner, repo, pr, token, new_body):
    url = f"{API}/repos/{owner}/{repo}/issues/{pr}"
    payload = json.dumps({"body": new_body}, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=payload,
        method="PATCH",
        headers={
            "Authorization": f"token {token}",
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json; charset=utf-8",
            "User-Agent": "evidence-rewriter",
        },
    )
    with urllib.request.urlopen(req) as resp:
        data = json.load(resp)
    with open(
        os.path.join(LOG_DIR, f"pr{pr}_patch_groups_resp.json"), "w", encoding="utf-8"
    ) as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    print(data.get("updated_at", ""))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--owner", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--pr", type=int, required=True)
    parser.add_argument(
        "--checkbox", choices=["checked", "unchecked"], default="checked"
    )
    parser.add_argument("--urls-file", default=os.path.join(LOG_DIR, "e2e_urls.json"))
    parser.add_argument(
        "--debug-diff",
        action="store_true",
        help="Print unified diff when change detected",
    )
    args = parser.parse_args()

    token = read_token()
    body = fetch_issue_body(args.owner, args.repo, args.pr, token)

    urls = None
    if os.path.exists(args.urls_file):
        try:
            with open(args.urls_file, "r", encoding="utf-8") as f:
                urls = json.load(f)
        except Exception:
            urls = None
    if not urls or len(urls) < 4:
        urls = extract_urls_from_body(body)

    group_block = build_group_block(urls, args.checkbox)
    new_body = replace_block(body, group_block)

    os.makedirs(LOG_DIR, exist_ok=True)
    with open(
        os.path.join(LOG_DIR, f"pr{args.pr}_body_groups.md"), "w", encoding="utf-8"
    ) as f:
        f.write(new_body)

    # 正规化函数：统一换行、移除行尾空白、规范中文冒号后的空格、压缩多余空行
    def _normalize(txt: str) -> str:
        txt = txt.replace("\r\n", "\n")
        lines = [re.sub(r"[ \t]+$", "", ln) for ln in txt.split("\n")]
        lines = [re.sub(r"：\s*", "： ", ln) for ln in lines]
        out = []
        blank = False
        for ln in lines:
            if ln.strip() == "":
                if not blank:
                    out.append("")
                blank = True
            else:
                out.append(ln)
                blank = False
        return "\n".join(out).strip()

    norm_new = _normalize(new_body)
    norm_old = _normalize(body)

    # 幂等：若正规化后无变化则跳过 PATCH，减少无谓编辑与速率消耗
    if norm_new == norm_old:
        print("no-op: evidence block already up-to-date")
        return

    if args.debug_diff:
        diff = difflib.unified_diff(
            norm_old.splitlines(),
            norm_new.splitlines(),
            fromfile="before",
            tofile="after",
            lineterm="",
        )
        print("\n".join(diff))

    patch_issue_body(args.owner, args.repo, args.pr, token, new_body)


if __name__ == "__main__":
    main()
