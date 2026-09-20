# new-api 接入「仓山 / cii-group」渠道 — 接口测试方案与完善接口文档

> 整理范围：
>
> - 上游接口文档：`dev-docs/cii-api.md`（解析自 `https://www.cii-group.com/docs.html`）。
> - 已接入模型清单：`dev-docs/模型定价.xlsx`（28 条，覆盖京东 / 仓山两个人工渠道、文本对话 / 图片生成 / 视频生成 / 视频编辑 / OCR 识别 五种类型、6 种计费单位）。
> - new-api 中已落地的 adaptor 与 controller：`relay/channel/task/doubao`、`relay/channel/task/ciisubtitleerase`、`controller/asset.go`、`controller/visual_validate.go`、`service/asset`、`service/visualvalidate`、`model/visual_validate_session.go` 等。
> - **目的**：① 给出一张「上游接口 ↔ new-api 路由 ↔ 内部组件 ↔ 已接入模型 ↔ 测试用例」的完整对照表，定位可能的接入疏漏；② 用 curl + Go 单测两种形态给出可执行的测试用例样板，模拟单元测试 + 集成测试两个层面的回归基线；③ 标出**未接入**或**部分接入**的上游接口，让网关接入新模型时按这张表执行。

---

## 0. 缩略词 / 一句话架构

- **上游 (Upstream)**：cii-group 网关 `https://www.cii-group.com/app-api`，以及它代理的火山方舟接口；鉴权统一为 `Authorization: Bearer <API Key>`。
- **平台层 (Platform)**：new-api 自身，对客户端暴露 OpenAI-兼容 / 自定义路由（`/v1/video/generations`、`/v1/videos`、`/v1/contents/generations/tasks`、`/v1/videos/subtitle-erase/tasks`、`/v1/asset-groups`、`/v1/visual-validate/...`）。
- **渠道 (Channel)**：new-api 中每个上游对应一个 `ChannelType` 与 `BaseURL` 配置项，存于 `constant.ChannelBaseURLs[ChannelType]`。
- **Adaptor**：实现 `TaskAdaptor` 接口的对象，负责把平台的统一任务请求转换为上游 HTTP 调用，并把上游响应解析回平台统一格式（见 `controller/task_content.go` / `dto/openai_video.go`）。

---

## 1. 对接总览（一张表看完全部边界）

| # | 上游接口（cii-api.md） | 上游路径 | new-api 用户面路由 | new-api 内部组件 | 已接入模型（来自 xlsx） | 状态 |
|---|---|---|---|---|---|---|
| 1 | Video 创建任务 | `POST /api/v1/contents/generations/tasks` | `POST /v1/video/generations` `/v1/videos` `/v1/videos/:video_id/remix` | `relay/channel/task/doubao/adaptor.go::TaskAdaptor`（默认上游 `/api/v3/...`，可通过 `OtherSettings.TaskApiPath` 切换到 v1） | `doubao-seedance-1.5-pro`、`doubao-seedance-2.0`、`doubao-seedance-2.0-fast`、`doubao-seedance-2.0-mini`、`doubao-seedance-2.5` | 已接入 |
| 2 | Video 查询任务 | `GET /api/v1/contents/generations/tasks/{id}` | `GET /v1/video/generations/:task_id` `/v1/videos/:task_id` | 同上 | 同上 | 已接入 |
| 3 | Video 列表任务 | `GET /api/v1/contents/generations/tasks?page_num=&...` | `GET /v1/contents/generations/tasks` | `controller/task_content.go::ListContentTasks`（**不代理**，走本地 `model.Task` 库） | 全部 task 类模型 | 已接入（仅本地视图） |
| 4 | Video 取消 / 删除任务 | `DELETE /api/v1/contents/generations/tasks/{id}` | `DELETE /v1/contents/generations/tasks/:task_id` | `controller/task_content.go::CancelContentTask` → `doubao.DeleteTask` | 同上 | 已接入 |
| 5 | Subtitle Erase 创建 | `POST /api/v1/videos/subtitle-erase/tasks` | `POST /v1/videos/subtitle-erase/tasks` | `relay/channel/task/ciisubtitleerase/adaptor.go` | `cii-subtitle-erase` | 已接入 |
| 6 | Subtitle Erase 查询 | `GET /api/v1/videos/subtitle-erase/tasks/{task_id}` | `GET /v1/videos/subtitle-erase/tasks/:task_id` | 同上 | 同上 | 已接入 |
| 7 | Asset Groups 创建 | `POST /api/v1/asset-groups` | `POST /v1/asset-groups` | `controller/asset.go::CreateAssetGroup` → `service/asset/service.go` | 平台通用（与模型无关） | 已接入 |
| 8 | Asset Groups 列表 | `POST /api/v1/asset-groups/list` | `GET /v1/asset-groups` | `controller/asset.go::ListAssetGroups` | 同上 | 已接入 |
| 9 | Asset Groups 详情 | `GET /api/v1/asset-groups/{groupId}` | `GET /v1/asset-groups/:group_id` | `controller/asset.go::GetAssetGroup` | 同上 | 已接入 |
| 10 | Asset Groups 更新 | `PUT /api/v1/asset-groups/{groupId}` | `PUT /v1/asset-groups/:group_id` | `controller/asset.go::UpdateAssetGroup` | 同上 | 已接入 |
| 11 | Asset Groups 删除 | `DELETE /api/v1/asset-groups/{groupId}`（文档未列） | `DELETE /v1/asset-groups/:group_id` | `controller/asset.go::DeleteAssetGroup` | 同上 | 已接入（路径在 dev-docs/dev进度.md 已被定义为"上游存在"） |
| 12 | Assets 创建 | `POST /api/v1/assets`（注意文档写 `/api/v1/asset-groups/:groupId/assets` 二选一） | `POST /v1/asset-groups/:group_id/assets` | `controller/asset.go::CreateAsset` | 同上 | 已接入 |
| 13 | Assets 列表 | `POST /api/v1/assets/list` | `GET /v1/asset-groups/:group_id/assets` | `controller/asset.go::ListAssets` | 同上 | 已接入 |
| 14 | Assets 详情 | `GET /api/v1/assets/{assetId}` | `GET /v1/assets/:asset_id` | `controller/asset.go::GetAsset` | 同上 | 已接入 |
| 15 | Assets 更新 | `PUT /api/v1/assets/{assetId}` | `PUT /v1/assets/:asset_id` | `controller/asset.go::UpdateAsset` | 同上 | 已接入 |
| 16 | Assets 删除 | `DELETE /api/v1/assets/{assetId}`（文档未列） | `DELETE /v1/assets/:asset_id` | `controller/asset.go::DeleteAsset` | 同上 | 已接入 |
| 17 | Visual Validate 创建会话 | `POST /api/v1/visual-validate/sessions` | `POST /v1/visual-validate/sessions` | `controller/visual_validate.go::CreateVisualValidateSession` | 平台通用 | 已接入 |
| 18 | Visual Validate 查会话列表 | （文档未列，由 new-api 自身维护） | `GET /v1/visual-validate/sessions` | `controller/visual_validate.go::ListVisualValidateSessions` | 平台通用 | 已接入 |
| 19 | Visual Validate 查会话详情 | （同上） | `GET /v1/visual-validate/sessions/:session_id` | `controller/visual_validate.go::GetVisualValidateSession` | 平台通用 | 已接入 |
| 20 | Visual Validate 拉结果 | `POST /api/v1/visual-validate/results` | `POST /v1/visual-validate/results` | `controller/visual_validate.go::GetVisualValidateResult` | 平台通用 | 已接入 |
| 21 | Visual Validate 浏览器回调 | **new-api 自身端点**：`GET /v1/visual-validate/callback?session=&sign=` | `GET /v1/visual-validate/callback` | `controller/visual_validate.go::VisualValidateCallback`（不挂 TokenAuth，验签 sign 防伪造） | 平台通用 | 已接入 |
| 22 | Image Generations (Seedream) | `POST /app-api/v1/images/generations` | —— | —— | `doubao-seedream-3.0/4.0/4.5/5.0-lite/5.0-pro` | **未对接**（见 §6.1） |
| 23 | Image Edits | `POST /app-api/v1/images/edits`（multipart） | —— | —— | （xlsx 中无 Image Edit 模型） | **未对接** |
| 24 | Gemini 转发：图生图 / 文生图 | `POST /app-api/3rd/v1beta/models/{model}:generateContent` | —— | —— | `doubao-3.1-flash-image-preview`、`doubao-banana-2/pro`、`doubao-3.5-flash` 等 | **未对接**（见 §6.2） |

