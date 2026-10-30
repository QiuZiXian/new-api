# Seedance 计费探测报告

- 环境：http://www.qsykai.com:33000
- 时间：2026-10-29 15:29:03
- 换算式：`¥ = quota / 500000`
- 说明：异步任务消费日志为「预扣 + 差额补扣」，下表 quota 已按 task_id 求和。

| 用例 | 分组 | 模型 | 档位 | 请求时长 | 成片时长 | 时长偏差 | 公式 tokens | 平台 tokens | token 偏差 | 档位价 | 净扣 quota | 金额 | 金额偏差 | 结论 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| C1 | 默认 | doubao-seedance-2.0 | 720p 16:9 | 5s | 5.088s | 0.088s | 109901 | 217392 | +97.81% | 46.0 | 5000016 | ¥10.0000 | +97.81% | FAIL |
| C3 | 默认 | doubao-seedance-2.0-mini | 720p 16:9 | 5s | 5.088s | 0.088s | 109901 | 108900 | -0.91% | 23.0 | 1252350 | ¥2.5047 | -0.91% | PASS |
| C5 | 默认 | doubao-seedance-2.0-fast | 720p 16:9 | 5s | 5.088s | 0.088s | 109901 | 108900 | -0.91% | 37.0 | 2014650 | ¥4.0293 | -0.91% | PASS |
| C6 | 默认 | doubao-seedance-2.0-fast | 720p 16:9 | 10s | 5.088s | -4.912s | 109901 | 108900 | -0.91% | 37.0 | 2014650 | ¥4.0293 | -0.91% | FAIL |

## 明细

### C1 @ 默认组
```json
{
  "case": "C1",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "has_video": false,
  "status": "FAIL",
  "expected_tokens": 108000,
  "task_id": "task_Z7Yx4GAqhkpX9xGqEcCj6tZmJkMR7dF9",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "consume_quota": 5000016,
  "refund_quota": 0,
  "net_quota": 5000016,
  "total_quota": 5000016,
  "pre_consumed_quota": 2484000,
  "actual_quota_field": 5000016,
  "log_model_ratio": 23,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=217392, modelRatio=23.00, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "log_types": [
    2,
    2
  ],
  "tokens_source": "log",
  "actual_tokens": 217392,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 217392,
  "tier_price": 46.0,
  "group_ratio": 1.0,
  "formula_tokens": 109901,
  "formula_basis_seconds": 5.088,
  "duration_dev": 0.088,
  "duration_ok": true,
  "expected_quota": 5000016.0,
  "token_dev": 0.9781,
  "amount_cny": 10.0,
  "amount_expected": 5.0554,
  "amount_dev": 0.9781,
  "platform_recompute": 5000016.0,
  "platform_dev": 0.0,
  "token_ok": false,
  "amount_ok": false
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
  "task_id": "task_5SXZof4ARRI8N0InLACETgcrTXdCycXi",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "consume_quota": 1252350,
  "refund_quota": 0,
  "net_quota": 1252350,
  "total_quota": 1252350,
  "pre_consumed_quota": 1242000,
  "actual_quota_field": 1252350,
  "log_model_ratio": 11.5,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=108900, modelRatio=11.50, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "log_types": [
    2,
    2
  ],
  "tokens_source": "log",
  "actual_tokens": 108900,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 108900,
  "tier_price": 23.0,
  "group_ratio": 1.0,
  "formula_tokens": 109901,
  "formula_basis_seconds": 5.088,
  "duration_dev": 0.088,
  "duration_ok": true,
  "expected_quota": 1252350.0,
  "token_dev": -0.0091,
  "amount_cny": 2.5047,
  "amount_expected": 2.5277,
  "amount_dev": -0.0091,
  "platform_recompute": 1252350.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": true
}
```

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
  "task_id": "task_SIz8WLMWcAmuYZAFELqFUpHUT7SDPXmf",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "consume_quota": 2014650,
  "refund_quota": 0,
  "net_quota": 2014650,
  "total_quota": 2014650,
  "pre_consumed_quota": 1998000,
  "actual_quota_field": 2014650,
  "log_model_ratio": 18.5,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：108000",
    "token重算：tokens=108900, modelRatio=18.50, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "log_types": [
    2,
    2
  ],
  "tokens_source": "log",
  "actual_tokens": 108900,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 108900,
  "tier_price": 37.0,
  "group_ratio": 1.0,
  "formula_tokens": 109901,
  "formula_basis_seconds": 5.088,
  "duration_dev": 0.088,
  "duration_ok": true,
  "expected_quota": 2014650.0,
  "token_dev": -0.0091,
  "amount_cny": 4.0293,
  "amount_expected": 4.0663,
  "amount_dev": -0.0091,
  "platform_recompute": 2014650.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": true
}
```

### C6 @ 默认组
```json
{
  "case": "C6",
  "account": "默认组",
  "group": "默认",
  "model": "doubao-seedance-2.0-fast",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 10,
  "has_video": false,
  "status": "FAIL",
  "expected_tokens": 216000,
  "task_id": "task_B0JojqUplUEjzYmCVcy5X7ESkO0MVzJ7",
  "submit_http": 200,
  "state": "completed",
  "content_ok": true,
  "video_seconds": 5.088,
  "log_lines": 2,
  "consume_quota": 3996000,
  "refund_quota": 1981350,
  "net_quota": 2014650,
  "total_quota": 2014650,
  "pre_consumed_quota": 3996000,
  "actual_quota_field": 2014650,
  "log_model_ratio": 18.5,
  "log_group_ratio": 1,
  "log_content": [
    "操作 generate, 预估 tokens：216000",
    "token重算：tokens=108900, modelRatio=18.50, groupRatio=1.00, otherMultiplier=1.0000"
  ],
  "log_types": [
    2,
    6
  ],
  "tokens_source": "log",
  "actual_tokens": 108900,
  "log_prompt_tokens": 0,
  "log_completion_tokens": 108900,
  "tier_price": 37.0,
  "group_ratio": 1.0,
  "formula_tokens": 109901,
  "formula_basis_seconds": 5.088,
  "duration_dev": -4.912,
  "duration_ok": false,
  "expected_quota": 2014650.0,
  "token_dev": -0.0091,
  "amount_cny": 4.0293,
  "amount_expected": 4.0663,
  "amount_dev": -0.0091,
  "platform_recompute": 2014650.0,
  "platform_dev": 0.0,
  "token_ok": true,
  "amount_ok": true
}
```
