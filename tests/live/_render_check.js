// 用最小 DOM 桩跑页面内联脚本，验证渲染结果（不依赖浏览器）
const fs = require("fs");
const vm = require("vm");

const html = fs.readFileSync(process.argv[2], "utf8");
const scripts = [...html.matchAll(/<script>([\s\S]*?)<\/script>/g)].map(m => m[1]);
const dataScript = fs.readFileSync(process.argv[3], "utf8");
const inline = scripts[scripts.length - 1];

const els = new Map();
function el(sel) {
  if (!els.has(sel)) {
    els.set(sel, {
      _sel: sel, textContent: "", innerHTML: "", value: "", disabled: false,
      style: {}, dataset: {}, className: "",
      classList: { add() {}, remove() {}, toggle() {}, contains: () => false },
      addEventListener() {}, closest: () => null,
      querySelector: () => el(sel + " *"), querySelectorAll: () => [],
    });
  }
  return els.get(sel);
}

const sandbox = {
  console,
  atob,
  TextDecoder,
  URLSearchParams,
  JSON,
  Math,
  Date,
  setTimeout,
  performance: { now: () => 0 },
  navigator: { clipboard: { writeText: async () => {} } },
  location: { search: "?model=qwen-image-3.0-pro", protocol: "http:", origin: "http://www.qsykai.com:33000", hostname: "www.qsykai.com" },
  history: { replaceState() {} },
  window: { addEventListener() {}, scrollTo() {} },
  IntersectionObserver: class { constructor() {} observe() {} },
  document: {
    querySelector: (s) => el(s),
    querySelectorAll: () => [],
    getElementById: (id) => el("#" + id),
    addEventListener() {},
  },
};
vm.createContext(sandbox);
vm.runInContext(dataScript, sandbox);   // 注入 window.__DOC
vm.runInContext(inline, sandbox);       // 执行页面脚本（含初始化）

function txt(sel) { return el(sel).textContent || el(sel).innerHTML; }
function has(sel, s) { return (el(sel).innerHTML || "").includes(s); }

let fail = 0;
function check(name, ok, extra) {
  console.log((ok ? "PASS  " : "FAIL  ") + name + (extra ? "  → " + extra : ""));
  if (!ok) fail++;
}

// ---- qwen-image-3.0-pro（URL 参数已选中）----
check("标题", txt("#docTitle").includes("Qwen Image 3.0 Pro"), txt("#docTitle"));
check("模型 ID", txt("#docModelId") === "qwen-image-3.0-pro", txt("#docModelId"));
check("计费徽标", txt("#billingBadge") === "按张计费", txt("#billingBadge"));
check("接口地址", txt("#epUrl").endsWith("/v1/images/generations"), txt("#epUrl"));
check("价格表头", has("#priceHead", "单价（CNY / 张）"));
check("价格含 1K 0.25", has("#priceBody", "0.25 CNY"));
check("价格含 2K 0.50", has("#priceBody", "0.50 CNY"));
check("价格含输入 0.02", has("#priceBody", "0.02 CNY"));
check("价格标注销售单价口径", txt("#priceNote").includes("以下为当前接口的销售单价") && txt("#priceNote").includes("最终费用以实际调用用量和消费记录为准"));
check("请求参数含 parameters.n", has("#paramBody", "parameters.n"));
check("响应参数含 output_image_count", has("#respBody", "output_image_count"));
check("示例含 input.messages", txt("#sampleCode").includes("input"), "");
check("示例 URL 用 images/generations", txt("#sampleCode").includes("/v1/images/generations"));
check("概览为同步说明", txt("#ovCalloutText").includes("同步返回"));
check("响应示例含 data[].url", txt("#respSample").includes("output_0.png"));

// ---- 切回视频模型，确认回落逻辑 ----
vm.runInContext('state.modelId="doubao-seedance-2.0"; state.endpoint=curEndpoints()[0].code; renderHeader(); renderEndpoints();', sandbox);
check("视频-计费徽标回落", txt("#billingBadge") === "按 Token 计费", txt("#billingBadge"));
check("视频-接口回落", txt("#epUrl").endsWith("/v1/video/generations"), txt("#epUrl"));
check("视频-价格表头回落", has("#priceHead", "百万 Token"));
check("视频-参数回落", has("#paramBody", "resolution"));
check("视频-概览回落为异步", txt("#ovCalloutText").includes("异步任务范式"));
check("视频-响应回落", has("#respBody", "progress"));

console.log(fail ? `\n${fail} 项失败` : "\n全部通过");
process.exit(fail ? 1 : 0);