> 表中所有 `ChannelType` 与 `BaseURL` 来自 `constant/channel.go`：
>
> - `ChannelTypeDoubaoVideo = 54`，`ChannelBaseURLs[54] = "https://ark.cn-beijing.volces.com"`（方舟 / cii-group 通过渠道 OtherSettings 配置覆盖）
> - `ChannelTypeCiiSubtitleErase = 61`，`ChannelBaseURLs[61] = "https://www.cii-group.com/app-api"`
> - 素材 / 视觉审核走独立 TokenAuth 路由组，无 `ChannelType`，由 controller 直接调用 `service` 层
>
> 已接入模型详情见 §2。

---

## 2. 已接入模型清单（与 `模型定价.xlsx` 对齐）

> 摘录自 xlsx 报表，重点列：模型名称 / model 参数值 / 类型 / 计费模式 / 单价 / 走哪个渠道。客户端 `model` 字段必须填 **E 列 "model 参数值"**，不是 C 列中文展示名。

### 2.1 仓山 / cii-group 主推（ChannelTypeDoubaoVideo=54 / CiiSubtitleErase=61）

| # | model 参数值（发请求用） | 类型 | 适配 adaptor | 计费模式 | 单价（来自 xlsx） |
|---|---|---|---|---|---|
| 1 | `DeepSeek-V4-Pro` | 文本对话 | （走方舟 DeepSeek 渠道？需实测） | 按 Token 峰谷 | 高峰 ¥9/¥27/¥0.30；空闲 ¥4.5/¥13.5/¥0.15 |
| 2 | `DeepSeek-V4-Flash` | 文本对话 | 同上 | 按 Token 峰谷 | 高峰 ¥3/¥9/¥0.10；空闲 ¥1.5/¥4.5/¥0.05 |
| 3 | `DeepSeek-V3.2` | 文本对话 | 同上 | 按 Token | ¥2 / ¥3 / ¥1 |
| 4 | `DeepSeek-V3-0324` | 文本对话 | 同上 | 按 Token | ¥2 / ¥8 / ¥0.50 |
| 5 | `DeepSeek-R1-0528` | 文本对话 | 同上 | 按 Token | ¥4 / ¥16（含思维链） / ¥0.50 |
| 6 | `DeepSeek-OCR2` | OCR 识别 | 同上 | 按 Token / 自部署 | 第三方托管 ~$0.15/M Token |
| 7 | `DeepSeek-OCR` | OCR 识别 | 同上 | 同上 | 同上 |
| 8 | `doubao-seedance-2.5` | 视频生成 | `doubao` adaptor | 按秒（按分辨率折算） | 720p ≈¥1.51/秒；480p ≈¥0.67/秒；1080p ≈¥3.74/秒 |
| 9 | `doubao-seedance-2.0` | 视频生成 | `doubao` adaptor | 按秒 | 480p ¥0.56；720p ¥1.21；1080p ¥3.01；4K ¥6.22 |
| 10 | `doubao-seedance-2.0-fast` | 视频生成 | `doubao` adaptor | 按秒 | 480p ¥0.44；720p ¥0.95（活动期 75 折） |
| 11 | `doubao-seedance-2.0-mini` | 视频生成 | `doubao` adaptor | 按秒 | 480p ¥0.28；720p ¥0.605（活动期 4 折） |
| 12 | `doubao-seedance-1.5-pro` | 视频生成 | `doubao` adaptor | 按 Token | 有声 ¥0.016/千 Token；无声 ¥0.008/千 Token（⚠️即将下线） |
| 13 | `doubao-seedream-5.0-pro` + `doubao-seedream-5.0-lite` | 图片生成 | （⚠️ 走方舟 volcengine=45 渠道，不是 cii adaptor） | 按张 | Pro：≤261 万像素 ¥0.3/张、>261 万像素 ¥0.6/张；Lite ¥0.22/张；首张免费 |
| 14 | `doubao-seedream-4.5` | 图片生成 | volcengine=45 | 按张 | ¥0.25/张；首张免费；角色图≈¥0.50/张 |
| 15 | `doubao-seedream-4.0` | 图片生成 | volcengine=45 | 按张 | ¥0.20/张；首张免费 |
| 16 | `doubao-seedream-3.0` | 图片生成 | volcengine=45 | 按张 | ¥0.259/张 |
| 17 | `cii-subtitle-erase`（即 xlsx 中的 `subtitle-erase` 模型，注意 new-api 的 `ModelList` 已经规范化为 `cii-subtitle-erase`） | 视频编辑 | `ciisubtitleerase` adaptor | 按视频时长 | 火山引擎 ¥0.40/分钟；第三方托管 $0.015/秒 |

