# 1 查询任务列表+删除任务
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
# 2 新增 素材组管理+素材管理

> 本节是基于上游 cii-group 素材管理 API（`POST/GET/PUT /api/v1/asset-groups`、`POST/GET/PUT /api/v1/assets`、`/list`）+ dev-docs/dev进度.md 中"OSS 才是事实源 / 上游是缓存映射"架构的本地实现设计。
> 由于 `gh` / WebFetch 在本环境均不可用，未在 QuantumNous/new-api PR 列表里交叉验证既有实现；落点完全沿用本仓库已落地的 `controller/task_content.go` + `relay/channel/task/doubao/adaptor.go` 范式（用户 Token 鉴权、本地为权威、上游为旁路、用 CAS + 退款保证本地一致性、`normalizeTaskApiPath` 防路径注入），便于 review。

## 2.1 上游 API 摘录（dev-docs/cii-api.md）

| 操作 | Method + Path | 鉴权 | 入参（关键） | 出参（关键） |
|---|---|---|---|---|
| 素材组-创建 | `POST /api/v1/asset-groups` | Bearer | `Name`(必填) / `Description` | 平铺：`{AssetGroupId, Name, Description, CreatedAt, UpdatedAt}` |
| 素材组-列表 | `POST /api/v1/asset-groups/list` | Bearer | `PageNum`(默认 1) / `PageSize`(默认 20) | 包裹：`{ResponseMetadata, Result:{Items:[{Id,Name,Description,GroupType,CreatedAt,UpdatedAt}], TotalCount}}` |
| 素材组-更新 | `PUT /api/v1/asset-groups/{groupId}` | Bearer | `Name?` / `Description?` | 平铺，同上 |
| 素材组-详情 | `GET /api/v1/asset-groups/{groupId}` | Bearer | — | 平铺 |
| 素材-创建 | `POST /api/v1/assets` | Bearer | `AssetGroupId` / `ImageUrl`(必填) / `AssetType=Image` / `Name?` | 平铺：`{AssetId, AssetGroupId, AssetType, Name, Status, CreatedAt, UpdatedAt}` |
| 素材-列表 | `POST /api/v1/assets/list` | Bearer | `AssetGroupId` / `Filter.Status[]` / `PageNum` / `PageSize` | 平铺：`{Items[], Total, PageNum, PageSize}` |
| 素材-更新 | `PUT /api/v1/assets/{assetId}` | Bearer | `Name?` | 平铺 |
| 素材-详情 | `GET /api/v1/assets/{assetId}` | Bearer | — | 平铺 |
| 素材组/素材-删除 | **文档未列出 DELETE** | — | — | 假设 `DELETE /api/v1/asset-groups/{groupId}` 与 `DELETE /api/v1/assets/{assetId}` 存在并 2xx 无 body |

要点：

- 上游 `list/asset-groups` 用 `Id`，create 用 `AssetGroupId`；list 字段名不一致 → 适配器层做归一化。
- create/list 响应 schema 不一致（一个平铺一个包裹） → 控制器层不能直接 Unmarshal 透传，必须在 adaptor 内 parse。
- 文档未列 DELETE → 设计阶段把上游 DELETE 视为"尽力而为"，本地成功是真理。

## 2.2 本地架构（沿用 dev进度.md ASCII 图）

```
用户素材  →  new-api  ──┬─→  本地 OSS  (Source of Truth: 文件本体)
                        ├─→  asset_groups / assets 表 (Source of Truth: 元数据 + 拥有权)
                        └─→  上游 CII / Seedance  (缓存映射: upstream_asset_group_id / upstream_asset_id)
```

约束：

- **本地的 `id`（`group_xxx` / `asset_xxx`）是对外公开 ID；上游 ID 仅作为 cache 提示**。
- 多 provider 视角下，每个 (group, asset) 可能在多个上游有同名映射；当前只接 doubao/volcengine 两个渠道 → 表里只存 1 列 upstream id，预留 `upstream_mappings JSON` 字段便于以后扩展。
- 文件二进制落本地 OSS（暂用本地磁盘 + `setting/operation` 已有的 upload dir，无独立 OSS 时退化为本地路径 URL）；URL 是给上游的 `ImageUrl`。

## 2.3 路由（`router/video-router.go` 新增分组）

