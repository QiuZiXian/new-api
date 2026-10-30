# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-29 15:20:31
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 视频实际时长 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 实扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C2 | 默认 | doubao-seedance-2.0 | 1080p 16:9 | -s | 243000 | - | - | - | - | - | - | FAIL |
| C3 | 默认 | doubao-seedance-2.0-mini | 720p 16:9 | 5.088s | 108000 | 108900 | +0.83% | 23.0 | 1252350 | ¥2.5047 | +0.83% | PASS |

## 明细

### C2 @ 默认组
```json
{
  "case": "C2",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0",
  "resolution": "1080p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "FAIL",
  "expected_tokens": 243000,
  "submit_http": 404,
  "ratios_header": null,
  "reason": "提交失败: {\"code\": \"fail_to_fetch_task\", \"message\": \"\", \"status_code\": 404, \"error\": {\"code\": \"fail_to_fetch_task\", \"message\": \"\", \"type\": \"new_api_error\"}}"
}
```

### C3 @ 默认组
```json
{
  "case": "C3",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0-mini",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "PASS",
  "expected_tokens": 108000,
  "submit_http": 200,
  "ratios_header": "{}",
  "task_id": "task_5SXZof4ARRI8N0InLACETgcrTXdCycXi",
  "state": "completed",
  "elapsed_sec": 134.0,
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "total_quota": 1252350,
  "pre_consumed_quota": 1242000,
  "actual_quota_field": 1252350,
  "log_model_ratio": 11.5,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=108900, modelRatio=11.50, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "tokens_source": "log",
  "actual_tokens": 108900,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 108900,
  "tier_price": 23.0,
  "group_ratio": 1.0,
  "expected_quota": 1252350.0,
  "quota_dev": 0.0,
  "token_dev": 0.0083,
  "amount_cny": 2.5047,
  "amount_expected": 2.484,
  "amount_dev": 0.0083,
  "platform_recompute": 1252350.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": true
}
```
