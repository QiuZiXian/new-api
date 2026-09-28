#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
真人审核 (Visual Validate) 端到端测试脚本

覆盖两段：
  A. 上游直连  —— 直接打 cii-group 的 /api/v1/visual-validate/{sessions,results}，
                  验证上游 key / 路径 / 响应字段是否和 dev-docs/cii-api.md 一致。
  B. 自身平台  —— 打 new-api 的 /v1/visual-validate/*，验证路由、鉴权、落库、
                  回调签名校验、失败/过期分支、结果幂等，以及"认证成功→自动注册
                  本地素材组→可传素材"的闭环。

两种运行模式：
  simulate（默认，无人值守）：用模拟回调跑通除"真人真的刷脸"以外的全部分支。
        - 坏 sign            → 期望 400
        - resultCode != 10000 → 期望会话 failed
        - resultCode == 10000 但 token 未真人认证 → upstream 404 → 期望 expired
        - 重复回调            → 幂等，状态不变
        - results 纯透传/参数校验 → 期望 400
  live：真人在手机 H5 完成刷脸。脚本创建会话、打印 H5 链接，然后轮询会话状态
        直到 succeeded/failed/expired 或超时（默认 30 分钟 = BytedToken 有效期），
        成功后自动校验本地素材组注册与素材上传闭环。

用法
----
    # 1) 配置：tests/e2e/config.json（与 newapi_e2e.py 共用）里保证有 base_url / token
    #    可选新增 "visual_validate" 段（见文件末尾 CONFIG 说明）
    cp config.example.json config.json

    # 2) 只测上游（需要 config.visual_validate.upstream_key）
    python visual_validate_e2e.py --upstream-only

    # 3) 模拟全分支（不刷脸，不花钱）
    python visual_validate_e2e.py

    # 4) 真人刷脸全流程
    python visual_validate_e2e.py --live

    # 5) 出报告
    python visual_validate_e2e.py --report reports/vv-2026-09-24.md

退出码: 0=全部通过(允许 WARN/SKIP)  1=有 FAIL  2=配置/环境错误
"""

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8")
        sys.stderr.reconfigure(encoding="utf-8")
    except Exception:
        pass

HERE = os.path.dirname(os.path.abspath(__file__))
PASS, FAIL, WARN, SKIP = "PASS", "FAIL", "WARN", "SKIP"
_COLOR = {PASS: "\033[92m", FAIL: "\033[91m", WARN: "\033[93m", SKIP: "\033[90m"}
_RESET = "\033[0m"

DEFAULT_UPSTREAM_PATH = "/api/v1/visual-validate"
RESULT_CODE_PASSED = "10000"


def _c(status, text):
    if not sys.stdout.isatty():
        return text
    return f"{_COLOR.get(status, '')}{text}{_RESET}"


# ============================================================
# HTTP
# ============================================================
def http_call(method, url, body=None, headers=None, timeout=60):
    """返回 (status, text, obj)。非 2xx 不抛异常——用例要自己断言状态码。"""
    data = None
    hdrs = dict(headers or {})
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        hdrs.setdefault("Content-Type", "application/json")
    hdrs.setdefault("Accept", "application/json")

    req = urllib.request.Request(url, data=data, headers=hdrs, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            text = resp.read().decode("utf-8", errors="replace")
            status = resp.status
    except urllib.error.HTTPError as e:
        status = e.code
        try:
            text = e.read().decode("utf-8", errors="replace")
        except Exception:
            text = ""
    except Exception as e:
        return None, str(e), None

    obj = None
    try:
        obj = json.loads(text)
    except Exception:
        pass
    return status, text, obj


class Client:
    def __init__(self, base_url, token):
        self.base = base_url.rstrip("/")
        self.token = token

    def call(self, method, path, body=None, auth=True, params=None):
        url = self.base + path
        if params:
            url += "?" + urllib.parse.urlencode(params)
        headers = {"Authorization": f"Bearer {self.token}"} if auth else {}
        return http_call(method, url, body=body, headers=headers)


class Upstream:
    def __init__(self, base_url, key, api_path):
        self.base = base_url.rstrip("/")
        self.key = key
        self.path = api_path or DEFAULT_UPSTREAM_PATH

    def call(self, sub, body):
        url = f"{self.base}{self.path}/{sub}"
        headers = {"Authorization": f"Bearer {self.key}"}
        return http_call("POST", url, body=body, headers=headers)


# ============================================================
# 用例记录
# ============================================================
class Results:
    def __init__(self):
        self.items = []

    def add(self, name, status, detail=""):
        self.items.append({"name": name, "status": status, "detail": str(detail)[:400]})
        print(f"  [{_c(status, status)}] {name}" + (f"  -- {detail}" if detail else ""))

    def expect(self, name, ok, detail=""):
        self.add(name, PASS if ok else FAIL, detail)
        return ok

    @property
    def failed(self):
        return [i for i in self.items if i["status"] == FAIL]


def section(title):
    print(f"\n=== {title} ===")


# ============================================================
# 配置
# ============================================================
def load_config(path):
    if not os.path.exists(path):
        print(f"配置文件不存在: {path}\n请先 copy config.example.json config.json 并填写 base_url / token")
        sys.exit(2)
    with open(path, "r", encoding="utf-8") as f:
        cfg = json.load(f)
    if not cfg.get("base_url") or not cfg.get("token"):
        print("config.json 缺少 base_url / token")
        sys.exit(2)
    return cfg


# ============================================================
# A. 上游直连
# ============================================================
def test_upstream(up: Upstream, r: Results):
    section("A. 上游直连 (cii-group visual-validate)")
    if not up.base or not up.key:
        r.add("上游 sessions / results", SKIP, "config.visual_validate.upstream_base_url / upstream_key 未配置")
        return None

    # A1 sessions
    status, text, obj = up.call("sessions", {"CallbackURL": "https://www.example.com/callback"})
    if not r.expect("A1 POST {path}/sessions -> 200", status == 200, f"HTTP {status} {text[:200]}"):
        return None
    token = (obj or {}).get("BytedToken", "")
    h5 = (obj or {}).get("H5Link", "")
    r.expect("A1 BytedToken 非空", bool(token), token[:16] + "..." if token else "")
    r.expect("A1 H5Link 非空", bool(h5), h5[:80] if h5 else "")
    if h5:
        print(f"      H5 链接(30 分钟内可用, 仅限认证一次): {h5}")

    # A2 results（真实 token 但未刷脸 -> 期望 404）
    status2, text2, obj2 = up.call("results", {"BytedToken": token})
    if status2 == 404:
        r.expect("A2 POST {path}/results(未认证 token) -> 404", True, "符合文档『认证记录不存在或已过期』")
    elif status2 == 200 and (obj2 or {}).get("GroupId"):
        r.expect("A2 POST {path}/results(未认证 token)", WARN,
                 f"返回了 GroupId={obj2.get('GroupId')}——该 token 之前已完成认证，请换新 token 重测")
    else:
        r.expect("A2 POST {path}/results(未认证 token) -> 404", False, f"HTTP {status2} {text2[:200]}")

    # A3 results 参数校验
    status3, text3, _ = up.call("results", {"BytedToken": ""})
    r.expect("A3 results 空 BytedToken -> 400", status3 == 400, f"HTTP {status3} {text3[:160]}")
    return token


# ============================================================
# B. 自身平台
# ============================================================
def create_session(c: Client, r: Results, downstream_cb="", project="", label="会话"):
    body = {}
    if downstream_cb:
        body["callback_url"] = downstream_cb
    if project:
        body["project_name"] = project
    status, text, obj = c.call("POST", "/v1/visual-validate/sessions", body)
    if status != 200:
        r.add(f"{label} 创建", FAIL, f"HTTP {status} {text[:200]}")
        return None
    return obj


def test_platform(c: Client, r: Results, downstream_cb="", project="", mode="simulate",
                  image_url="", wait_sec=1800):
    section("B. 自身平台 (new-api /v1/visual-validate)")

    # B1 鉴权
    status, _, _ = c.call("GET", "/v1/visual-validate/sessions", auth=False)
    r.expect("B1 无 token 访问 -> 401", status == 401, f"HTTP {status}")

    # B2 创建会话
    s1 = create_session(c, r, downstream_cb, project, "B2 会话")
    if not s1:
        return None
    sid = s1.get("id", "")
    r.expect("B2 返回 id/byted_token/h5_link", bool(sid and s1.get("byted_token") and s1.get("h5_link")),
             f"id={sid}")
    r.expect("B2 初始状态 pending", s1.get("status") == "pending", s1.get("status"))

    # B3 回调地址形态 + 可达性
    cb_url = s1.get("callback_url", "")
    qs = urllib.parse.parse_qs(urllib.parse.urlparse(cb_url).query)
    sign = (qs.get("sign") or [""])[0]
    r.expect("B3 callback_url 含 session/sign", bool((qs.get("session") or [""])[0]) and bool(sign), cb_url[:120])
    host = urllib.parse.urlparse(cb_url).netloc.split(":")[0]
    if host in ("localhost", "127.0.0.1") or host.startswith("192.168.") or host.startswith("10."):
        r.add("B3 callback_url 公网可达性", WARN,
              f"回调地址是内网/本机({host})——手机刷脸后浏览器跳不回来，需先配置 系统设置-服务器地址(ServerAddress)")
    else:
        r.add("B3 callback_url 公网可达性", PASS, host)

    # B4 详情 / 列表
    status, text, obj = c.call("GET", f"/v1/visual-validate/sessions/{sid}")
    r.expect("B4 GET sessions/:id -> 200", status == 200 and (obj or {}).get("id") == sid, f"HTTP {status}")
    status, text, obj = c.call("GET", "/v1/visual-validate/sessions", params={"page_num": 1, "page_size": 20})
    items = (obj or {}).get("items", [])
    r.expect("B4 GET sessions 列表包含该会话", status == 200 and any(i.get("id") == sid for i in items),
             f"total={(obj or {}).get('total')}")

    if mode == "live":
        return run_live(c, r, s1, image_url, wait_sec)

    # ---- simulate 分支 ----
    # B5 坏 sign
    bad = f"{c.base}/v1/visual-validate/callback?" + urllib.parse.urlencode(
        {"session": sid, "sign": "wrong-sign", "bytedToken": s1.get("byted_token", ""),
         "resultCode": RESULT_CODE_PASSED})
    status, _, _ = http_call("GET", bad)
    r.expect("B5 回调 sign 错误 -> 400", status == 400, f"HTTP {status}")

    # B6 失败分支
    s_fail = create_session(c, r, project=project, label="B6 会话")
    if s_fail:
        fid = s_fail.get("id", "")
        url = build_callback(c.base, s_fail, result_code="20001")
        status, _, _ = http_call("GET", url)
        _, _, o = c.call("GET", f"/v1/visual-validate/sessions/{fid}")
        r.expect("B6 resultCode!=10000 -> failed", (o or {}).get("status") == "failed",
                 f"HTTP {status}, status={(o or {}).get('status')}, result_code={(o or {}).get('result_code')}")
        # B6b 已结束会话查 results
        status, text, _ = c.call("POST", "/v1/visual-validate/results", {"session_id": fid})
        r.expect("B6b 已结束会话 results -> 400", status == 400, f"{text[:160]}")

    # B7 成功码但未真刷脸 -> 上游 404 -> expired
    s_exp = create_session(c, r, project=project, label="B7 会话")
    if s_exp:
        eid = s_exp.get("id", "")
        url = build_callback(c.base, s_exp, result_code=RESULT_CODE_PASSED)
        status, _, _ = http_call("GET", url)
        _, _, o = c.call("GET", f"/v1/visual-validate/sessions/{eid}")
        st = (o or {}).get("status")
        r.expect("B7 resultCode=10000 但未认证 -> expired", st == "expired",
                 f"HTTP {status}, status={st}（上游 results 返回 404 说明 token 未认证，符合预期）")
        # B7b 幂等：重复回调状态不变
        status2, _, _ = http_call("GET", url)
        _, _, o2 = c.call("GET", f"/v1/visual-validate/sessions/{eid}")
        r.expect("B7b 重复回调幂等", (o2 or {}).get("status") == st, f"HTTP {status2}, status={(o2 or {}).get('status')}")

    # B8 results 参数校验 / 纯透传
    status, text, _ = c.call("POST", "/v1/visual-validate/results", {})
    r.expect("B8 results 空 body -> 400", status == 400, text[:160])
    if s1.get("byted_token"):
        status, text, _ = c.call("POST", "/v1/visual-validate/results", {"byted_token": s1.get("byted_token")})
        r.expect("B8b results 纯透传(未认证 token) -> 400", status == 400 and "过期" in text, text[:200])

    # B9 下游通知（simulate 下走了 B6/B7 的 failed/expired，若配了 callback_url 应收到 POST）
    if downstream_cb:
        r.add("B9 下游通知", WARN, f"已传 callback_url={downstream_cb}——请到该地址确认收到 POST（best-effort，失败不影响主流程）")
    else:
        r.add("B9 下游通知", SKIP, "未配置 callback_url（可用 webhook.site 生成后再跑）")
    return None


def build_callback(base, session, result_code=RESULT_CODE_PASSED):
    qs = urllib.parse.parse_qs(urllib.parse.urlparse(session.get("callback_url", "")).query)
    return f"{base}/v1/visual-validate/callback?" + urllib.parse.urlencode({
        "session": session.get("id", ""),
        "sign": (qs.get("sign") or [""])[0],
        "bytedToken": session.get("byted_token", ""),
        "resultCode": result_code,
        "algorithmBaseRespCode": "0",
        "reqMeasureInfoValue": "1",
        "verify_type": "real_time",
    })


def run_live(c: Client, r: Results, session, image_url, wait_sec):
    """真人刷脸：打印 H5 链接后轮询会话状态。"""
    sid = session.get("id", "")
    h5 = session.get("h5_link", "")
    print("\n" + "-" * 68)
    print("请用手机浏览器打开下面的链接完成真人认证（可追加 &lng=zh 指定语言）：")
    print(f"  {h5}")
    print(f"认证完成后浏览器会跳回: {session.get('callback_url', '')}")
    print("-" * 68)
    print(f"等待会话状态变化，最长 {wait_sec} 秒（BytedToken 有效期 30 分钟）...")

    deadline = time.time() + wait_sec
    final = None
    while time.time() < deadline:
        _, _, o = c.call("GET", f"/v1/visual-validate/sessions/{sid}")
        if o and o.get("status") != "pending":
            final = o
            break
        time.sleep(5)
    if not final:
        r.add("L1 真人认证完成", FAIL, f"{wait_sec}s 内会话仍为 pending（token 可能已过期）")
        return None

    st = final.get("status")
    r.expect("L1 真人认证 -> succeeded", st == "succeeded", f"status={st}, result_code={final.get('result_code')}")
    if st != "succeeded":
        return None

    gid = final.get("upstream_group_id", "")
    local_gid = final.get("local_asset_group_id", "")
    r.expect("L2 上游 GroupId 已落库", bool(gid), gid)
    r.expect("L3 本地素材组已自动注册", bool(local_gid), local_gid)

    if local_gid:
        status, text, o = c.call("GET", f"/v1/asset-groups/{local_gid}")
        r.expect("L4 GET /v1/asset-groups/:id -> 200", status == 200, text[:160])
        if status == 200 and gid:
            r.expect("L4 本地组 upstream_asset_group_id 对齐",
                     (o or {}).get("upstream_asset_group_id") == gid,
                     f"{(o or {}).get('upstream_asset_group_id')} vs {gid}")

    if image_url and local_gid:
        status, text, o = c.call("POST", f"/v1/asset-groups/{local_gid}/assets", {
            "image_url": image_url, "asset_type": "Image", "name": "真人审核测试素材"})
        r.expect("L5 向真人素材组上传素材", status in (200, 201) and bool((o or {}).get("id")), f"HTTP {status} {text[:160]}")
    else:
        r.add("L5 上传素材", SKIP, "未配置 image_url（config.test_assets.image_url）")
    return final


# ============================================================
# 报告
# ============================================================
def write_report(path, items, meta):
    os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)
    lines = ["# 真人审核 (Visual Validate) 测试报告", "",
             f"- 时间: {time.strftime('%Y-%m-%d %H:%M:%S')}",
             f"- 平台: {meta.get('base_url')}",
             f"- 上游: {meta.get('upstream') or '未配置'}",
             f"- 模式: {meta.get('mode')}", "",
             "| 结果 | 用例 | 详情 |", "|---|---|---|"]
    for i in items:
        lines.append(f"| {i['status']} | {i['name']} | {i['detail'].replace('|', '/')} |")
    fails = [i for i in items if i["status"] == FAIL]
    lines += ["", f"合计 {len(items)} 项，FAIL {len(fails)} 项。"]
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")
    print(f"\n报告已写入: {path}")


def main():
    ap = argparse.ArgumentParser(description="真人审核 (visual-validate) 端到端测试")
    ap.add_argument("--config", default=os.path.join(HERE, "config.json"))
    ap.add_argument("--upstream-only", action="store_true", help="只跑上游直连用例")
    ap.add_argument("--live", action="store_true", help="真人刷脸全流程（否则跑模拟分支）")
    ap.add_argument("--callback-url", default="", help="下游通知地址（会用 POST 收到一次通知）")
    ap.add_argument("--image-url", default="", help="闭环测试用的公网图片 URL")
    ap.add_argument("--project-name", default="", help="上游 ProjectName（默认 default）")
    ap.add_argument("--wait-sec", type=int, default=1800, help="live 模式最长等待秒数")
    ap.add_argument("--report", default="", help="输出 markdown 报告路径")
    args = ap.parse_args()

    cfg = load_config(args.config)
    vv = cfg.get("visual_validate", {}) or {}
    up = Upstream(vv.get("upstream_base_url", ""), vv.get("upstream_key", ""),
                  vv.get("upstream_api_path", DEFAULT_UPSTREAM_PATH))
    c = Client(cfg["base_url"], cfg["token"])

    r = Results()
    print(f"平台: {cfg['base_url']}   模式: {'live' if args.live else 'simulate'}")

    if up.base and up.key:
        test_upstream(up, r)
    else:
        r.add("A. 上游直连", SKIP, "config.visual_validate.upstream_base_url / upstream_key 未配置")
        if args.upstream_only:
            print("\n未配置上游直连参数，无法只跑上游用例。请在 config.json 的 visual_validate 段填写。")
            sys.exit(2)

    if not args.upstream_only:
        project = args.project_name or vv.get("project_name", "")
        image_url = args.image_url or (cfg.get("test_assets", {}) or {}).get("image_url", "")
        test_platform(c, r, downstream_cb=args.callback_url or vv.get("downstream_callback_url", ""),
                      project=project, mode="live" if args.live else "simulate",
                      image_url=image_url, wait_sec=args.wait_sec)

    print(f"\n合计 {len(r.items)} 项，FAIL {len(r.failed)} 项，"
          f"WARN {len([i for i in r.items if i['status'] == WARN])} 项，"
          f"SKIP {len([i for i in r.items if i['status'] == SKIP])} 项")

    if args.report:
        write_report(args.report, r.items, {
            "base_url": cfg["base_url"],
            "upstream": f"{up.base}{up.path}" if up.base else "",
            "mode": "live" if args.live else "simulate",
        })
    sys.exit(1 if r.failed else 0)


if __name__ == "__main__":
    main()
