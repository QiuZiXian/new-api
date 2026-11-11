# Seedance API 文档页部署方案（237 / nginx）

> 目标地址：`http://www.qsykai.com:33000/doc.html`
> 部署节点：237（主），`/data/compose/237`
> 源文件：`tests/live/seedance-docs.html`（明文，编辑用）
> 构建产物：`tests/live/dist/doc.html` + `tests/live/dist/doc-data.js`
> 文档版本：2026-10-30（v2，数据外置 + 访问控制）

---

## 一、结论摘要

| 项 | 决定 |
|---|---|
| 托管方式 | **nginx 静态文件 + 精确匹配 location** |
| 挂载方式 | 复用已有的 `/data/compose/237/nginx/conf.d` 挂载，**不新增 volume** |
| 配置形式 | `conf.d/doc.locations`（片段）+ `newapi.conf` 内一行 `include`。**location 不能直接放进 `conf.d/*.conf`** |
| 是否重建容器 | **否**，仅 `nginx -t` + `nginx -s reload` |
| 停机时间 | **0**（reload 平滑切换，不中断现有连接） |
| 文件构成 | **两个**：`doc.html`（外壳）+ `doc-data.js`（混淆数据） |
| 访问控制 | `doc-data.js` 做 Referer 校验，站外/无 Referer 一律 403 |
| 对外路径 | `/doc.html`，另有 `/docs` 302 跳转 |
| 是否影响 API | 否，`location =` 精确匹配优先级高于 `location /` |

选择理由：237 的 nginx 只挂了 `nginx.conf`、`conf.d`、`ssl`、`logs` 四项，**没有挂载 html 目录**。若走标准 `root /usr/share/nginx/html`，必须改 compose 并重建 nginx 容器（会有 1–2 秒抖动）。把文件放进已挂载的 `conf.d/doc/` 下、用 `alias` 指到单文件，可以做到零改动、零重启。

---

## 二、现状核对（实测，非推测）

### 2.1 拓扑

237 是主节点，`docker-compose.yml` 的 `name: newapi-primary`，包含：

| 服务 | 容器名 | 端口 |
|---|---|---|
| mysql 8.2 | mysql-master | 3306 |
| redis 7 | redis-master | 6379 |
| new-api master | newapi-master | 3000 |
| new-api slave | newapi-slave-1 / -2 | 3001 / 3002 |
| nginx 1.27 | **nginx-lb** | **33000**（另有 80 / 443 预留） |

`/v1/*` 全部经 nginx-lb 反代到 `upstream new_api_cluster`（当前只有 `172.28.5.12:3000` 生效，其余标 `down`）。

### 2.2 nginx 挂载点

```yaml
volumes:
  - /data/compose/237/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
  - /data/compose/237/nginx/conf.d:/etc/nginx/conf.d:ro
  - /data/nginx/ssl:/etc/nginx/ssl:ro
  - /data/nginx/logs:/var/log/nginx
```

关键结论：

1. **没有** `/usr/share/nginx/html` 挂载 → 放进去的文件在容器重建后会丢。
2. `conf.d` 目录已挂载（容器内只读），宿主机可写；`nginx.conf` 里是 `include /etc/nginx/conf.d/*.conf;`，**只匹配 `.conf`**，因此 `conf.d/doc/doc.html` 不会被当成配置解析。
3. `nginx.conf` 已 `include mime.types`，但 `default_type` 是 `application/octet-stream`，所以 location 里必须显式 `default_type text/html`，否则浏览器会下载而不是渲染。

### 2.3 一个必须知道的坑：现在 `/doc.html` 已经返回 200

未部署前实测：

```
GET /doc.html  →  200  text/html  1166 字节  <title>New API</title>
```

这是 new-api 前端的 SPA 兜底：任何未命中静态资源的路径都返回 `index.html`。**所以"200 就算部署成功"是错的**，验收必须比对内容（标题、字节数、模型名）。

---

## 三、部署后的请求路径

