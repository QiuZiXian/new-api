#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
new-api 仓山 / cii-group 渠道 —— 端到端自动化测试 + 计量计费核对脚本

设计目标
--------
1. 功能回归：覆盖 apifox 实测文档(dev-docs/AI模型接口.md)里出现过的全部对外端点，
   同时也测源码 router/video-router.go 里注册但实测没覆盖到的别名路由。
2. 计费核对：用「模型定价表(tests/e2e/pricing.json)」独立复算一遍预期费用，
   再和 new-api 日志里该次请求真实扣的 quota 比对，超容差就报出来。
   这是本脚本最有价值的部分——它能在 ModelRatio 配错、倍率漏配、
   异步任务按实际时长二次结算算错的时候第一时间发现。
3. 响应格式兼容：实测文档里的响应是「上游方舟透传」形态
   ({"ResponseMetadata":{...},"Result":{"Id":"..."}})，而当前源码 asset.go
   已经改成本地归一化形态({"id":"...","items":[...]})。两种都兼容解析，
   并记录下实际命中哪一种，方便你确认部署的版本。

用法
----
    # 1) 复制配置并填写
    cp config.example.json config.json   # Windows: copy
    # 编辑 config.json: base_url / token / quota_per_unit / usd_exchange_rate

    # 2) 跑冒烟(不花钱的只读用例)
    python newapi_e2e.py --smoke

    # 3) 跑全量(含会真实扣费的视频/字幕擦除用例)
    python newapi_e2e.py

    # 4) 只跑计费核对
    python newapi_e2e.py --billing-only

    # 5) 输出报告
    python newapi_e2e.py --report report.md

退出码: 0=全部通过  1=有 FAIL  2=配置/环境错误
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

# Windows 控制台中文输出
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

HERE = os.path.dirname(os.path.abspath(__file__))

PASS, FAIL, WARN, SKIP = "PASS", "FAIL", "WARN", "SKIP"

# 终端配色(Windows 10+ 支持 ANSI)
_COLOR = {PASS: "\033[92m", FAIL: "\033[91m", WARN: "\033[93m", SKIP: "\033[90m"}
_RESET = "\033[0m"


def _c(status, text):
    if not sys.stdout.isatty():
        return text
    return f"{_COLOR.get(status, '')}{text}{_RESET}"


# ============================================================
# HTTP Client
# ============================================================
class ApiError(Exception):
    def __init__(self, status, body, url=""):
        super().__init__(f"HTTP {status} {url}: {str(body)[:300]}")
        self.status = status
        self.body = body


class ApiClient:
    """极简 HTTP client。只用标准库，避免 requests 依赖问题。"""

    def __init__(self, base_url, token, timeout=60):
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout
        self.calls = []  # 记录所有调用，供报告用

    def _request(self, method, path, body=None, params=None, raw=False):
        url = self.base_url + path
        if params:
            url += "?" + urllib.parse.urlencode({k: v for k, v in params.items() if v is not None})

        data = None
        headers = {"Authorization": f"Bearer {self.token}"}
        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"
        headers["Accept"] = "application/json"

        started = time.time()
        # new-api 在鉴权失败时可能没读完请求体就回 401 并关连接，
        # 客户端随后发 body 会撞 RST(WinError 10054)。这类瞬时重置重试一次即可。
        for attempt in range(3):
            req = urllib.request.Request(url, data=data, headers=headers, method=method)
            try:
                with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                    raw_body = resp.read().decode("utf-8", errors="replace")
                    status = resp.status
                break
            except urllib.error.HTTPError as e:
                # 状态码已经拿到了，body 拿不到就退化成空串，不能因此把用例判成异常。
                status = e.code
                try:
                    raw_body = e.read().decode("utf-8", errors="replace")
                except Exception:
                    raw_body = ""
                break
            except (ConnectionResetError, ConnectionAbortedError,
                    http.client.RemoteDisconnected) as e:
                if attempt < 2:
                    time.sleep(0.6 * (attempt + 1))
                    continue
                self.calls.append((method, url, 0, str(e)))
                raise ApiError(0, str(e), url)
            except Exception as e:
                self.calls.append((method, url, 0, str(e)))
                raise ApiError(0, str(e), url)

        elapsed = time.time() - started
        self.calls.append((method, url, status, elapsed))
        if raw:
            return status, raw_body
        try:
            return status, json.loads(raw_body) if raw_body.strip() else {}
        except json.JSONDecodeError:
            return status, {"_raw": raw_body}

    def get(self, path, params=None, raw=False):
        return self._request("GET", path, params=params, raw=raw)

    def post(self, path, body=None, raw=False):
        return self._request("POST", path, body=body, raw=raw)

    def put(self, path, body=None, raw=False):
        return self._request("PUT", path, body=body, raw=raw)

    def delete(self, path, body=None, raw=False):
        return self._request("DELETE", path, body=body, raw=raw)


# ============================================================
# 响应解析 —— 兼容「上游透传」与「本地归一化」两种形态
# ============================================================
def pick(d, *keys, default=None):
    """按多个候选 key 取值，兼容 PascalCase / snake_case / 包裹层。"""
    if not isinstance(d, dict):
        return default
    for k in keys:
        if k in d and d[k] not in (None, ""):
            return d[k]
    return default


def unwrap(resp):
    """
    把响应统一成"业务对象"。
    上游透传形态: {"ResponseMetadata":{...},"Result":{...}}  -> Result
    本地归一化形态: {"id":...,"items":[...],...}             -> 原样
    """
    if not isinstance(resp, dict):
        return resp
    if "Result" in resp:
        return resp["Result"]
    if "result" in resp:
        return resp["result"]
    return resp


def extract_id(resp):
    """从创建类响应里提取 id，兼容两种形态 + 各种字段名。"""
    obj = unwrap(resp)
    if isinstance(obj, dict):
        return pick(obj, "id", "Id", "ID", "asset_id", "group_id", "public_id", default="")
    return ""


def extract_items(resp):
    """从列表响应里提取 items 数组，兼容两种形态。"""
    obj = unwrap(resp)
    if isinstance(obj, dict):
        items = pick(obj, "items", "Items", default=None)
        if items is None:
            # 单个对象也当作只有一项的列表
            return [obj] if extract_id(obj) else []
        return items if isinstance(items, list) else []
    return []


def extract_total(resp):
    obj = unwrap(resp)
    if isinstance(obj, dict):
        return pick(obj, "total", "Total", "TotalCount", "total_count", default=None)
    return None


