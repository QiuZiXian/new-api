#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Seedance 可用性 / 计量计费 / 折扣准确性 探测脚本

三层核对：
1. 可用性：提交 -> 轮询终态 -> 产物可下载
2. 计量一致性：平台 quota 是否等于 平台 tokens × modelRatio × 分组倍率（应为 0 偏差）
3. 计量口径：平台 tokens 是否接近官方公式 (秒 × 宽 × 高 × 24 / 1024)
   以及最终金额是否接近定价表折算价

注意：异步任务的消费日志是「预扣 + 差额补扣」两条，必须按 task_id 求和才是真实扣费。

用法：
    python seedance_billing_probe.py --dry-run
    python seedance_billing_probe.py --case C1
    python seedance_billing_probe.py --reconcile task_xxx --case C1
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
    except Exception:
        pass

HERE = os.path.dirname(os.path.abspath(__file__))
PASS, FAIL, WARN, SKIP = "PASS", "FAIL", "WARN", "SKIP"

# 终态：平台对外文档写 succeeded，实际返回 completed，两者都算成功
SUCCESS_STATES = {"succeeded", "completed", "success"}
FAIL_STATES = {"failed", "cancelled", "expired", "canceled"}

# ---------------------------------------------------------------- 定价表
# 元 / 百万 token，来自 dev-docs/doubao-seedance-2.0定价.md
PRICE_TABLE = {
    "doubao-seedance-2.0": {
        "base": 46.0,
        "tiers": {
            ("720p", False): 46.0, ("480p", False): 46.0,
            ("720p", True): 28.0, ("480p", True): 28.0,
            ("1080p", False): 51.0, ("1080p", True): 31.0,
            ("4k", False): 26.0, ("4k", True): 16.0,
        },
    },
    "doubao-seedance-2.0-fast": {
        "base": 37.0,
        "tiers": {
            ("720p", False): 37.0, ("480p", False): 37.0,
            ("720p", True): 22.0, ("480p", True): 22.0,
        },
    },
    "doubao-seedance-2.0-mini": {
        "base": 23.0,
        "tiers": {
            ("720p", False): 23.0, ("480p", False): 23.0,
            ("720p", True): 14.0, ("480p", True): 14.0,
        },
    },
    "doubao-seedance-2.5": {
        "base": 85.75,
        "tiers": {
            ("720p", False): 85.75, ("480p", False): 85.75,
            ("720p", True): 51.45, ("480p", True): 51.45,
        },
    },
}

BASE_DIMS = {"480p": (864, 480), "720p": (1280, 720), "1080p": (1920, 1080), "4k": (3840, 2160)}
FPS = 24
TOKEN_DIVISOR = 1024
QUOTA_PER_UNIT = 500000.0  # /api/status 实测，汇率 1


def output_dims(resolution, ratio):
    bw, bh = BASE_DIMS[resolution.lower()]
    if not ratio or ratio.lower() in ("adaptive", "auto"):
        return bw, bh
    try:
        w, h = [float(x) for x in ratio.split(":")]
    except Exception:
        return bw, bh
    if w <= 0 or h <= 0:
        return bw, bh
    if w >= h:
        return int(round(bh * w / h)), bh
    return bh, int(round(bh * h / w))


def estimate_tokens(resolution, ratio, seconds):
    w, h = output_dims(resolution, ratio)
    return int(round(seconds * w * h * FPS / TOKEN_DIVISOR))


def tier_price(model, resolution, has_video):
    table = PRICE_TABLE.get(model)
    return table["tiers"].get((resolution.lower(), bool(has_video))) if table else None


# ---------------------------------------------------------------- HTTP
class Client:
    def __init__(self, base_url, token, timeout=60):
        self.base = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout

    def call(self, method, path, body=None, params=None, raw=False, timeout=None):
        url = self.base + path
        if params:
            url += "?" + urllib.parse.urlencode(params)
        headers = {"Authorization": "Bearer " + self.token, "Accept": "application/json"}
        data = None
        if body is not None:
            data = json.dumps(body).encode()
            headers["Content-Type"] = "application/json"
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=timeout or self.timeout) as resp:
                payload = resp.read()
                return resp.status, dict(resp.headers), (payload if raw else payload.decode("utf-8", "replace"))
        except urllib.error.HTTPError as e:
            payload = e.read()
            return e.code, dict(e.headers), (payload if raw else payload.decode("utf-8", "replace"))
        except Exception as e:
            return 0, {}, str(e)

    def json(self, method, path, **kw):
        status, headers, text = self.call(method, path, **kw)
        try:
            return status, headers, json.loads(text)
        except Exception:
            return status, headers, {"_raw": text}


