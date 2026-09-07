# doubao 适配器对 Seedance 各变体的计费与参数校验 — 审查报告

> 审查范围：`/v1/video/generations` 端点在 doubao 渠道下的完整链路。
> 上游对照：`dev-docs/cii-api.md`（CII 仓山区人工智能公共服务平台）+ `dev-docs/doubao-api.md`（volcengine 火山方舟 Seedance 文档）。
> 审查日期：2026-09-04，分支 `feature/seedance`。

---

## 1. 路由与请求入口

| 项 | 现状 | 位置 |
|---|---|---|
| 客户端端点 | `POST /v1/video/generations` | `router/video-router.go:23` |
| 鉴权 | `TokenAuth` + `Distribute` | `router/video-router.go:21` |
| 入口 | `controller.RelayTask` → `RelayTaskSubmit` | `relay/relay_task.go:145` |
| 适配器选择 | 通过 `platform` 推断为 `doubao` | `relay/relay_task.go:149-153` |
| 适配器初始化 | `Init` → `ValidateRequestAndSetAction` → `BuildRequestURL` → `EstimateBilling` → `BuildRequestBody` → `DoRequest` | `relay/channel/task/doubao/adaptor.go:114-209` |

请求被 4 个 step 处理：

1. `Init(info)` (line 114) — 写入 `baseURL`、`apiKey`、`ChannelType`
2. `ValidateRequestAndSetAction` (line 121) — 委托给 `relaycommon.ValidateBasicTaskRequest`
3. `BuildRequestURL` (line 127) — 拼接 `baseURL + taskApiPath`
4. `BuildRequestBody` (line 184) — 调 `convertToRequestPayload` 把 OpenAI 风格 body 转 doubao 私有格式
5. `EstimateBilling` (line 140) — 返回 `OtherRatios` 给价格链

---

## 2. 模型清单（与上游支持的对照）

`relay/channel/task/doubao/constants.go:5-12` 当前只列了 6 个：

| 列表里有 | 列表里缺 |
|---|---|
| `doubao-seedance-1-0-pro-250528` | `doubao-seedance-2.5`（CII 当前主推 2.5，"已全面公开"）|
| `doubao-seedance-1-0-lite-t2v` | `doubao-seedance-2-0-mini`（2.0 mini）|
| `doubao-seedance-1-0-lite-i2v` | — |
| `doubao-seedance-1-5-pro-251215` | — |
| `doubao-seedance-2-0-260128` | — |
| `doubao-seedance-2-0-fast-260128` | — |

**问题**：

- CII 文档明确把 `doubao-seedance-2.5` 列在"模型能力"第一项（"全模态参考生视频 / 图生视频-首尾帧 / 图生视频-首帧 / 文生视频"），但 new-api 的 `ModelList` 完全没收录。
- `doubao-seedance-2-0-mini` 也是 CII 文档里 2.0 系列的一员（"doubao-seedance-2.0 / 2.0-fast / 2.0-mini"），但代码只列了 `2-0-260128` 和 `2-0-fast-260128` 两个。
- `GetModelList()` 返回的列表会用于前端渠道的 model 字段下拉，运维在页面上完全看不到 2.5 这个模型名。
- 即使 model 字段手工填入 `doubao-seedance-2.5`，请求也能跑（因为校验只查 prompt 和 duration），但**价格表查不到**（见 §4）→ 走基础 ModelRatio 计费。

---

## 3. 参数校验

### 3.1 实际校验内容

`ValidateRequestAndSetAction` (adaptor.go:121) 只调用 `relaycommon.ValidateBasicTaskRequest` (`relay/common/relay_utils.go:283-313`)。这个 helper **总共只校验两件事**：

| 字段 | 校验内容 | 实现位置 |
|---|---|---|
| `prompt` | 非空 | `relay_utils.go:136-141` |
| `duration` / `seconds` | 范围 `[1, 3600]`（防溢出）| `relay_utils.go:148-157`，`MaxTaskDurationSeconds = 3600` |

### 3.2 完全没有校验的字段