> ⚠️ **重要不一致**（测试时务必包含此 case）：xlsx 表格第 32 行 model 参数值字符串是
>
> ```
> doubao-seedream-5.0-prodoubao-seedream-5.0-lite
> ```
>
> 把两个模型名拼一起了。这是上游录入问题，new-api 端会按 Distribute() 找不到匹配渠道而 fail。测试矩阵应当包含一条「**xlsx 中的 model 字符串原样透传给上游后，应当被 `/api/v1/images/generations` 拒绝（或平台侧 fail-fast）**」的回归用例，避免后期用户凭 xlsx 截图提单。

### 2.2 京东（已接入但与 cii-group 无关，此处仅列出供测试查证）

| model 参数值 | 类型 | 计费 | 单价 |
|---|---|---|---|
| `T-C-3-hq` ... `T-C-1 -2` | Claude 文本对话 | 按 Token | $3~$25 / M |
| `I-A-6` | gpt-image-1 | 按张（Token 折算） | ≈¥0.04–0.25/张 |
| `kimi-k3` | Kimi K3 | 按 Token | ¥20/¥100/¥2 |
| `jd/deepseek-v4-flash-0731` | deepseek | 按次 | ¥0.01/次 |

> 京东渠道走 `ChannelTypeAnthropic = 14`（claude）/ `ChannelTypeOpenAI = 1`（gpt）/ `ChannelTypeMoonshot = 25`（kimi）/ `ChannelTypeDeepSeek = 43`（deepseek），**BaseURL 走京东网关**，与 cii-group 完全无关——本测试方案仅覆盖京东 deepseek 渠道的「路径可访问性」回归；模型间计费/校验逻辑与仓山 deepseek 同名模型完全不同，按次 vs 按 Token 峰谷，不能混。

---

## 3. 接口行为详解（按"实现"维度）

### 3.1 Video 任务（Seedance）

**对应代码**：`relay/channel/task/doubao/`

**new-api 路由**：
- `POST /v1/video/generations`
- `POST /v1/videos`（OpenAI-compatible 别名）
- `POST /v1/videos/:video_id/remix`
- `GET /v1/video/generations/:task_id`
- `GET /v1/videos/:task_id`
- `GET /v1/contents/generations/tasks`（本地列表）
- `DELETE /v1/contents/generations/tasks/:task_id`

**关键实现要点**：

1. **请求转换**（`BuildRequestBody → convertToRequestPayload`）：
   - `req.Prompt` → `content[].type=text, text=prompt`
   - `req.Images[]` → `content[].type=image_url, image_url.url=...`
   - `req.Seconds` / `req.Duration` → `duration` 字段
   - `req.Metadata.content`（如果有图、视频、音频、工具调用）→ 通过 `taskcommon.UnmarshalMetadata` 并入 `requestPayload`
   - **拒绝混合场景**：上游明确禁止「首帧 / 首尾帧 / 全模态参考」混用，new-api 端**未做这一层校验**，请在 adapter test 里加用例断言「混用时上游会返回 400 / 异步报错」。
2. **路由 path 兼容**：`defaultDoubaoTaskApiPath = "/api/v3/contents/generations/tasks"`。若渠道走 cii-group v1，应在 `OtherSettings.TaskApiPath` 填 `/api/v1/contents/generations/tasks`（已有单测 `TestNormalizeTaskApiPath` 覆盖）。
3. **轮询状态映射**：`ParseTaskResult` 把上游 `pending/queued → TaskStatusQueued`，`running → InProgress`，`succeeded → Success`，`failed → Failure`。其余未知状态保守按 InProgress 处理（见 `adaptor.go:497-501`）。
4. **计费预估**：`EstimateBilling` 仅写 OtherRatio 字典（`video_input` × `seconds`），由平台层按 `ModelRatio × token 实际用量` 在 `AdjustBillingOnComplete` 阶段结算。
5. **取消**：`TaskAdaptor.DeleteTask` 直接调用上游 `DELETE /api/v1/contents/generations/tasks/{id}`，把状态码 + body 摘要包装后返回。

### 3.2 字幕擦除（Subtitle Erase）

**对应代码**：`relay/channel/task/ciisubtitleerase/`

**new-api 路由**：
- `POST /v1/videos/subtitle-erase/tasks`
- `GET /v1/videos/subtitle-erase/tasks/:task_id`

**关键实现要点**：

1. **请求体**：只透传 `{"video_url": "..."}`。`extractVideoURL` 兼容两种 metadata 形态：字符串或 `{video_url:{url:...}}`。
2. **状态映射**：`running → InProgress`；`completed → Success + result.video_url`；`failed → Failure + error.message/code`（注意 `success` 字段是 API 层结果，与任务状态无关——已有 `TestParseTaskResult` 覆盖）。
3. **计费**：**不预扣**。`AdjustBillingOnComplete` 按 `finished_at - created_at` 实际秒数 × 1 积分计费（`quotaPerSecond = 1`），超时或负值写 `FailReason`，由轮询循环走失败分支。
4. **幂等性**：上游支持按 API Key + 规范化请求做幂等；new-api 端未自己加去重，依赖上游幂等指纹。
5. **路由 path 默认值**：`/api/v1/videos/subtitle-erase/tasks`，可被 `OtherSettings.TaskApiPath` 覆盖。

### 3.3 素材组 / 素材 CRUD

**对应代码**：`controller/asset.go` + `service/asset/service.go`

**new-api 路由**（全部仅 TokenAuth，不挂 Distribute）：
- 素材组：`POST /v1/asset-groups`、`GET /v1/asset-groups`、`GET /v1/asset-groups/:group_id`、`PUT /v1/asset-groups/:group_id`、`DELETE /v1/asset-groups/:group_id`
- 素材：`POST /v1/asset-groups/:group_id/assets`、`GET /v1/asset-groups/:group_id/assets`、`GET /v1/assets/:asset_id`、`PUT /v1/assets/:asset_id`、`DELETE /v1/assets/:asset_id`

**关键实现要点**（参考 `dev-docs/上游素材组+素材接口示例.md`、`dev-docs/dev进度.md`）：

1. **字段名归一化**：上游 `asset-groups/list` 用 `Id`，`/asset-groups`（POST）响应却用 `AssetGroupId`，适配器层做归一化，落到 new-api 内部统一为 `group_id` / `asset_id`。
2. **OSS 同步删除**：按 dev 进度文档，本地 `asset_groups.assets` 表记录是上游缓存，写入成功后调用上游 `POST /api/v1/assets` 同步；删除走软删 + 调用上游 `DELETE`。
3. **删除组校验**：删除素材组前先校验 `COUNT(*) FROM assets WHERE group_id=? AND deleted_at=0` → 非空则 409。
4. **响应字段形变**：列表响应统一为 `{Items:[{Id,Name,...}], TotalCount}`，与上游差异在适配层抹平。

### 3.4 真人审核（Visual Validate）

**对应代码**：`controller/visual_validate.go` + `service/visualvalidate/service.go` + `model/visual_validate_session.go` + `relay/channel/task/doubao/visual_validate_adaptor.go`

