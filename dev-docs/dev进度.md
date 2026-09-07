# 查询任务列表+删除任务
## 新增 查询任务列表+删除任务

## 修复
变更摘要

  客户端接口（原先 /api/task/upstream 的 AdminAuth 管理接口替换为 /v1、模型 API Key 访问）
  - router/video-router.go：新增
    - GET /v1/contents/generations/tasks → controller.ListContentTasks
    - DELETE /v1/contents/generations/tasks/:task_id → controller.CancelContentTask
    - 中间件 RouteTag("relay") + TokenAuth()（即普通用户 sk-... 可用）。
  - controller/task_content.go（新）：列表读取调用者自己的本地 tasks（按 user_id 限定，仅 doubao/volcengine 平台 54/45），并以上游形态返回 {"items":[…], "total":N}（item.id = 本地公开 task_xxx，status 映射为
  queued/running/cancelled/succeeded/failed，支持 page_num/page_size/filter.status/filter.task_ids/filter.model）。
  - 取消为先本地后上游：pending（NOT_START/SUBMITTED/QUEUED）可取消 → CAS 置终态 CANCELLED(progress 100%) → service.RefundTaskQuota 退款 → 再用任务自带渠道的 TaskApiPath 调上游 DeleteTask；运行中/已结束返回 400。

  清理/基建
  - router/api-router.go：删除旧 /api/task/upstream 与 /api/task/upstream/:id（AdminAuth）。
  - 删除 controller/task_upstream.go（整文件）。
  - model/task.go：新增终态 TaskStatusCancelled = "CANCELLED"，并把它加入轮询/超时扫描的排除条件（GetAllUnFinishSyncTasks/GetTimedOutUnfinishedTasks/HasUnfinishedSyncTasks），新增 TaskGetUserVideoTasks 查询（跨库安全的 GORM IN）。
  - relay/channel/task/doubao/adaptor.go：移除已无调用方的上游列表代码（ListTask/ListTaskOptions/buildListQuery/responseTaskList/clampInt，net/url import），保留 DeleteTask/FetchTask/normalizeTaskApiPath。

  验证

  - go build ./...、go vet 通过；go test ./controller/ ./router/ ./relay/channel/task/doubao/ 全部通过（新增 controller/task_content_test.go：状态映射、分页钳制、buildContentTaskItem 本地覆盖/快照富化/坏快照兜底）。
  - git diff --check、gofmt 干净。

  说明

  - 未改动前端（此前无调用 /api/task/upstream）与 relaykit/。
  - 改动尚未提交；如需我提交或补充 dev-docs 说明可再告诉我。
  - 可用真实网关做端到端验证：GET /v1/contents/generations/tasks 用 Bearer sk-… 返回本用户 doubao 任务；DELETE /v1/contents/generations/tasks/<task_xxx> 取消并退款；旧路径 /api/task/upstream 已 404。
