# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-29 15:24:48
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 视频实际时长 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 实扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C5 | 默认 | doubao-seedance-2.0-fast | 720p 16:9 | 5.088s | 108000 | 108900 | +0.83% | 37.0 | 2014650 | ¥4.0293 | +0.83% | PASS |

## 明细

### C5 @ 默认组
```json
{
  "case": "C5",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0-fast",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "PASS",
  "expected_tokens": 108000,
  "submit_http": 200,
  "ratios_header": "{}",
  "task_id": "task_SIz8WLMWcAmuYZAFELqFUpHUT7SDPXmf",
  "state": "completed",
  "elapsed_sec": 131.6,
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "total_quota": 2014650,
  "pre_consumed_quota": 1998000,
  "actual_quota_field": 2014650,
  "log_model_ratio": 18.5,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=108900, modelRatio=18.50, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "tokens_source": "log",
  "actual_tokens": 108900,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 108900,
  "tier_price": 37.0,
  "group_ratio": 1.0,
  "expected_quota": 2014650.0,
  "quota_dev": 0.0,
  "token_dev": 0.0083,
  "amount_cny": 4.0293,
  "amount_expected": 3.996,
  "amount_dev": 0.0083,
  "platform_recompute": 2014650.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": true
}
```