```go
// 沿用 task_content.go 的 RouteTag("relay") + TokenAuth() 范式。
assetV1Router := router.Group("/v1")
assetV1Router.Use(middleware.RouteTag("relay"))
assetV1Router.Use(middleware.TokenAuth())
{
    // 素材组
    assetV1Router.POST("/asset-groups",           controller.CreateAssetGroup)
    assetV1Router.GET ("/asset-groups",           controller.ListAssetGroups)
    assetV1Router.GET ("/asset-groups/:group_id", controller.GetAssetGroup)
    assetV1Router.PUT ("/asset-groups/:group_id", controller.UpdateAssetGroup)
    assetV1Router.DELETE("/asset-groups/:group_id", controller.DeleteAssetGroup)

    // 素材（注意：素材强依赖素材组，path 上保留 group_id，便于权限校验）
    assetV1Router.POST("/asset-groups/:group_id/assets", controller.CreateAsset)
    assetV1Router.GET ("/asset-groups/:group_id/assets", controller.ListAssets)
    assetV1Router.GET ("/assets/:asset_id",              controller.GetAsset)
    assetV1Router.PUT ("/assets/:asset_id",              controller.UpdateAsset)
    assetV1Router.DELETE("/assets/:asset_id",            controller.DeleteAsset)
}
```

- 与 `task_content.go` 一致：仅 `TokenAuth()`，**不**挂 `Distribute()` —— 素材 CRUD 不需要按模型分发渠道。
- 鉴权形态：调用方用 `sk-…` token；`middleware.TokenAuth()` 解析后通过 `c.GetInt("id")` 拿到 userId，所有查询/写入都按 `user_id` 限定。
- 错误响应统一用 `taskdto.TaskError{Code, Message, StatusCode}`（与 task_content 复用）。

## 2.4 模型（`model/asset.go` 新增）

两张表，AutoMigrate 进 `model/main.go` 的 `DB.AutoMigrate(...)` 列表。

```go
// AssetGroup 素材组（用户拥有的逻辑分组）
type AssetGroup struct {
    ID                  uint      `gorm:"primaryKey" json:"-"`
    PublicID            string    `gorm:"type:varchar(64);uniqueIndex" json:"id"`            // 对外 group_xxx
    UserID              int       `gorm:"index"           json:"-"`
    ChannelID           int       `gorm:"index"           json:"-"`                            // 创建时选定的上游渠道
    ChannelType         int       `json:"-"`
    UpstreamAssetGroupID string   `gorm:"type:varchar(128);index" json:"-"`                    // 上游 ID，缓存/同步用
    Name                string    `gorm:"type:varchar(255)" json:"name"`
    Description         string    `gorm:"type:text"        json:"description"`
    GroupType           string    `gorm:"type:varchar(32)" json:"group_type,omitempty"`        // 默认 "AIGC"
    CreatedAt           int64     `json:"created_at"`
    UpdatedAt           int64     `json:"updated_at"`
    DeletedAt           int64     `gorm:"index"            json:"-"`                            // 软删时间戳（0=未删）
}

// Asset 素材（用户上传的图片/视频，逻辑挂在 AssetGroup 下）
type Asset struct {
    ID             uint   `gorm:"primaryKey" json:"-"`
    PublicID       string `gorm:"type:varchar(64);uniqueIndex" json:"id"`         // 对外 asset_xxx
    UserID         int    `gorm:"index"           json:"-"`
    ChannelID      int    `gorm:"index"           json:"-"`
    ChannelType    int    `json:"-"`
    GroupID        uint   `gorm:"index"           json:"-"`                         // 关联 asset_groups.id
    UpstreamAssetID string `gorm:"type:varchar(128);index" json:"-"`               // 上游 ID
    AssetType      string `gorm:"type:varchar(32)"  json:"asset_type"`             // Image / Video
    Name           string `gorm:"type:varchar(255)" json:"name"`
    SourceURL      string `gorm:"type:text"         json:"source_url"`              // 本地 OSS URL（=上游 ImageUrl）
    Status         string `gorm:"type:varchar(32);index" json:"status"`             // Active / Disabled
    SizeBytes      int64  `json:"size_bytes,omitempty"`
    MimeType       string `gorm:"type:varchar(64)" json:"mime_type,omitempty"`
    CreatedAt      int64  `json:"created_at"`
    UpdatedAt      int64  `json:"updated_at"`
    DeletedAt      int64  `gorm:"index"            json:"-"`                        // 软删
}
```

要点：

