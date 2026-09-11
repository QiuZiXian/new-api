package asset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	taskdoubao "github.com/QuantumNous/new-api/relay/channel/task/doubao"
)

// 素材组 / 素材 业务编排层
//
// 总体策略（沿用 controller/task_content.go 的范式）：
//   - 本地表是事实源（id、名称、拥有权、状态）
//   - 上游 CII/Seedance 仓库是缓存映射（UpstreamAssetGroupID / UpstreamAssetID）
//   - 创建：先调上游，拿到上游 id 后再写本地；上游 2xx 之外 → 不写本地
//   - 更新：先写本地（用户视角立即生效），再 best-effort 调上游 PUT；上游失败仅记日志
//   - 删除：先本地 CAS 软删，再 best-effort 调上游 DELETE；上游失败仅记日志
//   - 列表 / 详情：纯本地查询，不查上游（避免其它租户泄漏）
//
// 不引入 lockForUpdate：与 dev-docs/dev进度.md 2.5 "并发安全" 段落一致，
// 素材 CRUD 不存在多端竞争关键路径。

// assetPageMax 与上游 Seedance/CII 的分页上限保持一致（dev-docs/cii-api.md
// 中 PageSize 通常最大 500）。
const assetPageMax = 500

// assetGroupTypeDefault 是新素材组的默认 group_type，与 dev-docs 一致。
const assetGroupTypeDefault = "AIGC"

// assetStatusActive / assetStatusDisabled 是本地 assets.status 的取值；
// 留给后续任务/视频生成流程判断该素材是否仍可用。
const (
	assetStatusActive   = "Active"
	assetStatusDisabled = "Disabled"
)

// assetNameMaxLen 名称最大长度；与 model.AssetGroup.Name / model.Asset.Name
// 的 varchar(255) 对齐。
const assetNameMaxLen = 255

// PickAssetChannel 在已启用的 doubao / volcengine 渠道中选一个。
//
// 说明：当前 new-api 没有 "per-user enabled channel" 这一概念——所有用户共享
// 全局渠道池。这里只按"渠道类型"+"启用状态"做过滤，等价于 dev 文档 2.5
// "pickAssetChannel：只允许 doubao / volcengine" 的语义。
//
// 注意：GetChannelsByType 带 .Omit("key")，返回的候选里 Key 永远是空串，所以
// 必须先按 type+status 挑一个候选，再用 GetChannelById(ch.Id, true) 取回含
// 真实 Key 的完整记录，才能判断上游调用所需字段是否就绪。
//
// 优先级：priority desc, id desc；返回 nil 表示无可用渠道。
func PickAssetChannel() (*model.Channel, error) {
	candidates, err := model.GetChannelsByType(0, 50, false, constant.ChannelTypeDoubaoVideo)
	if err != nil {
		return nil, err
	}
	volc, err := model.GetChannelsByType(0, 50, false, constant.ChannelTypeVolcEngine)
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, volc...)

	var firstErr error
	for _, ch := range candidates {
		if ch.Status != common.ChannelStatusEnabled {
			continue
		}
		full, err := model.GetChannelById(ch.Id, true)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if full.Status == common.ChannelStatusEnabled && full.Key != "" {
			return full, nil
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, errors.New("no enabled doubao/volcengine channel")
}

// resolveAssetGroupApiPath 取 channel setting 里的 AssetGroupApiPath 覆盖值；
// 为空时返回 ""，由 doubao adaptor 回退到默认路径 /api/v1/asset-groups。
func resolveAssetGroupApiPath(ch *model.Channel) string {
	if ch == nil {
		return ""
	}
	return ch.GetOtherSettings().AssetGroupApiPath
}

// resolveAssetApiPath 取 channel setting 里的 AssetApiPath 覆盖值；
// 为空时返回 ""，由 doubao adaptor 回退到默认路径 /api/v1/assets。
func resolveAssetApiPath(ch *model.Channel) string {
	if ch == nil {
		return ""
	}
	return ch.GetOtherSettings().AssetApiPath
}

// CreateAssetGroup 创建一个素材组：
//  1. 校验 name
//  2. 选定上游渠道
//  3. POST /api/v1/asset-groups
//  4. 写本地表
func CreateAssetGroup(userID int, name, description string) (*model.AssetGroup, *taskdto.TaskError) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errAssetBadRequest("name 不能为空")
	}
	if len(name) > assetNameMaxLen {
		return nil, errAssetBadRequest("name 超过 255 字符")
	}
	description = strings.TrimSpace(description)

	ch, err := PickAssetChannel()
	if err != nil {
		common.SysError("asset: pick channel failed: " + err.Error())
		return nil, errAssetBadGateway("无可用 doubao/volcengine 渠道")
	}

	body := mustMarshal(map[string]any{
		"Name":        name,
		"Description": description,
	})
	upstreamBody, status, err := (&taskdoubao.TaskAdaptor{}).CreateGroup(
		ch.GetBaseURL(), ch.Key, resolveAssetGroupApiPath(ch), body, ch.GetSetting().Proxy,
	)
	if err != nil {
		common.SysError("asset: upstream create group failed: " + err.Error())
		return nil, errAssetBadGateway("上游创建素材组失败")
	}
	if status < 200 || status >= 300 {
		return nil, errAssetBadGateway("上游创建素材组失败: " + snippetFromBody(upstreamBody))
	}

	upstreamID, perr := parseAssetGroupID(upstreamBody)
	if perr != nil {
		common.SysError("asset: parse upstream group id failed: " + perr.Error())
		// 直接把上游/解析错误的原文塞到响应里，方便客户端定位；
		// 否则"上游响应缺少 AssetGroupId"这种兜底会把真正的错误原因吞掉。
		return nil, errAssetBadGateway("上游创建素材组失败: " + perr.Error())
	}

	now := time.Now().Unix()
	g := &model.AssetGroup{
		PublicID:             model.GenerateAssetGroupID(),
		UserID:               userID,
		ChannelID:            ch.Id,
		ChannelType:          ch.Type,
		UpstreamAssetGroupID: upstreamID,
		Name:                 name,
		Description:          description,
		GroupType:            assetGroupTypeDefault,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := g.Insert(); err != nil {
		common.SysError("asset: insert group failed: " + err.Error())
		return nil, errAssetBadGateway("本地写入素材组失败")
	}
	return g, nil
}