def detect_response_shape(resp):
    """判断响应是上游透传还是本地归一化，报告里会标注。"""
    if isinstance(resp, dict) and ("Result" in resp or "ResponseMetadata" in resp):
        return "upstream-passthrough(方舟原样)"
    return "local-normalized(new-api本地)"


# ============================================================
# 定价表 / 计费换算
# ============================================================
class Pricing:
    def __init__(self, data, quota_per_unit, usd_exchange_rate):
        self.d = data
        self.quota_per_unit = quota_per_unit
        self.usd_rate = usd_exchange_rate

    def cny_to_quota(self, cny):
        """
        按 controller/billing.go 的公式反推:
            人民币 = quota / QuotaPerUnit * USDExchangeRate
        =>  quota = 人民币 * QuotaPerUnit / USDExchangeRate
        """
        return cny * self.quota_per_unit / self.usd_rate

    def quota_per_cny(self):
        """1 元等于多少 quota"""
        return self.quota_per_unit / self.usd_rate

    # ---- 视频：元/秒 ----
    def video_expected_cny(self, model, resolution, duration_sec):
        tbl = self.d.get("video_per_second", {})
        row = tbl.get(model)
        if not row:
            return None, f"定价表缺少 video_per_second.{model}"
        # 分辨率没配就退到最接近的档
        if resolution in row:
            per_sec = row[resolution]
        else:
            order = ["480p", "720p", "1080p", "4k"]
            cand = [r for r in order if r in row]
            if not cand:
                return None, f"定价表 {model} 无任何分辨率档"
            per_sec = row[cand[0]]
            return per_sec * duration_sec, (
                f"定价表无 {resolution} 档，回退用 {cand[0]}({per_sec}元/秒)，"
                f"结果为粗略参考"
            )
        return per_sec * duration_sec, None

    # ---- 视频：元/千 token（1.5-pro 特殊）----
    def video_token_expected_cny(self, model, tokens, with_audio=True):
        tbl = self.d.get("video_per_ktok", {})
        row = tbl.get(model)
        if not row:
            return None, f"定价表缺少 video_per_ktok.{model}"
        key = "with_audio" if with_audio else "no_audio"
        per_ktok = row.get(key)
        if per_ktok is None:
            return None, f"定价表 {model} 缺 {key}"
        return per_ktok * (tokens / 1000.0), None

    # ---- 字幕擦除：元/秒 ----
    def subtitle_expected_cny(self, model, seconds, vendor="volcengine_standard"):
        tbl = self.d.get("subtitle_erase_per_second", {})
        row = tbl.get(model)
        if not row:
            return None, f"定价表缺少 subtitle_erase_per_second.{model}"
        per_sec = row.get(vendor)
        if per_sec is None:
            return None, f"定价表 {model} 缺 {vendor}"
        return per_sec * seconds, None

    # ---- 图片：元/张 ----
    def image_expected_cny(self, model, n, first_free=True):
        tbl = self.d.get("image_per_image", {})
        row = tbl.get(model)
        if row is None:
            return None, f"定价表缺少 image_per_image.{model}"
        if isinstance(row, dict):
            # 按像素档位，默认取低价档
            per_img = min(v for v in row.values() if isinstance(v, (int, float)))
        else:
            per_img = row
        if first_free and n >= 1:
            billable = max(0, n - 1)  # 首张免费
            return per_img * billable, "已按「首张免费」折算"
        return per_img * n, None

    # ---- 文本：元/百万 token ----
    def text_expected_cny(self, model, prompt_tok, completion_tok, cache_tok=0):
        cny_tbl = self.d.get("text_per_mtok_cny", {})
        usd_tbl = self.d.get("text_per_mtok_usd", {})
        peak = False
        if model in cny_tbl:
            row = cny_tbl[model]
            peak = bool(row.get("_peak", False))
        elif model in usd_tbl:
            row = usd_tbl[model]
            # USD 单价 * 汇率 => 元
            rate = self.usd_rate
            cost = ((prompt_tok * row["input"] + completion_tok * row["output"]
                     + cache_tok * row["cache"]) / 1_000_000.0) * rate
            return cost, None
        else:
            return None, f"定价表缺少文本模型 {model}"

        cost = (prompt_tok * row["input"] + completion_tok * row["output"]
                + cache_tok * row["cache"]) / 1_000_000.0
        note = None
        if peak:
            ratio = self.d.get("peak_offpeak_ratio", 0.5)
            note = (f"{model} 为峰谷计费，此处按高峰期算；"
                    f"空闲期约为其 {ratio} 倍 => 预期区间 "
                    f"[{cost*ratio:.6f}, {cost:.6f}] 元")
        return cost, note

    # ---- 按次 ----
    def per_call_expected_cny(self, model):
        tbl = self.d.get("per_call_cny", {})
        if model not in tbl:
            return None, f"定价表缺少 per_call_cny.{model}"
        return tbl[model], None