```
浏览器
  │
  ├─ GET /doc.html ─────► nginx-lb  精确匹配 location = /doc.html
  │                          └─ alias /etc/nginx/conf.d/doc/doc.html（不进后端）
  │      └─ GET /doc-data.js ─► nginx-lb  精确匹配 + Referer 校验
  │                              └─ alias .../doc-data.js
  │
  └─ POST /v1/video/generations ─► nginx-lb  location /
                                     └─ proxy_pass new_api_cluster
                                          └─ newapi-master:3000
```

同源（协议 + 域名 + 端口一致）→ 浏览器不触发 CORS 预检 → 文档页「在线调试」可以直接调接口。

---

## 四、实施步骤

### 4.1 本地构建

```bash
cd D:\d\project\2026\feature\seedance\tests\live
python build_doc.py                 # 产物输出到 ./dist
# 可选：python build_doc.py --watermark   加半透明溯源水印
```

产物：

| 文件 | 当前大小 | 内容 |
|---|---|---|
| `dist/doc.html` | 30,373 B | 页面外壳，无正文/无定价 |
| `dist/doc-data.js` | 9,242 B | 文档数据，base64+XOR 混淆 |
| （源）`seedance-docs.html` | 37,196 B | 明文全量，**不要直接部署** |

构建脚本自带泄漏检查：若产物 HTML 里还能搜到「字节跳动」「1080p · 输入不含视频」等正文特征词，直接以非 0 退出。

### 4.2 同步到部署目录

```bash
cp dist/doc.html      D:\d\qzx\yunke\deploy\compose\237\nginx\conf.d\doc\doc.html
cp dist/doc-data.js   D:\d\qzx\yunke\deploy\compose\237\nginx\conf.d\doc\doc-data.js
```

### 4.3 上传并执行（推荐：脚本）

```powershell
cd D:\d\qzx\yunke\deploy
$env:SSH_PASS="<root口令>"
& "C:\Users\lenovo\.workbuddy\binaries\python\envs\default\Scripts\python.exe" deploy-doc.py 237
```

脚本依次：

1. `mkdir -p /data/compose/237/nginx/conf.d/doc`
2. sftp 上传 `doc.html` / `doc-data.js` / `doc.locations` / `deploy-doc.sh`
3. 远端执行 `bash /data/compose/237/deploy-doc.sh`（含幂等的 include 注入）

### 4.4 手工方式（等价，逐条）

```bash
# ① 上传（顺序：先数据文件、后页面，避免出现「页面在、数据没到」）
scp deploy/compose/237/nginx/conf.d/doc/doc-data.js  root@36.248.63.237:/data/compose/237/nginx/conf.d/doc/
scp deploy/compose/237/nginx/conf.d/doc/doc.html     root@36.248.63.237:/data/compose/237/nginx/conf.d/doc/
scp deploy/compose/237/nginx/conf.d/doc.locations    root@36.248.63.237:/data/compose/237/nginx/conf.d/

# ② 237 上
mkdir -p /data/compose/237/nginx/conf.d/doc
chmod 644 /data/compose/237/nginx/conf.d/doc/doc.html \
          /data/compose/237/nginx/conf.d/doc/doc-data.js \
          /data/compose/237/nginx/conf.d/doc.locations

# ②.5 首次部署：把片段挂进 server 块（幂等，已包含则跳过）
cd /data/compose/237/nginx/conf.d
if ! grep -q "doc.locations" newapi.conf; then
  cp -a newapi.conf newapi.conf.bak.$(date +%Y%m%d%H%M%S)
  sed -i '/^[[:space:]]*server_name[[:space:]]*_;/a\    include /etc/nginx/conf.d/doc.locations;' newapi.conf
fi

# ③ 容器内可见性（ro 挂载，宿主机写入即可见）
docker exec nginx-lb ls -la /etc/nginx/conf.d/doc/

# ④ 语法检查 → 平滑重载
docker exec nginx-lb nginx -t
docker exec nginx-lb nginx -s reload
```

