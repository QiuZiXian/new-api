package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// LogTaskConsumption 记录任务消费日志和统计信息（仅记录，不涉及实际扣费）。
// 实际扣费已由 BillingSession（PreConsumeBilling + SettleBilling）完成。
func LogTaskConsumption(c *gin.Context, info *relaycommon.RelayInfo) {
	tokenName := c.GetString("token_name")
	logContent := fmt.Sprintf("操作 %s", info.Action)
	// 支持任务仅按次计费
	if common.StringsContains(constant.TaskPricePatches, info.OriginModelName) {
		logContent = fmt.Sprintf("%s，按次计费", logContent)
	} else {
		if otherRatios := info.PriceData.OtherRatios(); len(otherRatios) > 0 {
			var contents []string
			for key, ra := range otherRatios {
				if 1.0 != ra {
					// 仅预扣的乘数（如视频时长）在结算阶段不参与连乘，这里显式标注，
					// 避免对账时把它当成最终扣费因子。
					if info.PriceData.IsOtherRatioEstimateOnly(key) {
						contents = append(contents, fmt.Sprintf("%s: %.2f(仅预扣)", key, ra))
						continue
					}
					contents = append(contents, fmt.Sprintf("%s: %.2f", key, ra))
				}
			}
			if len(contents) > 0 {
				logContent = fmt.Sprintf("%s, 计算参数：%s", logContent, strings.Join(contents, ", "))
			}
		}
		// 预扣基数是按官方公式估算的 token（(入+出时长)×宽×高×帧率/1024）时，
		// 把估算值写进日志，便于和终态结算的真实 token 对账。
		if info.PriceData.EstimatedTokens > 0 {
			logContent = fmt.Sprintf("%s, 预估 tokens：%d", logContent, info.PriceData.EstimatedTokens)
		}
	}
	other := make(map[string]interface{})
	other["is_task"] = true
	other["request_path"] = c.Request.URL.Path
	other["task_id"] = info.PublicTaskID
	other["model_price"] = info.PriceData.ModelPrice
	if info.PriceData.ModelRatio > 0 {
		other["model_ratio"] = info.PriceData.ModelRatio
	}
	other["group_ratio"] = info.PriceData.GroupRatioInfo.GroupRatio
	if info.PriceData.GroupRatioInfo.HasSpecialRatio {
		other["user_group_ratio"] = info.PriceData.GroupRatioInfo.GroupSpecialRatio
	}
	if info.IsModelMapped {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = info.UpstreamModelName
	}
	attachQuotaSaturation(c, info, other)
	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: info.OriginModelName,
		TokenName: tokenName,
		Quota:     info.PriceData.Quota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, info.PriceData.Quota)
	model.UpdateChannelUsedQuota(info.ChannelId, info.PriceData.Quota)
}

// ---------------------------------------------------------------------------
// 异步任务计费辅助函数
// ---------------------------------------------------------------------------

// resolveTokenKey 通过 TokenId 运行时获取令牌 Key（用于 Redis 缓存操作）。
// 如果令牌已被删除或查询失败，返回空字符串。
func resolveTokenKey(ctx context.Context, tokenId int, taskID string) string {
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("获取令牌 key 失败 (tokenId=%d, task=%s): %s", tokenId, taskID, err.Error()))
		return ""
	}
	return token.Key
}

// taskIsSubscription 判断任务是否通过订阅计费。
func taskIsSubscription(task *model.Task) bool {
	return task.PrivateData.BillingSource == BillingSourceSubscription && task.PrivateData.SubscriptionId > 0
}

// taskSubscriptionUnit 返回任务提交时快照的订阅记账单位。
func taskSubscriptionUnit(task *model.Task) string {
	return model.NormalizePlanQuotaUnit(task.PrivateData.SubscriptionQuotaUnit)
}

