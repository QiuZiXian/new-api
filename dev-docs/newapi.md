分析一下视频状态查询接口的逻辑/v1/video/generations/{task_id} ；是否有和上游交互的逻辑                                                                                                                                                            

  Thought for 17s, searched for 5 patterns, read 4 files (ctrl+o to expand)

● GET /v1/video/generations/{task_id} 状态查询接口分析
     
  调用链路
     
  HTTP 路由
  router/video-router.go:24
    videoV1Router.GET("/video/generations/:task_id", controller.RelayTaskFetch)
          │
          ▼
  controller.RelayTaskFetch                  controller/relay.go:477
    GenRelayInfo → relay.RelayTaskFetch(...)
          │
          ▼
  middleware.Distribute                      middleware/distributor.go:322-338
    GET → relayMode = RelayModeVideoFetchByID
          shouldSelectChannel = false   // 不重选渠道,任务已绑定 channel
          modelRequest.Model = getTaskOriginModelName(c)  // #4834 回填原模型名
          │
          ▼
  relay.RelayTaskFetch                       relay/relay_task.go:296
    按 relayMode 派发到 videoFetchByIDRespBodyBuilder
          │
          ▼
  videoFetchByIDRespBodyBuilder              relay/relay_task.go:371

  核心逻辑 (relay/relay_task.go:371-425)

  1. taskId := c.Param("task_id")
  2. originTask := model.GetByTaskId(userId, taskId)   // 查本地 DB
  3. if !exist → 400 task_not_exist

  4. isOpenAIVideoAPI := strings.HasPrefix(c.Request.RequestURI, "/v1/videos/")
     // /v1/videos/:task_id 走 OpenAI Video API 格式
     // /v1/video/generations/:task_id 走通用 TaskDto 格式

  5. if realtimeResp := tryRealtimeFetch(originTask, isOpenAIVideoAPI); len > 0
        → 走实时分支

  6. else if isOpenAIVideoAPI
        → adaptor.ConvertToOpenAIVideo(originTask)   // 仅做字段映射,无网络 IO

  7. else
        → common.Marshal(TaskModel2Dto(originTask))  // 直接返回本地快照

  关键结论:是否与上游交互?

  默认不与上游交互,只有 Gemini/Vertex 渠道会实时调上游。

  ┌─────────────────────────────────────────────────────────────┬───────────────┬───────────────────────────────────────────────────────────────┐
  │                          渠道类型                           │   上游交互?   │                           状态来源                            │
  ├─────────────────────────────────────────────────────────────┼───────────────┼───────────────────────────────────────────────────────────────┤
  │ Gemini / Vertex AI                                          │ ✅ 是         │ tryRealtimeFetch 主动 HTTP 调用                               │
  ├─────────────────────────────────────────────────────────────┼───────────────┼───────────────────────────────────────────────────────────────┤
  │ Doubao / Kling / Sora / Vidu / Jimeng / Suno / Hailuo / Ali │ ❌ 否         │ 仅读本地 tasks 表,状态由 service/task_polling.go 后台轮询更新 │
  ├─────────────────────────────────────────────────────────────┼───────────────┼───────────────────────────────────────────────────────────────┤
  │ OpenAI Video 兼容渠道 (/v1/videos/:task_id)                 │ ❌ 否(转换层) │ adaptor.ConvertToOpenAIVideo 只做字段映射                     │
  └─────────────────────────────────────────────────────────────┴───────────────┴───────────────────────────────────────────────────────────────┘

  唯一会调上游的分支: tryRealtimeFetch (relay/relay_task.go:430-509)

  触发条件严格限制:

  if channelModel.Type != constant.ChannelTypeVertexAi &&
     channelModel.Type != constant.ChannelTypeGemini {
      return nil   // 其他渠道直接跳过,不调上游
  }

  resp, err := adaptor.FetchTask(baseURL, channelModel.Key, map[string]any{
      "task_id": task.GetUpstreamTaskID(),
      "action":  task.Action,
  }, proxy)

  实际行为:

  1. 取出 task.GetUpstreamTaskID()(提交时上游返回的真实任务 ID,不是 new-api 的本地 task_id)
  2. 调对应 adaptor 的 FetchTask 拿到上游最新 *http.Response
  3. adaptor.ParseTaskResult(body) 解析为统一 TaskInfo
  4. 把上游状态写回本地 tasks 表:
    - task.Status ← ti.Status
    - task.Progress ← ti.Progress
    - task.PrivateData.ResultURL ← ti.Url(data: URI 例外)
  5. OpenAI Video API 路径只更新 DB,然后 return nil 让外层走 ConvertToOpenAIVideo 输出
  6. 通用路径则用 {format, status, task_id, url, metadata} 自定义包装返回

  旁路:其他渠道的状态从哪里来?

  service/task_polling.go 后台轮询任务会周期性地对非 Gemini/Vertex 渠道(每个 adaptor 自己实现 FetchTask + ParseTaskResult)主动调上游拉状态并更新 tasks 表——所以"用户拉"和"系统轮询"是两套独立的状态刷新机制:

  - 用户拉(GET /v1/video/generations/:task_id):仅 Gemini/Vertex 实时
  - 系统轮询(service/task_polling.go):覆盖 Doubao/Kling/Sora/Vidu/Jimeng/Suno/Hailuo/Ali 等

  也就是说,如果一个 Doubao 任务轮询失败/停止,该接口只会反复返回本地过期的状态,不会自愈。这是当前实现的真实行为,不是 bug。

  几个值得注意的细节

  - shouldSelectChannel = false(distributor.go:333)—— 状态查询时不再重选渠道,直接信任任务记录里 ChannelId。
  - modelRequest.Model = getTaskOriginModelName(c)(distributor.go:334, 420-439)—— 修复 #4834:之前 GET 不解析 model 会导致 modelLimitEnable 校验误报"无权限"。
  - adaptor.FetchTask 在 relay/channel/task/{hailuo,suno,jimeng,doubao,ali,sora,kling,vidu,gemini,vertex}/adaptor.go 都有实现,但只有 gemini/vertex 的实现会在本接口的路径里被触发;其他 adaptor 的 FetchTask 只被 service/task_polling.go 消费。