**new-api 路由**：
- `POST /v1/visual-validate/sessions`
- `GET /v1/visual-validate/sessions`（本地列表，跨 token 隔离）
- `GET /v1/visual-validate/sessions/:session_id`
- `POST /v1/visual-validate/results`
- `GET /v1/visual-validate/callback?session=&sign=`（**唯一不挂 TokenAuth** 的端点，浏览器跳转）

**关键实现要点**（基于项目长期记忆 + 源码）：

1. **链路形态**：`CallbackURL` 固定指向 new-api 公开端点 `GET /v1/visual-validate/callback?session=<public_id>&sign=<token>`。`sign` token 存于会话表（`model/visual_validate_session.go`），防止伪造。
2. **认证结果回流**：浏览器回调命中后，new-api 自动调上游 `POST /api/v1/visual-validate/results`，取 `GroupId` 后通过 `service/asset.RegisterUpstreamGroup` 幂等注册为本地素材组。调用方（客户）后续直接 `GET /v1/asset-groups/:id/assets` 拉真人素材。
3. **`callback_url` 可选**：new-api 把客户传进来的 `callback_url` 落库后做 best-effort POST 通知一次（`notify_status` 字段），失败不重试。
4. **会话列表/详情**：走 new-api 本地表，**不代理上游**（上游文档未列这两个接口）；查询走 `service.ListVisualValidateSessions`。
5. **`bytedToken` 30 分钟有效**：new-api 端要在 30 分钟内完成回调与 results 拉取，否则认证失败。对应的回归用例见 §4.4。

---

## 4. 测试方案

### 4.1 通用策略：三层覆盖

| 层级 | 形态 | 触发方式 | 覆盖目标 |
|---|---|---|---|
| L1 单元 | Go test（`go test ./...`） | CI / 本地 `make test` | 纯函数：path 归一化、价格倍率、状态映射、字段归一化 |
| L2 适配器集成 | Go test + `httptest.NewServer` 自建伪上游 | CI | HTTP 边界：headers、body、auth、状态码、错误透传 |
| L3 端到端 | curl + 真实上游（仅烟测） | 手工 / 灰度 | 鉴权、限流、回调、实际计费 |

> L3 灰度强烈建议把测试链路绑在「仓山测试通道」上，不要拿真实账户跑烟测，避免误扣费。

### 4.2 测试用例矩阵（按接口 × 维度）

| 接口 | 用例号 | 场景 | 期望 |
|---|---|---|---|
| 通用 | TC-G-01 | TokenAuth 缺失或非法 `Authorization: Bearer xxx` | 401 + 统一错误体 |
| 通用 | TC-G-02 | Distribute 失败：没有可用渠道 | 503，error.code=`channel_not_available` |
| 视频创建 | TC-V-01 | happy path：model=`doubao-seedance-1.5-pro`，纯文 | 200，返回 `id` 字段，进入 `Queued` |
| 视频创建 | TC-V-02 | 单图 + 文（i2v）：model=`doubao-seedance-2.5` | 200，`content[0].type=image_url, role=first_frame`（隐式默认） |
| 视频创建 | TC-V-03 | 首尾帧：model=`doubao-seedance-2.5`，2 张图 + `role=first_frame/last_frame` | 200 |
| 视频创建 | TC-V-04 | 全模态参考：1 张参考 + 1 段参考视频 + 文 | 200，duration=-1 表示自适应 |
| 视频创建 | TC-V-05 | 互斥混用：1 张首帧图 + 1 张 reference_image | 上游 400 或新-api 端 fail-fast（**当前未拦截**，应在 adapter test 中暴露） |
| 视频创建 | TC-V-06 | 上游 `TaskApiPath` 配为 `/api/v1/...`（cii-group v1） | 200，发往 `<baseURL>/api/v1/contents/generations/tasks` |
| 视频创建 | TC-V-07 | 上游 `TaskApiPath` 含 `..` 段 | path 被归一化为默认 `/api/v3/...`（已有单测） |
| 视频创建 | TC-V-08 | `priority=5`、`service_tier=flex`、`draft=true` 仅 1.5-pro 支持 | 上游 200 / 400（按模型走） |
| 视频创建 | TC-V-09 | `duration=-1`（智能选择） | 上游 200，结算靠 `usage.completion_tokens` |
| 视频查询 | TC-V-10 | happy：状态 `running` | `TaskStatusInProgress`，progress=50% |
| 视频查询 | TC-V-11 | `succeeded` | `TaskStatusSuccess`，`metadata.url` 透传 |
| 视频查询 | TC-V-12 | `failed`，上游 `error.code=InvalidArgument` | `TaskStatusFailure`，`error.message` 透传 |
| 视频查询 | TC-V-13 | 未知 status 字符串（防御） | `TaskStatusInProgress`（保守按进行中处理） |
| 视频列表 | TC-V-14 | `filter.status=queued` | 仅返回本地 `TaskStatus{NotStart,Submitted,Queued}` |
| 视频列表 | TC-V-15 | `filter.status=expired` | 返回空（本地无对应状态），不报错 |
| 视频列表 | TC-V-16 | page_num>500 / 非法输入 | 截到 500 / 1 |
| 视频取消 | TC-V-17 | queued → DELETE | 上游 2xx，本地状态 `cancelled` |
| 视频取消 | TC-V-18 | running → DELETE | 上游 4xx（按 cii 文档规定不允许），new-api 透传 |
| 视频取消 | TC-V-19 | succeeded → DELETE | 上游 2xx（删除记录），本地状态保留 |
| 字幕擦除 | TC-S-01 | happy：第三方视频 URL，公网可访问 | 200，`task_id` 非空 |
| 字幕擦除 | TC-S-02 | `video_url` 缺失 | 400，`metadata.video_url is required` |
| 字幕擦除 | TC-S-03 | `video_url={...嵌套形态}` `{video_url:{url:"..."}}` | 200，提取正常 |
| 字幕擦除 | TC-S-04 | 平台生成视频 URL（免费通道） | 200，`success=true` |
| 字幕擦除 | TC-S-05 | `402 InsufficientBalance`（第三方） | new-api 端按 502/402 透传，error.code=InsufficientBalance |
| 字幕擦除 | TC-S-06 | `running` 状态持续轮询 | 一直 `InProgress` 直到 `completed/failed` |
| 字幕擦除 | TC-S-07 | `completed` 时 `finished_at-created_at=62s` | `AdjustBillingOnComplete` 返回 62 quota |
| 字幕擦除 | TC-S-08 | `finished_at < created_at`（异常数据） | 0 quota + `FailReason="duration invalid"` |
| 字幕擦除 | TC-S-09 | `finished_at - created_at > 4 × MaxTaskDurationSeconds` | 0 quota + `FailReason` |
| 字幕擦除 | TC-S-10 | `TaskApiPath` 注入 `..` | path 回退默认 |
| 字幕擦除 | TC-S-11 | 幂等：相同 API Key + 相同 video_url 重复提交 | 第二次返回既有 task_id，不重复扣费 |
| 素材组 | TC-A-01 | 创建 happy：`{Name, Description}` | 200，返回 `AssetGroupId`，落库 |
| 素材组 | TC-A-02 | Name 为空 | 400 |
| 素材组 | TC-A-03 | 列表：page_num/page_size | 200，Items 字段归一化为 `{Id,...}`（吞掉上游 `AssetGroupId`） |
| 素材组 | TC-A-04 | 更新：`PUT /v1/asset-groups/:id` | 200，`UpdatedAt` 更新 |
| 素材组 | TC-A-05 | 删除：组下有素材 | 409 "请先删除组内素材" |
| 素材组 | TC-A-06 | 删除：空组 | 200，软删 |
| 素材 | TC-A-07 | 创建：`POST /v1/asset-groups/:id/assets`，传入公网图片 URL | 200，返回 `AssetId`，落库 |
| 素材 | TC-A-08 | 列表：`GET /v1/asset-groups/:id/assets`，带 `Filter.Status=[Active]` | 200，仅返回 Active |
| 素材 | TC-A-09 | 详情：`GET /v1/assets/:id` | 200 |
| 素材 | TC-A-10 | 更新：`PUT /v1/assets/:id`，仅 Name | 200 |
| 素材 | TC-A-11 | 删除：软删 + 调上游 DELETE | 200，本地 `deleted_at` 写入 |
| 视觉审核 | TC-VV-01 | 创建会话 happy：`{CallbackURL}` | 200，返回 `BytedToken`（30 分钟有效）、`H5Link` |
| 视觉审核 | TC-VV-02 | `lng=zh-Hant` 通过 `H5Link` query 拼接 | H5Link 命中 `lng=zh-Hant` |
| 视觉审核 | TC-VV-03 | 回调伪造：`GET /v1/visual-validate/callback?session=X&sign=INVALID` | 401 / 400，不调上游 results |
| 视觉审核 | TC-VV-04 | 回调成功：`sign` 正确 + `resultCode=10000` | 自动 `POST /results` 拉 GroupId，注册为本地 `asset_groups` 行 |
| 视觉审核 | TC-VV-05 | 回调失败：`resultCode != 10000` | 不调 results，写入 `FailReason` |
| 视觉审核 | TC-VV-06 | `byted_token` 超过 30 分钟（伪造时间） | 拉 results 上游 4xx，session 标失败 |
| 视觉审核 | TC-VV-07 | 列表：仅本人 token 创建的会话 | 跨 token 隔离 |
| 视觉审核 | TC-VV-08 | `notify_status`：传 `callback_url`，回调后 best-effort POST 一次 | 触发 POST，状态码落库 |
| 视觉审核 | TC-VV-09 | 详情：`session_id` 不存在 | 404 |

