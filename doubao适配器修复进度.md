完成

  按照 dev-docs/doubao 适配器对 Seedance 各变体的计费与参数校验.md 末尾"建议修复顺序"的前三项，完成了如下改动：

  1. 补 ModelList（relay/channel/task/doubao/constants.go:5-14）

  新增 doubao-seedance-2-0-mini、doubao-seedance-2.5 两条；前端的 model 字段下拉会立即出现这两个模型名。

  2. 补齐 6 个模型的价格表（constants.go:32-77）

  为 2-0-mini / 2.5 / 1-5-pro-251215 / 1-0-pro-250528 / 1-0-lite-t2v / 1-0-lite-i2v 补齐 (分辨率档 × hasVideo) 组合价；结构与 2.0/2.0-fast 保持一致，按 CII 文档已知的分辨率支持范围（mini 仅 480p/720p；2.5 到 1080p 不含 4k；4k 仅 2.0）做精简。

  ▎ 重要：因为我无法离线访问 https://www.volcengine.com/docs/82379/1099320 ，这些单价值是按公开文档 + 现有 2.0/2.0-fast 结构的合理推断；部署前需在方舟控制台校对一次。已在表头加注释明确这一点。

  3. EstimateBilling 加 seconds OtherRatio（adaptor.go:139-190）

  - 抽出 resolveTaskSeconds helper，优先 req.Duration、兼容 req.Seconds（与 convertToRequestPayload 中构造上游 duration 的取数策略保持一致）。
  - 仅在客户端提供正数 duration 时写入 seconds 倍率；duration=-1（智能选择）或不传时不写入，留给后续 AdjustBillingOnComplete 用 usage.completion_tokens 真实值结算。
  - 与 sora remix 路径同款防御：作为计费乘数前钳制到 relaycommon.MaxTaskDurationSeconds（3600），避免 int 溢出。
  - video_input 与 seconds 复合（相乘），符合 OtherRatioMultiplier 的语义。

  4. 单元测试（adaptor_test.go）

  新增两组测试，全部通过：
  - TestGetVideoInputRatio：覆盖 19 条 case，包括所有 8 个模型在 480p/1080p/4k + hasVideo 组合下的倍率，含"未配置组合回退到基准价"和"未知模型"两个边界。
  - TestEstimateBilling_Duration：8 条 case，覆盖以下关键契约 ——
    - 缺省/无倍率时返回 nil
    - 正数 duration 写入 seconds
    - video_url 触发 video_input 复合
    - resolution=1080p 触发正确的 video_input 倍率
    - Seconds 字符串字段被识别
    - seconds="-1" / "auto" 等无效值不写入倍率
    - 1_000_000 的超大值被钳制到 MaxTaskDurationSeconds=3600

  go build ./... 与 go test ./relay/channel/task/doubao/ 全部通过。

  后续如果继续做 dev-doc 第 4-7 项（watermark 倍率、模型特定参数校验、AdjustBillingOnComplete 实现），可以基于这次搭好的 OtherRatio 管线继续叠加。