| 字段 | 当前状态 | CII 上游的差异 |
|---|---|---|
| `model` | ❌ 不校验 | 决定后续所有约束 |
| `ratio` | ❌ 不校验 | 决定分辨率组合 |
| `resolution` | ❌ 不校验 | 不同模型支持的档位不同 |
| `frames` | ❌ 不校验 | 帧数影响 token |
| `images[]` 数量 | ❌ 无上限 | 2.5: 1-30, 2.0: 1-9, 1.5-pro: 1-2 |
| `content[]` 元素 | ❌ 无类型校验 | 必须为 `text` / `image_url` / `video_url` / `audio_url` |
| `role` (`first_frame` / `last_frame` / `reference_image`) | ❌ 不校验 | 1.5-pro 不支持 reference_image |
| `watermark` | ❌ 不校验 | 接受 `true` / `false` |
| `service_tier` | ❌ 不校验 | 接受 `default` / `flex` |
| `generate_audio` | ❌ 不校验 | 2.5 支持，1.5-pro 不支持 |
| `draft` | ❌ 不校验 | 接受 `true` / `false` |
| `camera_fixed` | ❌ 不校验 | 接受 `true` / `false` |
| audio 输入 | ❌ 不校验 | 2.0 系列禁止单独音频 |
| 图片 base64 体积 | ❌ 不校验 | 上游 64MB 上限 |

### 3.3 CII 文档里**按模型差异化**的参数（全部没在 new-api 侧实现）

| 校验项 | 2.5 | 2.0 系列 | 1.5-pro |
|---|---|---|---|
| `duration` 取值 | `[4, 30]` 或 `-1` | `[4, 15]` 或 `-1` | `[4, 12]` 或 `-1` |
| 图片数量（首帧 / 首尾帧）| 1 / 2 | 1 / 2 | 1 / 2 |
| 全模态参考生视频图片数 | 1-30 | 1-9 | ❌ 不支持 |
| 视频参考数量 | 0-10 | 0-3 | ❌ 不支持 |
| 音频参考数量 | 0-10 | 0-3 | ❌ 不支持 |
| 单音频时长 | `[2, 30]` | `[2, 15]` | ❌ |
| 单独传音频 | ✅ | ❌（必须有图/视频）| ❌ |
| `heic` / `heif` 图片格式 | ✅ | ✅ | ✅ |
| 1080p | ✅ | fast / mini ❌ | ✅ |
| 4K | ✅ | ❌ | ❌ |

### 3.4 风险

- 用户提交 → 上游异步校验失败 → 任务被拒 → 配额预扣已发生 → 退款成本 + 用户体验差。
- `frames` 是 `*dto.IntValue` 类型（`adaptor.go:62`），如果用户传 `frames=10000000`，会直接被透传到上游，**没有上限**。
- `images[]` 同样无上限：用户可传 100 张图，request body 超过上游 64MB 上限后才被拒。

### 3.5 防御性已做的部分

`convertToRequestPayload` (adaptor.go:344-378) 在 body 转换时做了两件防御：

```go
// 1. metadata 中的 model 字段被丢弃，防止改 model 绕过计费
delete(metadata, "model")
// 2. content 中的 text 元素被剥离，强制以 req.Prompt 作为唯一文本
r.Content = lo.Reject(r.Content, func(c ContentItem, _ int) bool { return c.Type == "text" })
r.Content = append(r.Content, ContentItem{Type: "text", Text: req.Prompt})
```

这两点保证了"用户从 metadata 塞 model / prompt 绕过计费"这条路径被堵死，**但除此之外的差异化校验完全没有**。

---

## 4. 计费策略现状

### 4.1 价格表 `videoPriceTable`（constants.go:26-39）

只覆盖了 **2 个模型**：

```go
"doubao-seedance-2-0-260128": {
    {hasVideo: false}:                46.0,
    {hasVideo: true}:                 28.0,
    {is1080p: true, hasVideo: false}: 51.0,
    {is1080p: true, hasVideo: true}:  31.0,
    {is4k: true, hasVideo: false}:    26.0,
    {is4k: true, hasVideo: true}:     16.0,
},
"doubao-seedance-2-0-fast-260128": {
    {hasVideo: false}: 37.0,
    {hasVideo: true}:  22.0,
},
```