> 顺序注意：**先 `doc-data.js`，再 `doc.html`，最后 `doc.locations` + include + reload**。反过来会出现「location 已在、文件还没到」的短暂 404，虽然无害但会干扰排障。

### 4.5 nginx 配置内容

**踩过的坑（首次部署直接失败）**：`conf.d/*.conf` 是被 `nginx.conf` 在 `http {}` 层 `include` 的，而 `location` 只能出现在 `server {}` 里。所以不能把 location 写成 `doc.conf` 放进去，否则报：

```
nginx: [emerg] "location" directive is not allowed here in /etc/nginx/conf.d/doc.conf:22
```

正确做法：用**不带 `.conf` 后缀**的片段文件 `doc.locations`（不会被自动加载），再在 `newapi.conf` 的 server 块内显式 include。

片段文件 `/data/compose/237/nginx/conf.d/doc.locations`：

```nginx
# 页面外壳
location = /doc.html {
    alias /etc/nginx/conf.d/doc/doc.html;
    default_type text/html;
    charset utf-8;
    add_header Cache-Control "no-cache, must-revalidate";
    add_header X-Content-Type-Options "nosniff";
    add_header X-Frame-Options "SAMEORIGIN";
}

# 文档数据：只允许本站页面引用
location = /doc-data.js {
    alias /etc/nginx/conf.d/doc/doc-data.js;
    default_type application/javascript;
    charset utf-8;
    add_header Cache-Control "no-cache, must-revalidate";
    add_header X-Content-Type-Options "nosniff";

    # 防直接抓取：无 Referer 或站外 Referer 一律 403
    if ($http_referer !~ "^https?://(www\.)?qsykai\.com(:[0-9]+)?/") {
        return 403;
    }
}

location = /docs {
    return 302 /doc.html;
}
```

在 `newapi.conf` 的 server 块内插入一行（放在 `location /` 之前）：

```nginx
server {
    listen 33000 default_server;
    server_name _;
    include /etc/nginx/conf.d/doc.locations;   # ← 新增这一行

    location / {
        proxy_pass http://new_api_cluster;
        ...
    }
}
```

部署脚本用 `sed` 幂等注入（已存在则跳过，并留 `newapi.conf.bak.<时间戳>`）：

```bash
if ! grep -q "doc.locations" newapi.conf; then
  cp -a newapi.conf newapi.conf.bak.$(date +%Y%m%d%H%M%S)
  sed -i '/^[[:space:]]*server_name[[:space:]]*_;/a\    include /etc/nginx/conf.d/doc.locations;' newapi.conf
fi
```

四个细节，缺一不可：

1. **`= /doc.html`**：精确匹配优先级最高（`=` > `^~` > 正则 > 前缀），否则会被 `location /` 抢走并反代到 new-api，只能拿到 SPA 兜底页。
2. **`default_type text/html`**：主配置的 `default_type` 是 `application/octet-stream`，不加会被浏览器当附件下载。
3. **`charset utf-8`**：文档含大量中文，缺 charset 在部分环境会乱码。
4. **`doc-data.js` 的 Referer 校验**：这是唯一真正"服务端可执行"的防护点。注意通过 IP 访问（`http://36.248.63.237:33000/doc.html`）也会被拦，排障时临时注释掉那三行。

### 4.6 改动范围

- `docker-compose.yml`：**不改**（无新增 volume）
- `nginx/nginx.conf`：**不改**
- `conf.d/newapi.conf`：**改 1 行**（加 `include /etc/nginx/conf.d/doc.locations;`），部署脚本自动备份
- `conf.d/ssl.conf.disabled`、`conf.d/` 其他文件：不改
- new-api 镜像：**不改、不重建**

---

## 五、验收标准

