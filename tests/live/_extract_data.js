// 从 seedance-docs.html 中抽出数据常量，输出 JSON（供 build_doc.py 调用）
// 用法: node _extract_data.js <html路径>
const fs = require("fs");

const src = fs.readFileSync(process.argv[2], "utf8");
const start = src.indexOf("const ENDPOINTS");
if (start < 0) throw new Error("未找到 const ENDPOINTS");

// 数据块结束于 RESPONSE_PARAMS 之后的注释分隔线
const rp = src.indexOf("const RESPONSE_PARAMS");
const close = src.indexOf("];", rp);
const end = src.indexOf("/*", close);
if (rp < 0 || end < 0) throw new Error("未找到数据块边界");

const block = src.slice(start, end);
const data = eval(
  "(() => {" + block + "; return { ENDPOINTS, MODELS, TREE, HEADER_PARAMS, BODY_PARAMS, RESPONSE_PARAMS }; })()"
);

process.stdout.write(JSON.stringify(data));