// ListAssetGroups 读取用户素材组列表（按 id 倒序）。
func ListAssetGroups(userID, pageNum, pageSize int) ([]*model.AssetGroup, int64, *taskdto.TaskError) {
	pageNum, pageSize = clampAssetPage(pageNum, pageSize)
	items, total, err := model.ListAssetGroups(userID, pageNum, pageSize)
	if err != nil {
		common.SysError("asset: list groups failed: " + err.Error())
		return nil, 0, errAssetBadGateway("读取素材组失败")
	}
	return items, total, nil
}

// GetAssetGroup 通过公开 ID 拿素材组；ownerUserID <= 0 表示不限制。
func GetAssetGroup(ownerUserID int, publicID string) (*model.AssetGroup, *taskdto.TaskError) {
	g, err := model.GetAssetGroupByPublicID(publicID, ownerUserID)
	if err != nil {
		return nil, errAssetNotFound("素材组不存在")
	}
	return g, nil
}

// UpdateAssetGroup 更新素材组名称/描述。本地先写，再 best-effort 同步上游。
func UpdateAssetGroup(userID int, publicID, name, description string) (*model.AssetGroup, *taskdto.TaskError) {
	g, terr := GetAssetGroup(userID, publicID)
	if terr != nil {
		return nil, terr
	}
	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) > assetNameMaxLen {
			return nil, errAssetBadRequest("name 超过 255 字符")
		}
		g.Name = name
	}
	if description != "" {
		g.Description = strings.TrimSpace(description)
	}
	g.UpdatedAt = time.Now().Unix()
	if err := g.Update(); err != nil {
		common.SysError("asset: update group failed: " + err.Error())
		return nil, errAssetBadGateway("本地更新素材组失败")
	}

	// best-effort 同步上游 PUT
	go func(chID int, upstreamID, name, desc string) {
		ch, cerr := model.GetChannelById(chID, true)
		if cerr != nil || ch == nil {
			return
		}
		body := mustMarshal(map[string]any{
			"Name":        name,
			"Description": desc,
		})
		if _, status, err := (&taskdoubao.TaskAdaptor{}).UpdateGroup(
			ch.GetBaseURL(), ch.Key, resolveAssetGroupApiPath(ch), upstreamID, body, ch.GetSetting().Proxy,
		); err != nil {
			common.SysError("asset: upstream update group failed: " + err.Error())
		} else if status < 200 || status >= 300 {
			common.SysError(fmt.Sprintf("asset: upstream update group non-2xx: %d", status))
		}
	}(g.ChannelID, g.UpstreamAssetGroupID, g.Name, g.Description)

	return g, nil
}