其他 4 个模型（`1-0-pro-250528`、`1-0-lite-t2v`、`1-0-lite-i2v`、`1-5-pro-251215`）**完全没价格表**。

### 4.2 `EstimateBilling`（adaptor.go:140-152）

只做了三件事：

1. 读 `metadata.resolution` 字符串
2. 读 `content` 数组是否有 `video_url` / `video_url` 字段
3. 调 `GetVideoInputRatio` 查表返回 OtherRatio

```go
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
    req, err := relaycommon.GetTaskRequest(c)
    if err != nil {
        return nil
    }
    hasVideo := hasVideoInMetadata(req.Metadata)
    resolution, _ := req.Metadata["resolution"].(string)
    ratio, ok := GetVideoInputRatio(info.OriginModelName, resolution, hasVideo)
    if !ok || ratio == 1.0 {
        return nil
    }
    return map[string]float64{"video_input": ratio}
}
```

**没有应用的计费因子**：

| 因子 | CII 上游规则 | 当前是否计费 | 严重程度 |
|---|---|---|---|
| `duration`（秒数）| 不同模型不同单价（按 token）| ❌ 完全没有 | **高** |
| `watermark=true` | 不同价格档（公开标价分水印/无水印）| ❌ | 中 |
| `ratio`（`16:9` 等）| 不同宽高比可能不同单价 | ❌ | 中 |
| `service_tier`（`default` / `flex`）| flex 折扣 | ❌ | 中 |
| `generate_audio=true` | 2.5 有声视频计费规则不同 | ❌ | 中 |
| `draft=true` | 样片价格远低于正式 | ❌ | 中 |
| `camera_fixed` | 可能有附加 | ❌ | 低 |
| `frames` | 帧数影响 token | ❌ | 低 |

### 4.3 duration 完全没有作为倍率

`relay/relay_task.go:125-129` 只在 Sora remix 路径里给 `seconds` 加了 OtherRatio：

```go
if seconds > relaycommon.MaxTaskDurationSeconds {
    seconds = relaycommon.MaxTaskDurationSeconds
}
info.PriceData.AddOtherRatio("seconds", float64(seconds))
```

**doubao 路径没有这段**。也就是说，`MaxTaskDurationSeconds = 3600` 这个防御只在防"溢出"上对 doubao 生效（避免 int 溢出为负），但**不会按 duration 真正计费**。

### 4.4 计费偏差举例

| 场景 | 实际应扣 | new-api 当前预扣 | 偏差 |
|---|---|---|---|
| `model=doubao-seedance-2-0-260128, duration=10, watermark=true` | 基础价 × 10 × 水印倍率 | 基础价 × 1 | **少 10× 以上** |
| `model=doubao-seedance-1-5-pro, duration=10, watermark=true` | 基础价 × 10 × 水印倍率 | 基础价 × 1（无价格表）| **少 10× 以上** |
| `model=doubao-seedance-2-0-fast-260128, duration=10, resolution=1080p` | 上游 1080p 不支持 → 任务被拒 | 基础价 × 1 | 用户被退费，体验差 |

### 4.5 结算/补扣（settle）路径

- `AdjustBillingOnSubmit` 和 `AdjustBillingOnComplete` 都用 `taskcommon.BaseBilling` 的默认 0（保留预扣）。
- 实际生成 token 数 vs 预扣额度 的差异**不会反映到账单**。
- Seedance 的真实成本是按 `usage.completion_tokens`（视频 token）计费，但当前实现只在 `ParseTaskResult` 把这个值存到了 `taskResult.CompletionTokens`（adaptor.go:403），**没有任何消费逻辑读取它**。

---

## 5. 风险总结