// taskAdjustFunding 调整任务的资金来源（钱包或订阅），delta > 0 表示扣费，delta < 0 表示退还。
// 返回实际生效的额度：钱包扣费最多扣到 0，charged < delta 表示余额不足，
// 差额即欠费（退还路径不会钳制，返回值等于 delta）。
func taskAdjustFunding(task *model.Task, delta int) (int, error) {
	if taskIsSubscription(task) {
		// 按次/按 token 记账的订阅，账本在预扣时已按单位记好，
		// 金额差额（以及任务完成后的金额重算）不能直接加到次数或 token 上。
		if taskSubscriptionUnit(task) != model.SubscriptionQuotaUnitAmount {
			return delta, nil
		}
		if err := model.PostConsumeUserSubscriptionDelta(task.PrivateData.SubscriptionId, int64(delta)); err != nil {
			return 0, err
		}
		return delta, nil
	}
	if delta > 0 {
		return model.DecreaseUserQuotaUpTo(task.UserId, delta, false)
	}
	return delta, model.IncreaseUserQuota(task.UserId, -delta, false)
}

// taskAdjustTokenQuota 调整任务的令牌额度，delta > 0 表示扣费，delta < 0 表示退还。
// 需要通过 resolveTokenKey 运行时获取 key（不从 PrivateData 中读取）。
func taskAdjustTokenQuota(ctx context.Context, task *model.Task, delta int) {
	if task.PrivateData.TokenId <= 0 || delta == 0 {
		return
	}
	tokenKey := resolveTokenKey(ctx, task.PrivateData.TokenId, task.TaskID)
	if tokenKey == "" {
		return
	}
	var err error
	if delta > 0 {
		err = model.DecreaseTokenQuota(task.PrivateData.TokenId, tokenKey, delta)
	} else {
		err = model.IncreaseTokenQuota(task.PrivateData.TokenId, tokenKey, -delta)
	}
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("调整令牌额度失败 (delta=%d, task=%s): %s", delta, task.TaskID, err.Error()))
	}
}

// taskBillingOther 从 task 的 BillingContext 构建日志 Other 字段。
func taskBillingOther(task *model.Task) map[string]interface{} {
	other := make(map[string]interface{})
	if bc := task.PrivateData.BillingContext; bc != nil {
		other["model_price"] = bc.ModelPrice
		if bc.ModelRatio > 0 {
			other["model_ratio"] = bc.ModelRatio
		}
		other["group_ratio"] = bc.GroupRatio
		if priceData := taskBillingContextPriceData(bc); priceData != nil {
			for k, v := range priceData.OtherRatios() {
				other[k] = v
			}
		}
	}
	props := task.Properties
	if props.UpstreamModelName != "" && props.UpstreamModelName != props.OriginModelName {
		other["is_model_mapped"] = true
		other["upstream_model_name"] = props.UpstreamModelName
	}
	return other
}

func taskBillingContextPriceData(bc *model.TaskBillingContext) *types.PriceData {
	if bc == nil || len(bc.OtherRatios) == 0 {
		return nil
	}
	priceData := &types.PriceData{}
	if !priceData.ReplaceOtherRatios(bc.OtherRatios) {
		return nil
	}
	// 恢复「仅预扣」标记：结算时要把 seconds / video_variant 这类
	// 只服务于估算的乘数剔掉。
	priceData.ReplaceEstimateOnlyRatios(bc.EstimateOnlyRatios)
	if len(bc.EstimateOnlyRatios) == 0 {
		// 兼容升级前提交、尚未结算的任务（快照里没有标记）：seconds 属于时长维度，
		// 在按 token 结算时永远不该再乘一次；按秒模型的档位倍率同理。
		priceData.MarkOtherRatioEstimateOnly("seconds")
		if bc.UseVideoSecond {
			priceData.MarkOtherRatioEstimateOnly("video_variant")
		}
	}
	return priceData
}

// taskModelName 从 BillingContext 或 Properties 中获取模型名称。
func taskModelName(task *model.Task) string {
	if bc := task.PrivateData.BillingContext; bc != nil && bc.OriginModelName != "" {
		return bc.OriginModelName
	}
	return task.Properties.OriginModelName
}