```bash
# 1) 内容断言：必须是我们的文档页，不是 New API SPA
curl -s http://www.qsykai.com:33000/doc.html | grep -o '<title>[^<]*</title>'
# 期望：含 Seedance；若出现 <title>New API</title> 说明 location 没生效

# 2) 体积断言（外壳约 30KB，SPA 兜底页 1.1KB）
curl -s -o /dev/null -w '%{size_download} bytes  %{content_type}\n' http://www.qsykai.com:33000/doc.html

# 3) 源码不含正文（关键）：应为 0
curl -s http://www.qsykai.com:33000/doc.html | grep -c "输入不含视频"

# 4) 数据文件：带 Referer 200，不带 403
curl -s -o /dev/null -w 'with-referer: %{http_code}\n' \
  -H "Referer: http://www.qsykai.com:33000/doc.html" http://www.qsykai.com:33000/doc-data.js
curl -s -o /dev/null -w 'no-referer  : %{http_code}\n' http://www.qsykai.com:33000/doc-data.js

# 5) 短链与 API 未受影响
curl -s -o /dev/null -w '/docs: %{http_code}\n' http://www.qsykai.com:33000/docs
curl -s -o /dev/null -w 'api  : %{http_code}\n' http://www.qsykai.com:33000/api/status
```

浏览器侧：

1. 打开 `http://www.qsykai.com:33000/doc.html`，标题为 Seedance 文档，左侧模型树可切换 4 个模型。
2. `Ctrl+U` 查看源码：**搜不到任何定价与参数表内容**。
3. 「在线调试」填 `sk-` Key → 发一次 `GET /v1/models` → 能拿到 JSON（同源生效，无 CORS 报错）。

| 检查项 | 通过标准 | 2026-10-30 实测 |
|---|---|---|
| HTTP 状态 | 200 | ✅ 200 |
| Content-Type | `text/html`（不是 `application/octet-stream`） | ✅ `text/html; charset=utf-8` |
| 页面标题 | 含 Seedance，不是 `New API` | ✅ `Seedance 视频生成 · API 文档 \| 启算云枢` |
| 响应体积 | ≈30KB | ✅ 30,373 B |
| 页面源码含定价 | 0 处 | ✅ 0 处 |
| `doc-data.js` 带 Referer | 200 | ✅ 200（9,242 B） |
| `doc-data.js` 无 Referer | 403 | ✅ 403 |
| `doc-data.js` 站外 Referer | 403 | ✅ 403（`https://evil.example.com`） |
| `/docs` | 302 到 `/doc.html` | ✅ 302 |
| `/api/status` | 200（未被 location 影响） | ✅ 200 |
| 线上数据可解码 | 4 个模型 / 2.0 倍率 23 / 6 档价格 / 17 条参数 | ✅ 全部一致 |
| 在线调试 | 能收到接口响应，控制台无 CORS 报错 | ⏳ 需人工在浏览器点一次 |

公网实测命令与结果（2026-10-30 20:48）：

```text
doc.html              : 200  text/html; charset=utf-8  30373B
doc-data.js 有Referer : 200  9242B
doc-data.js 无Referer : 403
doc-data.js 站外Ref   : 403
/docs                 : 302 -> http://www.qsykai.com:33000/doc.html
api/status            : 200
页面残留正文条数       : 0
```

---

## 六、回滚

```bash
# ① 移除 server 块里的 include（幂等）
cd /data/compose/237/nginx/conf.d
sed -i '/include \/etc\/nginx\/conf.d\/doc.locations;/d' newapi.conf

# ② 删掉片段与静态文件
rm -f  /data/compose/237/nginx/conf.d/doc.locations
rm -rf /data/compose/237/nginx/conf.d/doc

# ③ 校验并重载
docker exec nginx-lb nginx -t && docker exec nginx-lb nginx -s reload
```

若想完全恢复 `newapi.conf` 原样，直接用部署时留下的备份还原：

```bash
ls -la /data/compose/237/nginx/conf.d/newapi.conf.bak.*
cp -a /data/compose/237/nginx/conf.d/newapi.conf.bak.<时间戳> \
      /data/compose/237/nginx/conf.d/newapi.conf
```

