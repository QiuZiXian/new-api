Final summary of what shipped for dev-docs/dev进度.md #2 素材组管理+素材管理:

  Code (8 files):
  - model/asset.go — AssetGroup / Asset / AssetMappings tables + soft-delete CAS
  - model/main.go — AutoMigrate wiring for both tables
  - model/task_cas_test.go — added asset tables to the shared in-memory test DB
  - model/asset_test.go — 8 unit tests (owner filter, soft-delete idempotency, status filter, paging, JSON round-trip, ID format)
  - relay/channel/task/doubao/asset_adaptor.go — 8 upstream HTTP methods + normalizeAssetApiPath + joinAssetID
  - relay/channel/task/doubao/asset_adaptor_test.go — 5 tests (path-injection fallback, id-injection rejection, Bearer passthrough, 404 idempotency, 5xx propagation)
  - service/asset/service.go — pickAssetChannel + 5 group + 5 asset business methods, local-first writes with best-effort upstream
  - controller/asset.go — 10 Gin handlers using the existing respondTaskError
  - router/video-router.go — /v1/asset-groups[/...] + /v1/assets/:asset_id route group, TokenAuth only
  
  Validation results::
  - go build ./... — clean
  - cd relaykit && GOWORK=off go build ./... — clean (no relaykit changes)
  - go vet ./... — clean
  - go test ./controller/ ./router/ ./service/ ./relay/channel/task/... ./model/ — all green
  - gofmt -l . — 0 files
  
  Notable design notes (also written into dev-docs 2.11):
  - Service placed in service/asset/ subpackage to break the import cycle service → relay → service.
  - pickAssetChannel is global (no per-user enabled-channel concept in this codebase), only filters by ChannelStatusEnabled + Type in (DoubaoVideo, VolcEngine).
  - Binary upload → OSS is out of scope per 2.10; CreateAsset requires the caller to supply a public image_url.
  - ChannelOtherSettings.AssetApiPath field not yet exposed in dto; resolveAssetApiPath returns "" to fall back to the adaptor's default.