// RefundTaskQuota 统一的任务失败退款逻辑。
// 当异步任务失败时，退还资金与令牌额度，并回减用户和渠道用量。
// 返回资金来源是否已成功退还；失败时保留 quota，供显式重试或人工对账。
func RefundTaskQuota(ctx context.Context, task *model.Task, reason string) bool {
	quota := task.Quota
	if quota == 0 {
		return true
	}

	// 1. 退还资金来源（钱包或订阅）
	if taskIsSubscription(task) {
		unit := taskSubscriptionUnit(task)
		reserved := task.PrivateData.SubscriptionReserved
		if unit == model.SubscriptionQuotaUnitAmount {
			// 金额口径退的是最终结算额度（台账已含差额结算）。
			reserved = int64(quota)
		}
		if err := model.RefundSubscriptionUsage(
			task.PrivateData.SubscriptionId,
			taskModelName(task),
			unit,
			reserved,
			task.PrivateData.SubscriptionTokens,
		); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("退还订阅额度失败 task %s: %s", task.TaskID, err.Error()))
			return false
		}
	} else if _, err := taskAdjustFunding(task, -quota); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("退还资金来源失败 task %s: %s", task.TaskID, err.Error()))
		return false
	}

	// 2. 退还令牌额度
	taskAdjustTokenQuota(ctx, task, -quota)

	// 3. 回减预扣时累计的用户和渠道用量，请求次数保持不变
	model.UpdateUserUsedQuota(task.UserId, -quota)
	model.UpdateChannelUsedQuota(task.ChannelId, -quota)

	// 4. 记录日志
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["reason"] = reason
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:           task.UserId,
		LogType:          model.LogTypeRefund,
		Content:          "",
		ChannelId:        task.ChannelId,
		ModelName:        taskModelName(task),
		Quota:            quota,
		PromptTokens:     task.PromptTokens,
		CompletionTokens: task.CompletionTokens,
		TokenId:          task.PrivateData.TokenId,
		Group:            task.Group,
		Other:            other,
	})

	// 5. 资金退款完成后再清除持久化标记。
	// 回写失败必须显式告警，避免漏掉潜在的重复退款风险。
	task.Quota = 0
	if err := task.UpdateQuota(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("退款成功但清除 task quota 失败 task %s: %s", task.TaskID, err.Error()))
	}
	return true
}

