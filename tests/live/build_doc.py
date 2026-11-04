#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
把明文的 seedance-docs.html 构建成可部署的产物：

  dist/doc.html       —— 页面外壳，不再包含任何文档数据
  dist/doc-data.js    —— 文档数据（混淆后），由页面通过 <script src> 加载

同时支持可选的溯源水印。

用法：
    python build_doc.py                  # 默认构建到 ./dist
    python build_doc.py --watermark      # 附带半透明水印（含构建日期）
    python build_doc.py --out ../dist    # 指定输出目录

为什么拆成两个文件：
    单文件时，Ctrl+U 复制走 HTML 就拿到了全部正文与定价。
    拆开后，页面源码只剩骨架，数据文件可以单独在 nginx 上做
    Referer 校验 / 限流 / 鉴权，防护才有落点。

注意：这不是加密，只是提高复制成本。真正的保密只能靠服务端鉴权。
"""
import argparse
import base64
import datetime
import json
import os
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
SRC = os.path.join(HERE, "seedance-docs.html")
EXTRACT = os.path.join(HERE, "_extract_data.js")

# 混淆用密钥（分成两段存放，抬高一点直接搜字符串的成本）
KEY_PARTS = ("cXN5a2Fp", "LTIwMjYtZG9j")  # atob 后拼成 "qsykai-2026-doc"

DATA_PLACEHOLDER = """/* ------------------------------------------------------------------
   文档数据已外置：由 doc-data.js 注入 window.__DOC
   明文源见 tests/live/seedance-docs.html（仅本地编辑用，不要直接部署）
   ------------------------------------------------------------------ */
const { ENDPOINTS, MODELS, TREE, HEADER_PARAMS, BODY_PARAMS, RESPONSE_PARAMS } = window.__DOC;"""


def extract_data():
    out = subprocess.run(
        ["node", EXTRACT, SRC], capture_output=True, check=True
    )
    raw = out.stdout.decode("utf-8")
    data = json.loads(raw)
    for key in ("ENDPOINTS", "MODELS", "TREE",
                "HEADER_PARAMS", "BODY_PARAMS", "RESPONSE_PARAMS"):
        if key not in data or not data[key]:
            raise SystemExit("抽取失败：缺少 %s" % key)
    return data


def xor_b64(payload: bytes, key: str) -> str:
    kb = key.encode("utf-8")
    mixed = bytes(b ^ kb[i % len(kb)] for i, b in enumerate(payload))
    return base64.b64encode(mixed).decode("ascii")


def build_data_js(data) -> str:
    key = base64.b64decode(KEY_PARTS[0]).decode() + base64.b64decode(KEY_PARTS[1]).decode()
    raw = json.dumps(data, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    blob = xor_b64(raw, key)
    return (
        "/* Seedance API 文档数据（构建产物，请勿手工编辑）\n"
        "   源：tests/live/seedance-docs.html\n"
        "   构建：python build_doc.py\n"
        "   说明：本文件由页面以 <script src> 方式加载，可单独做访问控制。 */\n"
        "(function () {\n"
        '  var p1 = "%s", p2 = "%s";\n'
        "  var k = atob(p1) + atob(p2);\n"
        '  var b = atob("%s");\n'
        "  var out = new Uint8Array(b.length);\n"
        "  for (var i = 0; i < b.length; i++) {\n"
        "    out[i] = b.charCodeAt(i) ^ k.charCodeAt(i %% k.length);\n"
        "  }\n"
        '  var json = new TextDecoder("utf-8").decode(out);\n'
        "  window.__DOC = JSON.parse(json);\n"
        "})();\n"
    ) % (KEY_PARTS[0], KEY_PARTS[1], blob)


def watermark_block() -> str:
    stamp = datetime.date.today().isoformat()
    svg = (
        "<svg xmlns='http://www.w3.org/2000/svg' width='360' height='200'>"
        "<text x='0' y='100' fill='#000' font-size='14' "
        "font-family='system-ui,sans-serif' transform='rotate(-22 0 100)'>"
        "qsykai.com · %s</text></svg>" % stamp
    )
    b64 = base64.b64encode(svg.encode("utf-8")).decode("ascii")
    return (
        '<div class="doc-wm" aria-hidden="true"></div>\n'
        "<style>\n"
        ".doc-wm{position:fixed;inset:0;pointer-events:none;z-index:2147483000;"
        "opacity:.055;background-repeat:repeat;"
        "background-image:url(\"data:image/svg+xml;base64,%s\")}\n"
        "</style>\n" % b64
    )


def build_html(html: str, use_watermark: bool) -> str:
    # 1) 用占位符替换数据块
    start = html.index("const ENDPOINTS")
    rp = html.index("const RESPONSE_PARAMS")
    close = html.index("];", rp)
    end = html.index("/*", close)
    html = html[:start] + DATA_PLACEHOLDER + "\n\n" + html[end:]

    # 2) 在主脚本前引入数据文件
    if html.count("<script>") != 1:
        raise SystemExit("预期只有 1 个内联 <script>，实际 %d 个，请检查源文件" % html.count("<script>"))
    html = html.replace("<script>", '<script src="doc-data.js"></script>\n<script>', 1)

    # 3) 可选水印
    if use_watermark:
        html = html.replace("</body>", watermark_block() + "</body>", 1)

    return html


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default=os.path.join(HERE, "dist"))
    ap.add_argument("--watermark", action="store_true", help="加半透明溯源水印")
    args = ap.parse_args()

    if not os.path.isfile(SRC):
        raise SystemExit("找不到源文件：%s" % SRC)

    data = extract_data()
    with open(SRC, encoding="utf-8") as f:
        html = f.read()

    out_html = build_html(html, args.watermark)
    out_js = build_data_js(data)

    os.makedirs(args.out, exist_ok=True)
    p_html = os.path.join(args.out, "doc.html")
    p_js = os.path.join(args.out, "doc-data.js")
    with open(p_html, "w", encoding="utf-8", newline="\n") as f:
        f.write(out_html)
    with open(p_js, "w", encoding="utf-8", newline="\n") as f:
        f.write(out_js)

    # 自检：产物里不应残留正文特征词
    leaks = [w for w in ("字节跳动", "1080p · 输入不含视频", "参考一致性") if w in out_html]
    print("doc.html      %6d bytes" % os.path.getsize(p_html))
    print("doc-data.js   %6d bytes" % os.path.getsize(p_js))
    print("watermark     %s" % ("on" if args.watermark else "off"))
    print("源码残留检查  %s" % ("发现泄漏: " + ", ".join(leaks) if leaks else "通过（正文已全部外置）"))
    if leaks:
        sys.exit(1)


if __name__ == "__main__":
    main()