def other_ratios_header(headers):
    for k, v in (headers or {}).items():
        if k.lower().endswith("-other-ratios"):
            return k, v
    return None, None


# ---------------------------------------------------------------- 任务
def submit(client, case):
    body = {
        "model": case["model"],
        "prompt": case.get("prompt", "一只橘猫在窗台上打哈欠，阳光斜照"),
        "resolution": case.get("resolution", "720p"),
        "ratio": case.get("ratio", "16:9"),
        "duration": case.get("duration", 5),
    }
    if case.get("content"):
        body["content"] = case["content"]
    if case.get("metadata"):
        body["metadata"] = case["metadata"]
    return client.json("POST", "/v1/video/generations", body=body)


def get_task(client, task_id):
    return client.json("GET", "/v1/videos/" + urllib.parse.quote(task_id))


def poll(client, task_id, interval, timeout):
    started = time.time()
    last = {}
    while time.time() - started < timeout:
        status, _, data = get_task(client, task_id)
        last = data if isinstance(data, dict) else {"_raw": data}
        state = str(last.get("status", "")).lower()
        if state in SUCCESS_STATES or state in FAIL_STATES:
            return state, last, time.time() - started
        time.sleep(interval)
    return "timeout", last, time.time() - started


def check_content(client, task_id):
    status, headers, payload = client.call(
        "GET", "/v1/videos/%s/content" % urllib.parse.quote(task_id), raw=True, timeout=180)
    ctype = ""
    for k, v in headers.items():
        if k.lower() == "content-type":
            ctype = v
    return status == 200 and bool(payload) and ("video" in ctype or len(payload) > 1024)


def fetch_logs(client, limit=100):
    for path in ("/api/log/token", "/api/log/self"):
        status, _, data = client.json("GET", path, params={"p": 1, "size": limit})
        if status == 200 and isinstance(data, dict):
            rows = data.get("data") or data.get("items") or []
            if isinstance(rows, list):
                return rows
    return []


def logs_for_task(rows, task_id, model, since_ts):
    """返回该任务的全部消费日志（预扣 + 差额），按时间升序。"""
    hit, fuzzy = [], []
    for row in rows:
        if row.get("model_name") != model:
            continue
        other = row.get("other") or {}
        if isinstance(other, str):
            try:
                other = json.loads(other)
            except Exception:
                other = {}
        same = (other.get("task_id") == task_id) or (row.get("task_id") == task_id)
        if same:
            hit.append(row)
            continue
        created = row.get("created_at") or 0
        if since_ts and created and created >= since_ts - 60:
            fuzzy.append(row)
    rows = hit or fuzzy
    return sorted(rows, key=lambda r: r.get("created_at") or 0)


def analyze_logs(rows):
    """type=2 消费、type=6 退款，净扣 = 消费 - 退款。"""
    consume = 0
    refund = 0
    prompt = 0
    completion = 0
    ratios = {}
    for r in rows:
        q = int(r.get("quota") or 0)
        if r.get("type") == 6:
            refund += q
        elif r.get("type") == 2:
            consume += q
        prompt += int(r.get("prompt_tokens") or 0)
        completion += int(r.get("completion_tokens") or 0)
        other = r.get("other") or {}
        if isinstance(other, str):
            try:
                other = json.loads(other)
            except Exception:
                other = {}
        for key in ("model_ratio", "group_ratio", "user_group_ratio", "pre_consumed_quota", "actual_quota"):
            if other.get(key) is not None:
                ratios[key] = other[key]
    return {"consume": consume, "refund": refund, "net": consume - refund,
            "prompt": prompt, "completion": completion, "ratios": ratios}