// RecalculateTaskQuota 通用的异步差额结算。
// actualQuota 是任务完成后的实际应扣额度，与预扣额度 (task.Quota) 做差额结算。
// reason 用于日志记录（例如 "token重算" 或 "adaptor调整"）。
// clamps 可选：若计算 actualQuota 时发生额度饱和，将其记入日志 admin_info（仅管理员可见）。
func RecalculateTaskQuota(ctx context.Context, task *model.Task, actualQuota int, reason string, clamps ...*common.QuotaClamp) {
	if actualQuota <= 0 {
		return
	}
	preConsumedQuota := task.Quota
	quotaDelta := actualQuota - preConsumedQuota

	if quotaDelta == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("任务 %s 预扣费准确（%s，%s）",
			task.TaskID, logger.LogQuota(actualQuota), reason))
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("任务 %s 差额结算：delta=%s（实际：%s，预扣：%s，%s）",
		task.TaskID,
		logger.LogQuota(quotaDelta),
		logger.LogQuota(actualQuota),
		logger.LogQuota(preConsumedQuota),
		reason,
	))

	// 调整资金来源。补扣最多扣到 0：余额不足的差额不再把余额打成负数，
	// 而是记为欠费写进日志，任务记录的实际扣费也按真实扣款回写。
	charged, fundingErr := taskAdjustFunding(task, quotaDelta)
	if fundingErr != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算资金调整失败 task %s: %s", task.TaskID, fundingErr.Error()))
		return
	}
	settledDelta := quotaDelta
	unpaidQuota := 0
	if quotaDelta > 0 && charged < quotaDelta {
		unpaidQuota = quotaDelta - charged
		settledDelta = charged
		logger.LogWarn(ctx, fmt.Sprintf("任务 %s 结算补扣余额不足：应收 %s，预扣 %s，实扣 %s，欠费 %s",
			task.TaskID,
			logger.LogQuota(actualQuota),
			logger.LogQuota(preConsumedQuota),
			logger.LogQuota(preConsumedQuota+settledDelta),
			logger.LogQuota(unpaidQuota),
		))
	}
	// 订阅按金额记账时，逐模型金额要跟着账本一起调整。
	if taskIsSubscription(task) && taskSubscriptionUnit(task) == model.SubscriptionQuotaUnitAmount {
		if name := taskModelName(task); name != "" {
			if err := model.AdjustSubscriptionModelAmount(task.PrivateData.SubscriptionId, name, int64(quotaDelta)); err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("调整订阅模型金额失败 task %s: %s", task.TaskID, err.Error()))
			}
		}
	}

	// 调整令牌额度
	taskAdjustTokenQuota(ctx, task, settledDelta)

	task.Quota = preConsumedQuota + settledDelta
	if err := task.UpdateQuota(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算回写 quota 失败 task %s: %s", task.TaskID, err.Error()))
	}

	// 提交阶段已经累计过一次请求；结算阶段只调整最终用量。
	model.UpdateUserUsedQuota(task.UserId, settledDelta)
	model.UpdateChannelUsedQuota(task.ChannelId, settledDelta)

	var logType int
	var logQuota int
	if settledDelta > 0 {
		logType = model.LogTypeConsume
		logQuota = settledDelta
	} else {
		logType = model.LogTypeRefund
		logQuota = -settledDelta
	}
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["pre_consumed_quota"] = preConsumedQuota
	other["actual_quota"] = actualQuota
	if unpaidQuota > 0 {
		// actual_quota 是应收，unpaid_quota 是余额不足没扣到的部分，
		// 便于对账时区分「少扣」与「单价配置错误」。
		other["billed_quota"] = actualQuota
		other["unpaid_quota"] = unpaidQuota
		reason = fmt.Sprintf("%s（余额不足，欠费 %s）", reason, logger.LogQuota(unpaidQuota))
	}
	for _, clamp := range clamps {
		attachQuotaSaturationToOther(other, clamp)
	}
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:           task.UserId,
		LogType:          logType,
		Content:          reason,
		ChannelId:        task.ChannelId,
		ModelName:        taskModelName(task),
		Quota:            logQuota,
		PromptTokens:     task.PromptTokens,
		CompletionTokens: task.CompletionTokens,
		TokenId:          task.PrivateData.TokenId,
		Group:            task.Group,
		Other:            other,
		NodeName:         task.PrivateData.NodeName,
	})
}

