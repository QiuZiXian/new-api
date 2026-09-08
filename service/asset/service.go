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
	for _, ch := range candidates {
		if ch.Status == common.ChannelStatusEnabled && ch.Key != "" {
			// 返回副本，避免外部修改影响后续查找
			out := *ch
			return &out, nil
		}
	}
	return nil, errors.New("no enabled doubao/volcengine channel")
}

// resolveAssetApiPath 优先取 channel 自身的 OtherSettings.AssetApiPath，缺失时
// 退到 doubao adaptor 的默认值。该字段尚未在 dto.ChannelOtherSettings 中暴露，
// 这里暂时直接读 OtherSettings 原文，避免反复修改跨模块 dto。
func resolveAssetApiPath(ch *model.Channel) string {
	settings := ch.GetOtherSettings()
	// ChannelOtherSettings 没有 AssetApiPath 字段——后续若要支持 per-channel
	// 覆盖，扩展 dto.ChannelOtherSettings 后再读对应字段。
	_ = settings
	return ""
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
		ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), body, ch.GetSetting().Proxy,
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
		return nil, errAssetBadGateway("上游响应缺少 AssetGroupId")
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
			ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), upstreamID, body, ch.GetSetting().Proxy,
		); err != nil {
			common.SysError("asset: upstream update group failed: " + err.Error())
		} else if status < 200 || status >= 300 {
			common.SysError(fmt.Sprintf("asset: upstream update group non-2xx: %d", status))
		}
	}(g.ChannelID, g.UpstreamAssetGroupID, g.Name, g.Description)

	return g, nil
}

// DeleteAssetGroup 删除素材组：本地 CAS 软删，best-effort 上游删除。
// 业务约束：组内还有未删素材时禁止删除。
func DeleteAssetGroup(userID int, publicID string) *taskdto.TaskError {
	g, terr := GetAssetGroup(userID, publicID)
	if terr != nil {
		return terr
	}
	activeCount, err := model.CountActiveAssetsInGroup(uint(userID), g.ID)
	if err != nil {
		common.SysError("asset: count active assets failed: " + err.Error())
		return errAssetBadGateway("查询组内素材失败")
	}
	if activeCount > 0 {
		return errAssetBadRequest("素材组内还有素材，请先清空")
	}

	won, err := model.SoftDeleteAssetGroup(userID, publicID)
	if err != nil {
		common.SysError("asset: soft delete group failed: " + err.Error())
		return errAssetBadGateway("本地删除素材组失败")
	}
	if !won {
		// 已被并发删除——幂等
		return nil
	}

	// best-effort 同步上游
	go func(chID int, upstreamID string) {
		ch, cerr := model.GetChannelById(chID, true)
		if cerr != nil || ch == nil {
			return
		}
		if err := (&taskdoubao.TaskAdaptor{}).DeleteGroup(
			ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), upstreamID, ch.GetSetting().Proxy,
		); err != nil {
			common.SysError("asset: upstream delete group failed: " + err.Error())
		}
	}(g.ChannelID, g.UpstreamAssetGroupID)
	return nil
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
		return nil, errAssetBadGateway("上游响应缺少 AssetId")
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

// DeleteAsset 删除素材：本地 CAS 软删，best-effort 上游删除。
func DeleteAsset(userID int, publicID string) *taskdto.TaskError {
	a, terr := GetAsset(userID, publicID)
	if terr != nil {
		return terr
	}
	won, err := model.SoftDeleteAsset(userID, publicID)
	if err != nil {
		common.SysError("asset: soft delete asset failed: " + err.Error())
		return errAssetBadGateway("本地删除素材失败")
	}
	if !won {
		return nil
	}

	go func(chID int, upstreamID string) {
		ch, cerr := model.GetChannelById(chID, true)
		if cerr != nil || ch == nil {
			return
		}
		if err := (&taskdoubao.TaskAdaptor{}).DeleteAsset(
			ch.GetBaseURL(), ch.Key, resolveAssetApiPath(ch), upstreamID, ch.GetSetting().Proxy,
		); err != nil {
			common.SysError("asset: upstream delete asset failed: " + err.Error())
		}
	}(a.ChannelID, a.UpstreamAssetID)
	return nil
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

// parseAssetGroupID 从上游创建素材组的响应中抠出 AssetGroupId。
// 响应 schema 文档为平铺：{AssetGroupId, Name, ...}。
func parseAssetGroupID(body []byte) (string, error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if v, ok := m["AssetGroupId"].(string); ok && v != "" {
		return v, nil
	}
	return "", errors.New("AssetGroupId missing")
}

// parseAssetID 从上游创建素材的响应中抠出 AssetId。
func parseAssetID(body []byte) (string, error) {
	var m map[string]any
	if err := common.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if v, ok := m["AssetId"].(string); ok && v != "" {
		return v, nil
	}
	return "", errors.New("AssetId missing")
}