### 4.3 关键测试用例样例（可直接落地）

#### 样例 A — 上游 path 归一化（已存在 `relay/channel/task/doubao/adaptor_test.go::TestNormalizeTaskApiPath`）

```go
func TestNormalizeTaskApiPath(t *testing.T) {
    cases := []struct {
        name string
        in   string
        want string
    }{
        {"empty falls back to default", "", defaultDoubaoTaskApiPath},
        {"ark v3 path passes through", defaultDoubaoTaskApiPath, defaultDoubaoTaskApiPath},
        {"cii v1 path passes through", "/api/v1/contents/generations/tasks", "/api/v1/contents/generations/tasks"},
        {"dot segment falls back to default", "/api/../v1/contents/generations/tasks", defaultDoubaoTaskApiPath},
        {"too long falls back to default", "/" + strings.Repeat("a", 257), defaultDoubaoTaskApiPath},
    }
    // ... assert ...
}
```

**补充建议**（可在同一个测试文件追加）：

```go
// 新增：ciisubtitleerase 同步走同样的安全规则
func TestCiiSubtitleEraseNormalizeTaskApiPath(t *testing.T) {
    // 与 doubao 单测同形态，校验 /api/v1/videos/subtitle-erase/tasks 默认值
}

// 新增：把"TaskApiPath 留空 → 走默认"+"TaskApiPath 含 ../" 一起跑
// 已覆盖；新增以下断言：
func TestTaskApiPathFallsBackOnInjection(t *testing.T) {
    bad := []string{"../../etc/passwd", "/api\\v1\\tasks", "/api:v1/tasks"}
    for _, p := range bad {
        assert.Equal(t, defaultDoubaoTaskApiPath, normalizeTaskApiPath(p))
    }
}
```

#### 样例 B — 视频任务请求/响应转换适配器集成测试（建议新增 `relay/channel/task/doubao/adaptor_integration_test.go`）

```go
package doubao

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/QuantumNous/new-api/relay/common"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// fakeCiiServer 模拟 cii-group 上游，命中 /api/v1/contents/generations/tasks 返回 task_id。
// - 含 Bearer 鉴权校验
// - 入参 model / content / duration 字段透传校验
func TestBuildRequestBody_TranslationToCiiV1(t *testing.T) {
    gin.SetMode(gin.TestMode)

    var capturedPath, capturedAuth string
    var capturedBody map[string]any

    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        capturedPath = r.URL.Path
        capturedAuth = r.Header.Get("Authorization")
        _ = json.NewDecoder(r.Body).Decode(&capturedBody)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id":"cgt-abcdef123456"}`))
    }))
    defer srv.Close()

    a := &TaskAdaptor{}
    a.Init(&common.RelayInfo{
        ChannelType:     54,
        ChannelBaseUrl:  srv.URL,
        ApiKey:          "test-key",
        OriginModelName: "doubao-seedance-2.5",
        UpstreamModelName: "doubao-seedance-2.5",
        ChannelOtherSettings: common.ChannelOtherSettings{
            TaskApiPath: "/api/v1/contents/generations/tasks",
        },
    })

    // 构造一次首帧图生视频请求
    req := common.TaskSubmitReq{
        Model:  "doubao-seedance-2.5",
        Prompt: "小猫打哈欠",
        Duration: 5,
        Images: []string{"https://cdn.example.com/cat.jpg"},
    }
    // 注入 gin ctx
    ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
    body, err := a.BuildRequestBody(ctx, &common.RelayInfo{
        OriginModelName: req.Model,
        UpstreamModelName: req.Model,
    })
    require.NoError(t, err)

    // 这里直接读 body 内容做断言，避免走真实 HTTP
    buf := new(strings.Builder)
    _, _ = buf.ReadFrom(body)
    payload := map[string]any{}
    require.NoError(t, json.Unmarshal([]byte(buf.String()), &payload))

    // TC-V-01 happy 断言：
    assert.Equal(t, "doubao-seedance-2.5", payload["model"])
    content, ok := payload["content"].([]any)
    require.True(t, ok)
    require.Len(t, content, 2, "1 张 image_url + 1 个 text")
    assert.Equal(t, "image_url", content[0].(map[string]any)["type"])
    assert.Equal(t, "https://cdn.example.com/cat.jpg", content[0].(map[string]any)["image_url"].(map[string]any)["url"])
    assert.Equal(t, "text", content[1].(map[string]any)["type"])
    assert.Equal(t, "小猫打哈欠", content[1].(map[string]any)["text"])
    assert.EqualValues(t, 5, payload["duration"])

    // TC-G-01 鉴权校验（不直接发 HTTP，复用 BuildRequestHeader 校验头）
    httpReq, _ := http.NewRequest("POST", srv.URL+"/api/v1/contents/generations/tasks", nil)
    require.NoError(t, a.BuildRequestHeader(ctx, httpReq, nil))
    assert.Equal(t, "Bearer test-key", httpReq.Header.Get("Authorization"))
    _ = capturedPath // 占位：实际跑 DoRequest 时再断言路径
}