- `PublicID` 用 `group_`/`asset_` 前缀 + nanoid（沿用 `task_xxx` 的实现方式），与 task 体系保持一致。
- `varchar(191)` 在 MySQL 5.7 utf8mb4 是索引上限，因此字段按需截短；JSON 列在 SQLite/MySQL/Postgres 通用，跨库一致。
- 软删 (`DeletedAt`)：上游 DELETE 失败时本地仍可继续 GC（参考"2.6 删除"）。
- GORM AutoMigrate：放在 `model/main.go` 的 `DB.AutoMigrate(...)` 列表；SQLite 端走 `ADD COLUMN` 安全路径（与 `ALTER TABLE … ADD COLUMN` 一致，遵循 `model/main.go` 已有的 `HasColumn` 检查模式）。
- 复合索引 `(user_id, deleted_at, created_at)` 在第一次查询热点上避免全表；按需在迁移后追加，不放进初始 schema。

## 2.5 服务层（`service/asset.go` 新增）

```go
// 渠道选择：与 task_content.go 同源——只允许 doubao / volcengine
func pickAssetChannel(ctx context.Context, userID int) (*model.Channel, *taskdto.TaskError)

// 上游调用：复用 doubao adaptor 的 normalizeTaskApiPath 思想，提供
// normalizeAssetApiPath(raw) 默认值 "/api/v1/asset-groups"。
type assetAdaptor interface {
    CreateGroup(baseURL, key, path string, body []byte, proxy string) (*http.Response, error)
    ListGroups  (baseURL, key, path string, pageNum, pageSize int, proxy string) (*http.Response, error)
    UpdateGroup (baseURL, key, path, groupID string, body []byte, proxy string) (*http.Response, error)
    GetGroup    (baseURL, key, path, groupID string, proxy string) (*http.Response, error)
    DeleteGroup (baseURL, key, path, groupID string, proxy string) error           // 404 视为成功
    CreateAsset / ListAssets / UpdateAsset / GetAsset / DeleteAsset 同形
}

// CreateAssetGroup:
//   1) 入参校验 (Name 非空、长度 ≤ 255)
//   2) pickAssetChannel  → 选一个可用 doubao/volcengine 渠道
//   3) POST /api/v1/asset-groups {Name, Description}，取响应里的 AssetGroupId
//   4) INSERT asset_groups（PublicID = group_<nanoid>，UpstreamAssetGroupID = 上一步返回值）
//   5) 返回上游+本地合并后的对象
//
// CreateAsset（带二进制上传）：
//   1) 校验 user 拥有 :group_id（user_id + deleted_at=0）
//   2) 接收 multipart 上传 → 落本地 OSS（pkg/oss 或 setting/operation.UploadDir）
//      → 生成可被上游访问的公网 URL（若本机无外网，用内网 callback：暂不支持，
//        上传者必须自己提供可公网访问的 URL，遵循上游 "ImageUrl 必填" 约束）
//   3) POST /api/v1/assets {AssetGroupId, ImageUrl, AssetType, Name}
//   4) INSERT assets（关联 group_id，存 UpstreamAssetID）
//
// List/Get/Update/Delete：
//   - 读路径：纯本地表（user_id + deleted_at 限定），不查上游
//   - 写路径：本地 UPDATE 后 异步/同步 调上游 PUT/DELETE，失败仅记日志
//   - 名称修改需传播到上游（PUT /{id} Name），但本地写入是权威——失败下次列表仍展示新名

// 失败语义：
//   - 上游 4xx 且非 404 → 本地回滚（或不写本地）+ 返回 502
//   - 上游 5xx / 超时    → 返回 502，本地不写
//   - 上游 404 on DELETE → 本地软删成功（幂等）
//   - 上游 404 on GET    → 重建本地占位（避免用户看到残影）
```

并发安全：

- 素材组/素材不被多端竞争修改（无共享锁需求），用 GORM 默认乐观即可。
- 不引入 `lockForUpdate`（参照 `model/main.go` 中"仅在账单关键路径上锁"的现有约束）。

## 2.6 删除流程（重点：本地权威 + 上游尽力）

```
客户端 DELETE /v1/assets/:asset_id
        │
        ▼
service.DeleteAsset
   ├─ 1) SELECT * FROM assets WHERE public_id=? AND user_id=? AND deleted_at=0
   │     └─ 404 → 直接 404
   │
   ├─ 2) UPDATE assets SET deleted_at=NOW() WHERE id=? AND deleted_at=0   ← 软删 (CAS)
   │     └─ affected=0 → 已删过，幂等返回 200
   │
   ├─ 3) (best-effort) channel.GetBaseURL() + adaptor.DeleteAsset(...)
   │     ├─ 2xx    → done
   │     ├─ 404    → done (上游已无记录，幂等)
   │     └─ 其他   → logger.LogWarn，记入 system log，本地状态不动
   │
   └─ 4) 返回 204
```