回滚后 `/doc.html` 会重新落到 SPA 兜底页（200 + `New API` 标题），属预期。

---

## 七、日常更新流程

```bash
# 1) 改明文源
#    tests/live/seedance-docs.html
# 2) 重建
cd tests/live && python build_doc.py
# 3) 发布
python ../../..//deploy/deploy-doc.py 237     # 或手工 scp 两个文件后 reload
```

因为加了 `Cache-Control: no-cache`，用户刷新即可看到新版本，不需要强刷或加版本参数。

---

## 八、多节点注意事项

237 与 47 都对公网放行了 33000，若入口可能落到 47：

- 47 上也需部署同样两个文件与 `doc.conf`（路径 `/data/compose/47/...`）；
- 两台内容必须一致，否则同一地址刷新两次会看到不同版本。

---

## 九、备选方案对比

| 方案 | 做法 | 是否重建 | 适用场景 |
|---|---|---|---|
| **A（已采用）** | `conf.d/doc/` + `alias` + 数据外置 | 否 | 当前结构下的最优解 |
| B | 新增 `/data/compose/237/nginx/html:/usr/share/nginx/html:ro` | 是（1–2s 抖动） | 文档变多、需要整目录时 |
| C | 放进 `web/public/`，随 new-api 构建 embed | 是（全量重建镜像） | 迭代成本高，不推荐 |
| D | 独立 `nginx:alpine` 容器 | 新增容器 | 文档与主站解耦、独立域名 |

---

## 十、风险与安全

| 风险 | 说明 | 处置 |
|---|---|---|
| 文档页暴露 | 页面本身不含密钥，但会把用户填的 Key 发到同源接口 | 可接受；如需收敛加 basic auth |
| 调试面板存 Key | Token 只存在页面内存，不写 localStorage | 已实现；关闭标签即清除 |
| 与 SPA 路由冲突 | `/doc.html` 被精确匹配接管 | 无影响，前端不依赖该路径 |
| 容器重建后文件丢失 | 文件在宿主机挂载目录内 | 已规避（这正是选 A 而非 `docker cp` 的原因） |
| Referer 校验误伤 | 用 IP 访问、或浏览器策略剥离 Referer 时数据加载失败 | 排障时注释 `if` 三行 |

---

## 十一、文件清单

```
D:\d\qzx\yunke\deploy\
├── compose\237\nginx\conf.d\
│   ├── doc.locations               # 片段：location 配置（含 Referer 校验）
│   ├── newapi.conf                 # 原有，改动 1 行 include
│   └── doc\
│       ├── doc.html                # 外壳 30,373 B（无正文）
│       └── doc-data.js             # 数据 9,242 B（混淆）
├── remote\deploy-doc-237.sh        # 远端执行脚本（含幂等 include 注入）
└── deploy-doc.py                   # 一键上传+执行

D:\d\project\2026\feature\seedance\tests\live\
├── seedance-docs.html              # 明文源（编辑用，勿部署）
├── build_doc.py                    # 构建脚本（拆分 + 混淆 + 校验）
├── _extract_data.js                # node 抽取数据常量
└── dist\
    ├── doc.html                    # 构建产物
    └── doc-data.js
```

服务器侧实际形态（2026-10-30 部署后）：

```
/data/compose/237/nginx/conf.d/
├── newapi.conf                      # 原有：upstream + location /
├── newapi.conf.bak.20261030204723   # 部署前自动备份
├── ssl.conf.disabled                # 原有，未启用
├── doc.locations                    # 新增：片段（无 .conf 后缀，不被自动加载）
└── doc/
    ├── doc.html                     # 30,373 B
    └── doc-data.js                  # 9,242 B
```

`newapi.conf` 的 server 块内多了这一行：

```nginx
server_name _;
include /etc/nginx/conf.d/doc.locations;   # ← 新增
```

---

## 十二、防爬与内容保护（重要）