# ============================================================
# 计费核对器
# ============================================================
class QuotaAuditor:
    """
    从 new-api 日志里读该 token 的真实扣费，与定价表复算值比对。

    日志来源: GET /api/log/token  (middleware.TokenAuthReadOnly —— 用同一个
    sk-token 即可访问，不需要 session 登录)。日志字段见 model/log.go:
    quota / model_name / prompt_tokens / completion_tokens / use_time。
    """

    def __init__(self, client, pricing, tolerance=0.25):
        self.client = client
        self.pricing = pricing
        self.tolerance = tolerance

    def fetch_recent_logs(self, size=30):
        status, body = self.client.get("/api/log/token", params={"p": 1, "size": size})
        if status != 200:
            return None, f"拉取 /api/log/token 失败: HTTP {status} {str(body)[:200]}"
        obj = unwrap(body)
        if isinstance(obj, dict):
            items = pick(obj, "items", "data", "logs", default=None)
            if items is None:
                # 也许直接是 {data:[...]}
                for v in obj.values():
                    if isinstance(v, list):
                        items = v
                        break
            return (items or []), None
        if isinstance(obj, list):
            return obj, None
        return None, f"无法解析日志响应: {str(body)[:200]}"

    @staticmethod
    def _log_matches(log, model, after_ts):
        if log.get("model_name") != model:
            return False
        created = log.get("created_at") or log.get("created_time") or 0
        # created_at 可能是秒或毫秒
        if isinstance(created, (int, float)) and created > 1e11:
            created = created / 1000.0
        return created >= after_ts - 5

    def find_charge(self, model, after_ts, max_wait=90, poll=5):
        """轮询日志，等这条请求的扣费记录落库。"""
        deadline = time.time() + max_wait
        last_err = None
        while time.time() < deadline:
            logs, err = self.fetch_recent_logs()
            if err:
                last_err = err
                time.sleep(poll)
                continue
            # 同一个任务会有两条日志：预扣 + 结算。
            # 只有结算那条带 actual_quota（真实总消耗），优先取它。
            fallback = None
            for log in logs:
                if not self._log_matches(log, model, after_ts):
                    continue
                if self.parse_other(log).get("actual_quota") is not None:
                    return log, None
                if fallback is None:
                    fallback = log
            if fallback is not None:
                return fallback, None
            time.sleep(poll)
        return None, last_err or f"等待 {max_wait}s 未找到 {model} 的扣费日志"

    @staticmethod
    def parse_other(log):
        """日志的 other 字段是内嵌 JSON 字符串，里面有 actual_quota / model_ratio 等。"""
        raw = log.get("other")
        if isinstance(raw, dict):
            return raw
        if isinstance(raw, str) and raw.strip():
            try:
                parsed = json.loads(raw)
                return parsed if isinstance(parsed, dict) else {}
            except ValueError:
                return {}
        return {}

    @classmethod
    def actual_quota(cls, log):
        """
        任务型请求(视频/字幕擦除)分两步扣费：提交时预扣 pre_consumed_quota，
        完成后按实际量结算。日志的 quota 字段记的是**结算差额**
        (actual_quota - pre_consumed_quota)，不是本次总消耗。
        真实总消耗在 other.actual_quota 里，核对计费必须用它。
        """
        actual = cls.parse_other(log).get("actual_quota")
        if actual is not None:
            return int(actual)
        return int(log.get("quota") or 0)

    def audit(self, label, model, expected_cny, after_ts, note=None):
        """
        执行一次计费核对。返回 dict 结果。
        expected_cny 为 None 表示定价表查不到，只报告实际值(SKIP 判定)。
        """
        log, err = self.find_charge(model, after_ts)
        result = {
            "label": label,
            "model": model,
            "note": note,
        }
        if err or log is None:
            result.update(status=WARN, actual_quota=None, expected_quota=None,
                          message=err or "未找到扣费记录")
            return result

        actual = self.actual_quota(log)
        result["actual_quota"] = actual
        result["log"] = {
            "model_name": log.get("model_name"),
            "quota": actual,
            "prompt_tokens": log.get("prompt_tokens"),
            "completion_tokens": log.get("completion_tokens"),
            "use_time": log.get("use_time"),
            "created_at": log.get("created_at") or log.get("created_time"),
        }

        if expected_cny is None:
            result.update(status=SKIP, expected_quota=None,
                          message="定价表无此模型单价，仅记录实际扣费")
            return result

        expected_quota = self.pricing.cny_to_quota(expected_cny)
        result["expected_quota"] = round(expected_quota, 2)
        result["expected_cny"] = round(expected_cny, 6)

        if expected_quota <= 0:
            # 首张免费之类，预期为 0
            ok = actual == 0
            result["status"] = PASS if ok else WARN
            result["message"] = (f"预期 0 quota(首张免费等)，实际 {actual} quota"
                                 if not ok else "预期 0，实际 0，一致")
            return result

        diff = abs(actual - expected_quota) / expected_quota
        result["diff_ratio"] = round(diff, 4)
        if diff <= self.tolerance:
            result["status"] = PASS
            result["message"] = (
                f"实际 {actual} quota vs 预期 {expected_quota:.1f} quota "
                f"(偏差 {diff*100:.1f}% ≤ 容差 {self.tolerance*100:.0f}%)"
            )
        elif diff <= 0.6:
            result["status"] = WARN
            result["message"] = (
                f"实际 {actual} quota vs 预期 {expected_quota:.1f} quota，"
                f"偏差 {diff*100:.1f}% 偏大。请核对 ModelRatio/分组倍率配置，"
                f"或确认定价表是否过期"
            )
        else:
            result["status"] = FAIL
            result["message"] = (
                f"实际 {actual} quota vs 预期 {expected_quota:.1f} quota，"
                f"偏差 {diff*100:.1f}% 严重超标 —— 高度怀疑计费配置错误"
            )
        return result