// DeleteAssetGroup 删除素材组：本地 CAS 软删，同步上游删除。
//
// 本地删除始终是事实源；上游删除失败不会回滚本地状态，但会作为
// upstreamErr 回传给 controller，由 controller 决定是否要把这个错误
// 写进响应体（让客户端知道"本地已删、上游未删"，方便后续对账）。
//
// 业务约束：组内还有未删素材时禁止删除。
func DeleteAssetGroup(userID int, publicID string) (upstreamErr string, terr *taskdto.TaskError) {
	g, terr := GetAssetGroup(userID, publicID)
	if terr != nil {
		return "", terr
	}
	activeCount, err := model.CountActiveAssetsInGroup(uint(userID), g.ID)
	if err != nil {
		common.SysError("asset: count active assets failed: " + err.Error())
		return "", errAssetBadGateway("查询组内素材失败")
	}
	if activeCount > 0 {
		return "", errAssetBadRequest("素材组内还有素材，请先清空")
	}

	won, err := model.SoftDeleteAssetGroup(userID, publicID)
	if err != nil {
		common.SysError("asset: soft delete group failed: " + err.Error())
		return "", errAssetBadGateway("本地删除素材组失败")
	}
	if !won {
		// 已被并发删除——幂等
		return "", nil
	}

	ch, cerr := model.GetChannelById(g.ChannelID, true)
	if cerr != nil || ch == nil {
		common.SysError("asset: lookup channel for upstream delete group failed: " + errIfNonNil(cerr))
		return "本地已删除，上游同步渠道不可用", nil
	}
	_, status, derr := (&taskdoubao.TaskAdaptor{}).DeleteGroup(
		ch.GetBaseURL(), ch.Key, resolveAssetGroupApiPath(ch), g.UpstreamAssetGroupID, ch.GetSetting().Proxy,
	)
	if derr != nil {
		common.SysError("asset: upstream delete group failed: " + derr.Error())
		return "本地已删除，上游删除失败: " + derr.Error(), nil
	}
	if status >= 200 && status < 300 {
		return "", nil
	}
	// 理论不会到这里——assetDoDelete 收到非 2xx 已经返回 err。
	return fmt.Sprintf("本地已删除，上游返回非 2xx 状态: %d", status), nil
}

// CreateAsset 在指定素材组下创建一条素材。
//
// 重要：上传者必须自己提供可被上游访问的公网 URL（dev-docs 2.5 已说明，
// OSS 暂不接入；本流程要求 client 上传完成后自己回填 url）。imageURL 即为
// 上游 API 的 ImageUrl 字段。
func CreateAsset(userID int, groupPublicID, imageURL, assetType, name string) (*model.Asset, *taskdto.TaskError) {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return nil, errAssetBadRequest("image_url 不能为空")
	}
	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) > assetNameMaxLen {
			return nil, errAssetBadRequest("name 超过 255 字符")
		}
	}
	if assetType == "" {
		assetType = "Image"
	}

	g, terr := GetAssetGroup(userID, groupPublicID)
	if terr != nil {
		return nil, terr
	}

	ch, err := model.GetChannelById(g.ChannelID, true)
	if err != nil || ch == nil {
		common.SysError("asset: lookup channel for group failed: " + errIfNonNil(err))
		return nil, errAssetBadGateway("素材组关联渠道不可用")
	}

	body := mustMarshal(map[string]any{
		"AssetGroupId": g.UpstreamAssetGroupID,
		"ImageUrl":     imageURL,
		"AssetType":    assetType,
		"Name":         name,
	})
	upstreamBody, status, err := (&taskdoubao.TaskAdaptor{}).CreateAsset(
		ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), body, ch.GetSetting().Proxy,
	)
	if err != nil {
		common.SysError("asset: upstream create asset failed: " + err.Error())
		return nil, errAssetBadGateway("上游创建素材失败")
	}
	if status < 200 || status >= 300 {
		return nil, errAssetBadGateway("上游创建素材失败: " + snippetFromBody(upstreamBody))
	}

	upstreamID, perr := parseAssetID(upstreamBody)
	if perr != nil {
		common.SysError("asset: parse upstream asset id failed: " + perr.Error())
		// 直接把上游/解析错误的原文塞到响应里，方便客户端定位；
		// 否则"上游响应缺少 AssetId"这种兜底会把真正的错误原因吞掉。
		return nil, errAssetBadGateway("上游创建素材失败: " + perr.Error())
	}

	now := time.Now().Unix()
	a := &model.Asset{
		PublicID:        model.GenerateAssetID(),
		UserID:          userID,
		ChannelID:       ch.Id,
		ChannelType:     ch.Type,
		GroupID:         g.ID,
		UpstreamAssetID: upstreamID,
		AssetType:       assetType,
		Name:            name,
		SourceURL:       imageURL,
		Status:          assetStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := a.Insert(); err != nil {
		common.SysError("asset: insert asset failed: " + err.Error())
		return nil, errAssetBadGateway("本地写入素材失败")
	}
	return a, nil
}