# ---------------------------------------------------------------- 结论
def build_record(case, account, cfg):
    return {
        "case": case["id"], "account": account["label"], "group": account.get("group", ""),
        "model": case["model"], "resolution": case.get("resolution", "720p"),
        "ratio": case.get("ratio", "16:9"), "duration": case.get("duration", 5),
        "has_video": bool(case.get("has_video")), "status": "INIT",
    }


def finish(rec, cfg):
    """填预期值并给出四层结论：时长透传 / 平台自洽 / 计量口径 / 金额。"""
    price = tier_price(rec["model"], rec["resolution"], rec["has_video"])
    group_ratio = cfg["group_ratios"].get(rec.get("group", ""), 1.0)
    rec["tier_price"] = price
    rec["group_ratio"] = group_ratio

    tokens = rec.get("actual_tokens") or 0
    quota = rec.get("net_quota") or 0
    seconds = rec.get("video_seconds")
    basis = seconds or rec.get("duration")
    formula = estimate_tokens(rec["resolution"], rec["ratio"], basis) if basis else 0
    rec["formula_tokens"] = formula
    rec["formula_basis_seconds"] = basis

    if seconds:
        rec["duration_dev"] = round(seconds - rec["duration"], 3)
        rec["duration_ok"] = abs(rec["duration_dev"]) <= max(0.3, rec["duration"] * 0.05)

    if price and tokens:
        rec["expected_quota"] = round(tokens * price * 0.5 * group_ratio, 1)
    if tokens and formula:
        rec["token_dev"] = round((tokens - formula) / formula, 4)
    if quota:
        rec["amount_cny"] = round(quota / QUOTA_PER_UNIT, 4)
    if price and formula:
        rec["amount_expected"] = round(formula * price / 1e6 * group_ratio, 4)
        if rec.get("amount_cny"):
            rec["amount_dev"] = round((rec["amount_cny"] - rec["amount_expected"]) / rec["amount_expected"], 4)

    tol = cfg["tolerance"]
    model_ratio = rec.get("log_model_ratio")
    if quota and tokens and model_ratio:
        recomputed = tokens * float(model_ratio) * float(rec.get("log_group_ratio") or group_ratio)
        rec["platform_recompute"] = round(recomputed, 1)
        rec["platform_dev"] = round((quota - recomputed) / recomputed, 4) if recomputed else None

    rec["token_ok"] = rec.get("token_dev") is not None and abs(rec["token_dev"]) <= tol["tokens"]
    rec["amount_ok"] = rec.get("amount_dev") is not None and abs(rec["amount_dev"]) <= tol["quota"]

    if rec.get("state") not in SUCCESS_STATES:
        rec["status"] = rec.get("status") if rec["status"] in (FAIL, SKIP) else FAIL
    elif rec.get("duration_ok") is False:
        rec["status"] = FAIL  # 请求时长未被正确执行
    elif not rec["token_ok"]:
        rec["status"] = FAIL
    elif not rec["amount_ok"]:
        rec["status"] = FAIL
    elif rec.get("content_ok") is False:
        rec["status"] = WARN
    else:
        rec["status"] = PASS
    return rec


def run_case(client, account, case, cfg):
    rec = build_record(case, account, cfg)
    rec["expected_tokens"] = estimate_tokens(rec["resolution"], rec["ratio"], rec["duration"])

    started = int(time.time())
    status, headers, data = submit(client, case)
    rec["submit_http"] = status
    _, rec["ratios_header"] = other_ratios_header(headers)
    if status not in (200, 201):
        rec["status"] = FAIL
        rec["reason"] = "提交失败: %s" % json.dumps(data, ensure_ascii=False)[:400]
        return rec

    task_id = data.get("id") or data.get("task_id")
    rec["task_id"] = task_id
    if not task_id:
        rec["status"] = FAIL
        rec["reason"] = "响应无任务 ID: %s" % json.dumps(data, ensure_ascii=False)[:400]
        return rec

    state, task, elapsed = poll(client, task_id, cfg["poll_interval"], cfg["timeout"])
    rec.update({"state": state, "elapsed_sec": round(elapsed, 1)})
    if state not in SUCCESS_STATES:
        rec["status"] = FAIL if state in FAIL_STATES else SKIP
        rec["reason"] = "终态 %s: %s" % (state, json.dumps(task.get("error"), ensure_ascii=False))
        return rec

    rec["content_ok"] = check_content(client, task_id)
    rec["video_seconds"] = mp4_seconds(client, task_id)

    rows = []
    for _ in range(cfg["log_retries"]):
        time.sleep(cfg["log_wait"])
        rows = logs_for_task(fetch_logs(client), task_id, rec["model"], started)
        if rows:
            break
    absorb_logs(rec, rows, task)
    return finish(rec, cfg)