// RecalculateTaskQuotaByTokens 根据实际 token 消耗重新计费（异步差额结算）。
// 当任务成功且返回了 totalTokens 时，根据模型倍率和分组倍率重新计算实际扣费额度，
// 与预扣费的差额进行补扣或退还。支持钱包和订阅计费来源。
// 返回值表示是否已按 token 完成结算（false 表示缺少按 token 单价，调用方可以回退到
// 其它结算口径，例如按秒配置下的真实时长）。
func RecalculateTaskQuotaByTokens(ctx context.Context, task *model.Task, totalTokens int) bool {
	if totalTokens <= 0 {
		return false
	}

	modelName := taskModelName(task)

	// 按 token 记账的订阅：账本和逐模型计数都按真实 token 补差，金额重算不参与。
	if taskIsSubscription(task) && taskSubscriptionUnit(task) == model.SubscriptionQuotaUnitToken {
		delta := int64(totalTokens) - task.PrivateData.SubscriptionTokens
		if delta == 0 {
			return true
		}
		if err := model.SettleSubscriptionTokens(task.PrivateData.SubscriptionId, modelName, delta); err != nil {
			logger.LogError(ctx, fmt.Sprintf("token 订阅差额结算失败 task %s: %s", task.TaskID, err.Error()))
			return false
		}
		task.PrivateData.SubscriptionTokens = int64(totalTokens)
		if err := model.DB.Model(&model.Task{}).Where("id = ?", task.ID).
			Update("private_data", task.PrivateData).Error; err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("回写任务订阅 token 快照失败 task %s: %s", task.TaskID, err.Error()))
		}
		return true
	}

	// 获取模型价格和倍率
	modelRatio, hasRatioSetting, _ := ratio_setting.GetModelRatio(modelName)
	// 只有配置了倍率(非固定价格)时才按 token 重新计费
	if !hasRatioSetting || modelRatio <= 0 {
		return false
	}

	// 获取用户和组的倍率信息
	group := task.Group
	if group == "" {
		user, err := model.GetUserById(task.UserId, false)
		if err == nil {
			group = user.Group
		}
	}
	if group == "" {
		return false
	}

	groupRatio := ratio_setting.GetGroupRatio(group)
	userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(group, group)

	var finalGroupRatio float64
	if hasUserGroupRatio {
		finalGroupRatio = userGroupRatio
	} else {
		finalGroupRatio = groupRatio
	}

	// 结算倍率：跳过「仅预扣」的乘数（如视频时长），避免上游 token 已含时长后再乘一次
	otherMultiplier := 1.0
	if priceData := taskBillingContextPriceData(task.PrivateData.BillingContext); priceData != nil {
		otherMultiplier = priceData.SettlementOtherRatioMultiplier()
	}

	// 计算实际应扣费额度: totalTokens * modelRatio * groupRatio * otherMultiplier（饱和转换，防止溢出成负数）
	actualQuota, clamp := common.QuotaFromFloatChecked(float64(totalTokens) * modelRatio * finalGroupRatio * otherMultiplier)

	reason := fmt.Sprintf("token重算：tokens=%d, modelRatio=%.2f, groupRatio=%.2f, otherMultiplier=%.4f", totalTokens, modelRatio, finalGroupRatio, otherMultiplier)
	RecalculateTaskQuota(ctx, task, actualQuota, reason, clamp)
	return true
}

// RecalculateTaskQuotaBySeconds 按视频真实时长重算按秒计费任务的额度。
// 提交阶段用户可能未显式指定时长（duration=-1 / 缺省），那时只能按基准单价预扣，
// 真实秒数要等上游在终态回报。这里用实际时长与生效分辨率档把差额补齐（多退少补）。
func RecalculateTaskQuotaBySeconds(ctx context.Context, task *model.Task, seconds int, variant string) {
	if task == nil || seconds <= 0 {
		return
	}
	// 时长参与计费前必须钳制，与提交阶段一致，防止异常值溢出成负数额度
	if seconds > relaycommon.MaxTaskDurationSeconds {
		logger.LogWarn(ctx, fmt.Sprintf("任务 %s 上游回报时长 %ds 超过上限，按 %ds 计费",
			task.TaskID, seconds, relaycommon.MaxTaskDurationSeconds))
		seconds = relaycommon.MaxTaskDurationSeconds
	}

	modelName := taskModelName(task)
	price, ok := ratio_setting.GetVideoSecondVariantPrice(modelName, variant)
	if !ok || price <= 0 {
		logger.LogWarn(ctx, fmt.Sprintf("任务 %s 缺少按秒计费价格 (model=%s, variant=%s)，保持预扣额度",
			task.TaskID, modelName, variant))
		return
	}

	bc := task.PrivateData.BillingContext
	if bc == nil {
		return
	}

	actualQuota, clamp := common.QuotaFromFloatChecked(price * float64(seconds) * common.QuotaPerUnit * bc.GroupRatio)
	reason := fmt.Sprintf("按秒重算：seconds=%d, variant=%s, price=%.6f, groupRatio=%.2f",
		seconds, variant, price, bc.GroupRatio)
	RecalculateTaskQuota(ctx, task, actualQuota, reason, clamp)
}