// ListAssets 列出指定素材组下的素材（带可选 status 过滤）。
func ListAssets(userID, pageNum, pageSize int, groupPublicID, status string) ([]*model.Asset, int64, *taskdto.TaskError) {
	pageNum, pageSize = clampAssetPage(pageNum, pageSize)
	g, terr := GetAssetGroup(userID, groupPublicID)
	if terr != nil {
		return nil, 0, terr
	}
	statuses := []string{}
	if s := strings.TrimSpace(status); s != "" {
		statuses = []string{s}
	}
	items, total, err := model.ListAssets(uint(userID), g.ID, statuses, pageNum, pageSize)
	if err != nil {
		common.SysError("asset: list assets failed: " + err.Error())
		return nil, 0, errAssetBadGateway("读取素材失败")
	}
	return items, total, nil
}

// GetAsset 通过公开 ID 拿素材。
func GetAsset(ownerUserID int, publicID string) (*model.Asset, *taskdto.TaskError) {
	a, err := model.GetAssetByPublicID(publicID, ownerUserID)
	if err != nil {
		return nil, errAssetNotFound("素材不存在")
	}
	return a, nil
}

// UpdateAsset 更新素材名称（dev-docs 仅支持 name）。本地先写，再 best-effort 同步上游。
func UpdateAsset(userID int, publicID, name string) (*model.Asset, *taskdto.TaskError) {
	a, terr := GetAsset(userID, publicID)
	if terr != nil {
		return nil, terr
	}
	if name != "" {
		name = strings.TrimSpace(name)
		if len(name) > assetNameMaxLen {
			return nil, errAssetBadRequest("name 超过 255 字符")
		}
		a.Name = name
	}
	a.UpdatedAt = time.Now().Unix()
	if err := a.Update(); err != nil {
		common.SysError("asset: update asset failed: " + err.Error())
		return nil, errAssetBadGateway("本地更新素材失败")
	}

	go func(chID int, upstreamID, name string) {
		ch, cerr := model.GetChannelById(chID, true)
		if cerr != nil || ch == nil {
			return
		}
		body := mustMarshal(map[string]any{"Name": name})
		if _, status, err := (&taskdoubao.TaskAdaptor{}).UpdateAsset(
			ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), upstreamID, body, ch.GetSetting().Proxy,
		); err != nil {
			common.SysError("asset: upstream update asset failed: " + err.Error())
		} else if status < 200 || status >= 300 {
			common.SysError(fmt.Sprintf("asset: upstream update asset non-2xx: %d", status))
		}
	}(a.ChannelID, a.UpstreamAssetID, a.Name)
	return a, nil
}

// DeleteAsset 删除素材：本地 CAS 软删，同步上游删除。
//
// 与 DeleteAssetGroup 一致：本地是事实源；上游失败不影响本地删除结果，
// 但会作为 upstreamErr 回传，便于客户端对账。
func DeleteAsset(userID int, publicID string) (upstreamErr string, terr *taskdto.TaskError) {
	a, terr := GetAsset(userID, publicID)
	if terr != nil {
		return "", terr
	}
	won, err := model.SoftDeleteAsset(userID, publicID)
	if err != nil {
		common.SysError("asset: soft delete asset failed: " + err.Error())
		return "", errAssetBadGateway("本地删除素材失败")
	}
	if !won {
		return "", nil
	}

	ch, cerr := model.GetChannelById(a.ChannelID, true)
	if cerr != nil || ch == nil {
		common.SysError("asset: lookup channel for upstream delete asset failed: " + errIfNonNil(cerr))
		return "本地已删除，上游同步渠道不可用", nil
	}
	_, status, derr := (&taskdoubao.TaskAdaptor{}).DeleteAsset(
		ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), a.UpstreamAssetID, ch.GetSetting().Proxy,
	)
	if derr != nil {
		common.SysError("asset: upstream delete asset failed: " + derr.Error())
		return "本地已删除，上游删除失败: " + derr.Error(), nil
	}
	if status >= 200 && status < 300 {
		return "", nil
	}
	return fmt.Sprintf("本地已删除，上游返回非 2xx 状态: %d", status), nil
}

