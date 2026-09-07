package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	taskdoubao "github.com/QuantumNous/new-api/relay/channel/task/doubao"
	"github.com/gin-gonic/gin"
)

// upstreamListTaskPageSize 是上游 cii-group 平台允许的最大 page_size；
// 超过 500 会被上游忽略，这里做一次硬钳制。
const upstreamListTaskPageSize = 500

// ListUpstreamTask 调用上游任务平台（目前仅 Doubao/VolcEngine Seedance）的
// 任务列表接口，并把响应原样回传给前端。
//
// 与 GetAllTask 的区别：这里不查本地数据库的任务表，而是直接代理到上游；
// 因此查询参数会原样转发到上游（filter.status / filter.task_ids / filter.model），
// 而本地特有的 startTime/endTime 在此接口上无效（上游没有时间区间参数）。
//
// 请求参数：
//   - channel_id（必填）：要查询的渠道 ID；用于取 BaseURL、Key、TaskApiPath 和 proxy
//   - page（默认 1）、page_size（默认 20，上限 500）
//   - status（可选）：queued/running/cancelled/succeeded/failed/expired
//   - task_ids（可选）：多个用英文逗号分隔
//   - model（可选）：精确搜索的模型/接入点 ID
//   - service_tier（可选）：default/flex
func ListUpstreamTask(c *gin.Context) {
	channelID, err := parseChannelID(c.Query("channel_id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	ch, err := model.CacheGetChannel(channelID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	adaptor, ok := buildUpstreamTaskAdaptor(ch)
	if !ok {
		common.ApiErrorMsg(c, "该渠道类型暂不支持查询上游任务列表")
		return
	}

	opts := taskdoubao.ListTaskOptions{
		PageNum:     parseIntDefault(c.Query("page"), 1),
		PageSize:    clampPageSize(parseIntDefault(c.Query("page_size"), 20)),
		Status:      strings.TrimSpace(c.Query("status")),
		Model:       strings.TrimSpace(c.Query("model")),
		ServiceTier: strings.TrimSpace(c.Query("service_tier")),
	}
	if ids := parseCSV(c.Query("task_ids")); len(ids) > 0 {
		opts.TaskIDs = ids
	}

	body, err := adaptor.ListTask(
		ch.GetBaseURL(),
		ch.Key,
		ch.GetOtherSettings().TaskApiPath,
		opts,
		ch.GetSetting().Proxy,
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	// 把上游 JSON 原文透传；不强行包成 PageInfo 是因为 pageInfo 的 total
	// 语义是"本地总数"，与上游 total 含义不同，混在一起容易让前端误读。
	common.ApiSuccess(c, gin.H{
		"channel_id": ch.Id,
		"raw":        json.RawMessage(body),
	})
}

// DeleteUpstreamTask 调用上游任务平台的取消/删除任务接口。
//
// 请求参数：
//   - id（路径参数）：上游 task_id（不是平台 task_xxx 公开 ID）
//   - channel_id（query，必填）：要操作的渠道 ID
//
// 副作用规则：参考 cii-group 平台约定 ——
//   - queued      → 取消，状态变 cancelled
//   - succeeded/failed/expired → 删除记录
//   - running/cancelled → 上游会拒绝，返回 4xx
func DeleteUpstreamTask(c *gin.Context) {
	channelID, err := parseChannelID(c.Query("channel_id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	taskID := strings.TrimSpace(c.Param("id"))
	if taskID == "" {
		common.ApiError(c, errors.New("缺少任务 ID"))
		return
	}
	ch, err := model.CacheGetChannel(channelID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	adaptor, ok := buildUpstreamTaskAdaptor(ch)
	if !ok {
		common.ApiErrorMsg(c, "该渠道类型暂不支持取消/删除上游任务")
		return
	}

	if err := adaptor.DeleteTask(
		ch.GetBaseURL(),
		ch.Key,
		ch.GetOtherSettings().TaskApiPath,
		taskID,
		ch.GetSetting().Proxy,
	); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"channel_id": ch.Id,
		"task_id":    taskID,
	})
}

// buildUpstreamTaskAdaptor 根据渠道类型构造一个实现了 ListTask/DeleteTask 的 adaptor。
// 目前只有 Doubao/VolcEngine 渠道（ChannelTypeDoubaoVideo / ChannelTypeVolcEngine）
// 实现了这两个方法；其它渠道类型返回 ok=false，controller 应当提示"暂不支持"。
func buildUpstreamTaskAdaptor(ch *model.Channel) (*taskdoubao.TaskAdaptor, bool) {
	if ch == nil {
		return nil, false
	}
	switch ch.Type {
	case constant.ChannelTypeDoubaoVideo, constant.ChannelTypeVolcEngine:
		// ListTask / DeleteTask 直接接收 baseURL/key 作为参数，不依赖 adaptor 上的字段，
		// 因此无需 Init；这样也避免引入本不需要的 RelayInfo 字段。
		return &taskdoubao.TaskAdaptor{}, true
	}
	return nil, false
}

func parseChannelID(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("缺少 channel_id 参数")
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("channel_id 必须是正整数")
	}
	return id, nil
}

func parseIntDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func clampPageSize(v int) int {
	if v <= 0 {
		return 20
	}
	if v > upstreamListTaskPageSize {
		return upstreamListTaskPageSize
	}
	return v
}

func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