为何本地软删而非硬删：上游 DELETE 偶发 5xx 也不影响"用户视角已删除"；后台 GC 任务可定期清 `deleted_at < NOW() - 30d` 的本地行（不在本节范围）。

`DELETE /asset-groups/:group_id` 同形，且额外：先校验 `COUNT(*) FROM assets WHERE group_id=? AND deleted_at=0` → 非空则 409 "请先删除组内素材"。

## 2.7 计费

素材 CRUD 自身不消耗 quota（与 Seedance 视频任务的 token 计费解耦）。配额只在用户**用素材触发视频生成**时由 Seedance 渠道按 token 结算——这条路径已在 `relay/channel/task/doubao/adaptor.go` 的 `EstimateBilling` 中覆盖，asset 本身只是被嵌入 `content[]` 引用，**不需要重复计费**。

→ 不动 `service/task_billing.go`、不动 `quota_math.go`、不动 `QuotaClamp` 链（符合 CLAUDE.md 中"Billing safety invariants"最小变更面原则）。

## 2.8 错误与 i18n

- 错误响应统一 `taskdto.TaskError{Code,Message,StatusCode}` 形态（与 task_content.go 完全一致）。
- 错误码：

| HTTP | code | 触发 |
|---|---|---|
| 400 | `invalid_request` | Name 空 / Description 超长 / 文件过大 |
| 400 | `group_not_empty` | 删素材组时组内仍有未删素材 |
| 401 | `auth_failed` | TokenAuth 失败（中间件自带） |
| 403 | `forbidden` | 资源 owner 不匹配（user_id 不等） |
| 404 | `not_found` | PublicID 不存在或已删 |
| 502 | `upstream_failed` | 上游 4xx/5xx/超时；message 带状态码 + 摘要 |
| 500 | `internal_error` | DB/解析失败 |

- i18n：仅在 message 字符串中暴露中文/英文；`Code` 机器可读，不翻译。前端 i18n key 在前端任务中处理，不在本节。

## 2.9 复用与新增清单

复用：

- `controller/task_content.go` 里的 `respondContentTaskError` 抽到 `controller/task_helpers.go` 共享（命名 `respondTaskError`），assets 控制器也用它；或直接在 `controller/task_content.go` 同包写 `respondAssetError`，与现有风格一致。
- `relay/channel/task/doubao/adaptor.go` 里的 `normalizeTaskApiPath` → 抽到 `relay/channel/task/taskcommon/path.go`，改名 `NormalizePath(default, raw string) string`；新 adaptor 复用。
- `service.GetHttpClientWithProxy`、channel 的 `GetBaseURL/GetOtherSettings/GetSetting`。
- `model/main.go` 的 AutoMigrate + SQLite ADD COLUMN 检查。
- `taskdto.TaskError` 响应壳。

新增：

- `model/asset.go`（AssetGroup、Asset、Indexes、PublicID 生成器）。
- `model/main.go` 注册 AutoMigrate（注意 SQLite 的 ADD COLUMN 分支）。
- `relay/channel/task/cii/asset_adaptor.go`（或并入 doubao adaptor 新增方法）：8 个 HTTP 调用 + JSON 解析（含 list/create 字段名归一化）。
- `service/asset.go`：CRUD 业务编排、上游调用、错误归一。
- `controller/asset.go`：`CreateAssetGroup/ListAssetGroups/GetAssetGroup/UpdateAssetGroup/DeleteAssetGroup/CreateAsset/ListAssets/GetAsset/UpdateAsset/DeleteAsset`。
- `router/video-router.go` 新增分组（同上）。
- `controller/asset_test.go`：表驱动测试，覆盖：
  - list 字段归一（`Id` → `AssetGroupId`）、分页钳制
  - buildAssetGroupItem / buildAssetItem（owner 越权时返回 403）
  - DeleteAsset 软删幂等（第二次返回 200，上游 404 不报错）
  - pickAssetChannel 在 user 无可用 doubao/volcengine 渠道时返回 503
- `dev-docs/dev进度.md` 增补"2.10 验证"小节，记录 `go build ./...` / `go vet ./...` / `go test ./controller/ ./router/ ./service/ ./relay/channel/task/...` / `gofmt -l` 的实际结果。