好的，继续分析一下视频任务创建逻辑/v1/video/generations                                                                                                                                                                                           

  Thought for 21s, searched for 7 patterns, read 6 files (ctrl+o to expand)

● POST /v1/video/generations 视频任务创建接口分析

  1. 路由注册

  // router/video-router.go:23
  videoV1Router.POST("/video/generations", controller.RelayTask)

  中间件链:RouteTag("relay") → TokenAuth() → Distribute() → 控制器

  2. 完整调用链

  Distribute()                                    middleware/distributor.go:33
    ├── getModelRequest                          middleware/distributor.go:253
    │     • POST /v1/video/generations → 解析 JSON/form body 里的 model
    │     • relayMode = RelayModeVideoSubmit
    │     • shouldSelectChannel = true         ← 创建任务需要选渠道
    ├── token model limit 校验
    ├── ChannelAffinity 亲和性(若启用)
    ├── getChannel() → 按 model 选可用渠道, 绑定 key/baseURL
    └── (失败则 4xx, 成功则继续)

  Controller.RelayTask                           controller/relay.go:492
    ├── GenRelayInfo
    ├── ResolveOriginTask (remix 专用)            relay/relay_task.go:38
    ├── 重试循环 (RetryTimes):
    │     ├── getChannel / LockedChannel
    │     ├── GetBodyStorage (body 复用)
    │     ├── RelayTaskSubmit                    relay/relay_task.go:145
    │     │     ├── InitChannelMeta (回填 ApiKey/BaseURL/ChannelType)
    │     │     ├── GetTaskPlatform → GetTaskAdaptor
    │     │     ├── adaptor.Init(info)
    │     │     ├── ValidateRequestAndSetAction
    │     │     ├── ModelMappedHelper (应用渠道模型映射)
    │     │     ├── ModelPriceHelperPerCall (基础价)
    │     │     ├── adaptor.EstimateBilling (估算 OtherRatios)
    │     │     ├── PreConsumeBilling (预扣费)
    │     │     ├── adaptor.BuildRequestBody
    │     │     ├── adaptor.DoRequest → channel.DoTaskApiRequest  ← 真实上游调用
    │     │     ├── adaptor.DoResponse (解析 task_id)
    │     │     └── adaptor.AdjustBillingOnSubmit (按上游响应校正)
    │     └── 失败则换渠道/换 key 重试
    └── 成功: SettleBilling + LogTaskConsumption + task.Insert()

  3. 关键环节

  3.1 Distribute 中间件(distributor.go:322-338)

  } else if strings.Contains(c.Request.URL.Path, "/v1/video/generations") {
      if c.Request.Method == http.MethodPost {
          req, _ := getModelFromRequest(c)              // 读 body 里的 model/group
          modelRequest.Model = req.Model
          relayMode = relayconstant.RelayModeVideoSubmit
      }
      c.Set("relay_mode", relayMode)
  }

  注意:

  - shouldSelectChannel = true:必须选择渠道(不是 remix/remix-by-id 那种继承原任务的)。
  - POST 路径没有像 GET 那样回填原模型名,因为没有 task_id 路径参数。
  - getModelFromRequest 内部区分 JSON / form / multipart(distributor.go:192)。

  3.2 渠道选择

  Distribute → getChannel()
    ├── tokenSpecificChannelId 优先(锁定单渠道)
    ├── channelSupportsRequestPath (按 path 过滤)
    ├── model.IsChannelEnabledForGroupModel (按 model+group 匹配)
    ├── service.SelectModelWithFirstAvailable 兜底轮询
    └── GetNextEnabledKey (多 key 渠道)

  channel.Type 会通过 ContextKeyChannelType 写入 context,后续 GetTaskPlatform 用它来选 adaptor:

  // relay/relay_adaptor.go:136
  func GetTaskPlatform(c *gin.Context) constant.TaskPlatform {
      channelType := c.GetInt("channel_type")
      if channelType > 0 {
          return constant.TaskPlatform(strconv.Itoa(channelType))
      }
      return constant.TaskPlatform(c.GetString("platform"))
  }

  GetTaskAdaptor 把 ChannelType 数字映射到具体 adaptor:

  ┌──────────────────────────┬────────────────────────┐
  │       ChannelType        │        Adaptor         │
  ├──────────────────────────┼────────────────────────┤
  │ 阿里 (Ali)               │ taskali.TaskAdaptor    │
  ├──────────────────────────┼────────────────────────┤
  │ Kling                    │ taskkling.TaskAdaptor  │
  ├──────────────────────────┼────────────────────────┤
  │ Jimeng                   │ taskjimeng.TaskAdaptor │
  ├──────────────────────────┼────────────────────────┤
  │ Vertex                   │ taskvertex.TaskAdaptor │
  ├──────────────────────────┼────────────────────────┤
  │ Vidu                     │ taskvidu.TaskAdaptor   │
  ├──────────────────────────┼────────────────────────┤
  │ DoubaoVideo / VolcEngine │ taskdoubao.TaskAdaptor │
  ├──────────────────────────┼────────────────────────┤
  │ Sora / OpenAI            │ tasksora.TaskAdaptor   │
  ├──────────────────────────┼────────────────────────┤
  │ Gemini                   │ taskgemini.TaskAdaptor │
  ├──────────────────────────┼────────────────────────┤
  │ MiniMax                  │ hailuo.TaskAdaptor     │
  └──────────────────────────┴────────────────────────┘

  3.3 平台默认 model 名 (relay/relay_task.go:163-167)

  modelName := info.OriginModelName
  if modelName == "" {
      modelName = service.CoverTaskActionToModelName(platform, info.Action)
  }

  CoverTaskActionToModelName 简单拼接:strings.ToLower(platform) + "_" + strings.ToLower(action),例如:

  - kling + generate → kling_generate
  - doubaovideo + generate → doubaovideo_generate

  这是请求体里没有 model 字段时的兜底值,真正的 model 名通常由请求体提供。

  3.4 验证 / 计费 / 预扣

  // 1) ValidateRequestAndSetAction
  adaptor.ValidateRequestAndSetAction(c, info)        // 例如 doubao 走 relaycommon.ValidateBasicTaskRequest
  //   └─ info.Action = constant.TaskActionGenerate

  // 2) ModelPriceHelperPerCall → info.PriceData (含 ModelPrice/ModelRatio/GroupRatio)
  priceData, _ := helper.ModelPriceHelperPerCall(c, info)

  // 3) 适配器估算 OtherRatios (视频时长/分辨率等乘数)
  if estimatedRatios := adaptor.EstimateBilling(c, info); len(estimatedRatios) > 0 {
      for k, v := range estimatedRatios { info.PriceData.AddOtherRatio(k, v) }
  }

  // 4) 应用 OtherRatios(用 QuotaFromFloatChecked 防饱和溢出)
  quotaWithRatios := info.PriceData.ApplyOtherRatiosToFloat(float64(info.PriceData.Quota))
  quota, clamp := common.QuotaFromFloatChecked(quotaWithRatios)
  info.PriceData.Quota = quota
  noteTaskQuotaClamp(info, clamp)

  // 5) 预扣费(ForcePreConsume = true)
  service.PreConsumeBilling(c, info.PriceData.Quota, info)

  3.5 上游调用 (relay_task.go:213-241)

  requestBody, _ := adaptor.BuildRequestBody(c, info)    // 平台特定 JSON
  resp, err    := adaptor.DoRequest(c, info, requestBody) // ↓
  otherRatios  := info.PriceData.OtherRatios()
  c.Header("X-New-Api-Other-Ratios", string(ratiosJSON))
  upstreamTaskID, taskData, taskErr := adaptor.DoResponse(c, resp, info)

  DoRequest 统一走 channel.DoTaskApiRequest(adaptor, c, info, requestBody),由 adaptor 提供:

  - BuildRequestURL(info) → {baseURL}{taskApiPath(info)} (路径与 model/Action 有关)
  - BuildRequestHeader(c, req, info) → 注入 Authorization: Bearer {apiKey} 等
  - 内部发 POST,返回 *http.Response

  例如 Doubao adaptor(relay/channel/task/doubao/adaptor.go:127-247):

  // BuildRequestURL
  return a.baseURL + a.taskApiPath(info), nil

  // BuildRequestHeader
  req.Header.Set("Authorization", "Bearer "+a.apiKey)

  // BuildRequestBody: 转换为豆包视频生成 API 格式
  // - model / content[] (text+image_url+video_url)
  // - ratio / resolution / duration / frames / seed / camera_fixed / watermark
  // - service_tier / priority / draft / generate_audio

  // DoResponse: 解析上游 { "id": "..." } → 返回 dResp.ID 作为 upstreamTaskID
  ov := dto.NewOpenAIVideo()
  ov.ID = info.PublicTaskID
  c.JSON(http.StatusOK, ov)         // ← 直接把 OpenAI Video 兼容对象返回给客户端
  return dResp.ID, responseBody, nil

  3.6 提交后计费调整 (relay_task.go:243-252)

  if adjustedRatios := adaptor.AdjustBillingOnSubmit(info, taskData); len(adjustedRatios) > 0 {
      if adjustedQuota, ok := recalcQuotaFromRatios(info, adjustedRatios); ok {
          finalQuota = adjustedQuota
          info.PriceData.ReplaceOtherRatios(adjustedRatios)
          info.PriceData.Quota = finalQuota
      }
  }

  AdjustBillingOnSubmit 让 adaptor 根据上游真实响应(比如 usage.completion_tokens、实际生成的时长)校正 OtherRatios,然后用 recalcQuotaFromRatios 基于基础价 × 新 ratios 重算 quota。

  3.7 成功后入库(controller/relay.go:581-607)

  if taskErr == nil {
      service.SettleBilling(c, relayInfo, result.Quota)   // 结算/差额
      service.LogTaskConsumption(c, relayInfo)           // 写 consume/task 日志

      task := model.InitTask(result.Platform, relayInfo)
      task.PrivateData.UpstreamTaskID = result.UpstreamTaskID  // 上游真实 task_id
      task.PrivateData.BillingContext = &model.TaskBillingContext{
          ModelPrice:      relayInfo.PriceData.ModelPrice,
          GroupRatio:      ...,
          ModelRatio:      ...,
          OtherRatios:     relayInfo.PriceData.OtherRatios(),
          OriginModelName: relayInfo.OriginModelName,
          PerCallBilling:  ...,
      }
      task.Quota = result.Quota
      task.Data   = result.TaskData        // 上游原始响应体
      task.Action = relayInfo.Action
      task.Insert()
  }

  任务表里保存:
  - TaskID(PublicTaskID):new-api 自己生成的对外 ID
  - PrivateData.UpstreamTaskID:上游真实 task_id,后续轮询/fetch 用
  - Data:上游原始响应体(可重建 task 状态)
  - Quota:最终计费额度
  - PrivateData.BillingContext:定价/比率/OtherRatios 快照(remix 路径里会被读出来复用)

  4. 关键不变量与边界

  - bill 不能为负:所有 quota 转换走 common.QuotaFromFloatChecked / common.QuotaRound / common.QuotaFromDecimal,饱和事件经 noteTaskQuotaClamp → attachQuotaSaturation 落到 admin_info.quota_saturation。
  - 视频时长上限:relaycommon.MaxTaskDurationSeconds,Doubao adaptor 的 EstimateBilling 和 remix 路径都强制 clamp(relay/channel/task/doubao/adaptor.go:160-162,relay/relay_task.go:125-127)。
  - OtherRatios 守卫:types.PriceData.AddOtherRatio 拒绝非正数 / NaN / +Inf(在 EstimateBilling 写回前自动校验)。
  - 预扣+结算:defer Refund 走 controller 第 510-514 行,失败路径不丢钱;ForcePreConsume = true 保证首次预扣。

  5. 总结

  POST /v1/video/generations 的真实工作流:

  1. 解析 model (JSON/form body)
  2. 选择渠道(token 限制、亲和性、按 model+group 匹配、多 key 轮换)
  3. 绑定 ApiKey/BaseURL
  4. 预扣费 (按 ModelPrice × GroupRatio × ModelRatio × ∏OtherRatios,带饱和保护)
  5. 构建平台特定请求体(Doubao / Kling / Sora / Vidu / Jimeng / Suno / Hailuo / Ali / Vertex / Gemini 各自实现 BuildRequestBody + BuildRequestURL + BuildRequestHeader)
  6. 真实 HTTP POST 到上游 (走 channel.DoTaskApiRequest,头部 Authorization: Bearer ...)
  7. 解析上游响应拿到 task_id(adaptor.DoResponse)
  8. 提交后校正计费 (AdjustBillingOnSubmit + recalcQuotaFromRatios)
  9. 结算差额 (SettleBilling),写消费日志,插入 tasks 表,保存 UpstreamTaskID 给后续轮询

  → 这个接口每请求都会真实地与上游交互一次,且会真实扣费/结算。

  GET(状态查询)与之对比:Gemini/Vertex 走实时上游,其他平台走本地缓存 + 后台轮询。