# ============================================================
# 测试运行器
# ============================================================
class Runner:
    def __init__(self, cfg, pricing):
        self.cfg = cfg
        self.client = ApiClient(cfg["base_url"], cfg["token"])
        self.pricing = pricing
        self.auditor = QuotaAuditor(self.client, pricing,
                                    cfg.get("quota_tolerance", 0.25))
        self.results = []
        self.billing_results = []
        self.shapes = {}
        self.ctx = {}  # 用例间传递数据，如 group_id / asset_id
        self.reachable = True

    # ---------- 断言辅助 ----------
    def check(self, case_id, name, fn):
        """执行一个用例，捕获异常并记结果。"""
        started = time.time()
        try:
            status, msg = fn()
        except ApiError as e:
            status, msg = FAIL, f"请求异常: {e}"
        except Exception as e:
            status, msg = FAIL, f"未预期异常: {type(e).__name__}: {e}"
        elapsed = time.time() - started
        # 服务不可达时，连接类失败会污染结果（一堆 FAIL 看着像平台 bug）。
        # 统一降级为 WARN 并标注，让人一眼看出该重跑而不是去改代码。
        if status == FAIL and not self.reachable:
            status = WARN
            msg = f"[预检：服务不可达，本条结果不可信] {msg}"
        self.results.append({
            "id": case_id, "name": name, "status": status,
            "message": msg, "elapsed": round(elapsed, 2),
        })
        print(f"  [{_c(status, status)}] {case_id} {name}"
              + (f"  -- {msg}" if msg and status != PASS else ""))
        return status

    def assert_status(self, status, expect, ctx=""):
        if status == expect:
            return PASS, f"HTTP {status} {ctx}".strip()
        return FAIL, f"期望 HTTP {expect}，实际 {status} {ctx}".strip()

    # ========================================================
    # 用例组 1: 鉴权 / 连通性
    # ========================================================
    def preflight(self):
        """
        开跑前先探一次连通性。服务不可达时给出明确提示，
        避免后面几十条用例全被记成 FAIL，让人误以为是平台 bug。
        """
        print("\n--- 连通性预检 ---")
        try:
            st, body = self.client.get("/v1/asset-groups", params={"page_num": 1})
        except Exception as e:
            print(f"  [!] 无法连接 {self.cfg['base_url']}: {e}")
            print("      请确认 base_url 正确、new-api 已启动、网络可达。")
            self.reachable = False
            return False
        if st >= 500 or st == 0:
            print(f"  [!] 服务返回 HTTP {st}，可能未启动或被网关拦截。")
            print(f"      {str(body)[:200]}")
            self.reachable = False
            return False
        print(f"  [OK] 服务可达 (HTTP {st})")
        self.reachable = True
        return True

    def group_auth(self):
        print("\n=== 1. 鉴权与连通性 ===")

        def t_no_token():
            c = ApiClient(self.cfg["base_url"], "")
            st, _ = c.get("/v1/asset-groups")
            if st >= 500 or st == 0:
                return WARN, f"服务不可达(HTTP {st})，无法验证鉴权"
            return (PASS, f"无 token 返回 HTTP {st}，鉴权已生效") \
                if st in (401, 403) else (FAIL, f"无 token 竟返回 HTTP {st}")

        def t_bad_token():
            c = ApiClient(self.cfg["base_url"], "sk-invalid-token-for-test")
            st, _ = c.get("/v1/asset-groups")
            if st >= 500 or st == 0:
                return WARN, f"服务不可达(HTTP {st})，无法验证鉴权"
            return (PASS, f"错误 token 返回 HTTP {st}") \
                if st in (401, 403) else (FAIL, f"错误 token 竟返回 HTTP {st}")

        def t_valid_token():
            st, body = self.client.get("/v1/asset-groups")
            self.shapes["asset-groups-list"] = detect_response_shape(body)
            if st >= 500 or st == 0:
                return WARN, f"服务不可达(HTTP {st})，无法验证"
            return self.assert_status(st, 200, "token 有效，素材组列表可达")

        def t_log_token_accessible():
            st, body = self.client.get("/api/log/token", params={"p": 1, "size": 5})
            if st != 200:
                return WARN, f"/api/log/token 返回 {st}（计费核对将不可用）: {str(body)[:150]}"
            return PASS, "日志接口可用，计费核对数据源就绪"

        self.check("AUTH-01", "无 token 应被拒绝", t_no_token)
        self.check("AUTH-02", "错误 token 应被拒绝", t_bad_token)
        self.check("AUTH-03", "有效 token 可访问素材组列表", t_valid_token)
        self.check("AUTH-04", "sk-token 可读消耗日志(计费核对前提)", t_log_token_accessible)

    # ========================================================
    # 用例组 2: 素材组 CRUD
    # ========================================================
    def group_asset_group(self, skip_destructive):
        print("\n=== 2. 素材组 CRUD ===")
        ts = datetime.now().strftime("%m%d%H%M%S")
        gname = f"e2e-group-{ts}"

        def t_create():
            st, body = self.client.post("/v1/asset-groups", {"Name": gname, "Description": "e2e"})
            self.shapes["asset-group-create"] = detect_response_shape(body)
            if st != 200:
                return FAIL, f"创建失败 HTTP {st}: {str(body)[:200]}"
            gid = extract_id(body)
            if not gid:
                return FAIL, f"创建成功但拿不到 id: {str(body)[:200]}"
            self.ctx["group_id"] = gid
            shape = self.shapes["asset-group-create"]
            return PASS, f"id={gid}  响应形态={shape}"

        def t_list():
            st, body = self.client.get("/v1/asset-groups", params={"page_num": 1, "page_size": 50})
            if st != 200:
                return FAIL, f"列表失败 HTTP {st}"
            items = extract_items(body)
            total = extract_total(body)
            gid = self.ctx.get("group_id")
            found = any(extract_id(it) == gid for it in items) if gid else False
            if gid and not found:
                return WARN, (f"列表返回 {len(items)} 条(total={total})，"
                              f"但没找到刚创建的 {gid} —— 可能分页太小或响应形态异常")
            return PASS, f"列表 {len(items)} 条, total={total}"

        def t_get():
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id（创建用例未通过）"
            st, body = self.client.get(f"/v1/asset-groups/{gid}")
            if st != 200:
                return FAIL, f"详情失败 HTTP {st}: {str(body)[:200]}"
            return PASS, f"详情可读, id={extract_id(body) or gid}"

        def t_update():
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id"
            st, body = self.client.put(f"/v1/asset-groups/{gid}", {"Name": gname + "-upd"})
            if st != 200:
                return FAIL, f"更新失败 HTTP {st}: {str(body)[:200]}"
            return PASS, "更新成功"

        def t_update_empty():
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id"
            st, _ = self.client.put(f"/v1/asset-groups/{gid}", {})
            # 源码 asset.go: 两个字段都空 -> 400
            if st == 400:
                return PASS, "空 body 正确返回 400"
            return WARN, f"空 body 返回 HTTP {st}（源码约定应 400）"

        def t_delete():
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id"
            if skip_destructive:
                return SKIP, "skip_destructive=true，跳过真实删除"
            st, body = self.client.delete(f"/v1/asset-groups/{gid}")
            if st != 200:
                return FAIL, f"删除失败 HTTP {st}: {str(body)[:200]}"
            # new-api 本地实现会带 upstream_error 字段
            ue = ""
            if isinstance(body, dict) and body.get("upstream_error"):
                ue = f"（上游同步删除失败: {body['upstream_error']}）"
            return PASS, f"删除成功{ue}"

        self.check("AG-01", "创建素材组", t_create)
        self.check("AG-02", "素材组列表", t_list)
        self.check("AG-03", "素材组详情", t_get)
        self.check("AG-04", "修改素材组", t_update)
        self.check("AG-05", "修改传空应 400", t_update_empty)
        self.check("AG-06", "删除素材组", t_delete)

    # ========================================================
    # 用例组 3: 素材 CRUD  (含实测发现的"列表返回组结构"疑点)
    # ========================================================
    def group_asset(self, skip_destructive):
        print("\n=== 3. 素材 CRUD ===")
        image_url = (self.cfg.get("test_assets", {}) or {}).get("image_url", "")

        def t_create():
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id"
            if not image_url:
                return SKIP, "config.test_assets.image_url 未配置，跳过"
            st, body = self.client.post(
                f"/v1/asset-groups/{gid}/assets",
                {"image_url": image_url, "asset_type": "Image", "name": f"e2e-{int(time.time())}"})
            self.shapes["asset-create"] = detect_response_shape(body)
            if st != 200:
                return FAIL, f"创建素材失败 HTTP {st}: {str(body)[:200]}"
            aid = extract_id(body)
            if not aid:
                return FAIL, f"创建成功但无 id: {str(body)[:200]}"
            self.ctx["asset_id"] = aid
            return PASS, f"asset_id={aid} 形态={self.shapes['asset-create']}"

        def t_list_shape():
            """
            实测文档 GET /v1/asset-groups/{gid}/assets 返回的 Items 里是
            {"Id","Name","Description","GroupType",...} —— 那是**素材组**的字段，
            素材应该是 {"asset_type","source_url","status",...}。
            这里专门检出这个形态异常。
            """
            gid = self.ctx.get("group_id")
            if not gid:
                return SKIP, "无 group_id"
            st, body = self.client.get(f"/v1/asset-groups/{gid}/assets",
                                       params={"page_num": 1, "page_size": 20})
            if st != 200:
                return FAIL, f"素材列表失败 HTTP {st}"
            items = extract_items(body)
            self.shapes["asset-list"] = detect_response_shape(body)
            if not items:
                return PASS, "素材列表为空（未创建素材或已删除）"
            first = items[0]
            group_only_keys = {"GroupType", "group_type", "Description", "description"}
            asset_keys = {"asset_type", "AssetType", "source_url", "SourceURL",
                          "status", "Status", "URL"}
            has_group_key = group_only_keys & set(first.keys())
            has_asset_key = asset_keys & set(first.keys())
            if has_group_key and not has_asset_key:
                return (WARN,
                        f"⚠ 素材列表返回的是【素材组】结构（字段 {sorted(has_group_key)}），"
                        f"疑似接口串了/未过滤 group，请核对 controller/asset.go ListAssets")
            return PASS, f"素材列表 {len(items)} 条，字段正常"

        def t_get():
            aid = self.ctx.get("asset_id")
            if not aid:
                return SKIP, "无 asset_id"
            st, body = self.client.get(f"/v1/assets/{aid}")
            if st != 200:
                return FAIL, f"素材详情失败 HTTP {st}: {str(body)[:200]}"
            return PASS, f"详情可读 id={extract_id(body) or aid}"

        def t_update():
            aid = self.ctx.get("asset_id")
            if not aid:
                return SKIP, "无 asset_id"
            st, body = self.client.put(f"/v1/assets/{aid}", {"name": f"e2e-upd-{int(time.time())}"})
            if st != 200:
                return FAIL, f"更新素材失败 HTTP {st}: {str(body)[:200]}"
            return PASS, "更新成功"

        def t_update_empty():
            aid = self.ctx.get("asset_id")
            if not aid:
                return SKIP, "无 asset_id"
            st, _ = self.client.put(f"/v1/assets/{aid}", {})
            return (PASS, "空 name 正确返回 400") if st == 400 \
                else (WARN, f"空 name 返回 HTTP {st}（源码约定应 400）")

        def t_delete():
            aid = self.ctx.get("asset_id")
            if not aid:
                return SKIP, "无 asset_id"
            if skip_destructive:
                return SKIP, "skip_destructive=true，跳过真实删除"
            st, body = self.client.delete(f"/v1/assets/{aid}")
            if st != 200:
                return FAIL, f"删除素材失败 HTTP {st}: {str(body)[:200]}"
            ue = ""
            if isinstance(body, dict) and body.get("upstream_error"):
                ue = f"（上游同步删除失败: {body['upstream_error']}）"
            return PASS, f"删除成功{ue}"

        self.check("AS-01", "创建素材", t_create)
        self.check("AS-02", "素材列表(检出『返回组结构』疑点)", t_list_shape)
        self.check("AS-03", "素材详情", t_get)
        self.check("AS-04", "修改素材", t_update)
        self.check("AS-05", "修改传空 name 应 400", t_update_empty)
        self.check("AS-06", "删除素材", t_delete)

    # ========================================================
    # 用例组 4: 视频生成路由别名探测
    # ========================================================
    def group_video_routes(self):
        """
        实测文档用的是 POST /v1/videos/generations；而 router/video-router.go
        注册的是 POST /v1/video/generations(单数) 与 POST /v1/videos。
        这里只探测路由是否存在(用非法 model 让它快速失败，不真花钱)，
        避免 404 说成是业务错误。
        """
        print("\n=== 4. 视频生成路由探测(不产生费用) ===")
        probe = {"model": "__probe_nonexistent_model__", "prompt": "probe"}

        def probe_path(path, method="POST"):
            if method == "POST":
                st, body = self.client.post(path, probe)
            else:
                st, body = self.client.get(path)
            return st, body

        # 路由探测判定标准：
        #   404        -> 路由确实不存在
        #   200/4xx    -> 路由存在且业务层响应了（非法 model 会被业务层挡下）
        #   5xx/0      -> 无法判定（服务不可达/网关错误），不能算"路由存在"
        def probe_judge(st, body, path, extra_note=""):
            if st == 404:
                return WARN, f"POST {path} 不存在(404)。{extra_note}"
            # 非法 model 走完鉴权后会进到分发层，拿不到渠道时返回 503 model_not_found。
            # 这也是「路由存在且已进业务层」的证据，不能当成服务不可达。
            if st == 503 and isinstance(body, dict) and \
                    "model_not_found" in str(pick(body, "error.code", default="")):
                return PASS, f"路由存在，非法 model 已进分发层并返回 503 model_not_found"
            if st in (200, 400, 401, 403, 409, 422):
                return PASS, f"路由存在，探测返回 HTTP {st}（非法 model 被业务层挡下）"
            return WARN, (f"无法判定 {path}：HTTP {st}"
                          f"（5xx/0 通常意味着服务不可达或网关错误，请先确认 base_url）"
                          f"{(': ' + str(body)[:120]) if body else ''}")

        def t_videos_generations():
            st, body = probe_path("/v1/videos/generations")
            return probe_judge(st, body, "/v1/videos/generations",
                               "实测文档用的就是这个路径，需确认部署版本")

        def t_video_generations():
            st, body = probe_path("/v1/video/generations")
            return probe_judge(st, body, "/v1/video/generations",
                               "与 router/video-router.go 注册不符")

        def t_videos():
            st, body = probe_path("/v1/videos")
            return probe_judge(st, body, "/v1/videos",
                               "与 router/video-router.go 注册不符")

        def t_fetch_404():
            st, _ = probe_path("/v1/videos/__not_exist_task__", method="GET")
            if st in (400, 401, 403, 404, 200):
                return PASS, f"不存在的 task 返回 HTTP {st}，未崩溃"
            return WARN, (f"返回 HTTP {st}（5xx 通常为服务不可达，"
                          f"无法验证该路径的健壮性）")

        self.check("VR-01", "POST /v1/videos/generations(实测路径)", t_videos_generations)
        self.check("VR-02", "POST /v1/video/generations(源码注册路径)", t_video_generations)
        self.check("VR-03", "POST /v1/videos(源码注册路径)", t_videos)
        self.check("VR-04", "GET /v1/videos/{不存在} 不崩溃", t_fetch_404)

    # ========================================================
    # 用例组 5: 字幕擦除
    # ========================================================
    def group_subtitle(self):
        print("\n=== 5. 字幕擦除 ===")
        video_url = (self.cfg.get("test_assets", {}) or {}).get("video_url", "")

        def t_missing_video_url():
            st, body = self.client.post("/v1/videos/subtitle-erase/tasks",
                                        {"model": "cii-subtitle-erase", "prompt": "remove"})
            # 缺 video_url: 可能是 400(参数校验) 或 500(adaptor 报错)
            if st in (400, 422):
                return PASS, f"缺 video_url 正确返回 {st}"
            if st >= 500:
                return WARN, f"缺 video_url 返回 {st}（建议改成 400 参数校验）: {str(body)[:150]}"
            return WARN, f"缺 video_url 竟返回 HTTP {st}: {str(body)[:150]}"

        def t_submit_and_poll():
            if not video_url:
                return SKIP, "config.test_assets.video_url 未配置，跳过真实提交"
            st, body = self.client.post("/v1/videos/subtitle-erase/tasks", {
                "model": "cii-subtitle-erase",
                "prompt": "remove subtitles",
                "metadata": {"video_url": video_url},
            })
            if st != 200:
                return FAIL, f"提交失败 HTTP {st}: {str(body)[:250]}"
            tid = extract_id(body) or pick(body, "task_id", "id", default="")
            if not tid:
                return FAIL, f"提交成功但无 task_id: {str(body)[:250]}"
            self.ctx["subtitle_task_id"] = tid
            return PASS, f"task_id={tid}"

        def t_poll():
            tid = self.ctx.get("subtitle_task_id")
            if not tid:
                return SKIP, "无 task_id"
            return self._poll_task("/v1/videos/subtitle-erase/tasks", tid,
                                   ["completed", "succeeded"], ["failed"])

        def t_get_unknown():
            st, _ = self.client.get("/v1/videos/subtitle-erase/tasks/__not_exist__")
            return (PASS, f"未知 task 返回 HTTP {st}") if st in (400, 401, 403, 404, 200) \
                else (WARN, f"返回 HTTP {st}")

        self.check("SE-01", "缺 video_url 参数校验", t_missing_video_url)
        self.check("SE-02", "提交字幕擦除任务", t_submit_and_poll)
        self.check("SE-03", "轮询至终态", t_poll)
        self.check("SE-04", "查询不存在的任务不崩溃", t_get_unknown)

    # ========================================================
    # 用例组 6: 聊天(completions / messages)
    # ========================================================
    def group_chat(self):
        print("\n=== 6. 聊天接口 ===")

        def t_openai():
            st, body = self.client.post("/v1/chat/completions", {
                "model": "claude-opus-4-6",
                "messages": [{"role": "user", "content": "说 ok"}],
                "stream": False,
            })
            if st != 200:
                return FAIL, f"HTTP {st}: {str(body)[:200]}"
            if isinstance(body, dict) and "choices" in body:
                return PASS, "OpenAI 风格返回 choices"
            return WARN, f"HTTP 200 但无 choices 字段: {str(body)[:200]}"

        def t_claude():
            st, body = self.client.post("/v1/messages", {
                "model": "claude-opus-4-6",
                "messages": [{"role": "user", "content": "说 ok"}],
                "stream": False,
            })
            if st != 200:
                return FAIL, f"HTTP {st}: {str(body)[:200]}"
            if isinstance(body, dict) and ("content" in body or "choices" in body):
                return PASS, "Claude 风格返回 content"
            return WARN, f"HTTP 200 但结构异常: {str(body)[:200]}"

        def t_bad_model():
            st, body = self.client.post("/v1/chat/completions", {
                "model": "__probe_nonexistent_model__",
                "messages": [{"role": "user", "content": "hi"}],
            })
            if st in (400, 404):
                return PASS, f"未知模型正确返回 {st}"
            return WARN, f"未知模型返回 HTTP {st}: {str(body)[:150]}"

        self.check("CH-01", "POST /v1/chat/completions", t_openai)
        self.check("CH-02", "POST /v1/messages (Claude 风格)", t_claude)
        self.check("CH-03", "未知模型应报错", t_bad_model)

    # ========================================================
    # 通用任务轮询
    # ========================================================
    def _poll_task(self, base_path, task_id, success_states, fail_states,
                   timeout=None, interval=None):
        timeout = timeout or self.cfg.get("task_timeout_sec", 600)
        interval = interval or self.cfg.get("task_poll_interval_sec", 5)
        deadline = time.time() + timeout
        last = None
        while time.time() < deadline:
            st, body = self.client.get(f"{base_path}/{task_id}")
            if st != 200:
                return FAIL, f"轮询失败 HTTP {st}: {str(body)[:200]}"
            obj = unwrap(body)
            status = str(pick(obj, "status", "Status", default="")).lower()
            last = status
            if status in success_states:
                url = pick(obj, "url", "URL", default="")
                if isinstance(obj, dict):
                    md = obj.get("metadata") or {}
                    url = url or (md.get("url") if isinstance(md, dict) else "")
                self.ctx["last_task_url"] = url
                return PASS, f"终态={status} url={str(url)[:80]}"
            if status in fail_states:
                err = obj.get("error") if isinstance(obj, dict) else None
                msg = err.get("message") if isinstance(err, dict) else str(err or "")
                return FAIL, f"任务失败 status={status} msg={msg}"
            time.sleep(interval)
        return WARN, f"轮询超时 {timeout}s，最后状态={last}"

    # ========================================================
    # 计费核对(核心)
    # ========================================================
    def group_billing(self):
        print("\n=== 7. 计量计费核对 ===")
        cases = self.cfg.get("billing_cases", [])
        if not cases:
            print("  (config.billing_cases 为空，跳过)")
            return
        print(f"  换算: 1 元 = {self.pricing.quota_per_cny():.2f} quota "
              f"(QuotaPerUnit={self.pricing.quota_per_unit}, "
              f"USDExchangeRate={self.pricing.usd_rate})")
        for i, case in enumerate(cases, 1):
            try:
                self._run_billing_case(i, case)
            except ApiError as e:
                # 服务端重置连接会让用例直接抛异常，兜住它，别让整轮测试崩掉
                self.billing_results.append({
                    "label": case.get("name", f"case-{i}"),
                    "model": case.get("model", ""), "status": WARN,
                    "message": f"请求异常(服务端重置连接，通常伴随鉴权失败): {str(e)[:160]}"})
                print(f"      [{_c(WARN, WARN)}] 请求异常: {str(e)[:120]}")

    def _run_billing_case(self, idx, case):
        kind = case.get("kind")
        model = case.get("model")
        name = case.get("name", f"case-{idx}")
        after_ts = time.time()

        print(f"\n  --- 计费用例 B{idx:02d}: {name} ---")

        if kind == "chat":
            st, body = self.client.post("/v1/chat/completions", {
                "model": model,
                "messages": [{"role": "user", "content": case.get("prompt", "hi")}],
                "stream": False,
            })
            if st != 200:
                self.billing_results.append({
                    "label": name, "model": model, "status": FAIL,
                    "message": f"请求失败 HTTP {st}: {str(body)[:200]}"})
                print(f"      [{_c(FAIL, FAIL)}] 请求失败 HTTP {st}")
                return
            usage = {}
            if isinstance(body, dict) and "usage" in body:
                usage = body["usage"] or {}
            pt = usage.get("prompt_tokens") or 0
            ct = usage.get("completion_tokens") or 0
            exp_cny, note = self.pricing.text_expected_cny(model, pt, ct)
            print(f"      tokens: prompt={pt} completion={ct}")

        elif kind == "video":
            payload = {
                "model": model,
                "prompt": case.get("prompt", "a cat"),
                "content": [{"type": "text", "text": case.get("prompt", "a cat")}],
                "resolution": case.get("resolution", "720p"),
                "ratio": case.get("ratio", "16:9"),
                "duration": case.get("duration", 5),
            }
            if case.get("seed") is not None:
                payload["seed"] = case["seed"]
            if case.get("camera_fixed") is not None:
                payload["camera_fixed"] = case["camera_fixed"]
            if case.get("watermark") is not None:
                payload["watermark"] = case["watermark"]

            st, body = self.client.post("/v1/videos/generations", payload)
            if st != 200:
                # 回退到源码注册的别名路径
                st, body = self.client.post("/v1/video/generations", payload)
            if st != 200:
                self.billing_results.append({
                    "label": name, "model": model, "status": FAIL,
                    "message": f"提交失败 HTTP {st}: {str(body)[:200]}"})
                print(f"      [{_c(FAIL, FAIL)}] 提交失败 HTTP {st}: {str(body)[:150]}")
                return
            tid = extract_id(body) or pick(body, "task_id", "id", default="")
            print(f"      task_id={tid}")
            if tid:
                stt, msg = self._poll_task("/v1/videos", tid,
                                           ["succeeded", "completed"], ["failed", "cancelled"])
                print(f"      [{_c(stt, stt)}] 轮询: {msg}")
                if stt == FAIL:
                    self.billing_results.append({
                        "label": name, "model": model, "status": FAIL,
                        "message": f"任务失败: {msg}"})
                    return
            dur = case.get("duration", 5)
            exp_cny, note = self.pricing.video_expected_cny(
                model, case.get("resolution", "720p"), dur)
            if exp_cny is None and note and "缺少" in note:
                # 试按 token 计费的模型(1.5-pro)
                exp_cny, note = self.pricing.video_token_expected_cny(model, 1000)

        elif kind == "subtitle_erase":
            vurl = (self.cfg.get("test_assets", {}) or {}).get("video_url", "")
            if not vurl:
                self.billing_results.append({
                    "label": name, "model": model, "status": SKIP,
                    "message": "未配置 test_assets.video_url"})
                print(f"      [{_c(SKIP, SKIP)}] 未配置 video_url")
                return
            st, body = self.client.post("/v1/videos/subtitle-erase/tasks", {
                "model": model, "prompt": "remove subtitles",
                "metadata": {"video_url": vurl}})
            if st != 200:
                self.billing_results.append({
                    "label": name, "model": model, "status": FAIL,
                    "message": f"提交失败 HTTP {st}: {str(body)[:200]}"})
                print(f"      [{_c(FAIL, FAIL)}] 提交失败 HTTP {st}")
                return
            tid = extract_id(body) or pick(body, "task_id", "id", default="")
            if tid:
                stt, msg = self._poll_task("/v1/videos/subtitle-erase/tasks", tid,
                                           ["completed", "succeeded"], ["failed"])
                print(f"      [{_c(stt, stt)}] 轮询: {msg}")
                if stt == FAIL:
                    self.billing_results.append({
                        "label": name, "model": model, "status": FAIL,
                        "message": f"任务失败: {msg}"})
                    return
            exp_cny, note = self.pricing.subtitle_expected_cny(
                model, case.get("seconds", 10))
            note = (note or "") + " (seconds 需在 config 里按实际视频时长填写)"
        else:
            self.billing_results.append({
                "label": name, "model": model, "status": SKIP,
                "message": f"未知计费类型 {kind}"})
            return

        if note:
            print(f"      note: {note}")
        if exp_cny is not None:
            print(f"      预期费用: {exp_cny:.6f} 元 "
                  f"= 约 {self.pricing.cny_to_quota(exp_cny):.1f} quota")

        res = self.auditor.audit(name, model, exp_cny, after_ts, note)
        self.billing_results.append(res)
        stt = res["status"]
        print(f"      [{_c(stt, stt)}] {res['message']}")
        if res.get("log"):
            lg = res["log"]
            print(f"      日志: quota={lg['quota']} "
                  f"prompt_tok={lg['prompt_tokens']} "
                  f"completion_tok={lg['completion_tokens']} "
                  f"use_time={lg['use_time']}s")

    # ========================================================
    # 报告
    # ========================================================
    def summary(self):
        s = {PASS: 0, FAIL: 0, WARN: 0, SKIP: 0}
        for r in self.results:
            s[r["status"]] = s.get(r["status"], 0) + 1
        bs = {PASS: 0, FAIL: 0, WARN: 0, SKIP: 0}
        for r in self.billing_results:
            bs[r["status"]] = bs.get(r["status"], 0) + 1
        return s, bs

    def write_report(self, path):
        s, bs = self.summary()
        lines = []
        lines.append("# new-api 仓山/cii-group 渠道 自动化测试报告\n")
        lines.append(f"- 生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append(f"- 目标: `{self.cfg['base_url']}`")
        lines.append(f"- 计费换算: 1 元 = {self.pricing.quota_per_cny():.2f} quota "
                     f"(QuotaPerUnit={self.pricing.quota_per_unit}, "
                     f"USDExchangeRate={self.pricing.usd_rate})")
        lines.append(f"- 计费容差: ±{self.cfg.get('quota_tolerance', 0.25)*100:.0f}%\n")

        lines.append("## 总览\n")
        lines.append("| 类别 | PASS | FAIL | WARN | SKIP |")
        lines.append("|---|---|---|---|---|")
        lines.append(f"| 功能用例 | {s[PASS]} | {s[FAIL]} | {s[WARN]} | {s[SKIP]} |")
        lines.append(f"| 计费核对 | {bs[PASS]} | {bs[FAIL]} | {bs[WARN]} | {bs[SKIP]} |\n")

        lines.append("## 响应形态探测\n")
        lines.append("实测文档(dev-docs/AI模型接口.md)里的响应是上游方舟透传形态，")
        lines.append("当前源码已改为本地归一化形态。本次实际命中:\n")
        lines.append("| 端点 | 命中形态 |")
        lines.append("|---|---|")
        for k, v in self.shapes.items():
            lines.append(f"| `{k}` | {v} |")
        lines.append("")

        lines.append("## 功能用例明细\n")
        lines.append("| 用例 | 名称 | 结果 | 耗时(s) | 说明 |")
        lines.append("|---|---|---|---|---|")
        for r in self.results:
            msg = (r["message"] or "").replace("|", "\\|").replace("\n", " ")
            lines.append(f"| {r['id']} | {r['name']} | {r['status']} | "
                         f"{r['elapsed']} | {msg} |")
        lines.append("")

        if self.billing_results:
            lines.append("## 计费核对明细\n")
            lines.append("| 用例 | 模型 | 预期(元) | 预期(quota) | 实际(quota) | 偏差 | 结果 |")
            lines.append("|---|---|---|---|---|---|---|")
            for r in self.billing_results:
                eq = r.get("expected_quota")
                aq = r.get("actual_quota")
                dr = r.get("diff_ratio")
                ec = r.get("expected_cny")
                lines.append(
                    f"| {r.get('label','')} | `{r.get('model','')}` | "
                    f"{ec if ec is not None else '-'} | "
                    f"{eq if eq is not None else '-'} | "
                    f"{aq if aq is not None else '-'} | "
                    f"{f'{dr*100:.1f}%' if dr is not None else '-'} | "
                    f"{r.get('status','')} |")
            lines.append("")
            for r in self.billing_results:
                if r.get("message"):
                    lines.append(f"- **{r.get('label')}**: {r['message']}")
                if r.get("note"):
                    lines.append(f"  - 备注: {r['note']}")
            lines.append("")

        lines.append("## 结论\n")
        if s[FAIL] == 0 and bs[FAIL] == 0:
            lines.append("功能用例与计费核对均无 FAIL。")
        else:
            lines.append(f"存在 {s[FAIL]+bs[FAIL]} 个 FAIL，请优先处理上表中 FAIL 行。")
        if s[WARN] or bs[WARN]:
            lines.append(f"另有 {s[WARN]+bs[WARN]} 个 WARN 需要人工确认。")

        with open(path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
        return path


# ============================================================
# main
# ============================================================
def load_json(path, required=True):
    if not os.path.exists(path):
        if required:
            print(f"[错误] 找不到配置文件: {path}")
            print("       请复制 config.example.json 为 config.json 并填写。")
            sys.exit(2)
        return {}
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def main():
    ap = argparse.ArgumentParser(description="new-api 仓山渠道 E2E 测试 + 计费核对")
    ap.add_argument("--config", default=os.path.join(HERE, "config.json"))
    ap.add_argument("--pricing", default=os.path.join(HERE, "pricing.json"))
    ap.add_argument("--smoke", action="store_true",
                    help="只跑不花钱的只读用例")
    ap.add_argument("--billing-only", action="store_true",
                    help="只跑计费核对")
    ap.add_argument("--report", default=None, help="Markdown 报告输出路径")
    args = ap.parse_args()

    cfg = load_json(args.config)
    pricing_data = load_json(args.pricing, required=False)

    for key in ("base_url", "token"):
        if not cfg.get(key) or "xxx" in str(cfg[key]):
            print(f"[错误] config.json 的 {key} 未填写。")
            sys.exit(2)

    pricing = Pricing(pricing_data,
                      cfg.get("quota_per_unit", 500),
                      cfg.get("usd_exchange_rate", 7.3))
    runner = Runner(cfg, pricing)

    print("=" * 70)
    print("new-api 仓山/cii-group 渠道 端到端测试 + 计费核对")
    print(f"目标: {cfg['base_url']}")
    print(f"时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 70)

    skip_destructive = cfg.get("skip_destructive", True) or args.smoke

    if args.billing_only:
        runner.preflight()
        runner.group_billing()
    else:
        runner.preflight()
        runner.group_auth()
        runner.group_asset_group(skip_destructive)
        runner.group_asset(skip_destructive)
        runner.group_video_routes()
        runner.group_subtitle()
        runner.group_chat()
        if not args.smoke:
            runner.group_billing()
        else:
            print("\n(smoke 模式，跳过计费核对——它会真实扣费)")

    s, bs = runner.summary()
    print("\n" + "=" * 70)
    print(f"功能用例: PASS={s[PASS]} FAIL={s[FAIL]} WARN={s[WARN]} SKIP={s[SKIP]}")
    print(f"计费核对: PASS={bs[PASS]} FAIL={bs[FAIL]} WARN={bs[WARN]} SKIP={bs[SKIP]}")
    print("=" * 70)

    if runner.shapes:
        print("\n响应形态探测:")
        for k, v in runner.shapes.items():
            print(f"  {k}: {v}")

    if args.report:
        p = runner.write_report(args.report)
        print(f"\n报告已写入: {p}")

    if not runner.reachable:
        print("\n" + "!" * 70)
        print("注意: 预检发现服务不可达，上面的 FAIL 很可能是连不上导致的，")
        print("并非平台缺陷。请确认 base_url / 服务状态后重跑。")
        print("!" * 70)

    total_fail = s[FAIL] + bs[FAIL]
    sys.exit(1 if total_fail else 0)


if __name__ == "__main__":
    main()