## 2.10 取舍与未决项

- **不接二进制上传代理**：上游 `POST /api/v1/assets` 只接受 `ImageUrl`，本地无法把内网文件转公网 URL 给上游，**必须由调用方提供可公网访问的 URL**。如果产品要支持"前端直接上传"，需要先在网关加 `image-proxy` 把内网 OSS 资源临时签成公网 URL 再调上游（不在本节范围）。本节素材创建接口暂设计为"调用方传 URL + 可选 size/mime 校验"，二进制的本地 OSS 副本由后续上传链路补。
- **不实现真人审核 (visual-validate) 的 H5 跳转**——属于独立产品链路，且涉及状态机/回调 url；本节仅做素材 CRUD。
- **不持久化上游 `TotalCount` 分页**——上游 list 没有 `next_page_token`，纯页码；当本地与上游不同步时，本地是权威，按本地分页返回。
- **是否需要本地侧 status 字段与上游 `Status` 同步？** —— 需要，但只在 `CreateAsset` 时从上游响应里取一次，运行时**不主动**回查上游（避免每次列表都打 N 次上游）。要拉新可加一个 `GET /v1/assets/:id?refresh=1` 的隐藏参数，不在初版。
- **跨数据库**：所有字段在 SQLite / MySQL 5.7 / Postgres 9.6 上 AutoMigrate 通过；GORM `uniqueIndex` 三库都支持；`type:text` 在三库均为可索引（资产描述不建全文索引）。
- **是否进 relaykit/？** —— 不进。`controller/task_content.go` / 本节的 `controller/asset.go` 都依赖 `model/`、`service/`、`relay/channel/task/...`，这些都是根模块；relaykit 只能装纯协议解析，按 CLAUDE.md "relaykit 必须独立可编译" 约束不引入。

## 2.11 验证计划（落地后填写）

- [x] `go build ./...` —— 通过（无输出）
- [x] `cd relaykit && GOWORK=off go build ./...` —— 通过（本节未触碰 relaykit）
- [x] `go vet ./...` —— 通过
- [x] `go test ./controller/ ./router/ ./service/ ./relay/channel/task/...` —— 全部通过
- [x] `go test ./model/` —— 全部通过（含新 `model/asset_test.go`）
- [x] `gofmt -l . | wc -l` = 0
- [ ] 端到端：用真实 doubao 渠道 token 走 POST /v1/asset-groups → list → 创建素材 → list → 软删 → 再次 list 不见；上游对应记录被删除（200 或 404）。
- [ ] owner 越权：另一 user 的 token 取不到资源（403）。
- [ ] 上游 502：本地上游路径返回 502，本地状态不被破坏。

### 验证记录（2026/09/08）

落点：

- `model/asset.go` 新增 AssetGroup / Asset / AssetMappings；`model/main.go` 接入 AutoMigrate。
- `relay/channel/task/doubao/asset_adaptor.go` 新增 8 个上游 HTTP 方法 + `normalizeAssetApiPath` + `joinAssetID` + `assetDoJSON` + `assetDoDelete`。
- `service/asset/service.go` 新增 pickAssetChannel + 5 个组方法 + 5 个素材方法。
- `controller/asset.go` 新增 10 个 Gin handler，复用 `controller/relay.go` 的 `respondTaskError`。
- `router/video-router.go` 新增 `/v1/asset-groups[/...]` 和 `/v1/assets/:asset_id` 路由分组，仅 TokenAuth，不挂 Distribute。
- `model/asset_test.go` 新增 8 个测试（owner 过滤、软删幂等、status 过滤、分页排序、JSON round-trip）。
- `relay/channel/task/doubao/asset_adaptor_test.go` 新增 5 个测试（path 注入回退、id 注入拦截、Bearer 透传、404 幂等、5xx 报错）。

注意：

- 二进制上传 → 本地 OSS → 公网 URL 链：未实现；当前 `CreateAsset` 要求调用方自己提供可公网访问的 `image_url`，与 dev-docs 2.10 取舍一致。
- `ChannelOtherSettings.AssetApiPath` 字段未在 dto 暴露；`resolveAssetApiPath` 暂返回 `""` 走默认路径；后续若要支持 per-channel 覆盖再扩展 dto。
- `pickAssetChannel` 不区分 user——所有用户共享全局 doubao/volcengine 渠道池，与现有 task_content 体系一致。