### 12.1 先把边界说清楚

**浏览器能显示，客户端就能拿到。** 前端不存在真正的保密，只有成本高低：

- 禁用右键、禁用 F12、禁止选中复制：浏览器快捷键和 DevTools 协议都能绕过，只能挡住完全不懂技术的人，还会误伤正常阅读体验。**不建议做。**
- 数据加密（把 key 也放在页面里）：有耐心的人半小时能解出来。
- 真正有效的是**服务端鉴权**：没登录/没口令，连内容都不下发。

### 12.2 已实施的四层（成本递增）

| # | 措施 | 挡住什么 | 挡不住什么 |
|---|---|---|---|
| 1 | **结构分离**：正文/定价移到 `doc-data.js` | `Ctrl+U` 复制 HTML 只能拿到空壳 | 会看 Network 面板的人 |
| 2 | **数据混淆**：base64 + XOR | 直接打开数据文件不可读，无法搜索关键词 | 会跑这段 JS 的人（约 30 秒） |
| 3 | **Referer 校验**：`doc-data.js` 仅本站可引 | `wget` / `curl` / 无头爬虫批量抓取 | 伪造 `Referer` 头 |
| 4 | **X-Frame-Options** | 第三方套壳成「镜像站」 | 直接复制文件自建站 |

第 1 层解决的就是你提的那个问题：**以前复制一份 HTML 就全拿走了，现在复制不到正文。**

### 12.3 可选：溯源水印

```bash
python build_doc.py --watermark
```

在页面底层铺一层 5.5% 透明度的斜向水印（含 `qsykai.com` 与构建日期），`pointer-events:none` 不影响交互。作用是**截图外传后可追溯到来源与时间**，不防复制。介意观感就别开。

### 12.4 真正要"不给看"：加口令

如果这个文档只想给客户看，用 nginx basic auth，两行配置：

```bash
# 生成口令文件（在 237 上）
docker run --rm httpd:2.4-alpine htpasswd -nbB doc '你的口令' > /data/compose/237/nginx/conf.d/doc.htpasswd
```

```nginx
location = /doc.html {
    auth_basic "Seedance API Docs";
    auth_basic_user_file /etc/nginx/conf.d/doc.htpasswd;
    alias /etc/nginx/conf.d/doc/doc.html;
    default_type text/html;
    charset utf-8;
}
```

生效后浏览器会弹登录框，未认证连 HTML 都拿不到——这才是真正的访问控制。

### 12.5 一个更值得注意的泄漏面

实测 `GET /api/pricing` **无需任何鉴权**，直接返回全部模型的 `model_ratio` 与分组倍率：

```bash
curl -s http://www.qsykai.com:33000/api/pricing | head -c 300
```

也就是说，**真要抄你的定价，根本不用爬这个文档页**，直接调公开接口就行。文档页做再厚的壳，也改变不了这件事。

如果定价是需要保护的商业信息，正确方向是收敛 `/api/pricing`（改为登录可见或按分组过滤），而不是在前端页面加壳。但注意前端价格页依赖这个接口，改动要一起评估。

### 12.6 我的建议

文档正文（接口路径、参数说明）公开对获客有利，藏起来反而增加接入门槛，不建议保护；真正需要在意的是**定价与未公开模型**，从接口层收敛更有效。当前这套（结构分离 + 混淆 + Referer）已经挡住"顺手复制"这个级别的抄袭，投入产出比合适。

---

## 十三、部署状态

| 阶段 | 状态 |
|---|---|
| 现状核对（挂载、SPA 兜底、端口） | ✅ 完成 |
| 数据外置构建（`build_doc.py`） | ✅ 完成，产物 30,373 B + 9,242 B |
| 产物自检（无正文残留、JS 语法、解码回环） | ✅ 通过 |
| 上传到 237 | ✅ 完成（2026-10-30 20:47） |
| include 注入 + 备份 | ✅ 完成（`newapi.conf.bak.20261030204723`） |
| `nginx -t` / reload | ✅ 通过，平滑重载无中断 |
| 公网验收（状态码/标题/体积/Referer/短链/API） | ✅ 全部通过 |
| 线上数据解码校验 | ✅ 4 模型 / 倍率 23 / 6 档价格 / 17 条参数 |
| 浏览器人工点一次「在线调试」 | ⏳ 建议你点一次确认（我本机无头浏览器当前不可用） |