// TC-V-05 用例：首帧与全模态参考混用应被服务端拒绝。
// new-api 端当前未拦截，要写断言：上游返回 4xx 时 new-api 返回恰当错误。
// 现状可写为"上游拒绝 → new-api 透传为 502 + upstream body 摘要"。
func TestDoResponse_TaskIDEmpty(t *testing.T) {
    // 上游返回 {"success":true,"id":""}
    // 期望 new-api 返回 invalid_response，且不进入轮询
    // ... body, err := ...
}

// TC-S-07 用例：字幕擦除按 finished_at - created_at 计费
// 见 ciisubtitleerase/adaptor_test.go
```

#### 样例 C — 字幕擦除适配器集成测试（已存在 `ciisubtitleerase/adaptor_test.go`，可参考其骨架扩展）

```go
func TestAdjustBillingOnComplete(t *testing.T) {
    body := []byte(`{
        "status": "completed",
        "result": {"video_url": "https://example.com/out.mp4"},
        "created_at": 1000, "finished_at": 1062,
        "expires_at": 999999999, "error": null
    }`)
    // 把 body 写入 task.Data，构造 model.Task...
    // 然后断言 a.AdjustBillingOnComplete(task, info) == 62
}

// TC-S-08: finished_at < created_at
body = []byte(`{"status":"running","created_at":2000,"finished_at":1000,...}`)
// 断言返回 0 + FailReason 写入 "duration invalid: negative"
```

#### 样例 D — 视觉审核回调安全（强烈建议补 L2 测试，避免回调伪造漏洞）

```go
package visualvalidate_test

func TestCallback_RejectsInvalidSign(t *testing.T) {
    // 通过 controller.VisualValidateCallback 模拟伪造回调
    // /v1/visual-validate/callback?session=xxx&sign=BAD
    // 期望 401，且没有调上游 /results
}

func TestCallback_OnSuccess_FetchesResultsAndRegistersGroup(t *testing.T) {
    // 用 httptest 上游：sessions 返回 OK；results 返回 GroupId
    // 回调命中后断言：service.asset.RegisterUpstreamGroup 被调用一次
}

func TestCallback_OnResultCodeNot10000_RecordsFailure(t *testing.T) {
    // resultCode=20000 表示认证失败
    // 期望：session.FailReason 写入，且不调 /results
}

func TestSessionListing_IsTokenIsolated(t *testing.T) {
    // 用户 A 创建 2 个会话，用户 B 创建 1 个
    // 期望：A 看不到 B 的会话
}
```

#### 样例 E — 端到端 curl（灰度环境用）

```bash
# 环境变量
API_HOST=https://newapi-gateway.example.com
TOKEN=sk-xxx

# 1) 文生视频（doubao-seedance-1.5-pro）
curl -X POST "$API_HOST/v1/videos" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "doubao-seedance-1.5-pro",
        "prompt": "小猫对着镜头打哈欠",
        "seconds": "5",
        "metadata": {
            "resolution": "720p",
            "ratio": "16:9",
            "watermark": false
        }
    }'

# 2) 图生视频（doubao-seedance-2.5 首帧）
curl -X POST "$API_HOST/v1/videos" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "doubao-seedance-2.5",
        "prompt": "城市夜景延时摄影",
        "seconds": "6",
        "images": ["https://cdn.example.com/seed.png"],
        "metadata": {
            "resolution": "1080p",
            "ratio": "adaptive",
            "generate_audio": false
        }
    }'

# 3) 查询任务
curl -X GET "$API_HOST/v1/videos/cgt-abcdef123456" \
    -H "Authorization: Bearer $TOKEN"

# 4) 取消任务（仅 queued 状态可取消）
curl -X DELETE "$API_HOST/v1/contents/generations/tasks/cgt-abcdef123456" \
    -H "Authorization: Bearer $TOKEN"

# 5) 字幕擦除
curl -X POST "$API_HOST/v1/videos/subtitle-erase/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "cii-subtitle-erase",
        "video_url": "https://cdn.example.com/source.mp4"
    }'

# 6) 创建素材组
curl -X POST "$API_HOST/v1/asset-groups" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"Name":"MyGroup","Description":"E2E test"}'