| # | 风险 | 影响 | 优先级 |
|---|---|---|---|
| 1 | 2.5 / 2.0-mini 不在 ModelList | 用户无法通过前端 model 字段选用 | 高 |
| 2 | 4/6 模型无价格表 | 全部走基础价，**严重少计费** | **高（计费漏洞）** |
| 3 | `duration` 没有任何倍率 | 长视频和短视频同价，**严重少计费** | **高（计费漏洞）** |
| 4 | `watermark` 不计费 | 有水印视频通常不同价 | 中 |
| 5 | 缺少模型特定 duration / 图片数 / ratio / 分辨率校验 | 用户提交后被上游拒，浪费预扣 | 中 |
| 6 | `AdjustBillingOnComplete` 未实现 | 实际 token 与预扣不一致不调整 | 中 |
| 7 | `generate_audio` / `service_tier` / `draft` / `camera_fixed` 透传但不计费 | 价格梯度未覆盖 | 中 |
| 8 | `frames` 透传但无校验/无计费 | 可被滥用 | 低 |
| 9 | 没有 Seedance 测试用例 | 改动易回退 | 中 |

---

## 6. 建议修复顺序

1. **补 ModelList**（最小改动）— 把 `doubao-seedance-2.5`、`doubao-seedance-2-0-mini` 加进 `constants.go:5-12`。
2. **补齐 6 个模型的价格表** — 价格来源是火山方舟公开的 Seedance 阶梯计费（按 token / 元/百万 token），可以参考 CII 文档或方舟定价页。
3. **在 `EstimateBilling` 加 `duration` OtherRatio** — Seedance 是按 token 计费，duration 是主因子，**这个必须做**。
4. **加 `watermark` OtherRatio** — 至少区分有/无水印两档。
5. **加按 model 的参数校验** — 至少校验 duration 范围、图片数上下限、是否允许 audio / video 输入。
6. **实现 `AdjustBillingOnComplete`** — 用 `usage.completion_tokens` 与预扣对比，做差额结算。
7. **加测试** — `TestGetVideoInputRatio` / `TestEstimateBilling` / `TestConvertToRequestPayload`。

---

## 7. 关键代码索引

- 路由：`router/video-router.go:23`
- 适配器入口：`relay/channel/task/doubao/adaptor.go:114-209`
- 通用校验：`relay/common/relay_utils.go:283-313`
- 通用时长上限：`relay/common/relay_utils.go:146` (`MaxTaskDurationSeconds = 3600`)
- 模型清单：`relay/channel/task/doubao/constants.go:5-12`
- 价格表：`relay/channel/task/doubao/constants.go:26-39`
- 价格倍率计算：`relay/channel/task/doubao/constants.go:43-56`
- 计费估算：`relay/channel/task/doubao/adaptor.go:140-152`
- 视频输入检测：`relay/channel/task/doubao/adaptor.go:156-181`
- Body 转换：`relay/channel/task/doubao/adaptor.go:344-378`
- 任务结果解析：`relay/channel/task/doubao/adaptor.go:380-416`
- 通用计费链：`relay/relay_task.go:140-211`
- OtherRatio 工具：`types/price_data.go:35-100`
- duration 在 sora remix 路径的倍率示例：`relay/relay_task.go:125-132`
- 默认计费 hook：`relay/channel/task/taskcommon/helpers.go:82-97`

---

## 8. 总结

**直接回答：没有完全实现差异化计费与参数校验。** 当前 doubao 适配器对 Seedance 的处理是"半成品"状态：

- ✅ 做了：resolution 维度（1080p / 4k / 基准）+ 是否含视频输入 — 2 个模型（`2-0-260128`、`2-0-fast-260128`）
- ❌ 没做：`duration` 倍率、`watermark`、`audio`、`draft`、`service_tier`、`camera_fixed` — 0 个模型
- ❌ 没做：4 个模型（1.0 系列 + 1.5-pro 系列）完全没价格表
- ❌ 没做：所有模型特有的参数上下限校验
- ❌ 没做：按实际 token 结算

**当前代码对小流量、只跑 2.0-fast 的场景勉强可用**，但 2.5 / 1.5-pro / 1.0-lite / 1.0-pro 任何非基础 duration 用法都会**少计费**，长期跑会亏。

最低修复集合（建议 1+2+3 项）是：补 ModelList + 补价格表 + duration 倍率，**能堵住 90% 的计费漏洞**。