### 访问地址

- 文档页：`http://www.qsykai.com:33000/doc.html`
- 短链：`http://www.qsykai.com:33000/docs`
- 数据文件：`http://www.qsykai.com:33000/doc-data.js`（仅本站 Referer 可访问）
- 指定模型：`http://www.qsykai.com:33000/doc.html?model=qwen-image-3.0-pro`

---

## 十四、文档页数据模型（新增模型必读）

页面数据都放在 `seedance-docs.html` 的 `MODELS` 里。**没配的字段会自动回落到视频模型的全局定义**（`ENDPOINTS` / `HEADER_PARAMS` / `BODY_PARAMS` / `RESPONSE_PARAMS`），所以老的 Seedance 模型不需要改一行。

| 字段 | 作用 | 缺省行为 |
|---|---|---|
| `endpoints` | 该模型的接口列表（`code/method/name/path/billing/hint`） | 回落到视频的 4 个接口 |
| `params.header` / `.body` / `.response` | 请求头、Body、响应参数表 | 回落到全局表 |
| `sampleBody` | 调用示例与在线调试的预填请求体 | 回落到视频 create 示例 |
| `respSample` | 响应示例 | 回落到视频任务示例 |
| `pricing` | `{note, head, rows}`，支持按张计价的多列表格 | 由 `price[][]` 生成「CNY / 百万 Token」 |
| `billingBadge` | 顶部计费徽标 | 按 Token 计费 |
| `callout` / `notes` | 概览提示框与要点列表 | 异步任务范式 |

### 新增一个模型的三步

1. `MODELS` 加条目（照 `qwen-image-3.0-pro` 抄结构）；
2. `TREE` 对应分组加节点，`id` 与 `MODELS` 的 key 一致；
3. `python build_doc.py` → 覆盖部署 → `nginx -s reload`。

### 已接入的图像模型

`qwen-image-3.0-pro`：`POST /v1/images/generations`，同步返回，按张计费（输出 1K 0.25、2K 0.50，输入参考图 0.02）。价格取自线上 `/api/pricing` 的 `image_count_price` 字段，不是手写；改价后需重新核对文档。

### parameters 的透传规则

`relay/channel/ali/dto.go` 的 `AliImageParameters` 现在声明了 `size / n / prompt_extend / prompt_extend_mode / negative_prompt / seed / watermark / steps / scale / thinking_mode / enable_sequential / bbox_list / color_palette`，并带 `Extra map[string]json.RawMessage` 兜底：**未声明的上游新参数会在序列化时平铺回 `parameters`**，不再被静默丢弃。

> 历史坑（2026-11-05 已修）：早期只有白名单结构体，`negative_prompt` 与 `prompt_extend_mode` 会被丢弃，且不报错、无日志。文档当时删掉了这两行；修复后已补回，并把概览里的"不会转发"提示换成"parameters 会原样转发到上游"。

仍存在的一条限制：该渠道走自定义 API 路径时，**顶层 OpenAI 风格的 `size` / `n` 不会自动映射**到 `parameters`，必须写在 `parameters` 内。

### 回归校验

```bash
cd tests/live
python build_doc.py
node _render_check.js dist/doc.html dist/doc-data.js
```

`_render_check.js` 用最小 DOM 桩在 node 里跑页面脚本，校验「切到图像模型 / 切回视频模型」两套渲染结果（标题、计费徽标、接口地址、价格表头、参数表、示例、响应示例共 19 项）。本机无头浏览器不可用时，用它可以替代肉眼检查。