// ============================
// helpers
// ============================

// clampAssetPage 把分页参数钳到 [1, assetPageMax]。
func clampAssetPage(pageNum, pageSize int) (int, int) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > assetPageMax {
		pageSize = assetPageMax
	}
	return pageNum, pageSize
}

// errAssetBadRequest 构造 400 错误。
func errAssetBadRequest(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "bad_request", Message: msg, StatusCode: 400}
}

// errAssetNotFound 构造 404 错误。
func errAssetNotFound(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "not_found", Message: msg, StatusCode: 404}
}

// errAssetBadGateway 构造 502 错误（上游/数据库不可用）。
func errAssetBadGateway(msg string) *taskdto.TaskError {
	return &taskdto.TaskError{Code: "upstream_error", Message: msg, StatusCode: 502}
}

// snippetFromBody 把上游响应体裁短到 256 字符，避免错误日志爆掉。
func snippetFromBody(body []byte) string {
	s := strings.TrimSpace(string(body))
	if len(s) > 256 {
		s = s[:256] + " ...(truncated)"
	}
	return s
}

// errIfNonNil 让 fmt-like 调用可以无脑写在日志里。
func errIfNonNil(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// mustMarshal 把 v 序列化为 JSON；错误视为不可恢复的内部错误并打日志后返回 nil。
// 上游调用方拿到 nil body 仍会自行报错，不会造成不可观察的状态变更。
func mustMarshal(v any) []byte {
	b, err := common.Marshal(v)
	if err != nil {
		common.SysError("asset: marshal request body failed: " + err.Error())
		return nil
	}
	return b
}

// parseAssetGroupID 从上游创建素材组的响应中抠出素材组 ID。
//
// 同一 Seedance 协议在不同网关下的响应 schema 不一致，且同一网关在
// "成功"和"业务失败"下 schema 也不一致：
//   - 成功 - 平铺形式（dev-docs/cii-api.md）：
//       {"AssetGroupId": "ag-xxx", "Name": "..."}
//   - 成功 - 信封形式（cii-group.com CII app-api）：
//       {"ResponseMetadata": {...}, "Result": {"Id": "group-xxx"}}
//   - 业务失败 - 信封形式（CII app-api，HTTP 仍 200）：
//       {"ResponseMetadata": {"Error": {"Code": "...", "Message": "..."}}}
//
// 这里按 "平铺成功 → 信封成功 → 信封错误" 顺序尝试。前两种拿到 ID 即返回；
// 第三种是上游业务失败，要把它原始的 Code/Message 透出去给客户端，
// 避免出现"上游响应缺少 AssetGroupId"这种把上游错误吞掉的误导信息。
func parseAssetGroupID(body []byte) (string, error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if v, ok := m["AssetGroupId"].(string); ok && v != "" {
		return v, nil
	}
	if result, ok := m["Result"].(map[string]any); ok {
		if v, ok := result["Id"].(string); ok && v != "" {
			return v, nil
		}
	}
	if errMsg := extractEnvelopeError(m); errMsg != "" {
		return "", errors.New(errMsg)
	}
	return "", errors.New("AssetGroupId missing")
}

// parseAssetID 从上游创建素材的响应中抠出素材 ID。
// schema 兼容性与 parseAssetGroupID 一致：平铺成功 / 信封成功 / 信封错误。
func parseAssetID(body []byte) (string, error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if v, ok := m["AssetId"].(string); ok && v != "" {
		return v, nil
	}
	if result, ok := m["Result"].(map[string]any); ok {
		if v, ok := result["Id"].(string); ok && v != "" {
			return v, nil
		}
	}
	if errMsg := extractEnvelopeError(m); errMsg != "" {
		return "", errors.New(errMsg)
	}
	return "", errors.New("AssetId missing")
}

// extractEnvelopeError 解析 cii-group CII app-api 的信封错误结构：
//   {"ResponseMetadata": {"Error": {"Code": "...", "Message": "..."}}}
// 命中则返回 "Code: Message" 形式的可读字符串；未命中返回 ""。
func extractEnvelopeError(m map[string]any) string {
	rm, ok := m["ResponseMetadata"].(map[string]any)
	if !ok {
		return ""
	}
	errObj, ok := rm["Error"].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := errObj["Code"].(string)
	msg, _ := errObj["Message"].(string)
	code = strings.TrimSpace(code)
	msg = strings.TrimSpace(msg)
	switch {
	case code != "" && msg != "":
		return code + ": " + msg
	case code != "":
		return code
	case msg != "":
		return msg
	default:
		return ""
	}
}