# 7) 创建视觉审核会话
curl -X POST "$API_HOST/v1/visual-validate/sessions" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"CallbackURL":"https://your-app.example.com/done"}'
```

### 4.4 重点边界 — 计费不变量

按 AGENTS.md 「Billing safety invariants」，所有涉及计费的 adaptor 必须遵守：

1. **绝不产生负扣费**：计费基数 `seconds × quotaPerSecond` 必须 `>=0`；`finished_at - created_at < 0` 走 0 扣费 + 写 FailReason。
2. **clamp 到 MaxTaskDurationSeconds**：seedance 的 seconds 倍率必须先 `if seconds > MaxTaskDurationSeconds: seconds = MaxTaskDurationSeconds`。
3. **整数溢出防御**：所有乘数 × 单价用 `common.QuotaFromFloat`（float64），不要用 int 强转。
4. **凭证幂等**：哪怕客户端传 0 元，平台层也不能因"0 优惠"反过来多扣。

---

## 5. 「完善接口文档」—— 上游 × new-api 双向对照表

> 这一节是给网关接入人员的"实现笔记"——把上游文档的字段反向映射回 new-api 暴露给客户端的字段，并标注每个字段在 new-api 内部的位置。
>
> 客户端不需要关心 new-api 内部字段；接入网关只需要按 new-api 的入口路由调，new-api 会把字段映射到上游。
>
> 上游字段名以下划线连接；new-api 字段名按 camelCase / snake_case 视具体 adaptor 而定。

### 5.1 Video 任务：`POST /v1/video/generations` / `POST /v1/videos` 等

| 上游字段（cii-api.md） | new-api 入口字段 | new-api 内部位置 | 备注 |
|---|---|---|---|
| `model`（必填） | `model`（body） | `controller.RelayTask → adaptor.requestPayload.Model` | 必填，字符串；走方舟时也需要方舟 Model ID |
| `content[].type` | `prompt` + `images[]` + `metadata.content[]` | `adaptor.convertToRequestPayload` | 平台层把 OpenAI 风格 → `content[]` |
| `content[].text` | `prompt` | 同上 | 总是追加到最后 |
| `content[].image_url.url` | `images[]` 元素 / `metadata.content[].image_url.url` | 同上 | 接受 URL / `data:image/...;base64,...` / `asset://...` |
| `content[].video_url.url` | `metadata.content[].video_url.url` | 同上 | only 2.5 / 2.0 系列 |
| `content[].audio_url.url` | `metadata.content[].audio_url.url` | 同上 | 仅 2.5 可独立，2.0 必带图/视频 |
| `content[].role` | `metadata.content[].role` | 同上 | `first_frame`/`last_frame`/`reference_image`/`reference_video`/`reference_audio` |
| `content[].draft_task.id`（1.5-pro） | `metadata.content[].draft_task.id` | 同上 | draft 模式 |
| `callback_url` | `metadata.callback_url` | 同上 | best-effort 通知 |
| `return_last_frame` | `metadata.return_last_frame` | `BoolValue` 包装 | 指针，`omitempty` |
| `service_tier` | `metadata.service_tier` | 字符串 | 仅 1.5-pro 支持 flex |
| `execution_expires_after` | `metadata.execution_expires_after` | `IntValue` 包装 | 默认 172800 |
| `generate_audio` | `metadata.generate_audio` | `BoolValue` 包装 | 默认 true |
| `draft` | `metadata.draft` | `BoolValue` 包装 | 仅 1.5-pro |
| `tools[].type` | `metadata.tools[]` | 结构体 | `web_search` |
| `safety_identifier` | `metadata.safety_identifier` | 字符串 | ≤64 |
| `priority` | `metadata.priority` | `IntValue` 包装 | [0,9] |
| `resolution` | `metadata.resolution` | 字符串 | 强校验，详见 §2.1 各模型支持 |
| `ratio` | `metadata.ratio` | 字符串 | 强校验 |
| `duration` | `seconds`（string） / `duration`（int） | `adaptor.convertToRequestPayload` line 455-457 | 秒；负数 = 智能选择 |
| `seed` | `metadata.seed` | `IntValue` 包装 | 默认 -1 |
| `camera_fixed` | `metadata.camera_fixed` | `BoolValue` 包装 | 仅 1.5-pro |
| `watermark` | `metadata.watermark` | `BoolValue` 包装 | 默认 false |
| `frames` | `metadata.frames` | `IntValue` 包装 | 与 duration 互斥 |
| `output_format` | `metadata.output_format` | 字符串 | 仅 2.5（mp4/mov） |
| `omni_reference_task_type` | `metadata.omni_reference_task_type` | 字符串 | 仅 2.5（auto/reference/edit/extend） |

**响应字段**：`id`、`status`（`queued/running/cancelled/succeeded/failed/expired`）、`content.video_url`、`content.last_frame_url`、`seed`、`resolution`、`ratio`、`duration`、`frames`、`framespersecond`、`usage.completion_tokens`、`usage.total_tokens`、`usage.tool_usage.web_search`、`error.code`、`error.message`、`created_at`、`updated_at`，全部由 `ParseTaskResult` 解析。

### 5.2 字幕擦除：`POST /v1/videos/subtitle-erase/tasks`

| 上游字段 | new-api 入口字段 | 备注 |
|---|---|---|
| `video_url`（必填） | `video_url`（body，最顶层）/ `metadata.video_url` | 接受字符串或 `{url:"..."}` |

**响应**：`success`、`task_id`、`request_id`、错误时 `error.{code,message,param,type}`。

### 5.3 素材组：`/v1/asset-groups`

| 上游 | new-api | 备注 |
|---|---|---|
| `Name`（必填） | `Name` | 字段大小写：上游 PascalCase |
| `Description` | `Description` | 可选 |
| `AssetGroupId`（响应） | `Id`（归一化） | `asset_groups.group_id` 表 |
| `GroupType`（list 响应） | `GroupType` | 透传 |
| `CreatedAt`/`UpdatedAt`（ISO8601） | `CreatedAt`/`UpdatedAt` | 本地库用 int64 unix 秒 |

### 5.4 素材：`/v1/assets` + `/v1/asset-groups/:group_id/assets`

| 上游 | new-api | 备注 |
|---|---|---|
| `AssetGroupId` | `group_id`（URL 段） | 必填 |
| `ImageUrl` | `ImageUrl` | 公网 URL |
| `AssetType`（默认 `Image`） | `asset_type` | 透传 |
| `Name` | `name` | |
| `AssetId`（响应） | `asset_id` | |

> dev-docs/上游素材组+素材接口示例.md 已记录完整 mock；接入网关时建议直接照抄那一节。

### 5.5 视觉审核：`/v1/visual-validate/*`

| 上游 | new-api 路由 | 备注 |
|---|---|---|
| `POST /api/v1/visual-validate/sessions` (`{CallbackURL, ProjectName?}`) | `POST /v1/visual-validate/sessions` | 返回 `BytedToken/H5Link/CallbackURL` |
| `POST /api/v1/visual-validate/results` (`{BytedToken, ProjectName?}`) | `POST /v1/visual-validate/results`（**由回调自动触发**，也可手动调用） | 返回 `GroupId` |
| —— | `GET /v1/visual-validate/sessions` | 仅本地列表，不代理 |
| —— | `GET /v1/visual-validate/sessions/:session_id` | 仅本地详情 |
| —— | `GET /v1/visual-validate/callback?session=&sign=` | new-api 自己的端点；浏览器跳转回调入口 |

> 客户端工作流（接入网关参考）：
>
> 1. 调 new-api `POST /sessions` 拿 H5Link + BytedToken。
> 2. 把 H5Link 发给终端用户（前端 iframe / webview）。
> 3. 浏览器在认证成功后跳到 `new-api /v1/visual-validate/callback?session=<your_session_id>&sign=<new-api 签发 token>`，并由 new-api 触发后端拉 results + 注册素材组。
> 4. 业务侧轮询 `GET /v1/visual-validate/sessions/:session_id`，等 `Status=success` 后取 `group_id`，再调 `GET /v1/asset-groups/:group_id/assets` 拿真人素材。
>
> 注：`callback_url` 是 new-api 在转发 sessions 时原本期望的字段（参考上一轮记录的"透传 + 落库"形态），平台层目前用 `CallbackURL` 字段名（PascalCase，与上游一致）传入。