def reconcile(client, account, case, task_id, cfg):
    """对已存在的任务补做对账，不重新提交（省钱）。"""
    rec = build_record(case, account, cfg)
    rec["expected_tokens"] = estimate_tokens(rec["resolution"], rec["ratio"], rec["duration"])
    rec["task_id"] = task_id
    status, _, task = get_task(client, task_id)
    rec["submit_http"] = status
    rec["state"] = str(task.get("status", "")).lower()
    rec["content_ok"] = check_content(client, task_id)
    rec["video_seconds"] = mp4_seconds(client, task_id)
    rows = logs_for_task(fetch_logs(client), task_id, rec["model"], 0)
    absorb_logs(rec, rows, task)
    return finish(rec, cfg)


def absorb_logs(rec, rows, task):
    agg = analyze_logs(rows)
    rec["log_lines"] = len(rows)
    rec["consume_quota"] = agg["consume"]
    rec["refund_quota"] = agg["refund"]
    rec["net_quota"] = agg["net"] or None
    rec["total_quota"] = rec["net_quota"]
    rec["pre_consumed_quota"] = agg["ratios"].get("pre_consumed_quota")
    rec["actual_quota_field"] = agg["ratios"].get("actual_quota")
    rec["log_model_ratio"] = agg["ratios"].get("model_ratio")
    rec["log_group_ratio"] = agg["ratios"].get("group_ratio")
    rec["log_content"] = [r.get("content") for r in rows]
    rec["log_types"] = [r.get("type") for r in rows]

    prompt, completion = agg["prompt"], agg["completion"]
    tokens = completion or prompt
    if tokens <= 0:
        usage = (task or {}).get("usage") or {}
        tokens = int(usage.get("total_tokens") or 0) or (
            int(usage.get("prompt_tokens") or 0) + int(usage.get("completion_tokens") or 0))
        rec["tokens_source"] = "task.usage"
    else:
        rec["tokens_source"] = "log"
    rec["actual_tokens"] = tokens
    rec["log_prompt_tokens"] = prompt
    rec["log_completion_tokens"] = completion


def mp4_seconds(client, task_id):
    """从 mp4 mvhd 原子读真实时长，作为时长证据。"""
    status, _, payload = client.call(
        "GET", "/v1/videos/%s/content" % urllib.parse.quote(task_id), raw=True, timeout=180)
    if status != 200 or not isinstance(payload, (bytes, bytearray)):
        return None
    i = payload.find(b"mvhd")
    if i < 0:
        return None
    try:
        version = payload[i + 4]
        off = i + 8
        if version == 1:
            off += 16
            timescale = int.from_bytes(payload[off:off + 4], "big")
            duration = int.from_bytes(payload[off + 4:off + 12], "big")
        else:
            off += 8
            timescale = int.from_bytes(payload[off:off + 4], "big")
            duration = int.from_bytes(payload[off + 4:off + 8], "big")
        return round(duration / timescale, 3) if timescale else None
    except Exception:
        return None


# ---------------------------------------------------------------- 报告
def pct(v):
    return "-" if v is None else "%+.2f%%" % (v * 100)


def money(quota):
    return "-" if quota is None else "¥%.4f" % (float(quota) / QUOTA_PER_UNIT)


def write_report(records, cfg, path):
    lines = [
        "# Seedance 计费探测报告",
        "",
        "- 环境：%s" % cfg["base_url"],
        "- 时间：%s" % datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "- 换算式：`¥ = quota / 500000`",
        "- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。",
        "",
        "| 用例 | 分组 | 模型 | 档位 | 请求时长 | 成片时长 | 时长偏差 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 净扣 quota | 金额 | 金额偏差 | 结论 |",
        "|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|",
    ]
    for r in records:
        lines.append("| %s | %s | %s | %s %s | %ss | %ss | %ss | %s | %s | %s | %s | %s | %s | %s | %s |" % (
            r["case"], r.get("group", ""), r["model"], r["resolution"], r["ratio"],
            r.get("duration", "-"), r.get("video_seconds", "-"), r.get("duration_dev", "-"),
            r.get("formula_tokens", "-"), r.get("actual_tokens", "-"), pct(r.get("token_dev")),
            r.get("tier_price", "-"), r.get("net_quota", "-"), money(r.get("net_quota")),
            pct(r.get("amount_dev")), r.get("status")))
    lines += ["", "## 明细", ""]
    for r in records:
        lines += ["### %s @ %s" % (r["case"], r["account"]), "```json",
                  json.dumps(r, ensure_ascii=False, indent=2), "```", ""]
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))
    return path


# ---------------------------------------------------------------- main
def load_cfg(path):
    with open(path, encoding="utf-8") as f:
        cfg = json.load(f)
    cfg.setdefault("poll_interval", 5)
    cfg.setdefault("timeout", 600)
    cfg.setdefault("log_wait", 5)
    cfg.setdefault("log_retries", 4)
    cfg.setdefault("tolerance", {"quota": 0.01, "tokens": 0.05, "discount": 0.005})
    cfg.setdefault("group_ratios", {"默认": 1.0, "9.9折扣": 0.99, "9折区": 0.9,
                                    "8.5折区": 0.85, "8折区": 0.8, "7.5折区": 0.75})
    return cfg


def find_case(cfg, case_id):
    for c in cfg["cases"]:
        if c["id"] == case_id:
            return c
    raise SystemExit("配置里没有用例 %s" % case_id)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--config", default=os.path.join(HERE, "config.json"))
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--case", action="append", default=[])
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--reconcile", metavar="TASK_ID")
    ap.add_argument("--report", default=os.path.join(HERE, "seedance-probe-report.md"))
    args = ap.parse_args()

    if not os.path.exists(args.config):
        print("缺少配置：%s" % args.config)
        return 2
    cfg = load_cfg(args.config)

    if args.reconcile:
        if not args.case:
            raise SystemExit("--reconcile 需要配合 --case <ID> 指定档位参数")
        acct = cfg["accounts"][0]
        rec = reconcile(Client(cfg["base_url"], acct["token"]), acct, find_case(cfg, args.case[0]),
                        args.reconcile, cfg)
        print(write_report([rec], cfg, args.report))
        return 0 if rec["status"] in (PASS, WARN, SKIP) else 1

    cases = [c for c in cfg["cases"] if args.all or not args.case or c["id"] in args.case]

    if args.dry_run:
        for c in cases:
            t = estimate_tokens(c.get("resolution", "720p"), c.get("ratio", "16:9"), c.get("duration", 5))
            p = tier_price(c["model"], c.get("resolution", "720p"), c.get("has_video"))
            print("%-4s %-26s %-6s %ss tokens=%-7d tier=%-6s 预期金额=%s" % (
                c["id"], c["model"], c.get("resolution", "720p"), c.get("duration", 5),
                t, p, money(t * p * 0.5 if p else None)))
        return 0

    records = []
    for acct in cfg["accounts"]:
        client = Client(cfg["base_url"], acct["token"])
        for c in cases:
            print(">> %s @ %s ..." % (c["id"], acct["label"]), flush=True)
            rec = run_case(client, acct, c, cfg)
            print("   %s tokens=%s quota=%s dev=%s" % (
                rec.get("status"), rec.get("actual_tokens"), rec.get("total_quota"),
                pct(rec.get("amount_dev"))), flush=True)
            records.append(rec)

    print("报告：%s" % write_report(records, cfg, args.report))
    return 0 if all(r["status"] in (PASS, WARN, SKIP) for r in records) else 1


if __name__ == "__main__":
    sys.exit(main())