---

## 6. 接入疏漏与下一步行动

### 6.1 未对接：`POST /app-api/v1/images/generations`（Seedream 文生图/组图）

- **xlsx 中已有 5 个模型**（seedream-3.0/4.0/4.5/5.0-lite/5.0-pro），但 new-api 中**没有专属 adaptor**把它们打到 cii-group 的 `app-api/v1/images/generations` 端点。
- **当前路径**：xlsx 的 seedream 模型经 `volcengine` 渠道（type=45）走方舟的 `/api/v3/images/generations`（见 `relay/channel/volcengine/adaptor.go:269`）。
- **影响**：客户用 cii-group 渠道时，Seedream 走方舟，成本/字段不一定是 cii 文档承诺；新接入的 `seedream-5.0-pro` 等若方舟不支持，就会变 404。
- **下一步**：
  1. 新增 `relay/channel/task/doubao/imageadaptor`（或归并到 doubao adaptor 下），覆盖 `POST /app-api/v1/images/generations` 的请求/响应转换，包含 `sequential_image_generation`/`sequential_image_generation_options.max_images`/`tools`/`response_format`/`watermark` 等独有字段。
  2. 把 `doubao-seedream-5.0-pro`、`doubao-seedream-5.0-lite` 加入 doubao 的 `ModelList`。
  3. xlsx 32 行的 `doubao-seedream-5.0-prodoubao-seedream-5.0-lite` 字符串错配，必须视为两个独立行，并校验「client 发错时如何 fail-fast」。

### 6.2 未对接：`POST /app-api/3rd/v1beta/models/{model}:generateContent`（Gemini 转发）

- xlsx 中没有列 doubao-3.1-flash-image-preview / banana 系列，但是 cii 文档（§Gemini 转发）把这些模型作为代理接入了 Gemini 协议。new-api 现在的 Gemini 协议实现见 `relay/channel/gemini` 与 `relay/channel/advancedcustom` 适配器，**没有适配 cii-group 的 `app-api/3rd/...` 前缀**。
- **下一步**：评估是否要为 cii-group Gemini 转发另开 adaptor，或在 `volcengine` adaptor 下加 path prefix；建议先与业务方确认"是否真的有 Gemini 转发需求"。

### 6.3 已对接但存在风险的边界

| 风险点 | 现状 | 建议 |
|---|---|---|
| `requestPayload.Model` 与 `UpstreamModelName` 的不一致 | 代码 line 234-238：若 `IsModelMapped` 则覆盖，否则保持 body.Model | 抽象出 `ResolveUpstreamModelName(req)`，便于单测覆盖 model 映射规则 |
| `command/seedance` adaptor 中 `duration` / `frames` 互斥 | docs 提示二者互斥，但代码未校验 | 在 `convertToRequestPayload` 加 early-return：不能同时给 |
| asset CRUD 软删后上游列表残留 | dev 进度文档曾标注 "OSS 才是事实源" | 加一条 L2 测试：删除素材后立即再 list，本地空、上游 200/404（取决于上游实现） |
| visual-validate 回调伪造保护 | 已用 sign token 校验，但未有专门的测试 | 必须补 `TestCallback_RejectsInvalidSign` |
| doubao adaptor 默认 path 是 v3（方舟），新接入仓山渠道时容易踩坑 | 单测已覆盖 path 归一化，但**没有端到端确认** | 在 channel-test 列表里加一条「仓山渠道，ExpectedPath=/api/v1/contents/generations/tasks」 |

---

## 7. 仓库落点参考（接入改源码时打开的文件）

| 关注点 | 路径 |
|---|---|
| 渠道常量 | `constant/channel.go` |
| doubao 视频任务 adaptor | `relay/channel/task/doubao/adaptor.go` |
| doubao 价格表 / 模型列表 | `relay/channel/task/doubao/constants.go` |
| doubao adaptor 单测 | `relay/channel/task/doubao/adaptor_test.go` |
| doubao 视觉审核 adaptor | `relay/channel/task/doubao/visual_validate_adaptor.go` 及其 test |
| 字幕擦除 adaptor | `relay/channel/task/ciisubtitleerase/adaptor.go` |
| 字幕擦除 adaptor 单测 | `relay/channel/task/ciisubtitleerase/adaptor_test.go` |
| 素材 CRUD controller | `controller/asset.go` |
| 素材 service | `service/asset/service.go` |
| 视觉审核 controller | `controller/visual_validate.go` |
| 视觉审核 service | `service/visualvalidate/service.go` |
| 视觉审核 session 模型 | `model/visual_validate_session.go` |
| 内容任务 controller（list/cancel） | `controller/task_content.go` + test `controller/task_content_test.go` |
| 路由总表 | `router/video-router.go` |
| 渠道图标（前端） | `web/src/features/channels/lib/__tests__/cii-subtitle-erase-channel.test.ts` |

---

## 8. 测试如何跑

### 8.1 单元 / 集成（CI 必跑）

```bash
# 仓山 / doubao 相关
go test ./relay/channel/task/doubao/... -run 'TestNormalizeTaskApiPath|TestGetVideoInputRatio|TestEstimateBilling_Duration|TestVisualValidate' -v
go test ./relay/channel/task/ciisubtitleerase/... -v

# 内容任务 controller（ListContentTasks / CancelContentTask / status 映射）
go test ./controller/... -run 'TestContent|TestBuildContentTaskItem' -v
```

### 8.2 端到端（手动 / 灰度）

参照 §4.3 样例 E，绑定 `assets/.env` 中的灰度 token，并在仓山测试通道接入，避免误扣真实配额。

### 8.3 覆盖率目标

- adaptor 单元测试覆盖率 ≥ 80%（`go test -cover ./relay/channel/task/...`）
- 视觉审核 sign 校验分支 **必须** 100%
- §6.1 / §6.2 未对接接口覆盖率 **N/A**，不得凑数

---

> 维护说明：本文档应当与 `dev-docs/cii-api.md`、`dev-docs/模型定价.xlsx`、`dev-docs/doubao 适配器对 Seedance 各变体的计费与参数校验.md` 同步刷新：
>
> - 上游新增字段 → §3 / §5 同步；
> - new-api 新增模型 → §2 同步，并在 §1 表格里追加"已接入模型"列；
> - 新增 adaptor → §6.3 与 §7 同步；
> - 计费单价变更 → §2 与 §4.4"计费不变量"同步；
>
> 这三个文档共同构成"new-api 接入仓山 / cii-group"的完整知识库。
